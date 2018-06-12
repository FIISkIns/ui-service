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
)

const templatePath = "template"

const sessionUserKey = "user"

func renderPage(w http.ResponseWriter, r *http.Request, page string, params map[string]interface{}) {
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

func basePageParams(w http.ResponseWriter, r *http.Request, public bool) map[string]interface{} {
	userId, _ := sessionManager.Load(r).GetString(sessionUserKey)
	userInfo, err := external.GetUserInfo(userId)
	if err != nil {
		log.Println("GetUserInfo:", err)
		logOut(w, r, r.URL.String())
		return nil
	}

	if userInfo == nil && !public {
		LoginPage(w, r, nil)
		return nil
	}

	if userInfo != nil {
		go external.UserStatsPing(userId)

		firstSeen, err := time.Parse(time.RFC3339, userInfo.FirstSeen)
		if err == nil {
			userInfo.FirstSeen = humanize.Time(firstSeen)
		}
	}

	courseListChan := external.GetCourseList()
	courseList := <-courseListChan
	if courseList.Err != nil {
		log.Println("GetCourseList:", courseList.Err)
		http.Error(w, "Could not access course manager", http.StatusInternalServerError)
		return nil
	}

	return map[string]interface{}{
		"User":    userInfo,
		"Courses": courseList.Val,
	}
}

func logOut(w http.ResponseWriter, r *http.Request, url string) {
	sessionManager.Load(r).Clear(w)
	http.Redirect(w, r, url, http.StatusFound)
}

func HomePage(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	params := basePageParams(w, r, true)
	if params == nil {
		return
	}
	params["Active"] = "home"

	renderPage(w, r, "home", params)
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

func CourseTaskPage(w http.ResponseWriter, r *http.Request, ps map[string]string) {
	params := basePageParams(w, r, false)
	if params == nil {
		return
	}
	params["Active"] = "course"

	courseTasksChan := external.GetCourseTasks(config.CourseUrl)
	taskInfoChan := external.GetTaskInfo(config.CourseUrl, ps["task"])

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

	renderPage(w, r, "task", params)
}

func ProfilePage(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	params := basePageParams(w, r, false)
	if params == nil {
		return
	}
	params["Active"] = "profile"

	renderPage(w, r, "profile", params)
}
