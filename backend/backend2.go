package main

import (
	"encoding/json"
	"fmt"
	"github.com/auth0/go-jwt-middleware"
	"github.com/codegangsta/negroni"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"time"
	// database
	"database/sql"
	"github.com/coopernurse/gorp"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

const (
	privKeyPath = "demo.rsa"     // openssl genrsa -out app.rsa keysize
	pubKeyPath  = "demo.rsa.pub" // openssl rsa -in app.rsa -pubout > app.rsa.pub
	API_KEY     = "http://localhost:3000/api/files/"
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

type Docs []Doc

type Doc struct {
	Title string
	Url   string
	Cat   string
	Date  int
}

func (s Docs) Len() int {
	return len(s)
}
func (s Docs) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s Docs) Less(i, j int) bool {
	return s[i].Date > s[j].Date
}

var docs Docs

type User struct {
	Id       int
	Username string `json: username`
	Password string `json: password`
	Hash     []byte
}

type Server struct {
	Ren *render.Render
	db  *gorp.DbMap
}

type Token struct {
	Token string
}

var admin User

func main() {

	s := Server{
		render.New(),
		initDb(),
	}

	docs = Docs{}

	doc1 := Doc{"Board Agenda May 4 2013", API_KEY + "BoardAgenda4May2013.pdf", "Meeting Minutes", 20130504}
	doc2 := Doc{"Board Agenda March 8 2014", API_KEY + "BoardAgenda8March2014.pdf", "Meeting Minutes", 20140308}
	doc3 := Doc{"Board Agenda October 6 2013", API_KEY + "BoardAgenda6October2013.pdf", "Meeting Minutes", 20131106}
	doc4 := Doc{"Board  Meeting Minutes July 20 2013", API_KEY + "BoardMeetingMinutesJuly202013.pdf", "Meeting Minutes", 20130720}
	doc5 := Doc{"Fall 2013 Newsletter", API_KEY + "Fall13Newsletter.pdf", "NewsLetter", 20131001}
	doc6 := Doc{"Spring 2013 Newsletter", API_KEY + "AnnualMeeting2013.pdf", "NewsLetter", 20130401}
	doc7 := Doc{"2013 Annual and Organizational Meeting Minutes", API_KEY + "2013AnnualandOrganizationalMeetingMinutes.pdf", "Meeting Minutes", 20130401}
	doc8 := Doc{"Mar 8 2014 Minutes", API_KEY + "Mar82014Minutes.pdf", "Meeting Minutes", 20140308}
	doc9 := Doc{"May  2013 Minutes", API_KEY + "May2013Minutes.pdf", "Meeting Minutes", 20130501}
	doc10 := Doc{"Oct  2013 Minutes", API_KEY + "Oct2013Minutes.pdf", "Meeting Minutes", 20131101}
	doc11 := Doc{"Sept 2013 Minutes", API_KEY + "Sept2013Minutes.pdf", "Meeting Minutes", 20130901}
	doc12 := Doc{"Welcometo Walden's Landing April 2013.pdf", API_KEY + "WelcometoWalden'sLandingApril2013.pdf", "Welcome To Walden", 20130401}
	doc13 := Doc{"Record of WLCA Actions Between Meeting Sept Oct 2012", API_KEY + "RecordofWLCAActionsBetweenMeetingsSepOct2012.pdf", "Meeting Minutes", 20120901}

	docs = append(docs, doc1, doc2, doc3, doc4, doc5, doc6, doc7, doc8, doc9, doc10, doc11, doc12, doc13)
	sort.Sort(docs)
	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: parseToken,
		CredentialsOptional: true,
		Debug:               false,
	})

	router := mux.NewRouter()
	apiRoutes := mux.NewRouter()
	n := negroni.New()

	n.Use(negroni.HandlerFunc(CorsMiddleware))

	router.HandleFunc("/login", s.LoginHandler).Methods("POST")
	router.HandleFunc("/register", s.RegisterHandler).Methods("POST")
	apiRoutes.HandleFunc("/api/me", s.GetUserHandler).Methods("GET")
	apiRoutes.HandleFunc("/api/docs/", s.GetDocsHandler).Methods("GET")
	apiRoutes.HandleFunc("/api/files/{doc}", s.GetFileHandler).Methods("GET")
	apiRoutes.HandleFunc("/api/files/", s.PostFileHandler).Methods("POST")

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
	username := parsedToken.Claims["user"].(string)

	user, err := s.checkUser(username)
	// fmt.Println(userAdmin)
	if err != nil {
		http.Error(w, "Invalid User", http.StatusInternalServerError)
		return
	}

	s.Ren.JSON(w, http.StatusOK, &user)

}

