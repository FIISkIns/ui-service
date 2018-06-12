package external

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type TaskProgress struct {
	CourseId string
	TaskId   string
	Progress string
}

type CourseProgressResult struct {
	Val map[string][]TaskProgress
	Err error
}

func GetCourseProgress(userId string, course string) <-chan CourseProgressResult {
	ret := make(chan CourseProgressResult, 1)

	go func() {
		defer close(ret)

		url := fmt.Sprintf("%v/%v", config.CourseProgressUrl, userId)
		if course != "" {
			url += fmt.Sprintf("/%v", course)
		}

		data, err := http.Get(url)
		if err != nil {
			ret <- CourseProgressResult{Err: err}
			return
		}
		defer data.Body.Close()

		var progress []TaskProgress
		err = json.NewDecoder(data.Body).Decode(&progress)
		if err != nil {
			ret <- CourseProgressResult{Err: err}
			return
		}

		courses := make(map[string][]TaskProgress)
		for _, item := range progress {
			courses[item.CourseId] = append(courses[item.CourseId], item)
		}

		ret <- CourseProgressResult{Val: courses}
	}()

	return ret
}

func GetAllCourseProgress(userId string) <-chan CourseProgressResult {
	return GetCourseProgress(userId, "")
}
