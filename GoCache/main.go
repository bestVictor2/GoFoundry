package main

import (
	"GoCache/geecache"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func createGroup() *geecache.Group {
	return geecache.NewGroupWithOptions("scores", geecache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}), 2<<10, geecache.WithTTL(30*time.Second))
}

func startCacheServer(addr string, addrs []string, gee *geecache.Group) {
	peers := geecache.NewHTTPPool(addr)
	peers.Set(addrs...)
	gee.RegisterPeers(peers)
	log.Println("geecache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startAPIServer(apiAddr string, gee *geecache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			if key == "" {
				http.Error(w, "key is required", http.StatusBadRequest)
				return
			}
			view, err := gee.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			_, _ = w.Write(view.ByteSlice())
		}))

	http.Handle("/api/stats", http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(gee.Stats())
		}))

	http.Handle("/api/healthz", http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "ok",
				"group":  gee.Name(),
				"stats":  gee.Stats(),
			})
		}))

	http.Handle("/api/batch", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			keys := strings.Split(r.URL.Query().Get("keys"), ",")
			filtered := make([]string, 0, len(keys))
			for _, key := range keys {
				key = strings.TrimSpace(key)
				if key != "" {
					filtered = append(filtered, key)
				}
			}
			if len(filtered) == 0 {
				http.Error(w, "keys is required", http.StatusBadRequest)
				return
			}
			values, errs := gee.GetMany(filtered...)
			resp := make(map[string]interface{}, 2)
			data := make(map[string]string, len(values))
			for k, v := range values {
				data[k] = v.String()
			}
			resp["data"] = data
			if errs != nil {
				errmap := make(map[string]string, len(errs))
				for k, err := range errs {
					errmap[k] = err.Error()
				}
				resp["errors"] = errmap
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))

	http.Handle("/api/delete", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodDelete {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			key := r.URL.Query().Get("key")
			if key == "" {
				http.Error(w, "key is required", http.StatusBadRequest)
				return
			}
			gee.Remove(key)
			w.WriteHeader(http.StatusNoContent)
		}))

	log.Println("frontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "Geecache server port")
	flag.BoolVar(&api, "api", false, "Start an api server")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	addrs := make([]string, 0, len(addrMap))
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	gee := createGroup()
	if api {
		go startAPIServer(apiAddr, gee)
	}
	startCacheServer(addrMap[port], addrs, gee)
}
