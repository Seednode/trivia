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

func registerReloadInterval(questions *Questions, quit <-chan struct{}, errorChannel chan<- error) {
	interval, err := time.ParseDuration(reloadInterval)
	if err != nil {
		errorChannel <- err

		return
	}

	ticker := time.NewTicker(interval)

	if verbose {
		next := time.Now().Add(interval).Truncate(time.Second)
		fmt.Printf("%s | Next scheduled rebuild will run at %s\n", time.Now().Format(logDate), next.Format(logDate))
	}

	go func() {
		for {
			select {
			case <-ticker.C:
				next := time.Now().Add(interval).Truncate(time.Second)

				if verbose {
					fmt.Printf("%s | Started scheduled index rebuild\n", time.Now().Format(logDate))
				}

				loadQuestions(questions, errorChannel)

				if verbose {
					fmt.Printf("%s | Next scheduled rebuild will run at %s\n", time.Now().Format(logDate), next.Format(logDate))
				}
			case <-quit:
				ticker.Stop()

				return
			}
		}
	}()
}

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
