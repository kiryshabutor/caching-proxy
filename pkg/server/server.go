package server

import (
	"CachingProxy/pkg/cache"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
)

type key int

const cacheKey key = 0

type Server struct {
	server *http.Server
	proxy  *httputil.ReverseProxy
	port   int
	origin string
	cache  *cache.Cache
}

func NewServer(port int, origin string, c *cache.Cache) *Server {
	s := &Server{
		port:   port,
		origin: origin,
		cache:  c,
	}

	target, _ := url.Parse(origin)
	rp := httputil.NewSingleHostReverseProxy(target)

	rp.ModifyResponse = func(resp *http.Response) error {
		resp.Header.Set("X-Cache", "MISS")
		log.Println("Cache miss")
		body, _ := io.ReadAll(resp.Body)

		var keyVal string
		if v := resp.Request.Context().Value(cacheKey); v != nil {
			keyVal = v.(string)
		} else {
			keyVal = generateCacheKey(resp.Request, body)
		}

		s.cache.Set(keyVal, cache.CacheResponse{
			Body:   body,
			Status: resp.StatusCode,
			Header: resp.Header,
		})
		resp.Body = io.NopCloser(bytes.NewReader(body))
		return nil
	}

	s.proxy = rp
	s.server = &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: s,
	}
	return s
}

func (s *Server) Run() error {
	return s.server.ListenAndServe()
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	bodyBytes, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	keyStr := generateCacheKey(r, bodyBytes)

	if resp, ok := s.cache.Get(keyStr); ok {
		w.WriteHeader(resp.Status)
		for k, v := range resp.Header {
			for _, val := range v {
				w.Header().Add(k, val)
			}
		}
		w.Header().Set("X-Cache", "HIT")
		log.Println("Cache hit")
		w.Write(resp.Body)
		return
	}

	ctx := context.WithValue(r.Context(), cacheKey, keyStr)

	s.proxy.ServeHTTP(w, r.WithContext(ctx))
}

func generateCacheKey(r *http.Request, body []byte) string {
	keyStr := r.Method + ":" + r.URL.Path
	if r.URL.RawQuery != "" {
		keyStr += "?" + r.URL.RawQuery
	}
	if len(body) > 0 {
		hash := sha256.Sum256(body)
		keyStr += "|" + hex.EncodeToString(hash[:])
	}
	return keyStr
}
