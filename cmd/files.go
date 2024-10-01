/*
Copyright © 2024 Seednode <seednode@seedno.de>
*/

package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

func normalizePath(path string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	if path == "~" {
		path = homeDir
	} else if strings.HasPrefix(path, "~/") {
		path = filepath.Join(homeDir, path[2:])
	}

	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}

	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	return absolutePath, nil
}

func validatePaths(args []string) ([]string, error) {
	var paths []string

	for i := 0; i < len(args); i++ {
		path, err := normalizePath(args[i])
		if err != nil {
			return nil, err
		}

		paths = append(paths, path)
	}

	return paths, nil
}

func walkPath(path string, list map[string]Trivia, errorChannel chan<- error) []string {
	index := []string{}

	nodes, err := os.ReadDir(path)
	switch {
	case errors.Is(err, syscall.ENOTDIR):
		if filepath.Ext(path) == extension {
			index = append(index, loadFromFile(path, list, errorChannel)...)
		}
	case err != nil:
		errorChannel <- err
	default:
		for _, node := range nodes {
			fullPath := filepath.Join(path, node.Name())

			switch {
			case !node.IsDir() && filepath.Ext(node.Name()) == extension:
				index = append(index, loadFromFile(fullPath, list, errorChannel)...)
			case node.IsDir() && recursive:
				index = append(index, walkPath(fullPath, list, errorChannel)...)
			}
		}
	}

	return index
}
