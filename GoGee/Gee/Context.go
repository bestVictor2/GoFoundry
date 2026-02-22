package Gee

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type H map[string]interface{}

type Context struct {
	Writer     http.ResponseWriter
	Rep        *http.Request
	Method     string
	Path       string
	Params     map[string]string
	Keys       map[string]interface{}
	engine     *Engine
	handles    []HandlerFunc
	index      int
	StatusCode int
}

func (c *Context) Param(key string) string {
	return c.Params[key]
}
func NewContext(w http.ResponseWriter, req *http.Request, engine *Engine) *Context {
	return &Context{
		Writer:  w,
		Method:  req.Method,
		Path:    req.URL.Path,
		Rep:     req,
		engine:  engine,
		Keys:    make(map[string]interface{}),
		handles: make([]HandlerFunc, 0),
		index:   -1,
	}
}
func (c *Context) Set(key string, value interface{}) {
	c.Keys[key] = value
}
func (c *Context) Get(key string) (interface{}, bool) {
	value, ok := c.Keys[key]
	return value, ok
}
func (c *Context) GetString(key string) string {
	if value, ok := c.Get(key); ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}
func (c *Context) PostForm(key string) string {
	//fmt.Println(key)
	return c.Rep.FormValue(key)
}
func (c *Context) Query(key string) string {
	return c.Rep.URL.Query().Get(key)
}
func (c *Context) QueryDefault(key string, defaultValue string) string {
	if value := c.Query(key); value != "" {
		return value
	}
	return defaultValue
}
func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}
func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}
func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}
func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}
}
func (c *Context) BindJSON(obj interface{}) error {
	decoder := json.NewDecoder(c.Rep.Body)
	if err := decoder.Decode(obj); err != nil {
		return err
	}
	return nil
}
func (c *Context) Data(code int, data []byte) {
	c.SetHeader("Content-Type", "application/octet-stream")
	c.Status(code)
	c.Writer.Write(data)
}
func (c *Context) HTML(code int, html string) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	c.Writer.Write([]byte(html))
}
func (c *Context) HTMLTemplate(code int, name string, data interface{}) {
	if c.engine == nil || c.engine.htmlTemplates == nil {
		http.Error(c.Writer, "html templates not configured", http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	if err := c.engine.htmlTemplates.ExecuteTemplate(&buf, name, data); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	_, _ = c.Writer.Write(buf.Bytes())
}
func (c *Context) Abort() {
	c.index = len(c.handles)
}
func (c *Context) Fail(code int, message string) {
	c.Abort()
	c.JSON(code, H{"message": message})
}
func (c *Context) Next() {
	c.index++
	handlersLen := len(c.handles)
	for ; c.index < handlersLen; c.index++ {
		c.handles[c.index](c)
	}
}
