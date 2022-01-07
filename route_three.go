package admin

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
)

var (
	RouteNodeWalkStop    = errors.New("route node walk stop")
	RouteNodeWalkSkipDir = errors.New("route node walk skip dir")
)

type RouteNode struct {
	Parent       *RouteNode
	Children     map[string]*RouteNode
	Menu         *Menu
	Handler      http.Handler
	Depth        int
	Resource     *Resource
	ResourceItem bool
	Param        string
	Action       *Action
	Scheme       *Scheme

	MethodHandlers map[string]http.Handler
}

func (this *RouteNode) IsIndex() bool {
	return this.Menu == nil && this.Handler == nil && this.Resource == nil && this.Action == nil && this.Scheme == nil
}

func (this *RouteNode) Handle(method string, handler http.Handler) *RouteNode {
	if this.MethodHandlers == nil {
		this.MethodHandlers = map[string]http.Handler{}
	}
	this.MethodHandlers[method] = handler
	return this
}

func (this *RouteNode) Get(handler http.Handler) *RouteNode {
	return this.Handle(http.MethodGet, handler)
}

func (this *RouteNode) Post(handler http.Handler) *RouteNode {
	return this.Handle(http.MethodPost, handler)
}

func (this *RouteNode) Put(handler http.Handler) *RouteNode {
	return this.Handle(http.MethodPut, handler)
}

func (this *RouteNode) Delete(handler http.Handler) *RouteNode {
	return this.Handle(http.MethodDelete, handler)
}

func (this *RouteNode) Head(handler http.Handler) *RouteNode {
	return this.Handle(http.MethodHead, handler)
}

func (this *RouteNode) Path() string {
	var (
		el  = this
		pth []string
	)

	for el != nil {
		pth = append(pth, el.Param)
		el = el.Parent
	}
	for i, j := 0, len(pth)-1; i < j; i, j = i+1, j-1 {
		pth[i], pth[j] = pth[j], pth[i]
	}
	return strings.Join(pth, "/")
}

func (this *RouteNode) updateChildren() {
	for _, chield := range this.Children {
		chield.Depth = this.Depth + 1
		chield.Parent = this
		chield.updateChildren()
	}
}

func (this *RouteNode) addChild(param string, node *RouteNode) *RouteNode {
	if this.Children == nil {
		this.Children = map[string]*RouteNode{}
	}
	this.Children[param] = node
	node.Param = param
	node.Depth = this.Depth + 1
	node.updateChildren()
	return node
}

func (this *RouteNode) Pair() (path, label string) {
	var args []string
	if this.Depth > 1 {
		args = append(args, strings.Repeat("  ", this.Depth))
	}
	path = this.Param
	if this.Resource != nil {
		if this.ResourceItem {
			path = "{id}"
			label = "Res: _"
		} else {
			label = "Res: " + this.Resource.ID
		}
	} else if this.Action != nil {
		label = "{Action}"
	} else if this.Scheme != nil {
		label = "{Scheme}"
	} else if this.Menu != nil {
		label = "Menu: " + this.Menu.Name
	} else if this.Handler != nil {
		label = fmt.Sprintf("{Handler %s}", this.Handler)
		args = append(args, this.Param)
	} else if this.Param != "" {
		args = append(args, this.Param)
	}
	if len(this.Children) > 0 {
		path += "/"
	}

	var methods []string
	for method := range this.MethodHandlers {
		methods = append(methods, method)
	}
	if len(methods) > 0 {
		sort.Strings(methods)
		if label != "" {
			label += " "
		}
		label += "→ " + strings.Join(methods, "|")
	}
	return
}

func (this *RouteNode) I18nPair(ctx *Context) struct{ Path, Label string } {
	var (
		args        []string
		path, label string
	)
	if this.Depth > 1 {
		args = append(args, strings.Repeat("  ", this.Depth))
	}
	path = this.Param
	if this.Resource != nil {
		if this.ResourceItem {
			path = "{id}"
			label = this.Resource.GetLabel(ctx, false)
		} else if this.Resource.Config.Singleton {
			label = this.Resource.GetLabel(ctx, false)
		} else {
			label = this.Resource.GetLabel(ctx, true)
		}
	} else if this.Action != nil {
		label = ctx.TtS(this.Action)
	} else if this.Scheme != nil {
		label = ctx.TtS(this.Scheme.defaultMenu)
	} else if this.Menu != nil {
		label = ctx.TtS(this.Scheme.defaultMenu)
	} else if this.Handler != nil {
	} else if this.Param != "" {
		args = append(args, this.Param)
	}
	return struct{ Path, Label string }{path, label}
}

type RouteNodeWalkerItem struct {
	First, Last bool
	Path        string
	Node        *RouteNode
}

func (this *RouteNode) Walk(cb func(parents []*RouteNodeWalkerItem, item *RouteNodeWalkerItem) (err error)) error {
	var (
		parents []*RouteNodeWalkerItem
	)

	var walk func(parents []*RouteNodeWalkerItem, path string, node *RouteNode) (err error)

	walk = func(parents []*RouteNodeWalkerItem, path string, node *RouteNode) (err error) {
		if node.Children == nil {
			return
		}

		var (
			params []string
			l      = len(node.Children)
		)

		for param := range node.Children {
			params = append(params, param)
		}

		sort.Slice(params, func(i, j int) bool {
			if node.Children[params[i]].ResourceItem {
				return true
			} else if node.Children[params[j]].ResourceItem {
				return false
			}
			return params[i] < params[j]
		})

	paramsLoop:
		for i, param := range params {
			child := node.Children[param]
			item := &RouteNodeWalkerItem{i == 0, i == l-1, path + "/" + child.Param, child}
			if err = cb(parents, item); err != nil {
				if err == RouteNodeWalkSkipDir {
					err = nil
					continue paramsLoop
				}
				return err
			}
			if err = walk(append(parents, item), item.Path, child); err != nil {
				return
			}
		}
		return
	}
	return walk(parents, "", this)
}

func (this *RouteNode) Add(pth string, node *RouteNode) *RouteNode {
	pth = strings.Trim(pth, "/")
	var (
		child     *RouteNode
		nodeParam string

		parts = strings.Split(pth, "/")
		el    = this
	)
	nodeParam, parts = parts[len(parts)-1], parts[0:len(parts)-1]

	for _, param := range parts {
		if el.Children == nil {
			el.Children = map[string]*RouteNode{}
		}
		if child = el.Children[param]; child == nil {
			child = el.addChild(param, &RouteNode{Parent: el})
		}
		el = child
	}
	return el.addChild(nodeParam, node)
}

type RouteTree struct {
	RouteNode
}
