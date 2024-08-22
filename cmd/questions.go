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
}

type Questions struct {
	mu   sync.RWMutex
	list []Trivia
}

func parseQuestions(questions string, errorChannel chan<- error) *Questions {
	startTime := time.Now()

	f, err := os.Open(questions)
	if err != nil {
		errorChannel <- err

		return nil
	}
	defer func() {
		err = f.Close()
		if err != nil {
			errorChannel <- err
		}
	}()

	retVal := Questions{
		list: []Trivia{},
	}

	s := bufio.NewScanner(f)
	b := make([]byte, 0, 64*1024)
	s.Buffer(b, 1024*1024)
	s.Split(bufio.ScanLines)

	retVal.mu.Lock()

	for s.Scan() {
		line := s.Text()

		question, answer, found := strings.Cut(line, "|")

		if found {
			retVal.list = append(retVal.list, Trivia{question, answer})
		}
	}

	retVal.mu.Unlock()

	if verbose {
		fmt.Printf("%s | Loaded trivia database in %dms\n",
			startTime.Format(logDate),
			time.Since(startTime).Milliseconds())
	}

	return &retVal
}

func getTrivia(questions *Questions) (string, string) {
	questions.mu.RLock()
	n := rand.IntN(len(questions.list))
	questions.mu.RUnlock()

	return questions.list[n].question, questions.list[n].answer
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

		question, answer := getTrivia(questions)

		var htmlBody strings.Builder

		htmlBody.WriteString(`<!DOCTYPE html><html lang="en"><head>`)
		htmlBody.WriteString(`<meta name="viewport" content="width=device-width, initial-scale=1">`)
		htmlBody.WriteString(`<style>html {text-align: center;} a { all: unset;}`)
		htmlBody.WriteString(`#answer {display: none;width: 100%;padding: 50px 0;text-align: center; background-color: lightblue; margin-top: 20px;}</style>`)
		htmlBody.WriteString(fmt.Sprintf("<title>%s</title></head>", "Trivia"))
		htmlBody.WriteString(fmt.Sprintf(`<body><a href="/"><p id="question">%s</p></a>`, question))
		htmlBody.WriteString(`<button onclick="toggleAnswer()">Show Answer</button>`)
		htmlBody.WriteString(fmt.Sprintf(`<div id="answer">%s</div>`, answer))
		htmlBody.WriteString(`<script>function toggleAnswer() {var x = document.getElementById("answer");`)
		htmlBody.WriteString(`if (x.style.display === "block") {x.style.display = "none";} else {x.style.display = "block";}}</script>`)
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
