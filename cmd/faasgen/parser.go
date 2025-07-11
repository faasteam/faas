package main

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"strings"
)

type FuncletType int

const (
	HTTPFunclet FuncletType = iota
	TimingFunclet
)

type HTTPAnnotation struct {
	Entry    string // "api" or "local"
	Type     string // "path" or "prefix"
	Path     string
	ParamCnt int    // 2 or 3
	Method   string // Default to GET for now, can be extended
}

type TimingAnnotation struct {
	Type     string // "repeat", "everyday", "once"
	Interval string // e.g., "5s", "13h"
}

type Funclet struct {
	Name             string
	ImportPath       string
	Package          string
	HTTPAnnotation   *HTTPAnnotation
	TimingAnnotation *TimingAnnotation
}
type MatchAnnotation func(fn *ast.FuncDecl, text string) (*Funclet, error)

var (
	httpRegex   = regexp.MustCompile(`^//\s*@(onHandleFunclet|onMessageFunclet)\s+(\w+)\s*\(\s*(\w+)\s*,\s*([^)]*)\s*\)`)
	timingRegex = regexp.MustCompile(`^//\s*@onTimingFunclet\s+time\s*\(\s*(repeat|everyday|once)\s*(?:,\s*([^)]+)\s*)?\)`)
	matchSlice  = []MatchAnnotation{matchHTTPAnnotation, matchTimingAnnotation}
)

func parseFile(filePath, modulePath string) ([]*Funclet, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var funclets []*Funclet
	for _, decl := range node.Decls {
		if fn, isFn := decl.(*ast.FuncDecl); isFn {
			if fn.Doc != nil {
				for _, comment := range fn.Doc.List {
					text := strings.TrimSpace(comment.Text)
					for _, mf := range matchSlice {
						f, err := mf(fn, text)
						if err != nil {
							return nil, errors.New("path: " + filePath + ",func: " + fn.Name.Name + ",err:" + err.Error())
						}
						if f != nil {
							f.Name = fn.Name.Name
							dir := filepath.Dir(filePath)
							if dir != "" {
								f.ImportPath = filepath.Join(modulePath, dir)
							}
							funclets = append(funclets, f)
						}
					}

				}
			}
		}
	}
	return funclets, nil
}

func matchHTTPAnnotation(fn *ast.FuncDecl, text string) (*Funclet, error) {
	matches := httpRegex.FindStringSubmatch(text)
	if len(matches) != 5 {
		return nil, nil
	}
	cnt := len(fn.Type.Params.List)
	if matches[1] == "onMessageFunclet" && cnt != 1 {
		return nil, errors.New("onMessageFunclet param err, not eq 1")
	} else if matches[1] == "onHandleFunclet" && cnt != 2 && cnt != 3 {
		return nil, errors.New("onHandleFunclet param err, not eq 2 or 3")
	}

	path := matches[4]
	if path == "" || path == "*" {
		path = "/"
	} else {
		path = strings.TrimRight(path, "/")
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
	}
	httpAnnot := &HTTPAnnotation{
		Entry:    matches[2],
		Type:     matches[3],
		Path:     path,
		ParamCnt: cnt,
	}
	if httpAnnot.Entry != "msg" && httpAnnot.Type != "path" && httpAnnot.Type != "prefix" {
		return nil, errors.New("Error http type " + httpAnnot.Type)
	}
	return &Funclet{HTTPAnnotation: httpAnnot}, nil
}

func matchTimingAnnotation(fn *ast.FuncDecl, text string) (*Funclet, error) {
	matches := timingRegex.FindStringSubmatch(text)
	if len(matches) < 2 {
		return nil, nil
	}
	if len(fn.Type.Params.List) != 1 {
		return nil, errors.New("func param err")
	}
	annotType := matches[1]
	interval := ""
	if len(matches) == 3 {
		interval = matches[2]
	}
	timingAnnot := &TimingAnnotation{
		Type:     annotType,
		Interval: interval,
	}
	return &Funclet{TimingAnnotation: timingAnnot}, nil
}
