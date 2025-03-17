/*
Copyright © 2025 Seednode <seednode@seedno.de>
*/

package main

import (
	"bufio"
	"crypto/sha1"
	"crypto/sha256"
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
	"slices"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

type Color struct {
	Hex  string
	Hash string
}

const (
	darkRed   string = "#ff000d"
	lightBlue string = "#add8e6"
)

var (
	DefaultColor = Color{
		Hex:  lightBlue,
		Hash: getChecksum(lightBlue),
	}

	ErrorColor = Color{
		Hex:  darkRed,
		Hash: getChecksum(darkRed),
	}
)

type Question struct {
	Version  string
	Theme    string
	Question any
	Answer   any
	Category Category
	Color    string
	Settings any
}

type Trivia struct {
	Question string
	Answer   string
	Category Category
}

func (t *Trivia) getId() QuestionId {
	sha1hash := sha1.New()
	sha1hash.Write([]byte(t.Question + t.Answer + t.Category.String()))
	sha1string := hex.EncodeToString(sha1hash.Sum(nil))

	return QuestionId(uuid.NewSHA1(uuid.NameSpaceURL, []byte(sha1string)).String())
}

type Category string

func (c Category) String() string {
	return string(c)
}

type QuestionId string

func (q QuestionId) String() string {
	return string(q)
}

type Questions struct {
	mu sync.RWMutex

	// Index is a mapping of a string representing a trivia category
	// to the UUIDv5 identifiers of all questions in that category
	index map[Category][]QuestionId

	// List is a mapping of a UUIDv5 string representing a trivia question
	// to a pointer to the struct itself
	list map[QuestionId]*Trivia
}

func (q *Questions) CategoryBytes() []byte {
	var list []byte

	q.mu.RLock()
	for category := range q.index {
		list = fmt.Appendf(list, "%s\n", category)
	}
	q.mu.RUnlock()

	return list
}

func (q *Questions) CategoryStrings() []string {
	var list []string

	q.mu.RLock()
	for category := range q.index {
		list = append(list, category.String())
	}
	q.mu.RUnlock()

	slices.Sort(list)

	return list
}

func (q *Questions) getRandomId(r *http.Request) QuestionId {
	categories := getCategories(r, q)

	ids := []QuestionId{}

	q.mu.RLock()
	defer q.mu.RUnlock()

	for _, i := range categories {
		ids = append(ids, q.index[Category(i)]...)
	}

	if len(ids) < 1 {
		return "00000000-0000-0000-0000-000000000000"
	}

	return ids[rand.IntN(len(ids))]
}

func (q *Questions) getTrivia(id QuestionId) *Trivia {
	q.mu.RLock()
	t, exists := q.list[id]
	q.mu.RUnlock()

	if !exists {
		return nil
	}

	return t
}

func getQuestionTemplate() string {
	return `<!DOCTYPE html>
<html lang="en-US">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <meta name="Description" content="A very basic trivia webapp." />
    <title>Trivia v{{.Version}}</title>
    <link rel="stylesheet" href="/css/{{.Theme}}.css" />
	<link rel="stylesheet" href="/css/trivia.css" />
    <style>.footer {background-color:{{.Color}};}</style>
    <script src="/js/toggleAnswer.js" defer></script>
    <link rel="apple-touch-icon" sizes="180x180" href="/favicons/apple-touch-icon.webp" />
    <link rel="icon" type="image/webp" sizes="32x32" href="/favicons/favicon-32x32.webp" />
    <link rel="icon" type="image/webp" sizes="16x16" href="/favicons/favicon-16x16.webp" />
    <link rel="manifest" href="/favicons/site.webmanifest" crossorigin="use-credentials" />
    <link rel="mask-icon" href="/favicons/safari-pinned-tab.svg" color="#5bbad5" />
    <meta name="msapplication-TileColor" content="#da532c" />
    <meta name="theme-color" content="#ffffff" />
    <meta property="og:site_name" content="https://github.com/Seednode/trivia" />
    <meta property="og:title" content="Trivia v{{.Version}}" />
    <meta property="og:description" content="A very basic trivia webapp." />
    <meta property="og:url" content="https://github.com/Seednode/trivia" />
    <meta property="og:type" content="website" />
    <meta property="og:image" content="/favicons/apple-touch-icon.webp" />
  </head>
  <body>
  {{.Settings}}
    <p id="hint">(Click on the question to load a new one)</p>
    <a href="/"><p id="question">{{.Question}}</p></a>
    <button id="toggle-answer">Show Answer</button>
    <div id="answer"><p>{{.Answer}}</p></div>
    <div class="footer"><p>{{.Category}}</p></div>
  </body>
</html>`
}

func getChecksum(hex string) string {
	h := sha256.New()
	h.Write(fmt.Appendf(nil, ".footer {background-color:%s;}", hex))

	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func loadColors(path string, errorChannel chan<- error) map[Category]Color {
	if colorsFile == "" {
		return map[Category]Color{}
	}

	startTime := time.Now()

	colors := map[Category]Color{}

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

		split := strings.Split(line, "|")

		if len(split) != 2 {
			if verbose {
				fmt.Printf("Invalid color mapping: `%s`. Skipping.\n", line)
			}

			continue
		}

		category := Category(strings.TrimSpace(split[0]))
		hex := strings.TrimSpace(split[1])

		colors[category] = Color{
			Hex:  hex,
			Hash: getChecksum(hex),
		}
	}

	if verbose {
		fmt.Printf("%s | Loaded %d color mappings in %s\n",
			startTime.Format(logDate),
			len(colors),
			time.Since(startTime))
	}

	return colors
}

func normalizePath(path string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	if path == "~" {
		path = homeDir
	} else if strings.HasPrefix(path, "~/") {
		path = filepath.Join(homeDir, path[2:])
	}

	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}

	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	return absolutePath, nil
}

