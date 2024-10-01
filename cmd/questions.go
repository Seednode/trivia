/*
Copyright © 2024 Seednode <seednode@seedno.de>
*/

package cmd

import (
	"bufio"
	random "crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"html/template"
	"math/rand/v2"
	"net/http"
	"os"
	"path"
	"path/filepath"
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
<html lang="en-US">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
	<meta name="Description" content="A very basic trivia webapp." />
	<title>Trivia v{{.Version}}</title>
    <link rel="stylesheet" href="/css/question.css" />
	<style nonce="{{.Nonce}}">.footer {background-color:{{.Color}};}</style>
	<script src="/js/darkMode.js"></script>
	<script src="/js/toggleAnswer.js" defer></script>
    <link rel="apple-touch-icon" sizes="180x180" href="/favicons/apple-touch-icon.png" />
	<link rel="icon" type="image/png" sizes="32x32" href="/favicons/favicon-32x32.png" />
	<link rel="icon" type="image/png" sizes="16x16" href="/favicons/favicon-16x16.png" />
	<link rel="manifest" href="/favicons/site.webmanifest" crossorigin="use-credentials" />
	<link rel="mask-icon" href="/favicons/safari-pinned-tab.svg" color="#5bbad5" />
	<meta name="msapplication-TileColor" content="#da532c" />
	<meta name="theme-color" content="#ffffff" />
    <meta property="og:site_name" content="https://github.com/Seednode/trivia" />
    <meta property="og:title" content="Trivia v{{.Version}}" />
    <meta property="og:description" content="A very basic trivia webapp." />
    <meta property="og:url" content="https://github.com/Seednode/trivia" />
    <meta property="og:type" content="website" />
    <meta property="og:image" content="/favicons/apple-touch-icon.png" />
  </head>

  <body>
    <p id="dark-mode">Toggle dark mode</p>
    <p id="hint">(Click on a question to load a new one)</p>
    <a href="/"><p id="question">{{.Question}}</p></a>
	
	<button id="toggle">Show Answer</button>
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

var (
	ErrInvalidFileCountValue = errors.New("no supported files found")

	Colors = map[string]string{
		"Geography":         " ",
		"Entertainment":     "#da6ab2",
		"History":           "#e5cb3a",
		"Arts & Literature": "#7a563c",
		"Science & Nature":  "#157255",
		"Sports & Leisure":  "#db6327",
		"Global View":       "#6d6b82",
		"Sound & Screen":    "#a04251",
		"News":              "#b37e00",
		"The Written Word":  "#7a4e34",
		"Innovations":       "#4f7144",
		"Game Time":         "#a66231",
	}
)

type Trivia struct {
	Question string
	Answer   string
	Category string
}

func (t *Trivia) getId() string {
	sha1hash := sha1.New()
	sha1hash.Write([]byte(t.Question + t.Answer + t.Category))
	sha1string := hex.EncodeToString(sha1hash.Sum(nil))

	return uuid.NewSHA1(uuid.NameSpaceURL, []byte(sha1string)).String()
}

type Questions struct {
	mu    sync.RWMutex
	index []string
	list  map[string]Trivia
}

func (q *Questions) getRandomId() string {
	q.mu.RLock()
	id := q.index[rand.IntN(len(q.index))]
	q.mu.RUnlock()

	return id
}

func (q *Questions) getTrivia(path string) *Trivia {
	q.mu.RLock()
	t := q.list[path]
	q.mu.RUnlock()

	return &t
}

func generateNonce() (string, error) {
	n := make([]byte, 32)
	_, err := random.Read(n)
	if err != nil {
		return "", fmt.Errorf("could not generate nonce")
	}

	return base64.URLEncoding.EncodeToString(n), nil
}

func loadFromFile(path string, list map[string]Trivia, errorChannel chan<- error) []string {
	index := []string{}

	f, err := os.Open(path)
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
			question = strings.TrimSpace(split[0])
			answer = strings.TrimSpace(split[1])
			category = "Uncategorized"
		case len(split) == 3:
			question = strings.TrimSpace(split[0])
			answer = strings.TrimSpace(split[1])
			category = strings.TrimSpace(split[2])
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

	return index
}

func walkPath(path string, list map[string]Trivia, errorChannel chan<- error) []string {
	index := []string{}

	nodes, err := os.ReadDir(path)
	if err != nil {
		errorChannel <- err

		return index
	}

	for _, node := range nodes {
		fullPath := filepath.Join(path, node.Name())

		switch {
		case !node.IsDir() && filepath.Ext(node.Name()) == extension:
			index = append(index, loadFromFile(fullPath, list, errorChannel)...)
		case node.IsDir() && recursive:
			index = append(index, walkPath(fullPath, list, errorChannel)...)
		}
	}

	return index
}

func loadQuestions(questions *Questions, errorChannel chan<- error) int {
	startTime := time.Now()

	index := []string{}
	list := map[string]Trivia{}

	if len(paths) > 0 {
		for i := 0; i < len(paths); i++ {
			index = append(index, walkPath(paths[i], list, errorChannel)...)
		}
	}

	if len(files) > 0 {
		for i := 0; i < len(files); i++ {
			index = append(index, loadFromFile(files[i], list, errorChannel)...)
		}
	}

	if len(index) < 1 || len(list) < 1 {
		fmt.Printf("%s | No supported files found. Exiting.\n", startTime.Format(logDate))

		os.Exit(1)
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
		newUrl := fmt.Sprintf("%s//%s/q/%s",
			r.URL.Scheme,
			r.Host,
			questions.getRandomId(),
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

			return
		}

		w.Header().Set("Content-Security-Policy", fmt.Sprintf("default-src 'self' 'nonce-%s'", nonce))

		if verbose {
			fmt.Printf("%s | %s => %s\n",
				startTime.Format(logDate),
				realIP(r),
				r.RequestURI)
		}

		q := questions.getTrivia(path.Base(r.URL.Path))

		color, exists := Colors[q.Category]
		if !exists {
			color = "lightblue"
		}

		data := Template{
			Version:  ReleaseVersion,
			Question: q.Question,
			Answer:   q.Answer,
			Category: q.Category,
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
