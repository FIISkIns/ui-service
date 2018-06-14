package main

import (
	"errors"
	"fmt"
	"github.com/FIISkIns/ui-service/external"
	"github.com/Masterminds/sprig"
	"github.com/dustin/go-humanize"
	"gopkg.in/russross/blackfriday.v2"
	"html/template"
	"log"
	"net/http"
	"path"
	"time"
)

const templatePath = "template"

const sessionUserKey = "user"

func renderPage(w http.ResponseWriter, page string, params map[string]interface{}) {
	pageFile := page + ".html"
	t, err := template.New(pageFile).Funcs(sprig.FuncMap()).
		ParseFiles(path.Join(templatePath, "layout.html"), path.Join(templatePath, pageFile))
	if err != nil {
		http.Error(w, "Could not parse template", http.StatusInternalServerError)
		log.Printf("While parsing template for page %v: %v\n", page, err)
		return
	}

	err = t.ExecuteTemplate(w, "layout", params)
	if err != nil {
		http.Error(w, "Could not render page", http.StatusInternalServerError)
		log.Printf("While rendering page %v: %v\n", page, err)
		return
	}
}

func renderCourseMarkdown(course string, markdown string) template.HTML {
	renderer := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
		AbsolutePrefix: getStaticRoot(course),
		Flags:          blackfriday.CommonHTMLFlags,
	})

	return template.HTML(blackfriday.Run([]byte(markdown), blackfriday.WithRenderer(renderer)))
}

func basePageParams(w http.ResponseWriter, r *http.Request, public bool) (map[string]interface{}, string) {
	userId, _ := sessionManager.Load(r).GetString(sessionUserKey)
	userInfo, err := external.GetUserInfo(userId)
	if err != nil {
		log.Println("GetUserInfo:", err)
		logOut(w, r, r.URL.String())
		return nil, ""
	}

	if userInfo != nil {
		firstSeen, err := time.Parse(time.RFC3339, userInfo.FirstSeen)
		if err == nil {
			userInfo.FirstSeen = humanize.Time(firstSeen)
		}
	}

	if userInfo == nil && !public {
		LoginPage(w, r, nil)
		userId = ""
		return nil, ""
	}

	if userInfo != nil {
		log.Printf("%v -> %v", userId, r.URL.String())
	} else {
		log.Printf("guest -> %v", r.URL.String())
	}

	var pingChan <-chan error
	var courseProgressChan <-chan external.CourseProgressResult
	if userInfo != nil {
		pingChan = external.UserStatsPing(userId)
		courseProgressChan = external.GetAllCourseProgress(userId)
	}

	courseListChan := external.GetCourseList()

	if pingChan != nil {
		err = <-pingChan
		if err != nil {
			log.Println("UserStatsPing:", err)
		}
	}

	courseList := <-courseListChan
	if courseList.Err != nil {
		log.Println("GetCourseList:", courseList.Err)
		http.Error(w, "Could not access course manager", http.StatusInternalServerError)
		return nil, userId
	}

	courseInfo := make(map[string]external.BasicCourseInfo)
	for _, info := range courseList.Val {
		courseInfo[info.Id] = info
	}

	startedCourses := make([]string, 0)
	availableCourses := make([]string, 0)

	var courseProgress external.CourseProgressResult
	if courseProgressChan != nil {
		courseProgress = <-courseProgressChan
		if courseProgress.Err != nil {
			log.Println("GetAllCourseProgress:", courseProgress.Err)
			http.Error(w, "Could not access course progress", http.StatusInternalServerError)
			return nil, userId
		}

		for course, progress := range courseProgress.Val {
			completed := 0
			for _, task := range progress {
				if task.Progress == "completed" {
					completed++
				}
			}
			if completed > 0 {
				startedCourses = append(startedCourses, course)
			} else {
				availableCourses = append(availableCourses, course)
			}
		}
	} else {
		for _, info := range courseList.Val {
			availableCourses = append(availableCourses, info.Id)
		}
	}

	return map[string]interface{}{
		"User":             userInfo,
		"Courses":          courseInfo,
		"StartedCourses":   startedCourses,
		"AvailableCourses": availableCourses,
		"CourseProgress":   courseProgress.Val,
	}, userId
}

func logOut(w http.ResponseWriter, r *http.Request, url string) {
	sessionManager.Load(r).Clear(w)
	http.Redirect(w, r, url, http.StatusFound)
}

func getTaskProgress(course string, task string, params map[string]interface{}) string {
	courseProgress := params["CourseProgress"].(map[string][]external.TaskProgress)
	if courseProgress[course] != nil {
		for _, progress := range courseProgress[course] {
			if progress.TaskId == task {
				return progress.Progress
			}
		}
	}

	return ""
}

func getNextTask(course string, task string, params map[string]interface{}) (string, error) {
	courseProgress := params["CourseProgress"].(map[string][]external.TaskProgress)
	if courseProgress[course] != nil {
		for i, progress := range courseProgress[course] {
			if progress.TaskId == task {
				if i+1 < len(courseProgress[course]) {
					return courseProgress[course][i+1].TaskId, nil
				} else {
					return "", nil
				}
			}
		}
	}

	return "", errors.New("task not found")
}

func setTaskProgressWithParams(userId string, course string, task string, progress string, params map[string]interface{}) error {
	err := external.SetTaskProgress(userId, course, task, progress)
	if err != nil {
		return err
	}

	updated := false
	for _, group := range params["CourseProgress"].(map[string][]external.TaskProgress) {
		for _, t := range group {
			if t.TaskId == task {
				t.Progress = progress
				updated = true
			}
		}
	}
	if !updated {
		return errors.New("task not found in progress")
	}

	return nil
}

