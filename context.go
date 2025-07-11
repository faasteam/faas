package faas

import (
	"net/http"
	"net/url"
	"os"
)

type Context struct {
	W http.ResponseWriter
	R *http.Request
	//沙箱注解开始的路径
	RelPath string
	//沙箱数据目录
	DataDir string
	//代码检出的工作目录
	WorkDir string
	//日志目录
	LogDir string
	//内部访问沙箱的路径前缀
	IscUrl string
	//内部访问gogs/gitea的路径前缀
	GitUrl string
	//入口名称
	Entry string
	//原始请求路径
	oriPath string
}

func newContext(w http.ResponseWriter, r *http.Request) *Context {
	c := new(Context)
	c.W = w
	c.R = r
	if r != nil {
		c.Entry = r.Header.Get("Faas-Gateway-Name")
		if c.Entry == "" {
			c.Entry = "api"
		}
		if suffix := r.Header.Get("Faas-Path-Suffix"); suffix != "" {
			c.RelPath, _ = url.QueryUnescape(suffix)
			r.Header.Set("Faas-Path-Suffix", c.RelPath)
		} else {
			c.RelPath = r.URL.Path
		}
		c.oriPath = r.URL.Path
	}
	c.DataDir = os.Getenv("DATA_PATH")
	c.WorkDir = os.Getenv("PROGRAM_PATH")
	c.LogDir = os.Getenv("LOG_PATH")
	c.IscUrl = os.Getenv("LOCAL_URL")
	c.GitUrl = os.Getenv("LOCAL_ROOT_URL")
	return c
}

var FAAS = newContext(nil, nil)
