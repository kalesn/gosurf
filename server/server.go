// `server.go` 为 web 程序入口。
package server

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"runtime/debug"
	"sync"
	"time"
)

func NewProxy(ctx context.Context, address string, C chan *Trace) *Proxy {
	p := &Proxy{
		rw:      new(sync.RWMutex),
		address: address,
		router:  make(map[string]http.Handler),
	}
	p.ctx, p.cancel = context.WithCancel(ctx)
	if C != nil {
		p.channel = C
		p.tracing = true
	}
	p.func404 = http.NotFound
	p.func405 = func(w http.ResponseWriter, r *http.Request) { http.Error(w, "405 method not allowed", 405) }
	p.func500 = func(w http.ResponseWriter, r *http.Request) { http.Error(w, "500 internal error", 500) }
	p.shutdown = func() {}
	return p
}

type Proxy struct {
	rw       *sync.RWMutex
	ctx      context.Context
	cancel   func()
	address  string
	router   map[string]http.Handler
	channel  chan *Trace
	tracing  bool
	shutdown func()

	// status method func
	func404, func405, func500 func(w http.ResponseWriter, r *http.Request)
}

func (p *Proxy) Run() {
	p.rw.Lock()
	defer p.rw.Unlock()
	// make server
	srv := &http.Server{
		Addr:         p.address,
		Handler:      p,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	go srv.ListenAndServe()
	p.shutdown = func() { srv.Shutdown(context.Background()) }
}

func (p *Proxy) Stop() {
	p.rw.Lock()
	defer p.rw.Unlock()
	p.cancel()
	if p.channel != nil {
		p.tracing = false
		close(p.channel)
	}
	p.shutdown()
}

func (p *Proxy) Handle(pattern string, handler http.Handler) {
	pattern = regex(pattern)

	p.rw.Lock()
	defer p.rw.Unlock()

	// set 405 method not allowed func for View type
	if view, ok := handler.(View); ok {
		view.func405 = p.func405
	}

	p.router[pattern] = handler
}

func (p *Proxy) send(t *Trace) {
	select {
	case <-p.ctx.Done():
	case p.channel <- t:
	}
}

func (p *Proxy) sendTrace(path string, err interface{}, withStack bool) {
	p.rw.RLock()
	send := p.tracing && p.channel != nil
	p.rw.RUnlock()
	if send {
		t := &Trace{Path: path, Err: err}
		if withStack {
			t.Stack = debug.Stack()
		}
		go p.send(t)
	}
}

func (p *Proxy) recoverHTTP(w http.ResponseWriter, r *http.Request) {
	if err := recover(); err != nil {
		// send trace message
		p.sendTrace(r.URL.Path, err, true)

		// call preset 500 function
		p.func500(w, r)
	}
}

// `ServeHTTP` 是开启服务器的入口方法，将自动匹配 URL 并执行相应的响应方法。
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer p.recoverHTTP(w, r)

	p.rw.RLock()
	defer p.rw.RUnlock()

	for pattern, handler := range p.router {
		if pm, ok := parseReq(r, pattern); ok {
			ctx := context.WithValue(r.Context(), CtxParamKey, pm)
			handler.ServeHTTP(w, r.WithContext(ctx))
			return
		}
	}

	// call preset 404 function
	p.func404(w, r)
}

func (p *Proxy) Set404(func404 func(w http.ResponseWriter, r *http.Request)) {
	if func404 == nil {
		panic("server: error nil 404 func")
	}
	p.rw.Lock()
	defer p.rw.Unlock()
	p.func404 = func404
}

func (p *Proxy) Set405(func405 func(w http.ResponseWriter, r *http.Request)) {
	if func405 == nil {
		panic("erver: error nil 405 func")
	}
	p.rw.Lock()
	defer p.rw.Unlock()
	p.func405 = func405
}

func (p *Proxy) Set500(func500 func(w http.ResponseWriter, r *http.Request)) {
	if func500 == nil {
		panic("server: error nil 500 func")
	}
	p.rw.Lock()
	defer p.rw.Unlock()
	p.func500 = func500
}

type Trace struct {
	Path  string
	Err   interface{}
	Stack []byte
}

func (t *Trace) PrintStack(w io.Writer) (n int64, err error) {
	if t.Stack == nil {
		return
	}
	var buf bytes.Buffer
	buf.WriteString("=== DEBUG - PRINT STACK - [" + time.Now().Format("2006-01-02 15:04:05") + "] ===\n")
	buf.Write(t.Stack)
	buf.WriteString("\n")
	return buf.WriteTo(w)
}

// new web handler
var defaultProxy = NewProxy(context.Background(), ":http", nil)

func Handle(pattern string, handler http.Handler) {
	defaultProxy.Handle(pattern, handler)
}

func SetAddress(addr string) {
	defaultProxy.rw.Lock()
	defer defaultProxy.rw.Unlock()
	defaultProxy.address = addr
}

func Watch(C chan *Trace) {
	if C == nil {
		panic("server: error nil channel")
	}
	defaultProxy.rw.Lock()
	defer defaultProxy.rw.Unlock()
	defaultProxy.channel = C
	defaultProxy.tracing = true
}

func Set404(func404 func(w http.ResponseWriter, r *http.Request)) {
	defaultProxy.Set404(func404)
}

func Set405(func405 func(w http.ResponseWriter, r *http.Request)) {
	defaultProxy.Set405(func405)
}

func Set500(func500 func(w http.ResponseWriter, r *http.Request)) {
	defaultProxy.Set500(func500)
}

func Run() {
	defaultProxy.Run()
}

func Stop() {
	defaultProxy.Stop()
}