func validatePaths(args []string) ([]string, error) {
	var paths []string

	for i := range args {
		path, err := normalizePath(args[i])
		if err != nil {
			return nil, err
		}

		paths = append(paths, path)
	}

	return paths, nil
}

func walkPath(path string, index map[Category][]QuestionId, list map[QuestionId]*Trivia, errorChannel chan<- error) ([]string, []string) {
	triviaIndex, categoryIndex := []string{}, []string{}

	nodes, err := os.ReadDir(path)
	switch {
	case errors.Is(err, syscall.ENOTDIR):
		if extension == "" || filepath.Ext(path) == extension {
			loadFromFile(path, index, list, errorChannel)
		}
	case err != nil:
		errorChannel <- err
	default:
		for _, node := range nodes {
			fullPath := filepath.Join(path, node.Name())

			switch {
			case !node.IsDir() && (extension == "" || filepath.Ext(node.Name()) == extension):
				loadFromFile(fullPath, index, list, errorChannel)
			case node.IsDir() && recursive:
				walkPath(fullPath, index, list, errorChannel)
			}
		}
	}

	return triviaIndex, categoryIndex
}

func loadFromFile(path string, index map[Category][]QuestionId, list map[QuestionId]*Trivia, errorChannel chan<- error) {
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

	l := 0

	for s.Scan() {
		l += 1

		line := s.Text()

		if line == "" {
			continue
		}

		var question, answer string
		var category Category

		split := strings.Split(line, "|")

		switch {
		case len(split) == 2 || len(split) == 3 && split[2] == "":
			question = strings.TrimSpace(split[0])
			answer = strings.TrimSpace(split[1])
			category = "Uncategorized"
		case len(split) == 3:
			question = strings.TrimSpace(split[0])
			answer = strings.TrimSpace(split[1])
			category = Category(strings.TrimSpace(split[2]))
		default:
			if verbose {
				fmt.Printf("%s | Skipped invalid entry at %s:%d\n",
					time.Now().Format(logDate),
					path,
					l)
			}

			continue
		}

		t := &Trivia{question, answer, category}

		id := t.getId()

		_, exists := list[id]
		if exists {
			if verbose {
				fmt.Printf("%s | Skipped duplicate entry at %s:%d\n",
					time.Now().Format(logDate),
					path,
					l)
			}

			continue
		}

		index[category] = append(index[category], id)

		list[id] = t
	}
}

