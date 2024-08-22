/*
Copyright © 2024 Seednode <seednode@seedno.de>
*/

package cmd

import (
	"net/http/pprof"

	"github.com/julienschmidt/httprouter"
)

func registerProfile(mux *httprouter.Router) {
	mux.Handler("GET", "/pprof/allocs", pprof.Handler("allocs"))
	mux.Handler("GET", "/pprof/block", pprof.Handler("block"))
	mux.Handler("GET", "/pprof/goroutine", pprof.Handler("goroutine"))
	mux.Handler("GET", "/pprof/heap", pprof.Handler("heap"))
	mux.Handler("GET", "/pprof/mutex", pprof.Handler("mutex"))
	mux.Handler("GET", "/pprof/threadcreate", pprof.Handler("threadcreate"))
	mux.HandlerFunc("GET", "/pprof/cmdline", pprof.Cmdline)
	mux.HandlerFunc("GET", "/pprof/profile", pprof.Profile)
	mux.HandlerFunc("GET", "/pprof/symbol", pprof.Symbol)
	mux.HandlerFunc("GET", "/pprof/trace", pprof.Trace)
}
