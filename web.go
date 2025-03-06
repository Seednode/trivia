/*
Copyright © 2025 Seednode <seednode@seedno.de>
*/

package main

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

func securityHeaders(w http.ResponseWriter) {
	w.Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
	w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
	w.Header().Set("Cross-Origin-Resource-Policy", "same-site")
	w.Header().Set("Permissions-Policy", "geolocation=(), midi=(), sync-xhr=(), microphone=(), camera=(), magnetometer=(), gyroscope=(), fullscreen=(), payment=()")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "SAMEORIGIN")
	w.Header().Set("X-Xss-Protection", "1; mode=block")
}

func realIP(r *http.Request) string {
	remoteAddr := strings.Split(r.RemoteAddr, ":")

	if len(remoteAddr) < 1 {
		return r.RemoteAddr
	}

	remotePort := remoteAddr[len(remoteAddr)-1]

	cfIp := r.Header.Get("Cf-Connecting-Ip")
	xRealIp := r.Header.Get("X-Real-Ip")

	requestor := ""

	switch {
	case cfIp != "":
		if net.ParseIP(cfIp).To4() == nil {
			cfIp = "[" + cfIp + "]"
		}

		requestor = cfIp + ":" + remotePort
	case xRealIp != "":
		if net.ParseIP(cfIp).To4() == nil {
			xRealIp = "[" + xRealIp + "]"
		}

		requestor = xRealIp + ":" + remotePort
	default:
		requestor = r.RemoteAddr
	}

	return requestor
}

func serverError(w http.ResponseWriter, r *http.Request, i any) {
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

func serverErrorHandler() func(http.ResponseWriter, *http.Request, any) {
	return serverError
}

func servePage(args []string) error {
	timeZone := os.Getenv("TZ")
	if timeZone != "" {
		var err error

		time.Local, err = time.LoadLocation(timeZone)
		if err != nil {
			return err
		}
	}

	fmt.Printf("%s | trivia v%s\n",
		time.Now().Format(logDate),
		ReleaseVersion)

	paths, err := validatePaths(args)
	if err != nil {
		return err
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
		IdleTimeout:  1 * time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
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
		list:  map[string]*Trivia{},
	}

	loadQuestions(paths, questions, errorChannel)

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
		registerReload(mux, paths, questions, errorChannel)
	}

	if reloadInterval != "" {
		quit := make(chan struct{})
		defer close(quit)

		registerReloadInterval(paths, questions, quit, errorChannel)
	}

	colors := loadColors(colorsFile, errorChannel)

	registerQuestions(mux, colors, questions, errorChannel)

	mux.GET("/version", serveVersion(errorChannel))

	if tlsKey != "" && tlsCert != "" {
		fmt.Printf("%s | Listening on https://%s/\n",
			time.Now().Format(logDate),
			srv.Addr)

		err = srv.ListenAndServeTLS(tlsCert, tlsKey)
	} else {
		fmt.Printf("%s | Listening on http://%s/\n",
			time.Now().Format(logDate),
			srv.Addr)

		err = srv.ListenAndServe()
	}

	return err
}
