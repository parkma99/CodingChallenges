package main

import (
	"encoding/json"
	"log"
	"mime"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisServer struct {
	store      *redis.Client
	expiration time.Duration
	host       string
}

const DEFAULT_HOST = "localhost:8080"

func renderJSON(w http.ResponseWriter, v interface{}) {
	js, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (rs *redisServer) createShortURLHandler(w http.ResponseWriter, req *http.Request) {

	type RequestURL struct {
		Url string `json:"url"`
	}

	type Response struct {
		Key      string `json:"key"`
		LongURL  string `json:"long_url"`
		ShortURL string `json:"short_url"`
	}

	contentType := req.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if mediatype != "application/json" {
		http.Error(w, "expect application/json Content-Type", http.StatusUnsupportedMediaType)
		return
	}
	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()
	var r RequestURL
	if err := dec.Decode(&r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	hashed := hash(r.Url)
	for {
		url, err := rs.store.Get(req.Context(), hashed).Result()
		if err != nil {
			break
		}
		if url == r.Url {
			break
		}
		hashed = hash(r.Url + strconv.FormatInt(time.Now().Unix(), 10))
	}
	err = rs.store.Set(req.Context(), hashed, r.Url, rs.expiration).Err()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	renderJSON(w, Response{Key: hashed, LongURL: r.Url, ShortURL: rs.host + "/" + hashed})
}

func (rs *redisServer) getShortURLHandler(w http.ResponseWriter, req *http.Request) {

	key := req.PathValue("key")

	val, err := rs.store.Get(req.Context(), key).Result()
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	http.Redirect(w, req, val, http.StatusFound)
}

func (rs *redisServer) deleteShortURLHandler(w http.ResponseWriter, req *http.Request) {
	key := req.PathValue("key")
	err := rs.store.Del(req.Context(), key).Err()
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
}

func NewRedisServer(addr string, expiration time.Duration, host string) *redisServer {
	server := &redisServer{}
	server.store = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})
	server.expiration = expiration
	server.host = host
	return server
}

func main() {
	host := "http://" + DEFAULT_HOST
	mux := http.NewServeMux()
	server := NewRedisServer("localhost:6379", time.Hour, host)

	mux.HandleFunc("POST /", server.createShortURLHandler)
	mux.HandleFunc("GET /{key}", server.getShortURLHandler)
	mux.HandleFunc("DELETE /{key}", server.deleteShortURLHandler)

	log.Fatal(http.ListenAndServe(DEFAULT_HOST, mux))
}
