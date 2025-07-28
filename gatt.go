package faas

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var gattHandlerMap map[string]func(http.ResponseWriter, *http.Request, *Context)

func GattEntry(entryName, gattPath string, handler func(http.ResponseWriter, *http.Request, *Context), resDir string) {
	var directoryTreeCache []byte
	var treeCacheErr error
	log.Printf("Registering GattEntry path: %s\n", gattPath)

	if gattHandlerMap != nil {
		panic("GattEntry can only be called once")
	}
	gattHandlerMap = make(map[string]func(http.ResponseWriter, *http.Request, *Context))

	absPath := filepath.Join(FAAS.WorkDir, resDir)
	tree, err := buildDirectoryTree(absPath)
	if err != nil {
		treeCacheErr = err
	} else {
		directoryTreeCache, treeCacheErr = json.Marshal(tree)
	}

	wrapHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			c.Fn = strings.TrimPrefix(path, "/") + "@" + f
			if h, ok := gattHandlerMap[c.Fn]; ok {
				h(w, r, c)
			} else if h, ok := gattHandlerMap["*"]; ok {
				h(w, r, c)
			} else {
				http.Error(w, "Not Found", http.StatusNotFound)
			}
		} else {
			serveStatic(w, r, c, absPath, handler)
		}
	})
	HandleFunc(entryName, "path", gattPath, wrapHandler)
	HandleFunc(entryName, "prefix", gattPath, wrapHandler)
}

func RegisterGattFnHandler(fn string, handler func(http.ResponseWriter, *http.Request, *Context)) {
	log.Printf("Registering gatt fn: %s\n", fn)
	if gattHandlerMap == nil {
		panic("Before registering an GattFnHandler, you must first call GattEntry")
	}
	if _, ok := gattHandlerMap[fn]; ok {
		panic(fn + " has already been registered.")
	}
	gattHandlerMap[fn] = handler
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
				if attmap, ok := data.(map[string]any); ok {
					if v, ok := attmap["@refer"].(string); ok && v != "" {
						item["url"] = v
					}
				}
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
	if c.w.Status() == 0 {
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}
