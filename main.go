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

const LLGOModuleIdentifyFile = "llpkg.cfg"

var currentSuffix = runtime.GOOS + "_" + runtime.GOARCH

func copyFile(originalFile string) {
	fileName := strings.TrimSuffix(originalFile, filepath.Ext(originalFile))

	newFile, err := os.Create(fmt.Sprintf("%s_%s.go", fileName, currentSuffix))
	must(err)
	defer newFile.Close()

	newFile.Write([]byte(fmt.Sprintf(`//go:build %s && %s
`, runtime.GOOS, runtime.GOARCH)))

	current, err := os.Open(originalFile)
	must(err)
	// remove old file
	defer os.Remove(originalFile)

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
		copyFile(match)
	}

	llpkgPath := filepath.Join(localPath, ".llpkg")
	os.Mkdir(llpkgPath, 0755)

	// be careful about llcppg config file here
	os.Rename(
		filepath.Join(localPath, "llcppg.cfg"),
		filepath.Join(llpkgPath, fmt.Sprintf("llcppg_%s.cfg", currentSuffix)))
	os.Rename(
		filepath.Join(localPath, "llcppg.pub"),
		filepath.Join(llpkgPath, fmt.Sprintf("llcppg_%s.pub", currentSuffix)))
	os.Rename(
		filepath.Join(localPath, "llcppg.symb.json"),
		filepath.Join(llpkgPath, fmt.Sprintf("llcppg_%s.symb.json", currentSuffix)))

	env, err := os.OpenFile(os.Getenv("GITHUB_ENV"), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	must(err)

	env.WriteString(fmt.Sprintf("ARTIFACT_NAME=%s%s\n", sc.Package.Name, sc.Package.Version))
	env.WriteString(fmt.Sprintf("LLCPPG_ABS_PATH=%s\n", absPath))
	env.Close()
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
