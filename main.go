package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	changes := os.Getenv("ALL_CHANGED_FILES")
	if changes == "" {
		panic("cannot find changes file!")
	}
	br := bufio.NewScanner(strings.NewReader(changes))

	for br.Scan() {
		fmt.Println(br.Text())
	}
}
