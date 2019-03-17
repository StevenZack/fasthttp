package fasthttp

import "strings"

type Router struct {
	r, mr  map[string]func(*RequestCtx)
	pre    []func(*RequestCtx)
	server *Server
}

func NewRouter() *Router {
	r := &Router{
		r:  make(map[string]func(*RequestCtx)),
		mr: make(map[string]func(*RequestCtx)),
	}
	r.server = &Server{
		Handler: r.handler,
	}
	r.server.MaxRequestBodySize = 20 * 1024 * 1024 * 1024
	r.server.ReduceMemoryUsage=true
	return r
}
func (r *Router) HandleFunc(s string, f func(*RequestCtx)) {
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
	url := strings.Split(cx.GetURI(), "?")[0]
	if h, ok := r.r[url]; ok {
		h(cx)
	} else if k, ok := hasPreffixInMap(r.mr, cx.GetURI()); ok {
		r.mr[k](cx)
	} else {
		cx.Response.SetStatusCode(404)
		cx.WriteString(`<!DOCTYPE html><html><head><title>404</title><meta charset="utf-8"><meta name="viewpos" content="width=device-width"></head><body>404 not found</body></html>`)
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
