package main

import (
	"flag"

	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	var allFunclets []*Funclet
	var (
		src    = flag.String("src", "", "Source file or directory to scan for annotations.")
		output = flag.String("output", "main.go", "Output file name for generated faas code.")
	)

	flag.Parse()
	log.Println("start generate faas code . . .")
	goModPath := "go.mod"
	goModContent, err := os.ReadFile(goModPath)
	if err != nil {
		log.Fatalf("Error reading go.mod: %v", err)
	}
	modulePath := findModulePath(string(goModContent))
	if modulePath == "" {
		log.Fatalf("Could not find module path in go.mod")
	}

	if *src == "" {
		fi, err := os.Stat("faas.go")
		if err == nil && !fi.IsDir() {
			*src = "faas.go"
		} else {
			*src = "server"
		}
	}
	fileInfo, err := os.Stat(*src)
	if err != nil {
		log.Fatalf("Could not find file or directory: %s", *src)
	}
	if !fileInfo.IsDir() {
		funclets, err := parseFile(*src, modulePath)
		if err != nil {
			log.Fatalf("Error parsing file %s: %v", *src, err)
		}
		allFunclets = append(allFunclets, funclets...)
	} else {
		srcDir := *src
		err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(path, ".go") {
				if strings.HasSuffix(path, "_test.go") {
					return nil
				}
				funclets, err := parseFile(path, modulePath)
				if err != nil {
					log.Fatalf("Error parsing file %s: %v", path, err)
				}
				allFunclets = append(allFunclets, funclets...)
			}
			return nil
		})
		if err != nil {
			log.Fatalf("Error walking src directory: %v", err)
		}
	}

	if len(allFunclets) == 0 {
		log.Println("No funclets found. Exiting.")
		return
	}

	if err := generateCode(allFunclets, *output); err != nil {
		log.Fatalf("Error generating code: %v", err)
	}
	log.Println("Code generation complete.")
}

func findModulePath(goModContent string) string {
	lines := strings.Split(goModContent, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return ""
}
