/*
Copyright © 2024 Seednode <seednode@seedno.de>
*/

package cmd

import (
	"context"
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
	logDate string        = `2006-01-02T15:04:05.000-07:00`
	timeout time.Duration = 10 * time.Second
)

type Error struct {
	Message error
	Host    string
	Path    string
}

func realIP(r *http.Request, includePort bool) string {
	fields := strings.SplitAfter(r.RemoteAddr, ":")

	host := strings.TrimSuffix(strings.Join(fields[:len(fields)-1], ""), ":")
	port := fields[len(fields)-1]

	if host == "" {
		return r.RemoteAddr
	}

	cfIP := r.Header.Get("Cf-Connecting-Ip")
	xRealIP := r.Header.Get("X-Real-Ip")

	switch {
	case cfIP != "" && includePort:
		return cfIP + ":" + port
	case cfIP != "":
		return cfIP
	case xRealIP != "" && includePort:
		return xRealIP + ":" + port
	case xRealIP != "":
		return xRealIP
	case includePort:
		return host + ":" + port
	default:
		return host
	}
}

func serverError(w http.ResponseWriter, r *http.Request, i interface{}) {
	if verbose {
		fmt.Printf("%s | %s => %s (Invalid request)\n",
			time.Now().Format(logDate),
			realIP(r, true),
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

	errorChannel := make(chan Error)

	go func() {
		for err := range errorChannel {
			if err.Host == "" {
				err.Host = "local"
			}

			fmt.Printf("%s | %s => %s (Error: `%v`)\n",
				time.Now().Format(`2006-01-02T15:04:05Z07:00`),
				err.Host,
				err.Path,
				err.Message)

			if exitOnError {
				fmt.Printf("%s | Error: Shutting down...\n", time.Now().Format(logDate))

				srv.Shutdown(context.Background())

				break
			}
		}
	}()

	if verbose {
		fmt.Printf("%s | Listening on http://%s/\n",
			time.Now().Format(logDate),
			srv.Addr)
	}

	err := srv.ListenAndServe()

	return err
}
