// `url.go` 实现了一个简易的路由系统，可以将 URL 转发到不同的响应方法，
// 并且可以通过正则表达式匹配方法中的参数（如 `ID`）并传给响应方法。
package server

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const CtxParamKey = "ctx_params"

type Params struct {
	m map[string]string
	s []string
}

func (p *Params) Get(key string) string {
	if key == "" {
		return ""
	}
	return p.m[key]
}

func (p *Params) GetByIndex(i int) string {
	if i < 0 || i >= len(p.s) {
		return ""
	}
	return p.s[i]
}

// `parseReq` 匹配 `URL` 中的正则表达式，并将匹配结果根据组名存储在 `map[string]string` 中。
func parseReq(r *http.Request, pattern string) (*Params, bool) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		panic(err)
	}

	var url string
	if strings.HasPrefix(pattern, "^/") {
		url = r.URL.Path
	} else {
		url = r.URL.Host + r.URL.Path
	}

	if !re.MatchString(url) {
		return nil, false
	}

	subName := re.SubexpNames()
	subMatch := re.FindStringSubmatch(url)

	p := &Params{
		m: make(map[string]string),
		s: make([]string, len(subName)),
	}

	for i := 0; i < len(subName); i++ {
		p.m[subName[i]] = subMatch[i]
		p.s[i] = subMatch[i]
	}

	return p, true
}

// `ServeStatic` 方法传入一个目录，返回针对该目录的静态文件处理方法。
func ServeStatic(dir string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			ctx := r.Context()

			p := ctx.Value(CtxParamKey).(*Params)

			if len(p.s) <= 1 || strings.HasSuffix(r.URL.Path, "/") {
				defaultProxy.func404(w, r)
				return
			}

			path := filepath.Join(dir, filepath.Join(p.s[1:]...))
			if _, err := os.Stat(path); os.IsNotExist(err) {
				defaultProxy.func404(w, r)
				return
			}

			http.ServeFile(w, r, path)
		default:
			defaultProxy.func405(w, r)
		}
	})
}

func regex(pattern string) string {
	if pattern == "/" {
		return `^/$`
	}

	if !strings.HasPrefix(pattern, "^") {
		pattern = `^` + pattern
	}

	if !strings.HasSuffix(pattern, "$") {
		if strings.HasSuffix(pattern, "/") {
			pattern = pattern + `?$`
		} else {
			pattern = pattern + `/?$`
		}
	}

	return pattern
}
