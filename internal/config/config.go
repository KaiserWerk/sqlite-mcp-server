package config

import (
	"errors"
	"os"
)

type Config struct {
	DatabasePath string
	Debug        bool
}

func validateDatabasePath(dbPath string) error {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		dir := dbPath[:len(dbPath)-len(dbPath[findLastSlash(dbPath):])]
		if dir == "" {
			dir = "."
		}
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return errors.New("database directory does not exist")
		}
	}
	return nil
}

func findLastSlash(path string) int {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return i
		}
	}
	return -1
}
