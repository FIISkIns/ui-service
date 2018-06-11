package external

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type LoginResponse struct {
	RedirectUrl string `json:"redirectUrl"`
	UserId      string `json:"userId"`
}

type UserInfo struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func DoLoginFlow(query string) (response *LoginResponse, err error) {
	res, err := http.Get(fmt.Sprintf("%v/login?%v", config.LoginUrl, query))
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
