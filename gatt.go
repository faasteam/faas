package faas

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func GattHandler(handler func(http.ResponseWriter, *http.Request, *Context), resDir string) http.HandlerFunc {
	var directoryTreeCache []byte
	var treeCacheErr error
	// 只加载并编码一次目录树
	absPath := filepath.Join(FAAS.WorkDir, resDir)
	tree, err := buildDirectoryTree(absPath)
	if err != nil {
		treeCacheErr = err
	} else {
		directoryTreeCache, treeCacheErr = json.Marshal(tree)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, _ := r.Context().Value(contextKey).(*Context)
		log.Printf("GattHandler relPath: %v,subPath: %v, r.URL.RawQuery: %v", c.RelPath, c.SubPath, r.URL.RawQuery)
		path := c.SubPath
		if path == "" {
			w.Header().Set("Content-Type", "application/json")
			if treeCacheErr != nil {
				http.Error(w, treeCacheErr.Error(), http.StatusInternalServerError)
				return
			}
			w.Write(directoryTreeCache)
		} else if f := r.URL.Query().Get("fn"); f != "" && strings.HasSuffix(path, ".att") {
			w.Header().Set("Content-Type", "application/json")
			c.Fn = filepath.Base(path) + "@" + f
			handler(w, r, c)
		} else {
			serveStatic(w, r, c, absPath, handler)
		}
	})
}

func StaticHandler(handler func(http.ResponseWriter, *http.Request, *Context), resDir string) http.HandlerFunc {
	absPath := filepath.Join(FAAS.WorkDir, resDir)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, _ := r.Context().Value(contextKey).(*Context)
		log.Printf("StaticHandler relPath: %v,subPath: %v", c.RelPath, c.SubPath)
		serveStatic(w, r, c, absPath, handler)
	})
}

// 构建目录树
func buildDirectoryTree(dir string) ([]map[string]any, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var tree []map[string]any
	for _, entry := range entries {
		item := make(map[string]any)
		item["name"] = entry.Name()
		if entry.IsDir() {
			item["type"] = "dir"
			content, err := buildDirectoryTree(filepath.Join(dir, entry.Name()))
			if err != nil {
				return nil, err
			}
			item["content"] = content
		} else {
			item["type"] = "file"
			if strings.HasSuffix(entry.Name(), ".att") {
				content, err := os.ReadFile(filepath.Join(dir, entry.Name()))
				if err != nil {
					return nil, err
				}
				var data any
				if err := json.Unmarshal(content, &data); err != nil {
					return nil, err
				}
				item["content"] = data
			}
		}
		tree = append(tree, item)
	}
	return tree, nil
}

func serveStatic(w http.ResponseWriter, r *http.Request, c *Context, rootPath string, handler404 func(http.ResponseWriter, *http.Request, *Context)) {
	filePath := rootPath + c.SubPath
	if strings.HasSuffix(filePath, "/") {
		filePath += "index.html"
	}
	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() {
		handler404(w, r, c)
	} else {
		if strings.HasSuffix(filePath, ".att") {
			w.Header().Set("Content-Type", "application/json")
		}
		http.ServeFile(w, r, filePath)
	}
}
