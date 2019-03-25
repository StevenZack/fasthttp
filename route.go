package fasthttp

import (
	"strings"

	"github.com/StevenZack/tools/strToolkit"
)

type Router struct {
	r, mr, pr map[string]func(*RequestCtx)
	pre       []func(*RequestCtx)
	server    *Server
}

func NewRouter() *Router {
	r := &Router{
		r:  make(map[string]func(*RequestCtx)),
		mr: make(map[string]func(*RequestCtx)),
	}
	r.server = &Server{
		Handler: r.handler,
	}
	r.server.ReduceMemoryUsage = true
	return r
}
func (r *Router) HandleFunc(s string, f func(*RequestCtx)) {
	if strings.Contains(s, "/:") {
		r.parsePathParams(s, f)
		return
	}
	r.r[s] = f
}
func (r *Router) HandleMultiReqs(s string, f func(*RequestCtx)) {
	r.mr[s] = f
}
func (r *Router) AddPreHandler(f func(*RequestCtx)) {
	r.pre = append(r.pre, f)
}
func (r *Router) ListenAndServe(addr string) error {

	return r.server.ListenAndServe(addr)
}
func (r *Router) handler(cx *RequestCtx) {
	for _, pre := range r.pre {
		pre(cx)
	}
	requestUri := string(cx.RequestURI())
	url := strings.Split(string(requestUri), "?")[0]
	if h, ok := r.r[url]; ok {
		h(cx)
	} else if k, ok := hasPreffixInMap(r.mr, requestUri); ok {
		r.mr[k](cx)
	} else {
		cx.NotFound()
	}
}
func (r *Router) GetServer() *Server {
	return r.server
}
func hasPreffixInMap(m map[string]func(*RequestCtx), p string) (string, bool) {
	for k, _ := range m {
		if len(p) >= len(k) && k == p[:len(k)] {
			return k, true
		}
	}
	return "", false
}
func (r *Router) parsePathParams(s string, f func(*RequestCtx)) {
	prefix := s[:strings.Index(s, "/:")+1]
	params := strings.Split(s, "/")
	indexes := []int{}
	keys := []string{}
	for index, param := range params {
		if strToolkit.StartsWith(param, ":") && len(param) > 1 {
			indexes = append(indexes, index)
			keys = append(keys, param[1:])
		}
	}
	r.mr[prefix] = func(cx *RequestCtx) {
		strs := strings.Split(cx.GetURI(), "/")
		if len(params) != len(strs) {
			cx.NotFound()
			return
		}
		if cx.pathParam == nil {
			cx.pathParam = make(map[string]string)
		}
		for keyIndex, index := range indexes {
			cx.pathParam[keys[keyIndex]] = strs[index]
		}
		f(cx)
	}
}
