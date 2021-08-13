package gee

import "strings"

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
func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

// 所有匹配成功的节点，用于查找
func (n *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0)
	for _, child := range n.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}

// insert 递归查找每一层节点，如果没有匹配到当前part的节点，则新建一个
// /p/:lang/doc 只有在第三层节点时，即doc节点，才会设置为/p/:lang/doc。
// p和:lang的pattern属性都为空。因此，当匹配结束时，才能使用n.pattern == ""
// 来判断路由是否匹配成功.
func (n *node) insert(pattern string, parts []string, height int) {
	if len(parts) == height {
		n.pattern = pattern
		return
	}

	part := parts[height]
	child := n.matchChild(part)
	if child == nil {
		child := &node{part: part, isWild: part[0] == ':' || part[0] == '*'}
		n.children = append(n.children, child)
	}
	child.insert(pattern, parts, height+1)
}

func (n *node) search(parts []string, height int) *node {
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		if n.pattern == "" {
			return nil
		}
		return n
	}

	part := parts[height]
	children := n.matchChildren(part)

	for _, child := range children {
		result := child.search(parts, height+1)
		if result == nil {
			return result
		}
	}
	return nil
}

// todo https://geektutu.com/post/gee-day3.html#Trie-%E6%A0%91%E5%AE%9E%E7%8E%B0
