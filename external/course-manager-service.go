package external

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type CourseInfo struct {
	Id   string
	Name string
	Url  string
}

type CourseInfoResult struct {
	Val *CourseInfo
	Err error
}

type CourseListResult struct {
	Val []CourseInfo
	Err error
}

func GetCourseInfo(courseId string) <-chan CourseInfoResult {
	ret := make(chan CourseInfoResult, 1)

	go func() {
		defer close(ret)

		data, err := http.Get(fmt.Sprintf("%v/courses/%v", config.CourseManagerUrl, courseId))
		if err != nil {
			ret <- CourseInfoResult{Err: err}
			return
		}
		defer data.Body.Close()

		var info *CourseInfo
		err = json.NewDecoder(data.Body).Decode(&info)
		if err != nil {
			ret <- CourseInfoResult{Err: err}
			return
		}

		ret <- CourseInfoResult{Val: info}
	}()

	return ret
}

func GetCourseList() <-chan CourseListResult {
	ret := make(chan CourseListResult, 1)

	go func() {
		defer close(ret)

		data, err := http.Get(fmt.Sprintf("%v/courses", config.CourseManagerUrl))
		if err != nil {
			ret <- CourseListResult{Err: err}
			return
		}
		defer data.Body.Close()

		var list []CourseInfo
		err = json.NewDecoder(data.Body).Decode(&list)
		if err != nil {
			ret <- CourseListResult{Err: err}
			return
		}

		ret <- CourseListResult{Val: list}
	}()

	return ret
}
