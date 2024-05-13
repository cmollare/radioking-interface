package main

import (
	"fmt"
	"net/http"
	"testing"
)

func TestSet(t *testing.T) {
	c := NewCache()
	resp := &http.Response{}
	c.Set("key", resp)

	if len(c.store) != 1 {
		t.Errorf("Expected 1, got %d", len(c.store))
	}
}

func TestGet(t *testing.T) {
	c := NewCache()
	resp := &http.Response{}
	c.Set("key", resp)

	if _, err := c.Get("key"); err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
}

func TestGetNonCached(t *testing.T) {
	c := NewCache()
	resp := &http.Response{}
	c.Set("key", resp)

	if _, err := c.Get("key2"); err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestGetRequestBuildingFail(t *testing.T) {
	c := NewCache()
	c.store["key"] = &CacheEntry{
		Response: []byte("invalid"),
	}

	if _, err := c.Get("key"); err == nil {
		t.Error("Expected error, got nil")
	}
}

type mockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func TestDo(t *testing.T) {
	c := NewCache()
	client := &mockClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{}, nil
		},
	}
	h := &HTTPClientCache{
		Client: client,
		Cache:  c,
	}

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	h.Do(req)

	if len(c.store) != 1 {
		t.Errorf("Expected 1, got %d", len(c.store))
	}

	h.Do(req)
	if len(c.store) != 1 {
		t.Errorf("Expected 1, got %d", len(c.store))
	}
}

func TestDoError(t *testing.T) {
	c := NewCache()
	client := &mockClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return nil, fmt.Errorf("error")
		},
	}
	h := &HTTPClientCache{
		Client: client,
		Cache:  c,
	}

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	h.Do(req)

	if len(c.store) != 0 {
		t.Errorf("Expected 0, got %d", len(c.store))
	}
}
