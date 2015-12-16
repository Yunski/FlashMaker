package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"appengine"
	"appengine/memcache"
)

func serveTemp(w http.ResponseWriter, r *http.Request, name string) {
	session := getSession(r)
	if len(session.Value) > 0 {
		var s SessionData
		json.Unmarshal(session.Value, &s)
		s.LoggedIn = true
		t.ExecuteTemplate(w, name, s)
	} else {
		c := appengine.NewContext(r)
		item, err := memcache.Get(c, name)
		if err != nil {
			buffer := new(bytes.Buffer)
			writer := io.MultiWriter(w, buffer)
			t.ExecuteTemplate(writer, name, SessionData{})
			memcache.Set(c, &memcache.Item{
				Key:   name,
				Value: buffer.Bytes(),
			})
			return
		}
		io.WriteString(w, string(item.Value))
	}
}

func servePage(w http.ResponseWriter, r *http.Request, name string) {
	session := getSession(r)
	if len(session.Value) > 0 {
		var s SessionData
		json.Unmarshal(session.Value, &s)
		s.LoggedIn = true
		t.ExecuteTemplate(w, name, s)
	} else {
		http.Redirect(w, r, "/result", http.StatusTemporaryRedirect)
	}
}

func getSession(r *http.Request) *memcache.Item {
	cookie, err := r.Cookie("session")
	if err != nil {
		return &memcache.Item{}
	}

	c := appengine.NewContext(r)
	item, err := memcache.Get(c, cookie.Value)
	if err != nil {
		return &memcache.Item{}
	}
	return item
}
