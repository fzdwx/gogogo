package cache

// ByteView 抽象了一个只读数据结构 ByteView 用来表示缓存值，是 GeeCache 主要的数据结构之一。
type ByteView struct {
	b []byte // 存储真实的缓存值，为了支持任意数据类型
}

// String toString impl Stringer
func (bv ByteView) String() string {
	return string(bv.b)
}

// Len 返回当前view的长度  impl Value
func (bv ByteView) Len() int {
	return len(bv.b)
}

// ByteSlice 返回一个拷贝，防止缓存值被外部程序修改。
func (bv ByteView) ByteSlice() []byte {
	return cloneBytes(bv.b)
}

// cloneBytes 克隆b中的数据，返回一个数据相同的byte数组
func cloneBytes(b []byte) []byte {
	bytes := make([]byte, len(b))
	copy(bytes, b)
	return bytes
}
