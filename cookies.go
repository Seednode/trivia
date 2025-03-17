/*
Copyright © 2025 Seednode <seednode@seedno.de>
*/

package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var (
	ErrValueTooLong = errors.New("cookie length must be <4096")
	ErrInvalidValue = errors.New("invalid value provided for cookie")
)

func setCookie(name, value string, w http.ResponseWriter) error {
	b := []byte(secret)

	cookie := http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   31556952,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
	}

	mac := hmac.New(sha256.New, b)
	mac.Write([]byte(cookie.Name))
	mac.Write([]byte(cookie.Value))
	signature := mac.Sum(nil)

	cookie.Value = base64.URLEncoding.EncodeToString([]byte(string(signature) + cookie.Value))

	http.SetCookie(w, &cookie)

	_, err := w.Write(fmt.Appendf(nil, "Set cookie for %s.\n", name))

	return err
}

func getCategories(r *http.Request, questions *Questions) []string {
	cookie, err := getCookie(r, "enabledCategories")
	if err != nil || cookie == "" {
		return questions.CategoryStrings()
	}

	return strings.Split(cookie, ",")
}

func getCookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", ErrInvalidValue
	}

	v, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return "", ErrInvalidValue
	}

	decodedValue := string(v)

	if len(cookie.Value) < sha256.Size {
		return "", ErrInvalidValue
	}

	signature := decodedValue[:sha256.Size]
	value := decodedValue[sha256.Size:]

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(name))
	mac.Write([]byte(value))
	expectedSignature := mac.Sum(nil)

	if !hmac.Equal([]byte(signature), expectedSignature) {
		return "", ErrInvalidValue
	}

	return value, nil
}

func getTheme(r *http.Request) string {
	value, err := getCookie(r, "colorTheme")

	if err != nil || value != "lightMode" {
		return "darkMode"
	}

	return "lightMode"
}
