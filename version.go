/*
Copyright © 2026 Seednode <seednode@seedno.de>
*/

package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

func serveVersion(errorChannel chan<- error) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		startTime := time.Now()

		data := fmt.Appendf(nil, "trivia v%s\n", ReleaseVersion)

		w.Header().Add("Content-Security-Policy", "default-src 'self';")

		w.Header().Set("Content-Type", "text/plain;charset=UTF-8")

		w.Header().Set("Content-Length", strconv.Itoa(len(data)))

		securityHeaders(w)

		_, err := w.Write(data)
		if err != nil {
			errorChannel <- err

			return
		}

		if verbose {
			fmt.Printf("%s | %s => %s\n",
				startTime.Format(logDate),
				realIP(r),
				r.RequestURI)
		}
	}
}
