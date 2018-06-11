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

func DoLoginFlow(query string) (response *LoginResponse, err error) {
	res, err := http.Get(fmt.Sprintf("%v/login?%v", config.LoginUrl, query))
	if err != nil {
		return
	}

	err = json.NewDecoder(res.Body).Decode(&response)
	return
}
