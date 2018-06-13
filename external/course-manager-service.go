package external

import (
	"encoding/json"
	"fmt"
	"github.com/dimiro1/health"
	"github.com/dimiro1/health/url"
	"net/http"
)

type BasicCourseInfo struct {
	Id   string
	Name string
	Url  string
}

type BasicCourseInfoResult struct {
	Val *BasicCourseInfo
	Err error
}

type CourseListResult struct {
	Val []BasicCourseInfo
	Err error
}

func GetBasicCourseInfo(courseId string) <-chan BasicCourseInfoResult {
	ret := make(chan BasicCourseInfoResult, 1)

	go func() {
		defer close(ret)

		data, err := http.Get(fmt.Sprintf("%v/courses/%v", config.CourseManagerUrl, courseId))
		if err != nil {
			ret <- BasicCourseInfoResult{Err: err}
			return
		}
		defer data.Body.Close()

		var info *BasicCourseInfo
		err = json.NewDecoder(data.Body).Decode(&info)
		if err != nil {
			ret <- BasicCourseInfoResult{Err: err}
			return
		}

		ret <- BasicCourseInfoResult{Val: info}
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

		var list []BasicCourseInfo
		err = json.NewDecoder(data.Body).Decode(&list)
		if err != nil {
			ret <- CourseListResult{Err: err}
			return
		}

		ret <- CourseListResult{Val: list}
	}()

	return ret
}

func GetCourseManagerHealthCheck() health.Checker {
	return url.NewChecker(fmt.Sprintf("%v/health", config.CourseManagerUrl))
}
