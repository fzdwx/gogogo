package gee_rpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"strings"
	"sync"
	"time"

	"gee-rpc/codec"
)

// Server rpc服务提供者
type Server struct {
	serviceMap sync.Map
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

// Accept 接收连接
func Accept(lis net.Listener) {
	DefaultServer.Accept(lis)
}

// ServeConn 处理conn的请求 主要是接收客户端的一次发包，解析魔数是否匹配，以及使用的编解协议
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

	// 魔数匹配，编码协议解析正确
	s.serveCodec(f(conn))
}

// Register publishes the receiver's methods in the DefaultServer.
func Register(rcvr interface{}) error { return DefaultServer.Register(rcvr) }

// Register publishes in the server the set of methods of the
func (s *Server) Register(rcvr interface{}) error {
	service := newService(rcvr)
	_, dup := s.serviceMap.LoadOrStore(service.name, service)
	if dup {
		return errors.New("rpc: service already defined: " + service.name)
	}
	return nil
}

func (s *Server) findService(serviceMethod string) (svc *service, mtype *methodType, err error) {
	dot := strings.LastIndex(serviceMethod, ".")
	if dot < 0 {
		err = errors.New("rpc server: service/method request ill-formed: " + serviceMethod)
		return
	}
	serviceName, methodName := serviceMethod[:dot], serviceMethod[dot+1:]

	svci, ok := s.serviceMap.Load(serviceName)
	if !ok {
		err = errors.New("rpc server: can't find service " + serviceName)
		return
	}
	svc = svci.(*service)
	mtype = svc.method[methodName]
	if mtype == nil {
		err = errors.New("rpc server: can't find method " + methodName)
	}
	return
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
		go s.handleRequest(c, req, sending, wg, time.Second*3)
	}
	// 等待处理完成，关闭连接
	wg.Wait()
	_ = c.Close()
}

// request stores all information of a call
type request struct {
	h                     *codec.RequestHeader // header of request
	argValues, replyValue reflect.Value        // argv and replyv of request
	mtType                *methodType
	service               *service
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
	req.service, req.mtType, err = s.findService(header.ServiceMethod)
	if err != nil {
		return req, err
	}
	req.argValues = req.mtType.newArgv()
	req.replyValue = req.mtType.newReplayValue()

	// make sure that argvi is a pointer, ReadBody need a pointer as parameter
	argvi := req.argValues.Interface()
	if req.argValues.Type().Kind() != reflect.Ptr {
		argvi = req.argValues.Addr().Interface()
	}
	if err = c.ReadBody(argvi); err != nil {
		log.Println("rpc server: read body err:", err)
		return req, err
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
func (s *Server) handleRequest(c codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup, timeout time.Duration) {
	defer wg.Done()
	called := make(chan struct{})
	sent := make(chan struct{})

	go func() {
		err := req.service.call(req.mtType, req.argValues, req.replyValue)
		called <- struct{}{}
		if err != nil {
			req.h.Error = err.Error()
			s.sendResponse(c, req.h, invalidRequest, sending)
			sent <- struct{}{}
			return
		}
		s.sendResponse(c, req.h, req.replyValue.Interface(), sending)
		sent <- struct{}{}
	}()

	if timeout == 0 {
		<-called
		<-sent
		return
	}

	select {
	case <-time.After(timeout):
		req.h.Error = fmt.Sprintf("rpc server: request handle timeout: expect within %s", timeout)
		s.sendResponse(c, req.h, invalidRequest, sending)
	case <-called:
		<-sent
	}
}

// invalidRequest is a placeholder for response argv when error occurs
var invalidRequest = struct{}{}
