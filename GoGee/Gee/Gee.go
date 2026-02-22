package Gee

import (
	"html/template"
	"net/http"
	"path"
	"strings"
)

type HandlerFunc func(*Context)
type RouterGroup struct {
	prefix      string
	middlewares []HandlerFunc
	parent      *RouterGroup
	engine      *Engine
}
type Engine struct {
	routerGroup   *RouterGroup
	router        *Router
	routerGroups  []*RouterGroup
	htmlTemplates *template.Template
	funcMap       template.FuncMap
	noRoute       HandlerFunc
	noMethod      HandlerFunc
}

func joinGroupPrefix(parentPrefix, childPrefix string) string {
	parentPrefix = strings.TrimSuffix(parentPrefix, "/")
	childPrefix = strings.TrimPrefix(childPrefix, "/")
	if childPrefix == "" {
		if parentPrefix == "" {
			return ""
		}
		return parentPrefix
	}
	if parentPrefix == "" {
		return "/" + childPrefix
	}
	return parentPrefix + "/" + childPrefix
}

func joinRoutePath(groupPrefix, comp string) string {
	prefix := strings.TrimSuffix(groupPrefix, "/")
	if comp == "" {
		comp = "/"
	}
	if !strings.HasPrefix(comp, "/") {
		comp = "/" + comp
	}
	if prefix == "" {
		return comp
	}
	if comp == "/" {
		return prefix
	}
	return prefix + comp
}

var allHTTPMethods = []string{
	http.MethodGet,
	http.MethodPost,
	http.MethodPut,
	http.MethodDelete,
	http.MethodPatch,
	http.MethodHead,
	http.MethodOptions,
}

func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.routerGroup = &RouterGroup{engine: engine}
	engine.routerGroups = []*RouterGroup{engine.routerGroup}
	return engine
}
func (engine *Engine) Group(prefix string) *RouterGroup {
	return engine.routerGroup.Group(prefix)
}
func (routerGroup *RouterGroup) Group(prefix string) *RouterGroup {
	engine := routerGroup.engine
	newGroup := &RouterGroup{prefix: joinGroupPrefix(routerGroup.prefix, prefix), parent: routerGroup, engine: engine}
	engine.routerGroups = append(engine.routerGroups, newGroup)
	return newGroup
}
func (routerGroup *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := joinRoutePath(routerGroup.prefix, comp)
	routerGroup.engine.router.addRouter(method, pattern, handler)
}
func (engine *Engine) addRouter(method string, pattern string, handler HandlerFunc) {
	engine.router.addRouter(method, joinRoutePath("", pattern), handler)
}
func (engine *Engine) Handle(method string, pattern string, handler HandlerFunc) {
	engine.addRouter(method, pattern, handler)
}
func (engine *Engine) GET(pattern string, handler HandlerFunc) {
	engine.addRouter("GET", pattern, handler)
}
func (routerGroup *RouterGroup) GET(pattern string, handler HandlerFunc) {
	routerGroup.addRoute("GET", pattern, handler)
}
func (routerGroup *RouterGroup) POST(pattern string, handler HandlerFunc) {
	routerGroup.addRoute("POST", pattern, handler)
}
func (routerGroup *RouterGroup) PUT(pattern string, handler HandlerFunc) {
	routerGroup.addRoute("PUT", pattern, handler)
}
func (routerGroup *RouterGroup) DELETE(pattern string, handler HandlerFunc) {
	routerGroup.addRoute("DELETE", pattern, handler)
}
func (routerGroup *RouterGroup) PATCH(pattern string, handler HandlerFunc) {
	routerGroup.addRoute("PATCH", pattern, handler)
}
func (routerGroup *RouterGroup) HEAD(pattern string, handler HandlerFunc) {
	routerGroup.addRoute("HEAD", pattern, handler)
}
func (routerGroup *RouterGroup) OPTIONS(pattern string, handler HandlerFunc) {
	routerGroup.addRoute("OPTIONS", pattern, handler)
}
func (routerGroup *RouterGroup) Any(pattern string, handler HandlerFunc) {
	for _, method := range allHTTPMethods {
		routerGroup.addRoute(method, pattern, handler)
	}
}
func (engine *Engine) POST(pattern string, handler HandlerFunc) {
	engine.addRouter("POST", pattern, handler)
}
func (engine *Engine) PUT(pattern string, handler HandlerFunc) {
	engine.addRouter("PUT", pattern, handler)
}
func (engine *Engine) DELETE(pattern string, handler HandlerFunc) {
	engine.addRouter("DELETE", pattern, handler)
}
func (engine *Engine) PATCH(pattern string, handler HandlerFunc) {
	engine.addRouter("PATCH", pattern, handler)
}
func (engine *Engine) HEAD(pattern string, handler HandlerFunc) {
	engine.addRouter("HEAD", pattern, handler)
}
func (engine *Engine) OPTIONS(pattern string, handler HandlerFunc) {
	engine.addRouter("OPTIONS", pattern, handler)
}
func (engine *Engine) Any(pattern string, handler HandlerFunc) {
	for _, method := range allHTTPMethods {
		engine.addRouter(method, pattern, handler)
	}
}
func (engine *Engine) Routes() []Route {
	return engine.router.listRoutes()
}
func (engine *Engine) Run(addr string) error {
	return http.ListenAndServe(addr, engine)
}
func (engine *Engine) Use(middleware ...HandlerFunc) {
	engine.routerGroup.middlewares = append(engine.routerGroup.middlewares, middleware...)
}
func (routerGroup *RouterGroup) Use(middleware ...HandlerFunc) {
	routerGroup.middlewares = append(routerGroup.middlewares, middleware...)
}
func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.funcMap = funcMap
}
func (engine *Engine) LoadHTMLGlob(pattern string) {
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}
func (engine *Engine) NoRoute(handler HandlerFunc) {
	engine.noRoute = handler
}
func (engine *Engine) NoMethod(handler HandlerFunc) {
	engine.noMethod = handler
}
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var middleWare []HandlerFunc
	for _, group := range engine.routerGroups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middleWare = append(middleWare, group.middlewares...)
		}
	}
	c := NewContext(w, req, engine)
	c.handles = middleWare
	engine.router.handle(c, engine.noRoute, engine.noMethod)
}
func (routerGroup *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(routerGroup.prefix, relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filepath")
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		fileServer.ServeHTTP(c.Writer, c.Rep)
	}

}
func (routerGroup *RouterGroup) Static(relativePath string, root string) {
	handler := routerGroup.createStaticHandler(relativePath, http.Dir(root))
	pattern := path.Join(relativePath, "/*filepath")
	routerGroup.GET(pattern, handler)
}

func Default() *Engine {
	engine := New()
	engine.Use(Logger(), Recovery())
	return engine
}
