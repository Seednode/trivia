/*
Copyright © 2025 Seednode <seednode@seedno.de>
*/

package main

import (
	"net/http"
)

const (
	ThemeCookie string = "colorTheme"
	DarkMode    string = "darkMode"
	LightMode   string = "lightMode"
)

func readCookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}

	return cookie.Value, nil
}

func getTheme(r *http.Request) string {
	value, err := readCookie(r, ThemeCookie)

	if err != nil || value != "lightMode" {
		return "darkMode"
	}

	return "lightMode"
}
