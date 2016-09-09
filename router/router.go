package router

import (
	"strings"
)

type RouteTree struct {
	base 		*node
	notFound 	string
}

type node struct {
	handler 	string
	key		string
	parameter 	string
	children 	map[string] *node
}

func (tree *RouteTree) Match(path string) (string, map[string] string) {
	paths := clean(path)
	params := make(map[string] string)

	current := tree.base
	for _, leaf := range paths {
		if child, ok := current.children[leaf]; ok {
			current = child
		} else if child, ok := current.children["*"]; ok {
			current = child
			return current.handler, params
		} else if child, ok := current.children[":"]; ok {
			current = child
			params[current.parameter] = leaf
		} else {
			return tree.notFound, params
		}
	}

	return current.handler, params
}

func (tree *RouteTree) Default(handler string) {
	tree.notFound = handler
}

func (tree *RouteTree) Add(path string, handler string) {
	paths := clean(path)

	current := tree.base
	for _, leaf := range paths {
		parameter := ""

		if strings.HasPrefix(leaf, ":") {
			parameter = leaf
			leaf = ":"
		}

		if child, ok := current.children[leaf]; ok {
			current = child
		} else {
			newChild := &node{
				key: leaf,
				handler: "",
				parameter: parameter,
				children: make(map[string] *node),
			}

			current.children[leaf] = newChild
			current = newChild
		}

	}

	current.handler = handler
}

func clean(path string) []string {
	path = strings.Replace(path, "//", "/", -1)
	path = strings.Replace(path, "///", "/", -1)
	path = strings.Replace(path, "////", "/", -1)
	path = strings.Split(path, "?")[0]
	if path[0] == '/' {
		path = strings.TrimPrefix(path, "/")
	}
	return strings.Split(path, "/")
}

func Router() *RouteTree {
	return &RouteTree{
		base: &node{children: make(map[string] *node)},
	}
}

