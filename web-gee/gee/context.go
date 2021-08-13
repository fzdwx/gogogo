package gee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type H map[string]interface{}

type Context struct {
	Writer  http.ResponseWriter
	request *http.Request
	// request info
	Path   string
	Method string
	Params map[string]string
	// response info
	status int
	// middleware
	handlers []HandlerFunc
	index    int
	// engine pointer
	engine *Engine
}

func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer:  w,
		request: req,
		Path:    req.URL.Path,
		Method:  req.Method,
		index:   -1,
	}
}
func (c *Context) Next() {
	c.index++

	s := len(c.handlers)
	for ; c.index > s; c.index++ {
		c.handlers[c.index](c)
	}
}

// ================================ request start

func (c Context) Param(key string) string {
	return c.Params[key]
}

// PostFrom 从post请求从获取key对应的value
func (c Context) PostFrom(key string) string {
	return c.request.FormValue(key)
}

// Query 从url路径从获取key的value
func (c Context) Query(key string) string {
	return c.request.URL.Query().Get(key)
}

// Status 设置当前响应的状态码
func (c Context) Status(code int) {
	c.status = code
	c.Writer.WriteHeader(code)
}

// ================================ response start

// SetHeader 设置当前响应的响应头
func (c Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

// String 返回string类型的响应
func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

// JSON 返回Json类型的响应
func (c Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)

	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}
func (c *Context) Fail(code int, err string) {
	c.index = len(c.handlers)
	c.JSON(code, H{"message": err})
}

// Data 返回二进制类型的数据
func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

// HTML 返回HTML页面
func (c *Context) HTML(code int, name string, data interface{}) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	if err := c.engine.htmlTemplates.ExecuteTemplate(c.Writer, name, data); err != nil {
		c.Fail(500, err.Error())
	}
}
