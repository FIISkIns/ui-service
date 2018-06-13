package external

import (
	"bytes"
	"encoding/json"
	"errors"
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

func SetTaskProgress(userId string, course string, task string, progress string) error {
	object := map[string]string{
		"progress": progress,
	}

	data, err := json.Marshal(object)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%v/%v/%v/%v", config.CourseProgressUrl, userId, course, task), bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	client := http.Client{}
	res, err := client.Do(req)
	defer res.Body.Close()
	if err != nil {
		return err
	} else if res.StatusCode/100 != 2 {
		return errors.New(fmt.Sprintf("progress set failed %v", res.StatusCode))
	} else {
		return nil
	}
}
