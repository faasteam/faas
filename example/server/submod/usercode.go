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

// 注解gatt的路径和资源路径
// @onGattEntry  api(/gatt,res)
func GattEntry(w http.ResponseWriter, r *http.Request, c *faas.Context) {
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	fmt.Fprintf(w, "GattEntry not found for path: %s fn: %s", r.URL.Path, c.Fn)
}

// 匹配gatt的fn为api.att@get_list
// @onGattFunclet api(api.att@get_list)
func GattGetList(w http.ResponseWriter, r *http.Request, c *faas.Context) {
	fmt.Fprintf(w, "GattGetList Request received for path: %s fn: %s", r.URL.Path, c.Fn)
}

// 匹配gatt的fn为abc/api.att@get_user
// @onGattFunclet api(abc/api.att@get_user)
func GattGetUser(w http.ResponseWriter, r *http.Request, c *faas.Context) {
	fmt.Fprintf(w, "GattGetUser Request received for path: %s fn: %s", r.URL.Path, c.Fn)
}

// gatt默认的handler
// @onGattFunclet api(*)
func GattDefault(w http.ResponseWriter, r *http.Request, c *faas.Context) {
	fmt.Fprintf(w, "GattDefault Request received for path: %s fn: %s", r.URL.Path, c.Fn)
}

// 希望匹配到前缀/file"的路由关联到res目录作为一个文件服务，当找不到文件的时候进入此函数
// @onStaticFunclet api(prefix,/file,res)
func StaticHandler(w http.ResponseWriter, r *http.Request, c *faas.Context) {
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	fmt.Fprintf(w, "StaticHandler Request received for path: %s not found", r.URL.Path)
}

// 希望入口为api的都经过此函数做鉴权
// @onAuthFunclet api()
func AuthHandler(w http.ResponseWriter, r *http.Request, c *faas.Context) {
	log.Printf("AuthHandler Request received for path: %s", r.URL.Path)
}
