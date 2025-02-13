package main

import (
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const LLGOModuleIdentifyFile = "llpkg.cfg"

var currentSuffix = runtime.GOOS + "_" + runtime.GOARCH

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

	llpkgPath := filepath.Join(localPath, ".llpkg")
	os.Mkdir(llpkgPath, 0755)

	// be careful about the paths of llcppg config file here
	// llcppg.cfg/symb.json is in absPath, while pub is in
	os.Rename(
		filepath.Join(absPath, "llcppg.cfg"),
		filepath.Join(localPath, "llcppg.cfg"),
	)
	os.Rename(
		filepath.Join(absPath, "llcppg.symb.json"),
		filepath.Join(llpkgPath, "llcppg.symb.json"))
	os.Rename(
		filepath.Join(localPath, "llcppg.pub"),
		filepath.Join(llpkgPath, "llcppg.pub"))

	env, err := os.OpenFile(os.Getenv("GITHUB_ENV"), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	must(err)

	// prevent duplicate name
	env.WriteString(fmt.Sprintf("ARTIFACT_NAME=%s%s_%s\n", sc.Package.Name, sc.Package.Version, currentSuffix))
	env.WriteString(fmt.Sprintf("LLCPPG_ABS_PATH=%s\n", localPath))
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
