/*
Copyright © 2025 Seednode <seednode@seedno.de>
*/

package main

import (
	"net/http"
)

func setCookie(value string, httpOnly, secure bool, w http.ResponseWriter) error {
	cookie := http.Cookie{
		Name:     "enabledCategories",
		Value:    value,
		Path:     "/",
		MaxAge:   31556952,
		HttpOnly: httpOnly,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(w, &cookie)

	_, err := w.Write([]byte("Set category cookie.\n"))

	return err
}

func getCookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}

	return cookie.Value, nil
}

func getTheme(r *http.Request) string {
	value, err := getCookie(r, "colorTheme")

	if err != nil || value != "lightMode" {
		return "darkMode"
	}

	return "lightMode"
}
