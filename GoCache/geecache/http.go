package geecache

import (
	"GoCache/geecache/consistenthash"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const defaultBasePath = "/_geecache/"
const defaultReplicas = 50
const defaultHTTPTimeout = 2 * time.Second

type HTTPPoolOption func(*HTTPPool)

type HTTPPool struct {
	self        string
	basePath    string
	replicas    int
	mu          sync.RWMutex
	peers       *consistenthash.Map
	httpGetters map[string]PeerGetter
	client      *http.Client
}

func NewHTTPPool(self string) *HTTPPool {
	return NewHTTPPoolWithOptions(self)
}

func NewHTTPPoolWithOptions(self string, opts ...HTTPPoolOption) *HTTPPool {
	pool := &HTTPPool{
		self:        self,
		basePath:    defaultBasePath,
		replicas:    defaultReplicas,
		httpGetters: make(map[string]PeerGetter),
		client:      &http.Client{Timeout: defaultHTTPTimeout},
	}
	for _, opt := range opts {
		if opt != nil {
			opt(pool)
		}
	}
	pool.basePath = normalizeBasePath(pool.basePath)
	if pool.replicas <= 0 {
		pool.replicas = defaultReplicas
	}
	if pool.client == nil {
		pool.client = &http.Client{Timeout: defaultHTTPTimeout}
	}
	return pool
}

func WithHTTPPoolBasePath(basePath string) HTTPPoolOption {
	return func(pool *HTTPPool) {
		if basePath != "" {
			pool.basePath = basePath
		}
	}
}

func WithHTTPPoolReplicas(replicas int) HTTPPoolOption {
	return func(pool *HTTPPool) {
		if replicas > 0 {
			pool.replicas = replicas
		}
	}
}

func WithHTTPPoolClient(client *http.Client) HTTPPoolOption {
	return func(pool *HTTPPool) {
		if client != nil {
			pool.client = client
		}
	}
}

func normalizeBasePath(basePath string) string {
	if basePath == "" {
		return defaultBasePath
	}
	if !strings.HasPrefix(basePath, "/") {
		basePath = "/" + basePath
	}
	if !strings.HasSuffix(basePath, "/") {
		basePath += "/"
	}
	return basePath
}

func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		http.Error(w, "bad request path", http.StatusBadRequest)
		return
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	groupName := parts[0]
	key := parts[1]
	if groupName == "" || key == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "group not found", http.StatusNotFound)
		return
	}
	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	_, _ = w.Write(view.ByteSlice())
}

type httpGetter struct {
	baseURL string
	client  *http.Client
}

func (g *httpGetter) Get(group string, key string) ([]byte, error) {
	u := fmt.Sprintf("%v%v/%v", g.baseURL, url.QueryEscape(group), url.QueryEscape(key))
	resp, err := g.client.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

var _ PeerGetter = (*httpGetter)(nil)

func (p *HTTPPool) Set(keys ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(p.replicas, nil)
	p.peers.Add(keys...)
	p.httpGetters = make(map[string]PeerGetter, len(keys))
	for _, key := range keys {
		p.httpGetters[key] = &httpGetter{baseURL: key + p.basePath, client: p.client}
	}
}

func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.peers == nil {
		return nil, false
	}
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		getter, ok := p.httpGetters[peer]
		return getter, ok
	}
	return nil, false
}

var _ PeerPicker = (*HTTPPool)(nil)
