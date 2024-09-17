/*
Copyright © 2024 Seednode <seednode@seedno.de>
*/

package main

import (
	"log"

	"seedno.de/seednode/trivia/cmd"
)

func main() {
	cmd := cmd.NewRootCommand()

	err := cmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
