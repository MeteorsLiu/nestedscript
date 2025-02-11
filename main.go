package main

import (
	"context"
	"fmt"
	"maps"
	"os"
	"os/exec"
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
	client          = github.NewClient(nil).WithAuthToken(os.Getenv("GH_TOKEN"))
)

func handle(path string, sc *Config) {
	absPath, _ := filepath.Abs(path)

	fmt.Println(absPath)

	os.WriteFile("conanfile.txt", []byte(sc.conanFile()), 0755)
	cmd := exec.Command("conan", "install", ".", "--build=missing")
	cmd.Dir = absPath
	cmd.Run()

	os.Setenv("PKG_CONFIG_PATH", absPath)

	// ok, we can generate
	cmd = exec.Command("llcppcfg", sc.Package.Name)
	cmd.Dir = absPath
	cmd.Run()

	cmd = exec.Command("llcppg", "llcppg.cfg")
	cmd.Dir = absPath
	cmd.Run()

}

func main() {
	changes := os.Getenv("ALL_CHANGED_FILES")
	if changes == "" {
		panic("cannot find changes file!")
	}

	pathMap := map[string][]string{}

	identifyFileCnt := 0

	for _, abs := range strings.Fields(changes) {
		if strings.Contains(abs, LLGOModuleIdentifyFile) {
			identifyFileCnt++
			if identifyFileCnt > 1 {
				panic("only one module in single pr")
			}
		}
		baseDir := filepath.Dir(abs)
		file := filepath.Base(abs)

		// build a file path
		pathMap[baseDir] = append(pathMap[baseDir], file)
	}

	if identifyFileCnt == 0 {
		panic("no identify file!")
	}
	fmt.Println(pathMap)
	// even though there's only one pending converting module,
	// we still use loop for further uses
	for path := range maps.Keys(pathMap) {
		cfg := Read(filepath.Join(path, LLGOModuleIdentifyFile))

		handle(path, cfg)
	}

	fmt.Println(pathMap)
}
