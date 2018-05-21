package main

import (
	"encoding/json"
	"fmt"
	"github.com/Masterminds/sprig"
	"github.com/julienschmidt/httprouter"
	"gopkg.in/russross/blackfriday.v2"
	"html/template"
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

func CourseTaskPage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	courseTasksChan := getCourseTasks(config.CourseUrl)
	taskInfoChan := getTaskInfo(config.CourseUrl, ps.ByName("task"))

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
		Body:     template.HTML(blackfriday.Run([]byte(taskInfo.Val.Body))),
	})
}

func main() {
	initConfig()

	router := httprouter.New()
	router.GET("/course/example/:task", CourseTaskPage)
	router.ServeFiles("/static/*filepath", http.Dir(staticPath))

	log.Printf("Listening on port %v...\n", config.Port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.Port), router))
}
