package main

import (
	"github.com/FIISkIns/ui-service/external"
	"github.com/Masterminds/sprig"
	"gopkg.in/russross/blackfriday.v2"
	"html/template"
	"log"
	"net/http"
	"path"
)

const templatePath = "template"

const sessionUserKey = "user"

func renderPage(w http.ResponseWriter, page string, params interface{}) error {
	pageFile := page + ".html"
	t, err := template.New(pageFile).Funcs(sprig.FuncMap()).
		ParseFiles(path.Join(templatePath, "layout.html"), path.Join(templatePath, pageFile))
	if err != nil {
		http.Error(w, "Could not parse template", http.StatusInternalServerError)
		log.Printf("While parsing template for page %v: %v\n", page, err)
		return err
	}

	err = t.ExecuteTemplate(w, "layout", params)
	if err != nil {
		http.Error(w, "Could not render page", http.StatusInternalServerError)
		log.Printf("While rendering page %v: %v\n", page, err)
		return err
	}

	return nil
}

func renderCourseMarkdown(course string, markdown string) template.HTML {
	renderer := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
		AbsolutePrefix: getStaticRoot(course),
		Flags:          blackfriday.CommonHTMLFlags,
	})

	return template.HTML(blackfriday.Run([]byte(markdown), blackfriday.WithRenderer(renderer)))
}

func HomePage(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	userId, _ := sessionManager.Load(r).GetString(sessionUserKey)
	log.Println("User:", userId)

	renderPage(w, "home", nil)
}

func LoginPage(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	res, err := external.DoLoginFlow(r.URL.RawQuery)
	if err != nil {
		log.Println("DoLoginFlow:", err)
		http.Error(w, "Could not access login service", http.StatusInternalServerError)
		return
	}

	sessionManager.Load(r).PutString(w, sessionUserKey, res.UserId)
	http.Redirect(w, r, res.RedirectUrl, http.StatusFound)
}

func CourseTaskPage(w http.ResponseWriter, _ *http.Request, ps map[string]string) {
	courseTasksChan := external.GetCourseTasks(config.CourseUrl)
	taskInfoChan := external.GetTaskInfo(config.CourseUrl, ps["task"])

	courseTasks := <-courseTasksChan
	if courseTasks.Err != nil {
		log.Println("GetCourseTasks:", courseTasks.Err)
		http.Error(w, "Could not access course service", http.StatusInternalServerError)
		return
	}

	taskInfo := <-taskInfoChan
	if taskInfo.Err != nil {
		log.Println("GetTaskInfo:", taskInfo.Err)
		http.Error(w, "Could not access course service", http.StatusInternalServerError)
		return
	}

	renderPage(w, "task", &struct {
		Tasks    []external.TaskGroup
		TaskInfo *external.TaskInfo
		Body     template.HTML
	}{
		Tasks:    courseTasks.Val,
		TaskInfo: taskInfo.Val,
		Body:     renderCourseMarkdown("example", taskInfo.Val.Body),
	})
}
