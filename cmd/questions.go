/*
Copyright © 2024 Seednode <seednode@seedno.de>
*/

package cmd

import (
	"bufio"
	"fmt"
	"math/rand/v2"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/julienschmidt/httprouter"
)

type Trivia struct {
	question string
	answer   string
	category string
}

type Questions struct {
	mu   sync.RWMutex
	list []Trivia
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

	list := []Trivia{}

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

			list = append(list, Trivia{question, answer, category})
		}
	}

	questions.mu.Lock()
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

func getTrivia(questions *Questions) (string, string, string) {
	questions.mu.RLock()
	n := rand.IntN(len(questions.list))
	q := questions.list[n]
	questions.mu.RUnlock()

	return q.question, q.answer, q.category
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

		question, answer, category := getTrivia(questions)

		color, exists := Colors[category]
		if !exists {
			color = "lightblue"
		}

		var htmlBody strings.Builder

		htmlBody.WriteString(`<!DOCTYPE html><html lang="en"><head>`)
		htmlBody.WriteString(`<meta name="viewport" content="width=device-width, initial-scale=1">`)
		htmlBody.WriteString(`<style>html {text-align: center;} a { all: unset;}`)
		htmlBody.WriteString(`#hint {font-size:12px;}`)
		htmlBody.WriteString(fmt.Sprintf(`.footer {background-color: %s; position: fixed; left: 0; bottom: 0; width: 100%%; text-align: center;}`, color))
		htmlBody.WriteString(`p, div {font-size: clamp(var(--min), var(--val), var(--max));}`)
		htmlBody.WriteString(`p, div {--min: 1em; --val: 2.5vw; --max: 1.5em;}`)
		htmlBody.WriteString(`#answer {display: none;width: 100%;padding: 50px 0;text-align: center;`)
		htmlBody.WriteString(fmt.Sprintf(`background-color: %s; margin-top: 20px; outline: ridge;}</style>`, "lightgrey"))
		htmlBody.WriteString(fmt.Sprintf("<title>%s</title></head>", "Trivia"))
		htmlBody.WriteString(`<p id="hint">(Click on the question to load a new one)</p>`)
		htmlBody.WriteString(fmt.Sprintf(`<body><a href="/"><p id="question">%s</p></a>`, question))
		htmlBody.WriteString(`<button onclick="toggleAnswer()">Show Answer</button>`)
		htmlBody.WriteString(fmt.Sprintf(`<div id="answer">%s</div>`, answer))
		htmlBody.WriteString(`<script>function toggleAnswer() {var x = document.getElementById("answer");`)
		htmlBody.WriteString(`if (x.style.display === "block") {x.style.display = "none";} else {x.style.display = "block";}}</script>`)
		htmlBody.WriteString(fmt.Sprintf(`<div class="footer"><p>%s</p>`, category))
		htmlBody.WriteString(`</body></html>`)

		_, err := w.Write([]byte(htmlBody.String() + "\n"))
		if err != nil {
			errorChannel <- err

			return
		}
	}
}

func registerQuestions(mux *httprouter.Router, questions *Questions, errorChannel chan<- error) {
	mux.GET("/", serveQuestion(questions, errorChannel))
}
