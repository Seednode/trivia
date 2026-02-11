/*
Copyright © 2026 Seednode <seednode@seedno.de>
*/

package main

import (
	"net/http"
	"strings"
)

func setCookie(name, value string, w http.ResponseWriter) {
	cookie := http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   31556952,
		HttpOnly: false,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(w, &cookie)
}

func getCategories(r *http.Request, questions *Questions) []string {
	cookie := getCookie(r, "enabledCategories")
	if cookie == "" || !settings {
		return questions.CategoryStrings()
	}

	return strings.Split(cookie, ",")
}

func getCookie(r *http.Request, name string) string {
	cookie, err := r.Cookie(name)

	if err != nil {
		return ""
	}

	return cookie.Value
}

func getTheme(r *http.Request) string {
	value := getCookie(r, "colorTheme")

	if value != "lightMode" {
		return "darkMode"
	}

	return "lightMode"
}