func HomePage(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	params, _ := basePageParams(w, r, true)
	if params == nil {
		return
	}
	params["Active"] = "home"

	courseChans := make(map[string]<-chan external.CourseInfoResult)
	for course, info := range params["Courses"].(map[string]external.BasicCourseInfo) {
		courseChans[course] = external.GetCourseInfo(info.Url)
	}

	courses := make(map[string]*external.CourseInfo)
	for course, infoChan := range courseChans {
		info := <-infoChan
		if info.Err != nil {
			log.Println("GetCourseInfo:", course, info.Err)
			continue
		}
		courses[course] = info.Val
	}
	params["CourseInfo"] = courses

	renderPage(w, "home", params)
}

func LoginPage(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	rawQuery := ""
	returnUrl := r.URL.String()
	if r.URL.Path == "/login" {
		rawQuery = r.URL.RawQuery
		returnUrl = "/"
	}
	res, err := external.DoLoginFlow(rawQuery, returnUrl)
	if err != nil {
		log.Println("DoLoginFlow:", err)
		http.Error(w, "Could not access login service", http.StatusInternalServerError)
		return
	}

	sessionManager.Load(r).PutString(w, sessionUserKey, res.UserId)
	http.Redirect(w, r, res.RedirectUrl, http.StatusFound)
}

func LogoutPage(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	logOut(w, r, "/")
}

func CourseRootPage(w http.ResponseWriter, r *http.Request, ps map[string]string) {
	params, _ := basePageParams(w, r, false)
	if params == nil {
		return
	}

	tasks := params["CourseProgress"].(map[string][]external.TaskProgress)[ps["course"]]

	nextTask := tasks[0].TaskId
	takeNext := false
	for _, task := range tasks {
		if task.Progress == "completed" {
			nextTask = task.TaskId
			takeNext = true
		} else if takeNext {
			nextTask = task.TaskId
			takeNext = false
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/course/%v/%v", ps["course"], nextTask), http.StatusFound)
}

func CourseTaskPage(w http.ResponseWriter, r *http.Request, ps map[string]string) {
	params, userId := basePageParams(w, r, false)
	if params == nil {
		return
	}
	params["Active"] = "course"
	params["CourseId"] = ps["course"]

	courseInfo := <-external.GetBasicCourseInfo(ps["course"])
	if courseInfo.Err != nil {
		log.Println("GetBasicCourseInfo:", courseInfo.Err)
		http.Error(w, "Could not access course manager", http.StatusInternalServerError)
		return
	}

	courseTasksChan := external.GetCourseTasks(courseInfo.Val.Url)
	taskInfoChan := external.GetTaskInfo(courseInfo.Val.Url, ps["task"])

	courseTasks := <-courseTasksChan
	if courseTasks.Err != nil {
		log.Println("GetCourseTasks:", courseTasks.Err)
		http.Error(w, "Could not access course service", http.StatusInternalServerError)
		return
	}
	params["Tasks"] = courseTasks.Val

	taskInfo := <-taskInfoChan
	if taskInfo.Err != nil {
		log.Println("GetTaskInfo:", taskInfo.Err)
		http.Error(w, "Could not access course service", http.StatusInternalServerError)
		return
	}
	params["TaskInfo"] = taskInfo.Val
	params["TaskBody"] = renderCourseMarkdown(ps["course"], taskInfo.Val.Body)

	next, err := getNextTask(ps["course"], ps["task"], params)
	if err != nil {
		log.Println("getNextTask:", err)
		http.Error(w, "Error getting next task", http.StatusInternalServerError)
		return
	}
	if next != "" {
		for _, group := range courseTasks.Val {
			for _, task := range group.Tasks {
				if task.Id == next {
					params["NextTaskInfo"] = task
					break
				}
			}
		}
	}

	progress := getTaskProgress(ps["course"], ps["task"], params)
	if progress != "completed" {
		setTaskProgressWithParams(userId, ps["course"], ps["task"], "started", params)
	}

	renderPage(w, "task", params)
}

func CourseNextPage(w http.ResponseWriter, r *http.Request, ps map[string]string) {
	params, userId := basePageParams(w, r, false)
	if params == nil {
		return
	}

	next, err := getNextTask(ps["course"], ps["task"], params)
	if err != nil {
		log.Println("getNextTask:", err)
		http.Error(w, "Error getting next task", http.StatusInternalServerError)
		return
	}

	err = setTaskProgressWithParams(userId, ps["course"], ps["task"], "completed", params)
	if err != nil {
		log.Println("SetTaskProgress:", err)
		http.Error(w, "Error saving task progress", http.StatusInternalServerError)
		return
	}

	if next != "" {
		http.Redirect(w, r, fmt.Sprintf("/course/%v/%v", ps["course"], next), http.StatusFound)
	} else {
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func ProfilePage(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	params, userId := basePageParams(w, r, false)
	if params == nil {
		return
	}
	params["Active"] = "profile"

	statsChan := external.GetUserStats(userId)
	achievementsChan := external.GetUserAchievements(userId)

	stats := <-statsChan
	if stats.Err != nil {
		log.Println("GetUserStats:", stats.Err)
		http.Error(w, "Could not access stats service", http.StatusInternalServerError)
		return
	}
	params["UserStats"] = stats.Val

	achievements := <-achievementsChan
	if achievements.Err != nil {
		log.Println("GetUserAchievements:", achievements.Err)
		http.Error(w, "Could not access achievement service", http.StatusInternalServerError)
		return
	}
	params["Achievements"] = achievements.Val

	renderPage(w, "profile", params)
}

func StatsPing(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	userId, _ := sessionManager.Load(r).GetString(sessionUserKey)
	external.UserStatsPing(userId)
	w.WriteHeader(http.StatusNoContent)
}
