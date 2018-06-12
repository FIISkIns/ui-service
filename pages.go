package main

import (
	"github.com/FIISkIns/ui-service/external"
	"github.com/Masterminds/sprig"
	"github.com/dustin/go-humanize"
	"gopkg.in/russross/blackfriday.v2"
	"html/template"
	"log"
	"net/http"
	"path"
	"time"
	"fmt"
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
	if userInfo != nil {
		pingChan = external.UserStatsPing(userId)
	}

	courseListChan := external.GetCourseList()
	courseProgressChan := external.GetAllCourseProgress(userId)

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

	courseInfo := make(map[string]*external.CourseInfo)
	for _, info := range courseList.Val {
		courseInfo[info.Id] = &info
	}

	courseProgress := <-courseProgressChan
	if courseProgress.Err != nil {
		log.Println("GetAllCourseProgress:", courseProgress.Err)
		http.Error(w, "Could not access course progress", http.StatusInternalServerError)
		return nil, userId
	}

	startedCourses := make([]string, 0)
	availableCourses := make([]string, 0)
	for course, progress := range courseProgress.Val {
		started := 0
		for _, task := range progress {
			if task.Progress != "not started" {
				started++
			}
		}
		if started > 1 {
			startedCourses = append(startedCourses, course)
		} else {
			availableCourses = append(availableCourses, course)
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

func HomePage(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	params, _ := basePageParams(w, r, true)
	if params == nil {
		return
	}
	params["Active"] = "home"

	renderPage(w, "home", params)
}

func LoginPage(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	rawQuery := ""
	if r.URL.Path == "/login" {
		rawQuery = r.URL.RawQuery
	}
	res, err := external.DoLoginFlow(rawQuery, r.URL.String())
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

	var nextTask string
	for _, task := range params["CourseProgress"].(map[string][]external.TaskProgress)[ps["course"]] {
		if task.Progress == "not started" && nextTask == "" {
			nextTask = task.TaskId
		} else if task.Progress != "not started" {
			nextTask = task.TaskId
		}
	}
	if nextTask == "" {
		log.Println("Course", ps["course"], "has no tasks")
		http.Error(w, "Course is broken, come back later", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/course/%v/%v", ps["course"], nextTask), http.StatusFound)
}

func CourseTaskPage(w http.ResponseWriter, r *http.Request, ps map[string]string) {
	params, _ := basePageParams(w, r, false)
	if params == nil {
		return
	}
	params["Active"] = "course"

	courseInfo := <-external.GetCourseInfo(ps["course"])
	if courseInfo.Err != nil {
		log.Println("GetCourseInfo:", courseInfo.Err)
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
	params["TaskBody"] = renderCourseMarkdown("example", taskInfo.Val.Body)

	renderPage(w, "task", params)
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
