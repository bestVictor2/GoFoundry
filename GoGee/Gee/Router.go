package Gee

import (
	"net/http"
	"sort"
	"strings"
)

type Router struct {
	roots    map[string]*node
	handlers map[string]HandlerFunc
	routes   []Route
}
type Route struct {
	Method  string `json:"method"`
	Pattern string `json:"pattern"`
}

func newRouter() *Router {
	return &Router{handlers: make(map[string]HandlerFunc), roots: make(map[string]*node), routes: make([]Route, 0)}
}
func parsePattern(pattern string) []string {
	vs := strings.Split(pattern, "/")
	parts := make([]string, 0)
	for _, part := range vs {
		if part != "" {
			parts = append(parts, part)
			if part[0] == '*' {
				break
			}
		}
	}
	return parts
}
func (router *Router) addRouter(method string, pattern string, handler HandlerFunc) {
	//fmt.Println(pattern)
	key := method + "-" + pattern
	if _, exists := router.handlers[key]; !exists {
		router.routes = append(router.routes, Route{Method: method, Pattern: pattern})
	}
	router.handlers[key] = handler
	_, ok := router.roots[method]
	if !ok {
		router.roots[method] = &node{}
	}
	parts := parsePattern(pattern)
	router.roots[method].Insert(pattern, parts, 0)
}
func (router *Router) listRoutes() []Route {
	routes := make([]Route, len(router.routes))
	copy(routes, router.routes)
	return routes
}
func (router *Router) handle(c *Context, noRoute HandlerFunc, noMethod HandlerFunc) {
	n, params := router.getRouter(c.Method, c.Path)
	if n != nil {
		c.Params = params
		key := c.Method + "-" + n.pattern
		if handler, ok := router.handlers[key]; ok {
			c.handles = append(c.handles, handler)
		}
	} else if router.pathExists(c.Path) {
		methods := router.allowedMethods(c.Path)
		c.handles = append(c.handles, func(c *Context) {
			if len(methods) > 0 {
				c.SetHeader("Allow", strings.Join(methods, ", "))
			}
			if noMethod != nil {
				noMethod(c)
				return
			}
			c.String(http.StatusMethodNotAllowed, "405 method not allowed")
		})
	} else {
		if noRoute != nil {
			c.handles = append(c.handles, noRoute)
		} else {
			c.handles = append(c.handles, func(c *Context) {
				c.String(http.StatusNotFound, "404 page not found")
			})
		}
	}
	c.Next()
}

func (router *Router) pathExists(path string) bool {
	for method := range router.roots {
		if n, _ := router.getRouter(method, path); n != nil {
			return true
		}
	}
	return false
}

func (router *Router) allowedMethods(path string) []string {
	methods := make([]string, 0)
	for method := range router.roots {
		if n, _ := router.getRouter(method, path); n != nil {
			methods = append(methods, method)
		}
	}
	sort.Strings(methods)
	return methods
}

func (router *Router) getRouter(method string, pattern string) (*node, map[string]string) {
	partsOpt := parsePattern(pattern)
	params := make(map[string]string)
	root, ok := router.roots[method]
	if !ok {
		return nil, nil
	}
	lastnode := root.Search(partsOpt, 0)
	if lastnode != nil {
		parts := parsePattern(lastnode.pattern)
		for index, part := range parts {
			if part[0] == '*' && len(part) > 1 {
				params[part[1:]] = strings.Join(partsOpt[index:], "/")
			}
			if part[0] == ':' {
				params[part[1:]] = partsOpt[index]
			}
		}
		return lastnode, params
	} else {
		return nil, nil
	}
}
