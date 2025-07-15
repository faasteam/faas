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
	FuncletType string // onHandleFunclet|onMessageFunclet|onAuthFunclet|onGattFunclet|onStaticFunclet
	Entry       string // "api" or "local"
	Type        string // "path" or "prefix"
	Path        string
	ResPath     string
	ParamCnt    int
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
	httpRegex   = regexp.MustCompile(`^//\s*@(onHandleFunclet|onMessageFunclet|onAuthFunclet|onGattFunclet|onStaticFunclet)\s+(\w+)\s*\((.*?)\)`)
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
							if dir != "" && dir != "." {
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

func parseParam(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	arr := strings.Split(s, ",")
	for i, v := range arr {
		arr[i] = strings.TrimSpace(v)
	}
	return arr
}
func matchHTTPAnnotation(fn *ast.FuncDecl, text string) (*Funclet, error) {
	matches := httpRegex.FindStringSubmatch(text)
	if len(matches) != 4 {
		return nil, nil
	}
	cnt := len(fn.Type.Params.List)
	param := parseParam(matches[3])
	httpAnnot := &HTTPAnnotation{
		FuncletType: matches[1],
		Entry:       matches[2],
		ParamCnt:    cnt,
	}
	if matches[1] == "onAuthFunclet" {
		if len(param) != 0 {
			return nil, errors.New("bad Annotation")
		}
		if cnt != 3 {
			return nil, errors.New("bad function param")
		}
	} else if matches[1] == "onMessageFunclet" {
		if len(param) != 2 {
			return nil, errors.New("bad Annotation")
		}
		if cnt != 1 {
			return nil, errors.New("bad function param")
		}
		httpAnnot.Type = param[0]
		httpAnnot.Path = param[1]
	} else if matches[1] == "onHandleFunclet" {
		if len(param) != 2 {
			return nil, errors.New("bad Annotation")
		}
		if cnt != 2 && cnt != 3 {
			return nil, errors.New("bad function param")
		}
		if param[0] != "path" && param[0] != "prefix" {
			return nil, errors.New("Error type " + httpAnnot.Type + ",only support path/prefix")
		}
		httpAnnot.Type = param[0]
		httpAnnot.Path = param[1]
	} else if matches[1] == "onGattFunclet" || matches[1] == "onStaticFunclet" {
		if len(param) != 3 {
			return nil, errors.New("bad Annotation")
		}
		if cnt != 3 {
			return nil, errors.New("bad function param")
		}
		if param[0] != "path" && param[0] != "prefix" {
			return nil, errors.New("Error type " + httpAnnot.Type + ",only support path/prefix")
		}
		httpAnnot.Type = param[0]
		httpAnnot.Path = param[1]
		httpAnnot.ResPath = param[2]
	}

	if httpAnnot.Path == "" || httpAnnot.Path == "*" {
		httpAnnot.Path = "/"
	} else {
		httpAnnot.Path = strings.TrimRight(httpAnnot.Path, "/")
		if !strings.HasPrefix(httpAnnot.Path, "/") {
			httpAnnot.Path = "/" + httpAnnot.Path
		}
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
