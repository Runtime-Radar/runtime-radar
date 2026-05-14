package utils

import (
	"bufio"
	"os"
)

func ReadPasswordDictionary(path string) (map[string]struct{}, error) {
	passDictionary, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer passDictionary.Close()

	scanner := bufio.NewScanner(passDictionary)

	m := make(map[string]struct{})
	for scanner.Scan() {
		line := scanner.Text()
		m[line] = struct{}{}
	}

	return m, nil
}
