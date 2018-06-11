package external

import (
	"encoding/json"
	"fmt"
	"net/http"
)

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

type CourseTasksResult struct {
	Val []TaskGroup
	Err error
}

type TaskInfoResult struct {
	Val *TaskInfo
	Err error
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
