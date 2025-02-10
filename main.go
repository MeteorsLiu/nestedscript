package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const LLGOModuleIdentifyFile = "llpkg.cfg"

func main() {
	changes := os.Getenv("ALL_CHANGED_FILES")
	if changes == "" {
		panic("cannot find changes file!")
	}

	pathMap := map[string][]string{}

	br := bufio.NewScanner(strings.NewReader(changes))

	identifyFileCnt := 0

	for br.Scan() {
		if strings.Contains(br.Text(), LLGOModuleIdentifyFile) {
			identifyFileCnt++
			if identifyFileCnt > 1 {
				panic("only one module in single pr")
			}
		}
		baseDir := filepath.Dir(br.Text())
		file := filepath.Base(br.Text())
		// build a file path
		pathMap[baseDir] = append(pathMap[baseDir], file)
	}

	if identifyFileCnt == 0 {
		panic("no identify file!")
	}

	fmt.Println(pathMap)
}
