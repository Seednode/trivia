/*
Copyright © 2025 Seednode <seednode@seedno.de>
*/

package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
)

type SelectedCategories struct {
	Categories []string `json:"categories"`
}

type CategoryToggle struct {
	Version    string
	Theme      string
	Categories any
}

func getSettingsTemplate() string {
	return `<!DOCTYPE html>
<html lang="en-US">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <meta name="Description" content="A very basic trivia webapp." />
    <title>Trivia v{{.Version}}</title>
    <script src="/js/toggleCategories.js" defer></script>
	<script src="/js/toggleTheme.js" defer></script>
    <link rel="stylesheet" href="/css/{{.Theme}}.css" />
	<link rel="stylesheet" href="/css/trivia.css" />
    <link rel="apple-touch-icon" sizes="180x180" href="/favicons/apple-touch-icon.webp" />
    <link rel="icon" type="image/webp" sizes="32x32" href="/favicons/favicon-32x32.webp" />
    <link rel="icon" type="image/webp" sizes="16x16" href="/favicons/favicon-16x16.webp" />
    <link rel="manifest" href="/favicons/site.webmanifest" crossorigin="use-credentials" />
    <link rel="mask-icon" href="/favicons/safari-pinned-tab.svg" color="#5bbad5" />
    <meta name="msapplication-TileColor" content="#da532c" />
    <meta name="theme-color" content="#ffffff" />
    <meta property="og:site_name" content="https://github.com/Seednode/trivia" />
    <meta property="og:title" content="Trivia v{{.Version}}" />
    <meta property="og:description" content="A very basic trivia webapp." />
    <meta property="og:url" content="https://github.com/Seednode/trivia" />
    <meta property="og:type" content="website" />
    <meta property="og:image" content="/favicons/apple-touch-icon.webp" />
  </head>
  <body>
    <p id="settings-link"><a href="/">Back to homepage</a></p>
	<div class="settings-container">
      <div class="settings-item">
	    <div class="settings-scrollable">
	      <h2>Categories</h2>
	      <ul>
{{.Categories}}
          </ul>
	      <input id="set-categories" type="submit"></input>
        </div>
	  </div>
	  <div class="settings-item">
	    <h2>Theme</h2>
	      <input type="radio" id="light-mode" name="theme" value="lightMode" checked />
		  <label for="light-mode">Light mode</label><br />
		  <input type="radio" id="dark-mode" name="theme" value="darkMode" />
		  <label for="dark-mode">Dark mode</label><br />
	  	<input id="set-theme" type="submit"></input>
	  </div>
	</div>
  </body>
</html>`
}

func serveSettingsPage(questions *Questions, tpl *template.Template, errorChannel chan<- error) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Header().Set("Content-Type", "text/html;charset=UTF-8")

		w.Header().Set("Content-Security-Policy", "default-src 'self';")

		securityHeaders(w)

		var toggles strings.Builder

		selected := getCategories(r, questions)

		for _, j := range questions.CategoryStrings() {
			if slices.Contains(selected, j) {
				toggles.WriteString(fmt.Sprintf("          <li><label><input type=\"checkbox\" name=\"%s\" checked>%s</label></li>\n", j, j))
			} else {
				toggles.WriteString(fmt.Sprintf("          <li><label><input type=\"checkbox\" name=\"%s\">%s</label></li>\n", j, j))
			}
		}

		categoryToggle := CategoryToggle{
			Version:    ReleaseVersion,
			Theme:      getTheme(r),
			Categories: template.HTML(toggles.String()),
		}

		err := tpl.Execute(w, categoryToggle)
		if err != nil {
			errorChannel <- err
		}
	}
}

func serveCategorySettings(questions *Questions, errorChannel chan<- error) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		startTime := time.Now()

		data, err := io.ReadAll(r.Body)
		if err != nil {
			errorChannel <- err

			return
		}

		var selected SelectedCategories
		err = json.Unmarshal(data, &selected)
		if err != nil {
			errorChannel <- err

			return
		}

		enabled := questions.CategoryStrings()

		c := []string{}

		for _, s := range selected.Categories {
			for _, e := range enabled {
				if s == e {
					c = append(c, s)
				}
			}
		}

		setCookie("enabledCategories", strings.Join(c, ","), w)

		if verbose {
			fmt.Printf("%s | %s => %s (Selected %d/%d categories)\n",
				startTime.Format(logDate),
				realIP(r),
				r.RequestURI,
				len(c),
				len(enabled))
		}
	}
}

func serveThemeSettings() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		startTime := time.Now()

		setCookie("colorTheme", p.ByName("theme"), w)

		if verbose {
			fmt.Printf("%s | %s => %s\n",
				startTime.Format(logDate),
				realIP(r),
				r.RequestURI)
		}
	}
}

func registerSettingsPage(mux *httprouter.Router, questions *Questions, errorChannel chan<- error) {
	template, err := template.New("settings").Parse(getSettingsTemplate())
	if err != nil {
		errorChannel <- err

		return
	}

	mux.GET("/settings", serveSettingsPage(questions, template, errorChannel))
	mux.POST("/settings/categories", serveCategorySettings(questions, errorChannel))
	mux.POST("/settings/theme/:theme", serveThemeSettings())
}
