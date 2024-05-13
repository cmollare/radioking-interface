package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

/*
 * Améliorations:
 * - Ajouter des stratégies pour vider le cache (timeout, nombre d'entrées, etc.)
 * - Ne pas cacher les erreurs (404, 500, etc.)
 * - Ne cacher que les requêtes GET
 */

// CacheEntry is an entry in the cache
type CacheEntry struct {
	Response []byte
}

// Cache is a cache of HTTP responses
type Cache struct {
	mu    sync.RWMutex
	store map[string]*CacheEntry
}

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewCache creates a new cache
func NewCache() *Cache {
	return &Cache{
		store: make(map[string]*CacheEntry),
	}
}

// Get retrieves an entry from the cache
func (c *Cache) Get(key string) (*http.Response, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.store[key]
	if !ok {
		return nil, fmt.Errorf("cache miss")
	}

	buf := bytes.NewBuffer(entry.Response)
	reader := bufio.NewReader(buf)

	httpResp, err := http.ReadResponse(reader, nil)
	if err != nil {
		fmt.Println("Error reading response from cache:", err)
		return nil, err
	}

	return httpResp, nil
}

// Set an entry in the cache
func (c *Cache) Set(key string, resp *http.Response) {
	c.mu.Lock()
	defer c.mu.Unlock()
	response, _ := httputil.DumpResponse(resp, true)
	c.store[key] = &CacheEntry{
		Response: response,
	}
}

// HTTPClientCache is an HTTP client with a cache
type HTTPClientCache struct {
	Client HttpClient
	Cache  *Cache
}

// NewHTTPClientCache creates a new HTTP client with a cache
func NewHTTPClientCache() *HTTPClientCache {
	return &HTTPClientCache{
		Client: &http.Client{},
		Cache:  NewCache(),
	}
}

// Do executes an HTTP request implementing the interface
func (h *HTTPClientCache) Do(req *http.Request) (*http.Response, error) {
	cacheKey := req.URL.Path

	if resp, err := h.Cache.Get(cacheKey); err == nil {
		fmt.Printf("Cached request : %s\n", cacheKey)
		return resp, nil
	}
	fmt.Printf("Non-cached request : %s\n", cacheKey)

	resp, err := h.Client.Do(req)
	if err != nil {
		return nil, err
	}

	h.Cache.Set(cacheKey, resp)

	return resp, nil
}

func main() {
	// Use case
	client := NewHTTPClientCache()

	CallDo(client, "/api/v1/employee/")
	CallDo(client, "/api/v1/employee/")
	CallDo(client, "/api/v1/employee/1")
	CallDo(client, "/api/v1/employee/1")
	CallDo(client, "/api/v1/employee/2")
	CallDo(client, "/api/v1/employee/2")
}

type ResponseBody struct {
	Data string `json:"data"`
}

func CallDo(client *HTTPClientCache, path string) {
	resp, err := client.Do(&http.Request{
		Method: http.MethodGet,
		URL:    &url.URL{Scheme: "https", Host: "dummy.restapiexample.com", Path: path},
	})
	if err != nil {
		fmt.Println("Erreur lors de la requête :", err)
		return
	}

	fmt.Println("Code de statut de la réponse :", resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Erreur lors de la lecture du corps de la réponse :", err)
		return
	}

	/*var responseBody ResponseBody
	err = json.Unmarshal(body, &responseBody)
	if err != nil {
		fmt.Println("Erreur lors de la désérialisation du corps de la réponse :", err)
		return
	}*/
	fmt.Printf("Contenu de la réponse : %s\n", body)
	resp.Body.Close()
}
