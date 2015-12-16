package main

import (
	"encoding/json"
	"html/template"
	"net/http"
	"time"

	"appengine"
	"appengine/datastore"
	"appengine/memcache"

	"github.com/dchest/uniuri"
	"github.com/satori/go.uuid"
)

var (
	t          *template.Template
	myClientID = ""
	mySecret   = ""
	havenKey   = ""
	authFront  = "https://quizlet.com/authorize?"
	client     = "client_id=" + myClientID + "&"
	response   = "response_type=code&"
	scope      = "scope=read write_set&"
	quizletURL = "https://quizlet.com"
	tokenURL   = "https://api.quizlet.com/oauth/token"
	redir      = "http://localhost:8080/auth"
	setURL     = "https://api.quizlet.com/2.0/sets"
	havenURL   = "https://api.havenondemand.com/1/api/sync/ocrdocument/v1"
)

func init() {
	http.HandleFunc("/", home)
	http.HandleFunc("/auth", auth)
	http.HandleFunc("/create", create)
	http.HandleFunc("/redirect", redirect)
	http.HandleFunc("/login", login)
	http.HandleFunc("/login/form", loginPage)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/define", defineText)
	http.HandleFunc("/result", result)
	t = template.Must(template.ParseGlob("templates/*.html"))
}

func home(w http.ResponseWriter, r *http.Request) {
	serveTemp(w, r, "index.html")
}

func create(w http.ResponseWriter, r *http.Request) {
	servePage(w, r, "create.html")
}

func result(w http.ResponseWriter, r *http.Request) {
	serveTemp(w, r, "result.html")
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	serveTemp(w, r, "login.html")
}

func redirect(w http.ResponseWriter, r *http.Request) {
	state := uniuri.New()
	user := User{
		State: state,
	}
	createSession(w, r, user)
	authURL := authFront + response + client + scope + "state=" + state
	http.Redirect(w, r, authURL, http.StatusFound)
}

func login(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	key := datastore.NewKey(c, "Users", r.FormValue("user-name"), 0, nil)
	var user User
	err := datastore.Get(c, key, &user)
	if err != nil {
		http.Redirect(w, r, "/redirect", http.StatusTemporaryRedirect)
		return
	}
	createSession(w, r, user)
	http.Redirect(w, r, "/create", http.StatusFound)
}

func createSession(w http.ResponseWriter, r *http.Request, user User) {
	c := appengine.NewContext(r)
	id := uuid.NewV4()
	cookie := &http.Cookie{
		Name:  "session",
		Value: id.String(),
		Path:  "/",
		//Secure: true,
		//HttpOnly: true,
	}
	http.SetCookie(w, cookie)
	json, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s := memcache.Item{
		Key:        id.String(),
		Value:      json,
		Expiration: time.Duration(3600 * time.Second),
	}
	memcache.Set(c, &s)
}

func logout(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	s := memcache.Item{
		Key:        cookie.Value,
		Value:      []byte(""),
		Expiration: time.Duration(time.Microsecond),
	}
	memcache.Set(c, &s)
	cookie.MaxAge = -1
	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
