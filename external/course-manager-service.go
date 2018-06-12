package external

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type CourseListEntry struct {
	Id   string
	Name string
	Url  string
}

type CourseListResult struct {
	Val []CourseListEntry
	Err error
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

		var list []CourseListEntry
		err = json.NewDecoder(data.Body).Decode(&list)
		if err != nil {
			ret <- CourseListResult{Err: err}
			return
		}

		ret <- CourseListResult{Val: list}
	}()

	return ret
}
