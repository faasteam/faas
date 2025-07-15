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

type Entry struct {
	name   string
	auth   func(http.ResponseWriter, *http.Request, *Context)
	router http.ServeMux
}

var entryMap map[string]*Entry = make(map[string]*Entry)

func beforeDispatch() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := newContext(w, r)
		ctx := context.WithValue(r.Context(), contextKey, c)
		r = r.WithContext(ctx)
		entry, ok := entryMap[c.Entry]
		if !ok {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		if entry.auth != nil {
			entry.auth(c.w, r, c)
			if c.w.Status() != 0 {
				return
			}
		}
		entry.router.ServeHTTP(c.w, r)
		if c.w.Status() == 0 {
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	})
}

func beforeHandle(next http.Handler, path string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c, ok := r.Context().Value(contextKey).(*Context); ok {
			r.URL.Path = c.oriPath
			c.SubPath = c.RelPath[len(path):]
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
	log.Fatal(http.ListenAndServe(addr, beforeDispatch()))
}

func HandleAuth(entryName string, handler func(http.ResponseWriter, *http.Request, *Context)) {
	log.Printf("Registering auth entryName:%s\n", entryName)
	entry, ok := entryMap[entryName]
	if !ok {
		entry = &Entry{name: entryName}
		entryMap[entryName] = entry
	}
	entry.auth = handler
}

func HandleFunc(entryName, handlerType, path string, handler func(http.ResponseWriter, *http.Request)) {
	log.Printf("Registering router entryName:%s type:%s path:%s\n", entryName, handlerType, path)
	entry, ok := entryMap[entryName]
	if !ok {
		entry = &Entry{name: entryName}
		entryMap[entryName] = entry
	}
	if entryName == "msg" {
		entry.router.HandleFunc(filepath.Join(handlerType, path), handler)
	} else {
		if handlerType == "path" {
			entry.router.Handle(path, beforeHandle(http.HandlerFunc(handler), path))
		} else if handlerType == "prefix" {
			if !strings.HasSuffix(path, "/") {
				path = path + "/"
			}
			entry.router.Handle(path, beforeHandle(http.HandlerFunc(handler), path))
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
