package faas

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

// 定义目录树缓存和锁
var (
	directoryTreeCache []byte
	treeCacheOnce      sync.Once
	treeCacheErr       error
	localDir           = path.Join(FAAS.WorkDir, "res")
	gattHandlerMap     sync.Map
)

type GattHandlerFunc func(w http.ResponseWriter, r *http.Request, c any)

func InitGattService(handlerMap map[string]GattHandlerFunc) {
	for k, v := range handlerMap {
		gattHandlerMap.Store(k, v)
	}
}

func HandleGatt(w http.ResponseWriter, r *http.Request, cc any) {
	reqPath := path.Clean(r.Header.Get("Faas-Path-Suffix"))
	log.Printf("HandleGatt reqPath: %v, r.URL.RawQuery: %v", reqPath, r.URL.RawQuery)

	path := reqPath[len("/gatt"):]
	if path == "" {
		path = "/"
	}

	if path == "/" {
		w.Header().Set("Content-Type", "application/json")
		// 只加载并编码一次目录树
		treeCacheOnce.Do(func() {
			tree, err := buildDirectoryTree(localDir)
			if err != nil {
				treeCacheErr = err
				return
			}
			directoryTreeCache, treeCacheErr = json.Marshal(tree)
		})
		if treeCacheErr != nil {
			http.Error(w, treeCacheErr.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(directoryTreeCache)
	} else if strings.HasSuffix(path, ".att") {
		w.Header().Set("Content-Type", "application/json")
		// 处理 .att 文件
		f := r.URL.Query().Get("fn")
		if f != "" {
			f = filepath.Base(path) + "@" + f
			log.Printf("HandleGatt: f = %s", f)
			if handler, exists := gattHandlerMap.Load(f); exists {
				handler.(GattHandlerFunc)(w, r, cc)
				return
			}
			http.Error(w, "Handler method not found: "+f, http.StatusNotFound)
			return
		}
		// 读取 .att 文件内容
		filePath := filepath.Join(localDir, path)
		content, err := os.ReadFile(filePath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.Write(content)
	} else {
		// 读取普通文件
		filePath := filepath.Join(localDir, path)
		log.Println("gatt read file: ", filePath)
		content, err := os.ReadFile(filePath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		contentType := http.DetectContentType(content)
		w.Header().Set("Content-Type", contentType)
		w.Write(content)
	}
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
