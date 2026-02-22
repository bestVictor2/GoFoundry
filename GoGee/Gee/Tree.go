package Gee

import "strings"

type node struct {
	pattern  string
	part     string
	children []*node
	isWild   bool
}

func (fa *node) MatchChild(part string) *node {
	for _, child := range fa.children {
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}
func (fa *node) MatchChildren(part string) []*node {
	nodes := make([]*node, 0)
	for _, child := range fa.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}
func (fa *node) Insert(pattern string, parts []string, height int) {
	if len(parts) == height {
		fa.pattern = pattern
		return
	}
	part := parts[height]
	child := fa.MatchChild(part)
	if child == nil {
		child = &node{part: part, isWild: part[0] == ':' || part[0] == '*'}
		fa.children = append(fa.children, child)
	}
	child.Insert(pattern, parts, height+1)
}
func (fa *node) Search(parts []string, height int) *node {
	if len(parts) == height || strings.HasPrefix(fa.part, "*") {
		if fa.pattern == "" {
			return nil
		}
		return fa
	}
	part := parts[height]
	children := fa.MatchChildren(part)
	for _, child := range children {
		result := child.Search(parts, height+1)
		if result != nil {
			return result
		}
	}
	return nil
}
