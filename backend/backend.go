package main

import (
	"fmt"
	"github.com/StephanDollberg/go-json-rest-middleware-jwt"
	"github.com/ant0ine/go-json-rest/rest"
	"log"
	"net/http"
	"time"
)

type Doc struct {
	Title string
	Url   string
	Cat   string
	Date  int
}

var docs []Doc

type User struct {
	username string
	password string
}

var admin User

func main() {

	docs = []Doc{}

	admin = User{"admin", "admin"}

	doc1 := Doc{"Board Agenda May 4 2013", "docs/BoardAgenda4May2013.pdf", "Meeting Minutes", 20130504}
	doc2 := Doc{"Board Agenda March 8 2014", "docs/BoardAgenda8March2014.pdf", "Meeting Minutes", 20140308}
	doc3 := Doc{"Board Agenda October 6 2013.pdf", "docs/BoardAgenda6October2013.pdf", "Meeting Minutes", 20131106}

	docs = append(docs, doc1, doc2, doc3)

	jwt_middleware := jwt.JWTMiddleware{
		Key:        []byte("secret key"),
		Realm:      "jwt auth",
		Timeout:    time.Hour,
		MaxRefresh: time.Hour * 24,
		Authenticator: func(userId string, password string) bool {
			if userId == "admin" && password == "admin" {
				return true
			}
			return false
		}}

	cors_middleware := rest.CorsMiddleware{
		RejectNonCorsRequests: false,
		OriginValidator: func(origin string, request *rest.Request) bool {
			return origin == "http://localhost:8080"
		},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders: []string{
			"Accept", "Content-Type", "X-Custom-Header", "Origin", "*"},
		AccessControlAllowCredentials: true,
		AccessControlMaxAge:           3600,
	}

	login_api := rest.NewApi()
	login_api.Use(&cors_middleware)
	login_api.Use(rest.DefaultDevStack...)
	login_router, err := rest.MakeRouter(
		&rest.Route{"POST", "/login", jwt_middleware.LoginHandler},
	)
	if err != nil {
		fmt.Println(err.Error())
	}
	login_api.SetApp(login_router)

	main_api := rest.NewApi()
	main_api.Use(&cors_middleware)
	main_api.Use(&jwt_middleware)
	main_api.Use(rest.DefaultDevStack...)
	main_api_router, err := rest.MakeRouter(
		&rest.Route{"GET", "/me", GetMe},
		&rest.Route{"GET", "/docs", GetDocs},
		&rest.Route{"GET", "/refresh_token", jwt_middleware.RefreshHandler})
	if err != nil {
		fmt.Println(err.Error())
	}
	main_api.SetApp(main_api_router)

	http.Handle("/", login_api.MakeHandler())
	http.Handle("/api/", http.StripPrefix("/api", main_api.MakeHandler()))

	log.Fatal(http.ListenAndServe(":3000", nil))
}

func GetMe(w rest.ResponseWriter, r *rest.Request) {
	w.WriteJson(map[string]string{"authed": r.Env["REMOTE_USER"].(string)})
}

func GetDocs(w rest.ResponseWriter, r *rest.Request) {
	w.WriteJson(&docs)
}
