package gee_rpc

import (
	"time"

	"gee-rpc/codec"
)

/*
协议的形式
| Option{MagicNumber: xxx, CodecType: xxx} | Header{ServiceMethod ...} | Body interface{} |
| <------      固定 JSON 编码      ------>  | <-------   编码方式由 CodeType 决定   ------->|

协议的通信结构
| ProtocolOption | Header1 | Body1 | Header2 | Body2 | ...
*/

// MagicNumber 协议的魔数 like cafe baby
const MagicNumber = 0x6c696b656c6f7665 // likelove

// ProtocolOption 当前rpc协议的一些固定配置
type ProtocolOption struct {
	MagicNumber       int           // 魔数 协议的标识
	CodecType         codec.Type    // 当前协议的序列化方式
	ConnectionTimeOut time.Duration // 连接超时时间
	HandlerTimeOut    time.Duration // 远程调用超时时间
}

// DefaultProtocolOption 默认协议的配置
var DefaultProtocolOption = &ProtocolOption{
	MagicNumber:       MagicNumber,      // 默认魔数
	CodecType:         codec.GobType,    // 默认使用 gob 解析器
	ConnectionTimeOut: time.Second * 10, // 默认十秒
}
