package go_concurrent

import (
	"fmt"
	"sync"
	"time"
)

type (
	// subscriber 订阅者所拥有的管道
	subscriber chan interface{}
	// message 消息
	message struct {
		topic string      // 当前消息的主题
		v     interface{} // 消息的内容
	}
	// publisher 发布者
	publisher struct {
		rmu         sync.RWMutex          // 读写锁
		buffer      int                   // 订阅队列的缓存大小
		timeout     time.Duration         // 发布超时时间
		subscribers map[subscriber]string // 维护订阅者的管道和对应的主题
	}
)

func newMessage(topic string, v interface{}) *message {
	return &message{topic: topic, v: v}
}

func newPublisher(buffer int, publishTimeOut time.Duration) *publisher {
	return &publisher{
		buffer:      buffer,
		timeout:     publishTimeOut,
		subscribers: make(map[subscriber]string),
	}
}

// subscribe 添加一个新的订阅者，订阅全部主题
func (p *publisher) subscribe() chan interface{} {
	return p.subscribeTopic("")
}

// SubscribeTopic 添加一个新的订阅者，订阅过滤器筛选后的主题
func (p *publisher) subscribeTopic(topic string) chan interface{} {
	ch := make(chan interface{}, p.buffer)
	p.rmu.Lock()
	p.subscribers[ch] = topic
	p.rmu.Unlock()

	return ch
}

// Evict 退出订阅
func (p *publisher) evict(sub chan interface{}) {
	p.rmu.Lock()
	defer p.rmu.Unlock()

	delete(p.subscribers, sub)
	close(sub)
}

// Publish 发布一个消息，给所有订阅者
func (p *publisher) publish(message *message) {
	p.rmu.RLock()
	defer p.rmu.RUnlock()

	var wg sync.WaitGroup

	for sub, topic := range p.subscribers {
		wg.Add(1)
		go p.sendTopic(sub, topic, message, &wg)
	}
	wg.Wait()
}

// Close  关闭发布者对象，同时关闭所有的订阅者管道。
func (p *publisher) close() {
	p.rmu.Lock()
	defer p.rmu.Unlock()

	for sub := range p.subscribers {
		delete(p.subscribers, sub)
		close(sub)
	}
}

// sendTopic 发送主题，可以容忍一定的超时
func (p *publisher) sendTopic(sub subscriber, destTopic string, message *message, wg *sync.WaitGroup) {
	defer wg.Done()

	// 当目的主题不为空时判断两个主题是否相同
	if destTopic != "" && destTopic != message.topic {
		return
	}

	select {
	case sub <- message.v:
	case <-time.After(p.timeout):
	}
}

func RunPubSubDemo() {
	p := newPublisher(10, 100*time.Millisecond)
	defer p.close()

	all := p.subscribe()
	golang := p.subscribeTopic("golang")

	p.publish(newMessage("golang", "发给订阅了golang频道的人"))
	p.publish(newMessage("all", "发给订阅了all频道的人"))

	go func() {
		for msg := range all {
			fmt.Println("all:", msg)
		}
	}()

	go func() {
		for msg := range golang {
			fmt.Println("golang:", msg)
		}
	}()

	// 运行一定时间后退出
	time.Sleep(3 * time.Second)
}
