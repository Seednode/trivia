/*
Copyright © 2024 Seednode <seednode@seedno.de>
*/

package cmd

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
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
)

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

var Colors = map[string]string{
	"Geography":         "#329cd8",
	"Entertainment":     "#da6ab2",
	"History":           "#e5cb3a",
	"Arts & Literature": "#7a563c",
	"Science & Nature":  "#157255",
	"Sports & Leisure":  "#db6327",
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

func serveNewQuestion(questions *Questions) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		q := questions.getRandomId()

		newUrl := fmt.Sprintf("http://%s/q/%s",
			r.Host,
			q,
		)

		http.Redirect(w, r, newUrl, redirectStatusCode)
	}
}

func serveQuestion(questions *Questions, errorChannel chan<- error) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		startTime := time.Now()

		w.Header().Set("Content-Type", "text/html;charset=UTF-8")

		if verbose {
			fmt.Printf("%s | %s => %s\n",
				startTime.Format(logDate),
				realIP(r),
				r.RequestURI)
		}

		id := path.Base(r.URL.Path)

		q := questions.list[id]

		color, exists := Colors[q.category]
		if !exists {
			color = "lightblue"
		}

		var htmlBody strings.Builder

		htmlBody.WriteString(`<!DOCTYPE html><html lang="en"><head>`)
		htmlBody.WriteString(`<meta name="viewport" content="width=device-width, initial-scale=1">`)
		htmlBody.WriteString(`<style>html {text-align: center;} a { all: unset;}`)
		htmlBody.WriteString(`body {color: #c9d1d9; background: #0d1117;}`)
		htmlBody.WriteString(`#hint {font-size:12px;}`)
		htmlBody.WriteString(fmt.Sprintf(`.footer {background-color: %s; color: #0d1117; position: fixed; left: 0; bottom: 0; width: 100%%; text-align: center;}`, color))
		htmlBody.WriteString(`p, div {font-size: clamp(var(--min), var(--val), var(--max));}`)
		htmlBody.WriteString(`p, div {--min: 1em; --val: 2.5vw; --max: 1.5em;}`)
		htmlBody.WriteString(`#question {line-height: 1.4; margin-left: auto; margin-right: auto; max-width: 80%; padding-top: 2rem; padding-bottom: 2rem;}`)
		htmlBody.WriteString(`#answer {display: none; margin-left: auto; margin-right: auto; max-width: 80%; padding: 50px 0; text-align: center; width: 100%`)
		htmlBody.WriteString(fmt.Sprintf(`background-color: %s; margin-top: 20px; outline: ridge;}</style>`, "lightgrey"))
		htmlBody.WriteString(fmt.Sprintf("<title>Trivia v%s</title></head>", ReleaseVersion))
		htmlBody.WriteString(`<p id="hint">(Click on the question to load a new one)</p>`)
		htmlBody.WriteString(fmt.Sprintf(`<body><a href="/"><p id="question">%s</p></a>`, q.question))
		htmlBody.WriteString(`<button onclick="toggleAnswer()">Show Answer</button>`)
		htmlBody.WriteString(fmt.Sprintf(`<div id="answer">%s</div>`, q.answer))
		htmlBody.WriteString(`<script>function toggleAnswer() {var x = document.getElementById("answer");`)
		htmlBody.WriteString(`if (x.style.display === "block") {x.style.display = "none";} else {x.style.display = "block";}};`)
		htmlBody.WriteString(`</script>`)
		htmlBody.WriteString(fmt.Sprintf(`<div class="footer"><p>%s</p>`, q.category))
		htmlBody.WriteString(`</body></html>`)

		_, err := w.Write([]byte(htmlBody.String() + "\n"))
		if err != nil {
			errorChannel <- err

			return
		}
	}
}

func registerQuestions(mux *httprouter.Router, questions *Questions, errorChannel chan<- error) {
	mux.GET("/", serveNewQuestion(questions))
	mux.GET("/q/:id", serveQuestion(questions, errorChannel))
}
