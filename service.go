package main

import (
	"encoding/json"
	"fmt"
	"github.com/Masterminds/sprig"
	"github.com/dimfeld/httptreemux"
	"gopkg.in/russross/blackfriday.v2"
	"html/template"
	"io"
	"log"
	"net/http"
	"path"
	"strconv"
)

const templatePath = "template"
const staticPath = "static"

type CourseTasksResult struct {
	Val []TaskGroup
	Err error
}

type TaskGroup struct {
	Title string
	Tasks []*BaseTaskInfo
}

type BaseTaskInfo struct {
	Id    string
	Title string
}

type TaskInfo struct {
	BaseTaskInfo
	Body string
}

type TaskInfoResult struct {
	Val *TaskInfo
	Err error
}

func getCourseTasks(courseUrl string) <-chan CourseTasksResult {
	ret := make(chan CourseTasksResult, 1)

	go func() {
		defer close(ret)

		data, err := http.Get(fmt.Sprintf("%v/tasks", courseUrl))
		if err != nil {
			ret <- CourseTasksResult{Err: err}
			return
		}
		defer data.Body.Close()

		var info []TaskGroup
		err = json.NewDecoder(data.Body).Decode(&info)
		if err != nil {
			ret <- CourseTasksResult{Err: err}
			return
		}

		ret <- CourseTasksResult{Val: info}
	}()

	return ret
}

func getTaskInfo(courseUrl string, taskId string) <-chan TaskInfoResult {
	ret := make(chan TaskInfoResult, 1)

	go func() {
		defer close(ret)

		data, err := http.Get(fmt.Sprintf("%v/tasks/%v", courseUrl, taskId))
		if err != nil {
			ret <- TaskInfoResult{Err: err}
			return
		}
		defer data.Body.Close()

		var info TaskInfo
		err = json.NewDecoder(data.Body).Decode(&info)
		if err != nil {
			ret <- TaskInfoResult{Err: err}
			return
		}

		ret <- TaskInfoResult{Val: &info}
	}()

	return ret
}

func renderPage(w http.ResponseWriter, page string, params interface{}) error {
	pageFile := page + ".html"
	t, err := template.New(pageFile).Funcs(sprig.FuncMap()).
		ParseFiles(path.Join(templatePath, pageFile))
	if err != nil {
		http.Error(w, "Could not parse template", 500)
		log.Printf("While parsing template for page %v: %v\n", page, err)
		return err
	}

	err = t.Execute(w, params)
	if err != nil {
		http.Error(w, "Could not render page", 500)
		log.Printf("While rendering page %v: %v\n", page, err)
		return err
	}

	return nil
}

func getStaticRoot(service string) string {
	return "/static/" + service + "/"
}

func renderCourseMarkdown(course string, markdown string) template.HTML {
	renderer := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
		AbsolutePrefix: getStaticRoot(course),
		Flags:          blackfriday.CommonHTMLFlags,
	})

	return template.HTML(blackfriday.Run([]byte(markdown), blackfriday.WithRenderer(renderer)))
}

func CourseTaskPage(w http.ResponseWriter, _ *http.Request, ps map[string]string) {
	courseTasksChan := getCourseTasks(config.CourseUrl)
	taskInfoChan := getTaskInfo(config.CourseUrl, ps["task"])

	courseTasks := <-courseTasksChan
	if courseTasks.Err != nil {
		log.Println("getCourseTasks:", courseTasks.Err)
		http.Error(w, "Could not access required services", 500)
		return
	}
	taskInfo := <-taskInfoChan
	if taskInfo.Err != nil {
		log.Println("getTaskInfo:", taskInfo.Err)
		http.Error(w, "Could not access required services", 500)
		return
	}

	renderPage(w, "task", &struct {
		Tasks    []TaskGroup
		TaskInfo *TaskInfo
		Body     template.HTML
	}{
		Tasks:    courseTasks.Val,
		TaskInfo: taskInfo.Val,
		Body:     renderCourseMarkdown("example", taskInfo.Val.Body),
	})
}

func getServiceDirectStaticUrl(service string) string {
	if service == "achievements" {
		// TODO
		return ""
	}

	// TODO: call course manager
	return fmt.Sprintf("%v/static/", config.CourseUrl)
}

func StaticResourceProxy(w http.ResponseWriter, r *http.Request, ps map[string]string) {
	res, err := http.Get(fmt.Sprintf("%v/%v", getServiceDirectStaticUrl(ps["service"]), ps["filepath"]))
	if err != nil {
		if res != nil && res.StatusCode == 404 {
			http.NotFound(w, r)
			return
		} else {
			http.Error(w, "Upstream error", 500)
			log.Println("StaticResourceProxy:", err)
			return
		}
	}
	defer res.Body.Close()

	io.Copy(w, res.Body)
}

func StaticResource(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	http.StripPrefix("/static/", http.FileServer(http.Dir(staticPath))).ServeHTTP(w, r)
}

func main() {
	initConfig()

	router := httptreemux.New()
	router.GET("/course/example/:task", CourseTaskPage)
	router.GET("/static/:service/*filepath", StaticResourceProxy)
	router.GET("/static/*filepath", StaticResource)

	log.Printf("Listening on port %v...\n", config.Port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.Port), router))
}
