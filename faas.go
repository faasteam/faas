package faas

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type contextKeyType string

const contextKey contextKeyType = "contextkey"

var router = http.NewServeMux()

func makeContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nc := newContext(w, r)
		ctx := context.WithValue(r.Context(), contextKey, nc)
		r.URL.Path = "/" + nc.Entry + nc.RelPath
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func restorePathMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c, ok := r.Context().Value(contextKey).(*Context); ok {
			r.URL.Path = c.oriPath
		}
		next.ServeHTTP(w, r)
	})
}

func WithContextHandler(handler func(http.ResponseWriter, *http.Request, *Context)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, _ := r.Context().Value(contextKey).(*Context)
		handler(w, r, c)
	})
}

func MessageHandler(handler func(string)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r, err := io.ReadAll(r.Body); err == nil {
			handler(string(r))
		}
		w.Write([]byte("ok"))
	})
}

func Run() {
	addr := os.Getenv("SU_SERVER_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(addr, makeContextMiddleware(router)))
}

func HandleFunc(gateway, handlerType, path string, handler func(http.ResponseWriter, *http.Request)) {
	log.Printf("Registering router gateway:%s type:%s path:%s\n", gateway, handlerType, path)
	if gateway == "msg" {
		router.HandleFunc(filepath.Join("/", gateway, handlerType, path), handler)
	} else {
		if handlerType == "path" {
			router.Handle("/"+gateway+path, restorePathMiddleware(http.HandlerFunc(handler)))
		} else if handlerType == "prefix" {
			if !strings.HasSuffix(path, "/") {
				path = path + "/"
			}
			router.Handle("/"+gateway+path, restorePathMiddleware(http.HandlerFunc(handler)))
		}
	}
}

func TimingFunc(timingType, interval string, handler func(env map[string]any)) {
	log.Printf("Registering timing type:%s interval: %s\n", timingType, interval)
	env := make(map[string]any)
	env["trigertype"] = timingType
	env["interval"] = interval
	switch timingType {
	case "repeat":
		go func() {
			duration, err := time.ParseDuration(interval)
			if err != nil {
				log.Printf("Error parsing duration for repeat timing function: %v\n", err)
				return
			}
			timer := time.NewTicker(duration)
			handler(env)
			for range timer.C {
				handler(env)
			}
		}()
	case "everyday":
		go func() {
			duration, err := time.ParseDuration("11h53m3s")
			if err != nil {
				log.Printf("Error parsing duration for everyday timing function: %v\n", err)
				return
			}
			for {
				now := time.Now()
				sec := int(duration / time.Second)
				h := sec / 3600
				m := sec / 60 % 60
				s := sec % 60
				targetTime := time.Date(now.Year(), now.Month(), now.Day(), h, m, s, 0, now.Location())
				if now.After(targetTime) {
					targetTime = targetTime.Add(24 * time.Hour)
				}
				sleepDuration := targetTime.Sub(now)
				log.Printf("Next execution for everyday timing function at: %v (sleeping for %v)\n", targetTime, sleepDuration)
				time.Sleep(sleepDuration)
				handler(env)
			}
		}()
	case "once":
		go func() {
			handler(env)
		}()
	}
}
