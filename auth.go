package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"appengine"
	"appengine/datastore"
	"appengine/urlfetch"
)

func auth(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)
	query := r.URL.Query()
	var s SessionData
	if len(session.Value) > 0 {
		json.Unmarshal(session.Value, &s)
	}
	state := query.Get("state")
	if state == s.State {
		c := appengine.NewContext(r)
		client := urlfetch.Client(c)
		accessCode := query.Get("code")
		data := url.Values{}
		data.Set("grant_type", "authorization_code")
		data.Add("code", accessCode)
		req, err := http.NewRequest("POST", tokenURL, bytes.NewBufferString(data.Encode()))
		req.Header.Add("Authorization", "Basic "+basicAuth(myClientID, mySecret))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		resp, err := client.Do(req)
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// put Error Page here, foundDefs.Execute(w, "Error")
			return
		}
		response := string(body)
		accessToken := parsePost(response, "\"access_token\":\"", "\",\"expires_in")
		username := parsePost(response, "\"user_id\":\"", "\"}")
		user := User{
			UserName:    username,
			AccessToken: accessToken,
		}
		key := datastore.NewKey(c, "Users", username, 0, nil)
		_, err = datastore.Put(c, key, &user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		createSession(w, r, user)
		http.Redirect(w, r, "/create", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func parsePost(response string, front string, end string) string {
	s := strings.Split(response, front)[1]
	return strings.Split(s, end)[0]
}
