package codec

import "io"

// RequestHeader 请求头
type RequestHeader struct {
	ServiceMethod string // 请求的服务以及具体的方法 service.method
	Seq           uint64 // 请求的唯一标识
	Error         string
}

// Codec 编码器的顶级接口
type Codec interface {
	io.Closer
	ReadHeader(header *RequestHeader) error
	ReadBody(body interface{}) error
	Write(header *RequestHeader, body interface{}) error
}

// ============= codec 构造函数相关

// NewCodecFunc 抽象出Codec的构造函数
type NewCodecFunc func(rwc io.ReadWriteCloser) Codec

type Type string // codec的类型

// 预定义codec的类型
const (
	GobType  Type = "application/gob"
	JsonType Type = "application/json"
)

// NewCodecFuncMap 存放codec类型对应的构造函数
var NewCodecFuncMap map[Type]NewCodecFunc

func init() {
	NewCodecFuncMap = make(map[Type]NewCodecFunc)
	NewCodecFuncMap[GobType] = NewGobCodecFunc
}
