package external

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type CourseInfo struct {
	Title       string
	Description string
	Picture     string
}

type BaseTaskInfo struct {
	Id    string
	Title string
}

type TaskInfo struct {
	BaseTaskInfo
	Body string
}

type TaskGroup struct {
	Title string
	Tasks []*BaseTaskInfo
}

type CourseInfoResult struct {
	Val *CourseInfo
	Err error
}

type CourseTasksResult struct {
	Val []TaskGroup
	Err error
}

type TaskInfoResult struct {
	Val *TaskInfo
	Err error
}

func GetCourseInfo(courseUrl string) <-chan CourseInfoResult {
	ret := make(chan CourseInfoResult, 1)

	go func() {
		defer close(ret)

		data, err := http.Get(courseUrl)
		if err != nil {
			ret <- CourseInfoResult{Err: err}
			return
		}
		defer data.Body.Close()

		var info CourseInfo
		err = json.NewDecoder(data.Body).Decode(&info)
		if err != nil {
			ret <- CourseInfoResult{Err: err}
			return
		}

		ret <- CourseInfoResult{Val: &info}
	}()

	return ret
}

func GetCourseTasks(courseUrl string) <-chan CourseTasksResult {
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

func GetTaskInfo(courseUrl string, taskId string) <-chan TaskInfoResult {
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