func loadQuestions(paths []string, questions *Questions, errorChannel chan<- error) (int, int) {
	startTime := time.Now()

	index := map[Category][]QuestionId{}
	list := map[QuestionId]*Trivia{}

	for i := range paths {
		walkPath(paths[i], index, list, errorChannel)
	}

	if len(index) < 1 || len(list) < 1 {
		fmt.Printf("%s | No supported files found.\n", startTime.Format(logDate))
	}

	for i := range index {
		slices.Sort(index[i])
	}

	questions.mu.Lock()
	questions.index = index
	questions.list = list
	categoryCount := len(questions.index)
	triviaCount := len(questions.list)
	questions.mu.Unlock()

	fmt.Printf("%s | Loaded %d questions across %d categories in %s\n",
		startTime.Format(logDate),
		triviaCount,
		categoryCount,
		time.Since(startTime))

	return triviaCount, categoryCount
}

func serveHome(questions *Questions) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		newUrl := fmt.Sprintf("%s//%s/q/%s",
			r.URL.Scheme,
			r.Host,
			questions.getRandomId(r),
		)

		http.Redirect(w, r, newUrl, http.StatusSeeOther)
	}
}

func serveQuestion(questions *Questions, colors map[Category]Color, tpl *template.Template, errorChannel chan<- error) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		startTime := time.Now()

		w.Header().Set("Content-Type", "text/html;charset=UTF-8")

		if verbose {
			fmt.Printf("%s | %s => %s\n",
				startTime.Format(logDate),
				realIP(r),
				r.RequestURI)
		}

		color := DefaultColor

		q := questions.getTrivia(QuestionId(path.Base(r.URL.Path)))

		if q == nil || len(questions.index) < 1 {
			color = ErrorColor
		} else {
			c, exists := colors[q.Category]
			if exists {
				color = c
			}
		}

		w.Header().Set("Content-Security-Policy", fmt.Sprintf("default-src 'self'; style-src-elem 'self' 'sha256-%s'", color.Hash))

		securityHeaders(w)

		question := Question{
			Version:  ReleaseVersion,
			Theme:    getTheme(r),
			Question: "",
			Answer:   "",
			Category: Category(""),
			Color:    color.Hex,
			Settings: "",
		}

		switch {
		case len(questions.index) < 1:
			question.Question = "How do I load questions into Trivia?"
			question.Answer = template.HTML("See <a id=\"help\" href=\"https://github.com/Seednode/trivia?tab=readme-ov-file#file-format\">the docs</a>.")
			question.Category = "Usage"
		case q == nil:
			question.Question = "Are you sure this URL is correct?"
			question.Answer = template.HTML("If not, please go back to the <a id=\"help\" href=\"/\">homepage</a> and try again.")
			question.Category = "Error"
		case html:
			question.Question = template.HTML(q.Question)
			question.Answer = template.HTML(q.Answer)
			question.Category = q.Category
		default:
			question.Question = q.Question
			question.Answer = q.Answer
			question.Category = q.Category
		}

		if settings {
			question.Settings = template.HTML("<p id=\"settings-link\"><a href=\"/settings\">Settings</a></p>")
		}

		err := tpl.Execute(w, question)
		if err != nil {
			errorChannel <- err
		}
	}
}

func serveCategories(questions *Questions, errorChannel chan<- error) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		_, err := w.Write(questions.CategoryBytes())
		if err != nil {
			errorChannel <- err

			return
		}
	}
}

func registerQuestions(mux *httprouter.Router, colors map[Category]Color, questions *Questions, errorChannel chan<- error) {
	template, err := template.New("question").Parse(getQuestionTemplate())
	if err != nil {
		errorChannel <- err

		return
	}

	mux.GET("/", serveHome(questions))
	mux.GET("/q/*id", serveQuestion(questions, colors, template, errorChannel))
	mux.GET("/categories", serveCategories(questions, errorChannel))
}
