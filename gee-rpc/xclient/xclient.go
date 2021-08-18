package xclient

import (
	"context"
	"io"
	"reflect"
	"sync"

	geeRpc "gee-rpc"
	"gee-rpc/xclient/discovery"
)

// XClient 支持负载均衡的客户端
type XClient struct {
	discovery discovery.Discovery       // discovery 服务发现的实现
	mode      discovery.SelectMode      // mode 负载均衡策略
	opt       *geeRpc.ProtocolOption    // opt 协议选项
	clients   map[string]*geeRpc.Client // clients 保存客户端
	mu        sync.Mutex
}

func NewXClient(d discovery.Discovery, mode discovery.SelectMode, opt *geeRpc.ProtocolOption) *XClient {
	return &XClient{discovery: d, mode: mode, opt: opt, clients: make(map[string]*geeRpc.Client)}
}

var _ io.Closer = (*XClient)(nil)

// Close 关闭已经所有建立的连接
func (xc *XClient) Close() error {
	xc.mu.Lock()
	defer xc.mu.Unlock()

	for rpcAddr, client := range xc.clients {
		_ = client.Close()
		delete(xc.clients, rpcAddr)
	}

	return nil
}

// Call 调用目标方法，等待完成，最后返回结果或错误
// 会根据定义的模式，选择服务器
func (xc *XClient) Call(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	rpcAddr, err := xc.discovery.Get(xc.mode)
	if err != nil {
		return err
	}
	return xc.call(rpcAddr, ctx, serviceMethod, args, reply)
}

// Broadcast 将请求广播到所有的服务实例，如果任意一个实例发生错误，则返回其中一个错误；如果调用成功，则返回其中一个的结果。
// 有以下几点需要注意：
//
// 为了提升性能，请求是并发的。
//
// 并发情况下需要使用互斥锁保证 error 和 reply 能被正确赋值。
//
// 借助 context.WithCancel 确保有错误发生时，快速失败。
func (xc *XClient) Broadcast(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	servers, err := xc.discovery.GetAll()
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	var mu sync.Mutex // protect e and replyDone
	var e error
	replyDone := reply == nil // if reply is nil, don't need to set value
	ctx, cancel := context.WithCancel(ctx)
	for _, rpcAddr := range servers {
		wg.Add(1)
		go func(rpcAddr string) {
			defer wg.Done()
			var cloneReply interface{}
			if reply != nil {
				cloneReply = reflect.New(reflect.ValueOf(reply).Elem().Type()).Interface()
			}
			err := xc.call(rpcAddr, ctx, serviceMethod, args, cloneReply)
			mu.Lock()
			if err != nil && e == nil {
				e = err
				cancel() // if any call failed, cancel unfinished calls
			}
			if err == nil && !replyDone {
				reflect.ValueOf(reply).Elem().Set(reflect.ValueOf(cloneReply).Elem())
				replyDone = true
			}
			mu.Unlock()
		}(rpcAddr)
	}
	wg.Wait()
	return e
}

func (xc *XClient) dial(rpcAddr string) (*geeRpc.Client, error) {
	xc.mu.Lock()
	defer xc.mu.Unlock()
	client, ok := xc.clients[rpcAddr]
	if ok && !client.IsAvailable() {
		_ = client.Close()
		delete(xc.clients, rpcAddr)
		client = nil
	}
	if client == nil {
		var err error
		client, err = geeRpc.XDial(rpcAddr, xc.opt)
		if err != nil {
			return nil, err
		}
		xc.clients[rpcAddr] = client
	}
	return client, nil
}

func (xc *XClient) call(rpcAddr string, ctx context.Context, serviceMethod string, args, reply interface{}) error {
	client, err := xc.dial(rpcAddr)
	if err != nil {
		return err
	}
	return client.Call(ctx, serviceMethod, args, reply)
}
