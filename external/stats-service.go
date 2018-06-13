package external

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type UserStats struct {
	StartedCourses   int
	CompletedCourses int
	LastLoggedIn     string
	TimeSpent        int
	LongestStreak    int
}

type StatsResult struct {
	Val *UserStats
	Err error
}

func GetUserStats(userId string) <-chan StatsResult {
	ret := make(chan StatsResult, 1)

	go func() {
		defer close(ret)

		data, err := http.Get(fmt.Sprintf("%v/stats/%v", config.StatsUrl, userId))
		if err != nil {
			ret <- StatsResult{Err: err}
			return
		}
		defer data.Body.Close()

		var stats UserStats
		err = json.NewDecoder(data.Body).Decode(&stats)
		if err != nil {
			ret <- StatsResult{Err: err}
			return
		}

		ret <- StatsResult{Val: &stats}
	}()

	return ret
}

func UserStatsPing(userId string) <-chan error {
	ret := make(chan error, 1)

	go func() {
		res, err := http.Post(fmt.Sprintf("%v/stats/%v/ping", config.StatsUrl, userId), "", nil)
		if res != nil {
			res.Body.Close()
		}
		ret <- err
	}()

	return ret
}
