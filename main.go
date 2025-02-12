package main

import (
	"fmt"
	"io"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type workflow struct {
	Repo  string
	Owner string
	RunID int64
}

const LLGOModuleIdentifyFile = "llpkg.cfg"

func copyFile(originalFile string) {
	fileName := strings.TrimSuffix(originalFile, filepath.Ext(originalFile))

	newFile, err := os.Create(fmt.Sprintf("%s_%s_%s.go", fileName, runtime.GOOS, runtime.GOARCH))
	must(err)
	defer newFile.Close()

	newFile.Write([]byte(fmt.Sprintf(`//go:build %s && %s
`, runtime.GOOS, runtime.GOARCH)))

	current, err := os.Open(originalFile)
	must(err)

	defer current.Close()

	io.Copy(newFile, current)
}

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

	localPath := filepath.Join(absPath, sc.Package.Name)
	// done, rename all file, and upload to artifact
	matches, _ := filepath.Glob(filepath.Join(localPath, "*.go"))

	for _, match := range matches {
		fmt.Println(match)
		copyFile(match)
	}
	outputs, err := os.Open(os.Getenv("GITHUB_OUTPUT"))
	must(err)

	outputs.Write([]byte(fmt.Sprintf("llcppg_current_dir=%s%s\n", sc.Package.Name, sc.Package.Version)))
	outputs.Write([]byte(fmt.Sprintf("current_artifact_name=%s\n", absPath)))
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
