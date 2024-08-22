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
	logDate string        = `2006-01-02T15:04:05.000-07:00`
	timeout time.Duration = 10 * time.Second
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

func humanReadableSize(bytes int) string {
	unit := 1000

	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := unit, 0

	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB",
		float64(bytes)/float64(div),
		"kMGTPE"[exp])
}

func serveVersion(errorChannel chan<- error) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		startTime := time.Now()

		data := []byte(fmt.Sprintf("roulette v%s\n", ReleaseVersion))

		w.Header().Add("Content-Security-Policy", "default-src 'self';")

		w.Header().Set("Content-Type", "text/plain;charset=UTF-8")

		w.Header().Set("Content-Length", strconv.Itoa(len(data)))

		written, err := w.Write(data)
		if err != nil {
			errorChannel <- err

			return
		}

		if verbose {
			fmt.Printf("%s | SERVE: Version page (%s) to %s in %s\n",
				startTime.Format(logDate),
				humanReadableSize(written),
				realIP(r),
				time.Since(startTime).Round(time.Microsecond),
			)
		}
	}
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
		list: []Trivia{},
	}

	loadQuestions(questions, errorChannel)

	if profile {
		registerProfile(mux)
	}

	if reload {
		registerReload(mux, questions, errorChannel)
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
