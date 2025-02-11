package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Package struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Upstream struct {
	Name   string          `json:"name"`
	Config json.RawMessage `json:"config"`
}

type Toolchain struct {
	Name    string          `json:"name"`
	Version string          `json:"version"`
	Config  json.RawMessage `json:"config"`
}

type Config struct {
	Package   `json:"package"`
	Upstream  `json:"upstream"`
	Toolchain `json:"toolchain"`
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func Read(file string) *Config {
	b, err := os.ReadFile(file)
	must(err)

	var c Config
	json.Unmarshal(b, &c)

	return &c
}

func (c *Config) conanFile() string {
	return fmt.Sprintf(`[requires]
%s/%s
	
[options]
*:shared=True

[generators]
PkgConfigDeps
`, c.Package.Name, c.Package.Version)
}
