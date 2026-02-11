/*
Copyright © 2026 Seednode <seednode@seedno.de>
*/

package main

import (
	"embed"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
)

//go:embed favicons/*
var favicons embed.FS

func serveFavicons(errorChannel chan<- error) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		fname := strings.TrimPrefix(r.URL.Path, "/")

		data, err := favicons.ReadFile(fname)
		if err != nil {
			return
		}

		w.Header().Set("Content-Length", strconv.Itoa(len(data)))

		w.Header().Add("Content-Security-Policy", "default-src 'self';")

		securityHeaders(w)

		_, err = w.Write(data)
		if err != nil {
			errorChannel <- err

			return
		}
	}
}

func registerFavicons(mux *httprouter.Router, errorChannel chan<- error) {
	mux.GET("/favicons/:favicon", serveFavicons(errorChannel))
	mux.GET("/favicon.ico", serveFavicons(errorChannel))
}
