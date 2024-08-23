/*
Copyright © 2024 Seednode <seednode@seedno.de>
*/

package cmd

import (
	"embed"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
)

//go:embed js/*
var js embed.FS

func serveJs(errorChannel chan<- error) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Header().Set("Content-Type", "text/javascript; charset=utf-8")

		fname := strings.TrimPrefix(r.URL.Path, "/")

		data, err := js.ReadFile(fname)
		if err != nil {
			return
		}

		w.Header().Set("Content-Length", strconv.Itoa(len(data)))

		_, err = w.Write(data)
		if err != nil {
			errorChannel <- err

			return
		}
	}
}

func registerJs(mux *httprouter.Router, errorChannel chan<- error) {
	mime.AddExtensionType(".js", "application/javascript; charset=utf-8")

	mux.GET("/js/:js", serveJs(errorChannel))
}
