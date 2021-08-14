package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash 根据data数据返回对应的hash
type Hash func(data []byte) uint32

type Map struct {
	hash     Hash           // 计算hash的函数
	replicas int            // 虚拟节点倍数
	keys     []int          // 哈希环
	hashMap  map[int]string // 虚拟节点和真实节点的映射表 key:hash,value:真实节点的名称
}

// New create Map Instance
func New(replicas int, fn Hash) *Map {
	m := &Map{
		hash:     fn,
		replicas: replicas,
		hashMap:  make(map[int]string),
	}

	// default algorithm
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}

	return m
}

// Add 添加节点 key:节点的名字
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			// 根据key计算出hash
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			// add key
			m.keys = append(m.keys, hash)
			// 增加虚拟节点和真实节点的映射关系。
			m.hashMap[hash] = key
		}
		sort.Ints(m.keys)
	}
}

func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	// get hash
	hash := int(m.hash([]byte(key)))

	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	return m.hashMap[m.keys[idx%len(m.keys)]]
}
