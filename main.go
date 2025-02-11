package main

import (
	"bufio"
	"context"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/go-github/v69/github"
)

type workflow struct {
	Repo  string
	Owner string
	RunID int64
}

const LLGOModuleIdentifyFile = "llpkg.cfg"

func getCurrentWorkflow() *workflow {
	repo := os.Getenv("GITHUB_REPOSITORY")
	owner := os.Getenv("GITHUB_REPOSITORY_OWNER")
	runid, _ := strconv.ParseInt(os.Getenv("GITHUB_RUN_ID"), 10, 64)

	return &workflow{Repo: repo, Owner: owner, RunID: runid}
}

var (
	cb              = context.TODO()
	currentWorkflow = getCurrentWorkflow()
	client          = github.NewClient(nil).WithAuthToken(os.Getenv("github-token"))
)

func handle(path string, sc *Config) {
	fmt.Println(currentWorkflow.Owner, currentWorkflow.Repo)
	fmt.Println(client.Actions.ListWorkflowRunArtifacts(cb, currentWorkflow.Owner, currentWorkflow.Repo, currentWorkflow.RunID, &github.ListOptions{}))
}

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

	// even though there's only one pending converting module,
	// we still use loop for further uses
	for path := range maps.Keys(pathMap) {
		cfg := Read(filepath.Join(path, LLGOModuleIdentifyFile))

		handle(path, cfg)
	}

	fmt.Println(pathMap)
}
