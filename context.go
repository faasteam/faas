package faas

import (
	"net/http"
	"net/url"
	"os"
)

type Context struct {
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
	/*******************以下内容全局变量FAAS上不存在**************************/
	w ResponseWriter
	r *http.Request
	//原始请求路径
	oriPath string
	//沙箱注解开始的路径
	RelPath string
	//前缀模式匹配下可用
	SubPath string
	//入口名称
	Entry string
	//gatt 函数名
	Fn string
	//auth 鉴权过后设置的内容
	Ctx any
}

func newContext(w http.ResponseWriter, r *http.Request) *Context {
	c := new(Context)
	if r != nil {
		c.w = NewResponse(w)
		c.r = r
		c.Entry = r.Header.Get("Faas-Gateway-Name")
		if c.Entry == "" {
			c.Entry = "api"
		}
		if suffix := r.Header.Get("Faas-Path-Suffix"); suffix != "" {
			c.RelPath, _ = url.QueryUnescape(suffix)
		} else {
			c.RelPath = r.URL.Path
		}
		r.Header.Set("Faas-Path-Suffix", c.RelPath)
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
