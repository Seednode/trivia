/*
Copyright © 2024 Seednode <seednode@seedno.de>
*/

package cmd

import (
	"fmt"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

func serveExport(questions *Questions, errorChannel chan<- error) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		startTime := time.Now()

		w.Header().Set("Content-Type", "text/plain;charset=UTF-8")

		if verbose {
			fmt.Printf("%s | %s => %s\n",
				startTime.Format(logDate),
				realIP(r),
				r.RequestURI)
		}

		questions.mu.RLock()
		defer questions.mu.RUnlock()

		for i := 0; i < len(questions.index); i++ {
			entry := questions.list[questions.index[i]]
			_, err := w.Write([]byte(fmt.Sprintf("Category: %s\nQuestion: %s\nAnswer: %s\n\n", entry.Category, entry.Question, entry.Answer)))
			if err != nil {
				errorChannel <- err

				return
			}
		}

	}
}

func registerExport(mux *httprouter.Router, questions *Questions, errorChannel chan<- error) {
	mux.GET("/export", serveExport(questions, errorChannel))
}
