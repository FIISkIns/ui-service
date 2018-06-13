package external

import (
	"encoding/json"
	"fmt"
	"github.com/dimiro1/health"
	"github.com/dimiro1/health/url"
	"net/http"
	"sort"
)

type Achievement struct {
	Completion  int
	Description string
	Icon        string
	Title       string
}

type AchievementsResult struct {
	Val []Achievement
	Err error
}

func GetUserAchievements(userId string) <-chan AchievementsResult {
	ret := make(chan AchievementsResult)

	go func() {
		defer close(ret)

		data, err := http.Get(fmt.Sprintf("%v/achievements/%v", config.AchievementsUrl, userId))
		if err != nil {
			ret <- AchievementsResult{Err: err}
			return
		}
		defer data.Body.Close()

		var achievements []Achievement
		err = json.NewDecoder(data.Body).Decode(&achievements)
		if err != nil {
			ret <- AchievementsResult{Err: err}
			return
		}

		sort.Slice(achievements, func(i, j int) bool {
			return achievements[i].Completion > achievements[j].Completion
		})

		ret <- AchievementsResult{Val: achievements}
	}()

	return ret
}

func GetAchievementStaticUrl() string {
	return fmt.Sprintf("%v/static", config.AchievementsUrl)
}

func GetAchievementHealthCheck() health.Checker {
	return url.NewChecker(fmt.Sprintf("%v/health", config.AchievementsUrl))
}
