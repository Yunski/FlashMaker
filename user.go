package main

//User ...quizlet user
type User struct {
	UserName    string
	AccessToken string `json:"-"`
	State       string `datastore:"-"`
	Key         string `datastore:"-"`
}

//SessionData ...stores session data
type SessionData struct {
	User
	LoggedIn bool
}
