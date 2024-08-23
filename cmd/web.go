/*
Copyright © 2024 Seednode <seednode@seedno.de>
*/

package cmd

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
)

const (
	logDate string = `2006-01-02T15:04:05.000-07:00`
)

func realIP(r *http.Request) string {
	remoteAddr := strings.SplitAfter(r.RemoteAddr, ":")

	if len(remoteAddr) < 1 {
		return r.RemoteAddr
	}

	remotePort := remoteAddr[len(remoteAddr)-1]

	cfIp := r.Header.Get("Cf-Connecting-Ip")
	xRealIp := r.Header.Get("X-Real-Ip")

	switch {
	case cfIp != "":
		return cfIp + ":" + remotePort
	case xRealIp != "":
		return xRealIp + ":" + remotePort
	default:
		return r.RemoteAddr
	}
}

func serverError(w http.ResponseWriter, r *http.Request, i interface{}) {
	if verbose {
		fmt.Printf("%s | %s => %s (Invalid request)\n",
			time.Now().Format(logDate),
			realIP(r),
			r.RequestURI)
	}

	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Add("Content-Type", "text/plain")

	w.Write([]byte("500 Internal Server Error\n"))
}

func serverErrorHandler() func(http.ResponseWriter, *http.Request, interface{}) {
	return serverError
}

func servePage() error {
	timeZone := os.Getenv("TZ")
	if timeZone != "" {
		var err error

		time.Local, err = time.LoadLocation(timeZone)
		if err != nil {
			return err
		}
	}

	if verbose {
		fmt.Printf("%s | trivia v%s\n",
			time.Now().Format(logDate),
			ReleaseVersion)
	}

	bindAddr := net.ParseIP(bind)
	if bindAddr == nil {
		return errors.New("invalid bind address provided")
	}

	mux := httprouter.New()

	mux.PanicHandler = serverErrorHandler()

	srv := &http.Server{
		Addr:         net.JoinHostPort(bind, strconv.Itoa(int(port))),
		Handler:      mux,
		IdleTimeout:  10 * time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Minute,
	}

	errorChannel := make(chan error)

	go func() {
		for err := range errorChannel {
			switch {
			case exitOnError:
				fmt.Printf("%s | FATAL: %v\n", time.Now().Format(logDate), err)
			case errors.Is(err, os.ErrNotExist) || errors.Is(err, os.ErrPermission):
				continue
			default:
				fmt.Printf("%s | ERROR: %v\n", time.Now().Format(logDate), err)
			}
		}
	}()

	questions := &Questions{
		index: []string{},
		list:  map[string]Trivia{},
	}

	loadQuestions(questions, errorChannel)

	registerFavicons(mux, errorChannel)

	registerCss(mux, errorChannel)

	registerJs(mux, errorChannel)

	if export {
		registerExport(mux, questions, errorChannel)
	}

	if profile {
		registerProfile(mux)
	}

	if reload {
		registerReload(mux, questions, errorChannel)
	}

	if reloadInterval != "" {
		quit := make(chan struct{})
		defer close(quit)

		registerReloadInterval(questions, quit, errorChannel)
	}

	registerQuestions(mux, questions, errorChannel)

	mux.GET("/version", serveVersion(errorChannel))

	if verbose {
		fmt.Printf("%s | Listening on http://%s/\n",
			time.Now().Format(logDate),
			srv.Addr)
	}

	err := srv.ListenAndServe()

	return err
}
