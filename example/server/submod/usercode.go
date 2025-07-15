package submod

import (
	"fmt"
	"log"
	"net/http"

	"github.com/faasteam/faas"
)

// 希望匹配到前缀/b"路由的函数进入此函数
// @onHandleFunclet api(path,/b)
func ContextPathHandler(w http.ResponseWriter, r *http.Request, c *faas.Context) {
	fmt.Fprintf(w, "ContextPathHandler Request received for path: %s", r.URL.Path)
}

// 希望匹配到前缀/b"路由的函数进入此函数
// @onHandleFunclet api(prefix,/b)
func ContextPrefixHandler(w http.ResponseWriter, r *http.Request, c *faas.Context) {
	fmt.Fprintf(w, "ContextPrefixHandler Request received for path: %s", r.URL.Path)
}

// 希望匹配到前缀/gatt"路由的函数进入gatt的流程res为资源目录
// @onGattFunclet api(prefix,/gatt,res)
func GattHandler(w http.ResponseWriter, r *http.Request, c *faas.Context) {
	fmt.Fprintf(w, "GattHandler Request received for path: %s fn: %s", r.URL.Path, c.Fn)
}

// 希望匹配到前缀/file"的路由关联到res目录作为一个文件服务，当找不到文件的时候进入此函数
// @onStaticFunclet api(prefix,/file,res)
func StaticHandler(w http.ResponseWriter, r *http.Request, c *faas.Context) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "StaticHandler Request received for path: %s not found", r.URL.Path)
}

// 希望入口为api的都经过此函数做鉴权
// @onAuthFunclet api()
func AuthHandler(w http.ResponseWriter, r *http.Request, c *faas.Context) {
	log.Printf("AuthHandler Request received for path: %s", r.URL.Path)
}
