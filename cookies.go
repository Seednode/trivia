/*
Copyright © 2025 Seednode <seednode@seedno.de>
*/

package main

import (
	"net/http"
)

func readCookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}

	return cookie.Value, nil
}

func getTheme(r *http.Request) string {
	value, err := readCookie(r, "colorTheme")

	if err != nil || value != "lightMode" {
		return "darkMode"
	}

	return "lightMode"
}
