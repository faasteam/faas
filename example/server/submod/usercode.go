package submod

import (
	"fmt"
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