func (s *Server) getUserAndAuth(username string, password string) (User, error) {
	user := User{}
	err := s.db.SelectOne(&user, "select * from users where Username=?", username)
	fmt.Println(user.Username)
	if err != nil {
		return user, err
	}

	if user.Password != password {
		return user, err
	}
	err = bcrypt.CompareHashAndPassword(user.Hash, []byte(password))
	if err != nil {
		return user, err
	}

	return user, nil
}

func (s *Server) checkUser(username string) (User, error) {
	user := User{}
	err := s.db.SelectOne(&user, "select * from users where Username=?", username)
	fmt.Println(user.Username)
	if err != nil {
		fmt.Println(err.Error())
		return user, err
	}
	return user, nil
}

func (s *Server) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	user := User{}
	jsonDecoder := json.NewDecoder(r.Body)
	err := jsonDecoder.Decode(&user)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("User decoded")
	// check to see if user exists
	// user2 := User{}
	// err = s.db.SelectOne(&user2, "select * from users where Username=?", user.Username)
	// fmt.Println(user2.Username)
	// if err == nil {
	// 	fmt.Println(err.Error())
	// }
	// insert user in db

	_, err = s.checkUser(user.Username)
	if err != nil {
		fmt.Println(err.Error())
		// http.Error(w, "User already has an account", http.StatusInternalServerError)
		// return
	}
	hash := []byte(user.Password)
	user.Hash, err = bcrypt.GenerateFromPassword(hash, bcrypt.DefaultCost)
	checkErr(err, "hash failed")
	user.Password = ""
	err = s.db.Insert(&user)
	checkErr(err, "Insert failed")

	// Token
	t := jwt.New(jwt.GetSigningMethod("RS256"))
	t.Claims["user"] = user.Username
	t.Claims["exp"] = time.Now().Add(time.Minute * 60 * 24).Unix()
	tokenString, err := t.SignedString(signKey)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tok := Token{
		tokenString,
	}

	// Send token
	s.Ren.JSON(w, http.StatusOK, &tok)

}

func (s *Server) LoginHandler(w http.ResponseWriter, r *http.Request) {
	user := User{}
	jsonDecoder := json.NewDecoder(r.Body)
	err := jsonDecoder.Decode(&user)
	if err != nil {
		fmt.Println(err.Error())
	}

	user2, err := s.getUserAndAuth(user.Username, user.Password)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
	}
	fmt.Println(user2)
	// token
	t := jwt.New(jwt.GetSigningMethod("RS256"))
	t.Claims["user"] = user.Username
	t.Claims["exp"] = time.Now().Add(time.Minute * 60 * 24).Unix()
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

func (s *Server) PostFileHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "hit file server", http.StatusInternalServerError)
}

func initDb() *gorp.DbMap {
	// connect to db using standard Go database/sql API
	// use whatever database/sql driver you wish
	db, err := sql.Open("sqlite3", "db.sql")
	checkErr(err, "sql.Open failed")

	// construct a gorp DbMap
	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}

	// add a table, setting the table name to 'posts' and
	// specifying that the Id property is an auto incrementing PK
	dbmap.AddTableWithName(User{}, "users").SetKeys(true, "Id")

	// create the table. in a production system you'd generally
	// use a migration tool, or create the tables via scripts
	err = dbmap.CreateTablesIfNotExists()
	checkErr(err, "Create tables failed")

	return dbmap
}

func checkErr(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
		fmt.Println(msg, err)
	}
}
