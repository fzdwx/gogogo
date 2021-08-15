package gee_rpc

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"sync"

	"gee-rpc/codec"
)

// Server rpc服务提供者
type Server struct {
}

// NewServer 构造函数
func NewServer() *Server {
	return &Server{}
}

// DefaultServer 默认server
var DefaultServer = NewServer()

// Accept 接收连接 会一直监听，然后启动一个协程，进行处理
func (s *Server) Accept(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("rpc server: accept error:", err)
			return
		}
		go s.ServeConn(conn)
	}
}

// ServeConn 处理conn的请求
func (s *Server) ServeConn(conn io.ReadWriteCloser) {
	defer func() { _ = conn.Close() }()

	var opt ProtocolOption
	// 对conn中的数据进行解码，然后封装成ProtocolOption
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Println("rpc server: options error: ", err)
		return
	}
	// 匹配魔数
	if opt.MagicNumber != MagicNumber {
		log.Printf("rpc server: invalid magic number %x", opt.MagicNumber)
		return
	}
	// 获取对应的解码器
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		log.Printf("rpc server: invalid codec type %s", opt.CodecType)
		return
	}
	s.serveCodec(f(conn))
}

func (s *Server) serveCodec(c codec.Codec) {
	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup)

	for {
		req, err := s.readRequest(c)
		// header 解析失败
		if err != nil {
			if req == nil {
				break
			}
			req.h.Error = err.Error()
			s.sendResponse(c, req.h, invalidRequest, sending)
			continue
		}
		wg.Add(1)
		go s.handleRequest(c, req, sending, wg)
	}
	// 等待处理完成，关闭连接
	wg.Wait()
	_ = c.Close()
}

// request stores all information of a call
type request struct {
	h                     *codec.RequestHeader // header of request
	argValues, replyValue reflect.Value        // argv and replyv of request
}

// readRequest 解析请求
func (s *Server) readRequest(c codec.Codec) (*request, error) {
	header, err := s.readRequestHeader(c)
	if err != nil {
		return nil, err
	}
	req := &request{
		h: header,
	}
	// todo unKnow type 读取请求中的请求参数以及返回值类型
	req.argValues = reflect.New(reflect.TypeOf(""))
	err = c.ReadBody(req.argValues.Interface())
	if err != nil {
		log.Println("rpc server: read argv err:", err)
	}

	return req, nil

}

// readRequestHeader 解析请求头
func (s *Server) readRequestHeader(c codec.Codec) (*codec.RequestHeader, error) {
	var h codec.RequestHeader
	// 对conn进行解码，封装成request header
	if err := c.ReadHeader(&h); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println("rpc server: read header error:", err)
		}
		return nil, err
	}
	return &h, nil
}

// sendResponse 返回响应
func (s *Server) sendResponse(c codec.Codec, h *codec.RequestHeader, body interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()

	err := c.Write(h, body)
	if err != nil {
		log.Println("rpc server: write response error:", err)
	}
}

// handleRequest 处理请求
func (s *Server) handleRequest(c codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println(req.h, req.argValues.Elem())
	req.replyValue = reflect.ValueOf(fmt.Sprintf("rpc reponse %d", req.h.Seq))
	s.sendResponse(c, req.h, req.replyValue.Interface(), sending)
}

// Accept 接收连接
func Accept(lis net.Listener) {
	DefaultServer.Accept(lis)
}

// invalidRequest is a placeholder for response argv when error occurs
var invalidRequest = struct{}{}
