package main

import (
	"fmt"
	"github.com/FIISkIns/ui-service/external"
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
	return "/static/" + service
}

func getServiceDirectStaticUrl(service string) (string, error) {
	if service == "achievements" {
		return external.GetAchievementStaticUrl(), nil
	}

	course := <-external.GetCourseInfo(service)
	if course.Err != nil {
		return "", course.Err
	}
	return fmt.Sprintf("%v/static", course.Val.Url), nil
}

func StaticResourceProxy(w http.ResponseWriter, r *http.Request, ps map[string]string) {
	staticUrl, err := getServiceDirectStaticUrl(ps["service"])
	if err != nil {
		http.Error(w, "Course manager error", http.StatusNotFound)
		log.Println("getServiceDirectStaticUrl:", err)
		return
	}

	res, err := http.Get(fmt.Sprintf("%v/%v", staticUrl, ps["filepath"]))
	if err != nil {
		if res != nil && res.StatusCode == http.StatusNotFound {
			http.NotFound(w, r)
			log.Println("Static asset not found on %v: %v", staticUrl, ps["filepath"])
			return
		} else {
			http.Error(w, "Upstream error", http.StatusInternalServerError)
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
	router.GET("/course/:course", CourseRootPage)
	router.GET("/course/:course/:task", CourseTaskPage)
	router.GET("/profile", ProfilePage)
	router.GET("/ping", StatsPing)
	router.GET("/static/*filepath", StaticResource)
	router.GET("/static/:service/*filepath", StaticResourceProxy)

	log.Printf("Listening on port %v...\n", config.Port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.Port), router))
}
