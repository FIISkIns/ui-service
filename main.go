package main

import (
	"fmt"
	"github.com/alexedwards/scs"
	"github.com/dimfeld/httptreemux"
	"io"
	"log"
	"net/http"
	"strconv"
)

const staticPath = "static"

var sessionManager *scs.Manager

func getStaticRoot(service string) string {
	return "/static/" + service + "/"
}

func getServiceDirectStaticUrl(service string) string {
	if service == "achievements" {
		// TODO
		return ""
	}

	// TODO: call course manager
	return fmt.Sprintf("%v/static/", config.CourseUrl)
}

func StaticResourceProxy(w http.ResponseWriter, r *http.Request, ps map[string]string) {
	res, err := http.Get(fmt.Sprintf("%v/%v", getServiceDirectStaticUrl(ps["service"]), ps["filepath"]))
	if err != nil {
		if res != nil && res.StatusCode == 404 {
			http.NotFound(w, r)
			return
		} else {
			http.Error(w, "Upstream error", 500)
			log.Println("StaticResourceProxy:", err)
			return
		}
	}
	defer res.Body.Close()

	io.Copy(w, res.Body)
}

func StaticResource(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	http.StripPrefix("/static/", http.FileServer(http.Dir(staticPath))).ServeHTTP(w, r)
}

func main() {
	initConfig()

	sessionManager = scs.NewCookieManager(config.SessionKey)

	router := httptreemux.New()
	router.GET("/", HomePage)
	router.GET("/login", LoginPage)
	router.GET("/logout", LogoutPage)
	router.GET("/course/example/:task", CourseTaskPage)
	router.GET("/profile", ProfilePage)
	router.GET("/static/:service/*filepath", StaticResourceProxy)
	router.GET("/static/*filepath", StaticResource)

	log.Printf("Listening on port %v...\n", config.Port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.Port), router))
}
