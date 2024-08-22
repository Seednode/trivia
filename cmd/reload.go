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

func serveReload(questions *Questions, errorChannel chan<- error) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		startTime := time.Now()

		w.Header().Set("Content-Type", "text/html;charset=UTF-8")

		if verbose {
			fmt.Printf("%s | %s => %s\n",
				startTime.Format(logDate),
				realIP(r),
				r.RequestURI)
		}

		count := loadQuestions(questions, errorChannel)

		_, err := w.Write([]byte(fmt.Sprintf("Loaded %d questions in %s.\n", count, time.Since(startTime))))
		if err != nil {
			errorChannel <- err

			return
		}
	}
}

func registerReload(mux *httprouter.Router, questions *Questions, errorChannel chan<- error) {
	mux.GET("/reload", serveReload(questions, errorChannel))
}
