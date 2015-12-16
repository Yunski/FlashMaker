package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"appengine"
	"appengine/datastore"
	"appengine/urlfetch"
)

var (
	fdictURL = "http://dictionary.reference.com/browse/"
	edictURL = "?s=t"
)

func define(w http.ResponseWriter, r *http.Request, title string, words []string, ch chan string) {
	session := getSession(r)
	var s SessionData
	json.Unmarshal(session.Value, &s)

	c := appengine.NewContext(r)
	key := datastore.NewKey(c, "Users", s.UserName, 0, nil)
	var user User
	err := datastore.Get(c, key, &user)

	var terms []string
	var definitions []string
	fDefs := make(map[string]string)
	defs := make(chan map[string]string)
	finish := make(chan bool)
	for i := 0; i < len(words); i++ {
		go search(words[i], retrieve(w, r, words[i]), defs, finish)
	}

	index := 0
	for index < len(words) {
		select {
		case d := <-defs:
			for k, v := range d {
				fDefs[k] = v
			}
		case <-finish:
			index++
		}
	}
	index = 0
	for word, def := range fDefs {
		terms = append(terms, word)
		definitions = append(definitions, def)
		index++
	}

	//c := appengine.NewContext(r)
	client := urlfetch.Client(c)

	data := url.Values{}
	data.Set("title", title)

	for i, term := range terms {
		data.Set("terms["+strconv.Itoa(i)+"]", term)
	}
	for i, defn := range definitions {
		data.Set("definitions["+strconv.Itoa(i)+"]", defn)
	}

	data.Set("lang_terms", "en")
	data.Set("lang_definitions", "en")
	req, err := http.NewRequest("POST", setURL, bytes.NewBufferString(data.Encode()))
	req.Header.Add("Authorization", "Bearer "+user.AccessToken)
	resp, err := client.Do(req)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ch <- "Failed to create quizlet set."
		return
	}

	/*
		list := bytes.NewBufferString("")
		for i := 0; i < len(terms); i++ {
			list.WriteString(strconv.Itoa(i+1) + ". " + terms[i] + "\n" + definitions[i] + "\n")
		}*/
	ch <- string(body)
}

func defineText(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)
	var s SessionData
	if len(session.Value) > 0 {
		json.Unmarshal(session.Value, &s)
		s.LoggedIn = true

		title := r.FormValue("title")
		var words []string
		text := strings.Fields(r.FormValue("content"))
		for _, word := range text {
			words = append(words, word)
		}
		file, _, err := r.FormFile("file")
		if err == nil {
			chw := make(chan []string)
			var fileText []string
			go retrieveFile(w, r, file, chw)
			fileText = <-chw
			for _, word := range fileText {
				words = append(words, word)
			}
		}

		ch := make(chan string)
		go define(w, r, title, words, ch)
		for {
			select {
			case result := <-ch:
				newURL := parsePost(result, "\"url\":\"", "\",\"title")
				url := quizletURL + newURL
				fmt.Fprint(w, url)
				return
			}
		}
	}

	http.Redirect(w, r, "/result", http.StatusTemporaryRedirect)
}

func retrieveFile(w http.ResponseWriter, r *http.Request, file multipart.File, ch chan []string) {
	c := appengine.NewContext(r)
	client := urlfetch.Client(c)

	fileContents, err := ioutil.ReadAll(file)
	if err != nil {
		return
	}
	bodyBuf := &bytes.Buffer{}
	writer := multipart.NewWriter(bodyBuf)
	part, err := writer.CreateFormFile("file", "img")
	if err != nil {
		return
	}
	part.Write(fileContents)
	writer.WriteField("apikey", havenKey)
	writer.Close()

	req, err := http.NewRequest("POST", havenURL, bodyBuf)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	resp, err := client.Do(req)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	text := parsePost(string(body), "\"text\": \"", "\",")
	rep := strings.NewReplacer("\\n", " ")
	text = rep.Replace(text)
	words := strings.Split(text, " ")
	var wordList []string
	for _, s := range words {
		wordList = append(wordList, s)
	}

	ch <- wordList
}

func search(word string, page string, c chan map[string]string, end chan bool) {
	defer func() {
		end <- true
	}()
	wordDef := make(map[string]string)
	content := strings.Split(page, "def-content")
	var twoDefs []string
	if len(content) > 1 {
		twoDefs = content[1:2]
	} else {
		wordDef[word] = "Definition not found."
		c <- wordDef
		return
	}
	var buffer bytes.Buffer

	for _, s := range twoDefs {
		buffer.WriteString(s)
	}
	str := buffer.String()
	lines := strings.Split(str, "\n")
	endTag := regexp.MustCompile(`</.*>`)
	html := regexp.MustCompile(`<.*>`)
	start := regexp.MustCompile(`^\w`)
	parenth := regexp.MustCompile(`^\(`)
	r := strings.NewReplacer("(", "", ")", "")

	var defs []string
	for _, line := range lines {
		if start.MatchString(line) || parenth.MatchString(line) {
			parseEnd := endTag.ReplaceAllString(line, "")
			content := html.ReplaceAllString(parseEnd, "")
			defs = append(defs, r.Replace(content))
			break
		}
	}
	var def string
	if len(defs) > 0 {
		def = defs[0]
	} else {
		def = "Definition not found."
	}
	wordDef[word] = def
	c <- wordDef
}

func retrieve(w http.ResponseWriter, r *http.Request, word string) string {
	c := appengine.NewContext(r)
	client := urlfetch.Client(c)
	resp, err := client.Get(fdictURL + word + edictURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return "Error"
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return string(body)
}
