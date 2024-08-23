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

//go:embed css/*
var css embed.FS

func serveCss(errorChannel chan<- error) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		fname := strings.TrimPrefix(r.URL.Path, "/")

		data, err := css.ReadFile(fname)
		if err != nil {
			return
		}

		w.Header().Set("Content-Length", strconv.Itoa(len(data)))

		w.Header().Set("Content-Type", "text/css; charset=utf-8")

		_, err = w.Write(data)
		if err != nil {
			errorChannel <- err

			return
		}
	}
}

func registerCss(mux *httprouter.Router, errorChannel chan<- error) {
	mime.AddExtensionType(".css", "text/css; charset=utf-8")

	mux.GET("/css/:css", serveCss(errorChannel))
}
