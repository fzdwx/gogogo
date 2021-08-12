package gee

type node struct {
	pattern  string  // 待匹配路由 eg: /qwe/:name
	part     string  // 路由中的一部分 eg: :lang
	children []*node // 子节点

	/*
		为了实现动态路由匹配，加上了isWild这个参数。即当我们匹配 /p/go/doc/这个路由时，
		第一层节点，p精准匹配到了p，
		第二层节点，go模糊匹配到:lang，那么将会把lang这个参数赋值为go，继续下一层匹配。
		我们将匹配的逻辑，包装为一个辅助函数。
	*/
	isWild bool // 是否精确匹配，part含有 : 或者 * 时为true
}

// 第一个匹配成功的节点，用于插入
func (n node) matchChild(part string) *node {
	for _, child := range n.children {
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

// 所有匹配成功的节点，用于查找
func (n node) matchChildren(part string) []*node {
	nodes := make([]*node, 0)
	for _, child := range n.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}

// todo https://geektutu.com/post/gee-day3.html#Trie-%E6%A0%91%E5%AE%9E%E7%8E%B0
