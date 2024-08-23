/*
Copyright © 2024 Seednode <seednode@seedno.de>
*/

package cmd

import (
	"bufio"
	"crypto/md5"
	random "crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"html/template"
	"math/rand/v2"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

const (
	redirectStatusCode int = http.StatusSeeOther

	tpl = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta name="viewport" content="width=device-width, initial-scale=1">
	<meta name="Description" content="A very basic trivia webapp." />
    <meta charset="utf-8" />
	<title>Trivia v{{.Version}}</title>
    <link rel="stylesheet" href="/css/question.css" />
	<style nonce="{{.Nonce}}">.footer {background-color:{{.Color}};}</style>
	<script src="/js/toggleAnswer.js"></script>
    <link rel="apple-touch-icon" sizes="180x180" href="/favicons/apple-touch-icon.png">
	<link rel="icon" type="image/png" sizes="32x32" href="/favicons/favicon-32x32.png">
	<link rel="icon" type="image/png" sizes="16x16" href="/favicons/favicon-16x16.png">
	<link rel="manifest" href="/favicons/site.webmanifest">
	<link rel="mask-icon" href="/favicons/safari-pinned-tab.svg" color="#5bbad5">
	<meta name="msapplication-TileColor" content="#da532c">
	<meta name="theme-color" content="#ffffff">
    <meta property="og:site_name" content="https://github.com/Seednode/trivia"/>
    <meta property="og:title" content="Trivia v{{.Version}}"/>
    <meta property="og:description" content="A very basic trivia webapp."/>
    <meta property="og:url" content="https://github.com/Seednode/trivia"/>
    <meta property="og:type" content="website"/>
    <meta property="og:image" content="https://raw.githubusercontent.com/Seednode/trivia/master/cmd/favicons/apple-touch-icon.png"/>
  </head>

  <body>
    <p id="hint">(Click on a question to load a new one)</p>
    <a href="/"><p id="question">{{.Question}}</p></a>
	<button onclick="toggleAnswer()">Show Answer</button>
    <div id="answer"><p>{{.Answer}}</p></div>
    <div class="footer"><p>{{.Category}}</p></div>
  </body>
</html>
`
)

type Template struct {
	Version  string
	Question string
	Answer   string
	Category string
	Color    string
	Nonce    string
}

var Colors = map[string]string{
	"Geography":         "#329cd8",
	"Entertainment":     "#da6ab2",
	"History":           "#e5cb3a",
	"Arts & Literature": "#7a563c",
	"Science & Nature":  "#157255",
	"Sports & Leisure":  "#db6327",
}

type Trivia struct {
	question string
	answer   string
	category string
}

func (t *Trivia) getId() string {
	md5hash := md5.New()
	md5hash.Write([]byte(t.question + t.answer + t.category))
	md5string := hex.EncodeToString(md5hash.Sum(nil))

	return uuid.NewMD5(uuid.NameSpaceURL, []byte(md5string)).String()
}

type Questions struct {
	mu    sync.RWMutex
	index []string
	list  map[string]Trivia
}

func (q *Questions) getRandomId() string {
	q.mu.RLock()
	n := rand.IntN(len(q.index))
	id := q.index[n]
	q.mu.RUnlock()

	return id
}

func generateNonce() (string, error) {
	nonceBytes := make([]byte, 32)
	_, err := random.Read(nonceBytes)
	if err != nil {
		return "", fmt.Errorf("could not generate nonce")
	}

	return base64.URLEncoding.EncodeToString(nonceBytes), nil
}

func loadQuestions(questions *Questions, errorChannel chan<- error) int {
	startTime := time.Now()

	index := []string{}
	list := map[string]Trivia{}

	for i := 0; i < len(files); i++ {
		f, err := os.Open(files[i])
		if err != nil {
			errorChannel <- err
		}
		defer func() {
			err = f.Close()
			if err != nil {
				errorChannel <- err
			}
		}()

		s := bufio.NewScanner(f)
		b := make([]byte, 0, 64*1024)
		s.Buffer(b, 1024*1024)
		s.Split(bufio.ScanLines)

		for s.Scan() {
			line := s.Text()

			if line == "" {
				continue
			}

			var question, answer, category string

			split := strings.Split(line, "|")

			switch {
			case len(split) == 2 || len(split) == 3 && split[2] == "":
				question = split[0]
				answer = split[1]
				category = "Uncategorized"
			case len(split) == 3:
				question = split[0]
				answer = split[1]
				category = split[2]
			default:
				if verbose {
					fmt.Printf("Invalid trivia entry: `%s`. Skipping.\n", line)
				}

				continue
			}

			t := Trivia{question, answer, category}

			id := t.getId()

			index = append(index, id)
			list[id] = t
		}
	}

	questions.mu.Lock()
	questions.index = index
	questions.list = list
	length := len(questions.list)
	questions.mu.Unlock()

	if verbose {
		fmt.Printf("%s | Loaded %d questions in %s\n",
			startTime.Format(logDate),
			length,
			time.Since(startTime))
	}

	return length
}

func serveHome(questions *Questions) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		q := questions.getRandomId()

		newUrl := fmt.Sprintf("%s//%s/q/%s",
			r.URL.Scheme,
			r.Host,
			q,
		)

		http.Redirect(w, r, newUrl, redirectStatusCode)
	}
}

func serveQuestion(questions *Questions, template *template.Template, errorChannel chan<- error) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		startTime := time.Now()

		w.Header().Set("Content-Type", "text/html;charset=UTF-8")

		nonce, err := generateNonce()
		if err != nil {
			errorChannel <- err
		}

		w.Header().Set("Content-Security-Policy", fmt.Sprintf("default-src 'self'; style-src 'self' 'nonce-%s'", nonce))

		if verbose {
			fmt.Printf("%s | %s => %s\n",
				startTime.Format(logDate),
				realIP(r),
				r.RequestURI)
		}

		q := questions.list[path.Base(r.URL.Path)]

		color, exists := Colors[q.category]
		if !exists {
			color = "lightblue"
		}

		data := Template{
			Version:  ReleaseVersion,
			Question: q.question,
			Answer:   q.answer,
			Category: q.category,
			Color:    color,
			Nonce:    nonce,
		}

		err = template.Execute(w, data)
		if err != nil {
			errorChannel <- err

			return
		}
	}
}

func registerQuestions(mux *httprouter.Router, questions *Questions, errorChannel chan<- error) {
	template, err := template.New("question").Parse(tpl)
	if err != nil {
		errorChannel <- err

		return
	}

	mux.GET("/", serveHome(questions))
	mux.GET("/q/:id", serveQuestion(questions, template, errorChannel))
}
