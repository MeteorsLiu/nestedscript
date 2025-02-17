package main

import (
	"flag"
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

func generate(path string, sc *Config) {
	absPath, _ := filepath.Abs(path)

	fmt.Println(absPath)

	exec.Command("conan", "profile", "detect").Run()

	os.WriteFile(filepath.Join(absPath, "conanfile.txt"), []byte(sc.conanFile()), 0755)
	cmd := exec.Command("conan", "install", ".", "--build=missing")
	cmd.Dir = absPath
	output, err := cmd.CombinedOutput()
	fmt.Println(string(output), err)

	os.Setenv("PKG_CONFIG_PATH", absPath)

	// ok, we can generate
	cmd = exec.Command("llcppcfg", sc.Package.Name)
	cmd.Dir = absPath
	output, err = cmd.CombinedOutput()
	fmt.Println(string(output), err)

	cmd = exec.Command("llcppg", "llcppg.cfg")
	cmd.Dir = absPath
	output, err = cmd.CombinedOutput()
	fmt.Println(string(output), err)

	fmt.Println(exec.Command("ls", absPath).Output())
	localPath := filepath.Join(absPath, sc.Package.Name)
	dirs, err := os.ReadDir(localPath)
	must(err)

	for _, file := range dirs {
		fmt.Println(file.Name())
		os.Rename(filepath.Join(localPath, file.Name()), filepath.Join(absPath, file.Name()))
	}

	os.Remove(localPath)

	llpkgPath := filepath.Join(absPath, ".llpkg")
	os.Mkdir(llpkgPath, 0755)

	// be careful about the paths of llcppg config file here
	// llcppg.cfg/symb.json is in absPath, while pub is in
	os.Rename(
		filepath.Join(absPath, "llcppg.symb.json"),
		filepath.Join(llpkgPath, "llcppg.symb.json"))
	os.Rename(
		filepath.Join(absPath, "llcppg.pub"),
		filepath.Join(llpkgPath, "llcppg.pub"))

	env, err := os.OpenFile(os.Getenv("GITHUB_ENV"), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	must(err)

	// prevent duplicate name
	env.WriteString(fmt.Sprintf("ARTIFACT_NAME=%s%s_%s\n", sc.Package.Name, sc.Package.Version, currentSuffix))
	env.WriteString(fmt.Sprintf("LLCPPG_ABS_PATH=%s\n", absPath))
	env.WriteString(fmt.Sprintf("LLCPPG_PATH=%s\n", path))
	env.Close()
}

func main() {
	var genVersion bool
	flag.BoolVar(&genVersion, "version", false, "Gen version")
	flag.Parse()

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

		generate(path, cfg)
	}

	fmt.Println(pathMap)
}
