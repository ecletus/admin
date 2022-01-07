package admin

import (
	"strings"

	"github.com/ecletus/core"
)

const TreeNodeDepthStringMeta = "TREE_NODE_DEPTH_STRING"

var DefaultTreeNodeDepthFormatter = StdTreeNodeDepthFormatter{"├── ", "└── ", "│   ", "    "}

type (
	ModelTreeNoder interface {
		IsLast() bool
		Label() string
		Parent() ModelTreeNoder
		Node() interface{}
	}

	ModelTreeNode struct {
		node   interface{}
		isLast bool
		label  string
		parent func() ModelTreeNoder
	}

	TreeNodeDepthFormatter interface {
		Format(context *core.Context, n ModelTreeNoder) *TreeNodeDepthString
	}

	TreeNodeDepthString struct {
		Depth, Label         string
		SafeDepth, SafeLabel string
	}

	TreeNodeDepthFormatterFunc func(context *core.Context, n ModelTreeNoder) *TreeNodeDepthString
)

func (t TreeNodeDepthFormatterFunc) Format(context *core.Context, n ModelTreeNoder) *TreeNodeDepthString {
	return t(context, n)
}

func (t TreeNodeDepthString) String() string {
	return t.Depth + " " + t.Label
}

func NewModelTreeNode(node interface{}, isLast bool, label string, parent func() ModelTreeNoder) *ModelTreeNode {
	return &ModelTreeNode{node: node, isLast: isLast, label: label, parent: parent}
}

func (n *ModelTreeNode) Node() interface{} {
	return n.node
}

func (n *ModelTreeNode) Label() string {
	return n.label
}

func (n *ModelTreeNode) IsLast() bool {
	return n.isLast
}

func (n *ModelTreeNode) Parent() ModelTreeNoder {
	return n.parent()
}

type StdTreeNodeDepthFormatter struct {
	Child, LastChild, Dir, Empty string
}

func (s *StdTreeNodeDepthFormatter) Format(_ *core.Context, n ModelTreeNoder) *TreeNodeDepthString {
	var (
		e      = n
		p      = e.Parent()
		chunks []string
	)

	if p != nil {
		if e.IsLast() {
			chunks = append(chunks, s.LastChild)
		} else {
			chunks = append(chunks, s.Child)
		}

		for p != nil && p.Parent() != nil {
			if p.IsLast() {
				chunks = append(chunks, s.Empty)
			} else {
				chunks = append(chunks, s.Dir)
			}
			p = p.Parent()
		}
		chunks = append(chunks, "  ")
	}

	for i, j := 0, len(chunks)-1; i < j; i, j = i+1, j-1 {
		chunks[i], chunks[j] = chunks[j], chunks[i]
	}
	return &TreeNodeDepthString{Depth: strings.Join(chunks, ""), Label: e.Label()}
}

func TreeNodeDepthStringMetaOf(res *Resource, fmtr TreeNodeDepthFormatter, getNode func(r interface{}) ModelTreeNoder, cb ...func(meta *Meta)) {
	meta := &Meta{
		Name: TreeNodeDepthStringMeta,
		Type: "tree_node_depth_string",
		Valuer: func(record interface{}, context *core.Context) interface{} {
			var e = getNode(record)
			return fmtr.Format(context, e)
		},
		FormattedValuer: func(record interface{}, context *core.Context) (fv *FormattedValue) {
			var (
				e  = getNode(record)
				ds = fmtr.Format(context, e)
			)

			fv = &FormattedValue{Record: record, Raw: ds}
			if ds.Depth == "" {
				fv.Value = ds.Label
				if ds.SafeLabel != "" {
					fv.SafeValue = ds.SafeLabel
				}
				return
			}
			return
		},
	}

	for _, cb := range cb {
		cb(meta)
	}

	res.Meta(meta)
}
