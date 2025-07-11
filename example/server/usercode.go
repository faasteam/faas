package server

import (
	"fmt"
	"log"
	"net/http"
)

// 希望完整匹配到/a路由的函数进入此函数
// @onHandleFunclet api(path,/a)
func PathHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "PathHandler Request received for path: %s", r.URL.Path)
}

// 希望匹配到前缀/a"路由的函数进入此函数
// @onHandleFunclet api(prefix,/a)
func PrefixHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "PrefixHandler Request received for path: %s", r.URL.Path)
}

// 5s重复执行的注解
// @onTimingFunclet time(repeat,5s)
func RepeatFunclet(env map[string]any) {
	log.Println("RepeatFunclet in 5s......")
}

// 每天中午1点重复执行的注解
// @onTimingFunclet time(everyday,13h)
func EveryDayFunclet(env map[string]any) {
	log.Println("EveryDayFunclet in 13h......")
}

// 启动的时候执行一次的注解
// @onTimingFunclet time(once)
func OnceFunclet(env map[string]any) {
	log.Println("OnceFunclet in......")
}

// @onMessageFunclet msg(potter,exit)
func RecvExitMsg(msgstr string) {
	log.Println("recv:", msgstr)
}
