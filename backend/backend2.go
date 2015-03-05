package main

import (
	// "bytes"
	"encoding/json"
	"fmt"
	"github.com/auth0/go-jwt-middleware"
	"github.com/codegangsta/negroni"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"io/ioutil"
	"log"
	// "mime"
	"net/http"
	// "os"
	// "string"
	"time"
)

const (
	privKeyPath = "demo.rsa"     // openssl genrsa -out app.rsa keysize
	pubKeyPath  = "demo.rsa.pub" // openssl rsa -in app.rsa -pubout > app.rsa.pub
)

// keys are held in global variables
// i havn't seen a memory corruption/info leakage in go yet
// but maybe it's a better idea, just to store the public key in ram?
// and load the signKey on every signing request? depends on  your usage i guess
var (
	verifyKey, signKey []byte
)

// read the key files before starting http handlers
func init() {
	var err error

	signKey, err = ioutil.ReadFile(privKeyPath)
	if err != nil {
		log.Fatal("Error reading private key")
		return
	}

	verifyKey, err = ioutil.ReadFile(pubKeyPath)
	if err != nil {
		log.Fatal("Error reading private key")
		return
	}
}

type Doc struct {
	Title string
	Url   string
	Cat   string
	Date  int
}

var docs []Doc

type User struct {
	Username string `json: username`
	Password string `json: password`
}

type Server struct {
	Ren *render.Render
}

type Token struct {
	Token string
}

var admin User

func main() {

	s := Server{render.New()}

	docs = []Doc{}

	doc1 := Doc{"Board Agenda May 4 2013", "http://localhost:3000/api/files/BoardAgenda4May2013.pdf", "Meeting Minutes", 20130504}
	doc2 := Doc{"Board Agenda March 8 2014", "http://localhost:3000/api/files/BoardAgenda8March2014.pdf", "Meeting Minutes", 20140308}
	doc3 := Doc{"Board Agenda October 6 2013.pdf", "http://localhost:3000/api/files/BoardAgenda6October2013.pdf", "Meeting Minutes", 20131106}

	docs = append(docs, doc1, doc2, doc3)

	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: parseToken,
		CredentialsOptional: true,
		Debug:               false,
	})

	admin = User{"admin", "admin"}

	router := mux.NewRouter()
	apiRoutes := mux.NewRouter()
	n := negroni.New()

	n.Use(negroni.HandlerFunc(CorsMiddleware))

	router.HandleFunc("/login", s.LoginHandler).Methods("POST")
	apiRoutes.HandleFunc("/api/me", s.GetUserHandler).Methods("GET")
	apiRoutes.HandleFunc("/api/docs/", s.GetDocsHandler).Methods("GET")
	apiRoutes.HandleFunc("/api/files/{doc}", s.GetFileHandler)
	// n.Use(CorsMiddleware)
	router.PathPrefix("/api").Handler(negroni.New(
		negroni.HandlerFunc(jwtMiddleware.HandlerWithNext),
		negroni.Wrap(apiRoutes),
	))
	n.UseHandler(router)
	n.Run(":3000")
}

func parseToken(token *jwt.Token) (interface{}, error) {
	return verifyKey, nil
}

func CorsMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {

	// Stop here if its Preflighted OPTIONS request
	if origin := r.Header.Get("Origin"); origin == "http://localhost:8080" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	}
	// Stop here if its Preflighted OPTIONS request
	if r.Method == "OPTIONS" {
		return
	}

	next(w, r)
}

func (s *Server) GetDocsHandler(w http.ResponseWriter, r *http.Request) {
	s.Ren.JSON(w, http.StatusOK, docs)
}

func (s *Server) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	token, err := jwtmiddleware.FromAuthHeader(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	parsedToken, err := jwt.Parse(token, parseToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	userAdmin := parsedToken.Claims["user"]
	user := User{
		"admin",
		"admin",
	}

	s.Ren.JSON(w, http.StatusOK, &user)

}

func (s *Server) LoginHandler(w http.ResponseWriter, r *http.Request) {
	user := User{}
	jsonDecoder := json.NewDecoder(r.Body)
	err := jsonDecoder.Decode(&user)
	if err != nil {
		fmt.Println(err.Error())
	}
	t := jwt.New(jwt.GetSigningMethod("RS256"))
	t.Claims["user"] = user.Username
	t.Claims["exp"] = time.Now().Add(time.Minute * 1).Unix()
	tokenString, err := t.SignedString(signKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tok := Token{
		tokenString,
	}
	s.Ren.JSON(w, http.StatusOK, &tok)

}

func (s *Server) GetFileHandler(w http.ResponseWriter, r *http.Request) {

	doc := mux.Vars(r)["doc"]
	// temp := string.Split(doc, ".")
	// length := temp.len()
	// w.Header().Set("Content-Type", mime.TypeByExtension(temp[length-1]))

	http.ServeFile(w, r, "docs/"+doc)
}
