package fileio

import "os"
import "log"
import "strings"

func ReadFiles() []string {
	files := []string{}
	fileNames, err := os.ReadDir("./")

	if err != nil {
		log.Fatal(err)
	}

	for _, file := range fileNames {
		if strings.Contains(file.Name(), ".") {
			files = append(files, file.Name())
		}
	}

	return files
}
