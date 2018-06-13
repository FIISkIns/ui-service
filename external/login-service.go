package external

import (
	"encoding/json"
	"fmt"
	"github.com/dimiro1/health"
	healthurl "github.com/dimiro1/health/url"
	"net/http"
	"net/url"
)

type LoginResponse struct {
	RedirectUrl string
	UserId      string
}

type UserInfo struct {
	Id        string
	Name      string
	Picture   string
	FirstSeen string
}

func DoLoginFlow(query string, returnUrl string) (response *LoginResponse, err error) {
	q, _ := url.ParseQuery(query)
	q.Add("return_url", returnUrl)

	res, err := http.Get(fmt.Sprintf("%v/login?%v", config.LoginUrl, q.Encode()))
	if err != nil {
		return
	}

	err = json.NewDecoder(res.Body).Decode(&response)
	return
}

func GetUserInfo(userId string) (response *UserInfo, err error) {
	if userId == "" {
		return
	}

	res, err := http.Get(fmt.Sprintf("%v/%v", config.LoginUrl, userId))
	if err != nil {
		return
	}

	err = json.NewDecoder(res.Body).Decode(&response)
	return
}

func GetLoginHealthCheck() health.Checker {
	return healthurl.NewChecker(fmt.Sprintf("%v/health", config.LoginUrl))
}
