package utils

import (
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/xlab/treeprint"
)

type Node struct {
	name   string
	leaf   bool
	nested []*Node
}

func NewNode(name string, leaf bool) *Node {
	return &Node{name: name, leaf: leaf}
}

func (n *Node) GetName() string {
	return n.name
}

func (n *Node) IsLeaf() bool {
	return n.leaf
}

func (n *Node) GetNested() []*Node {
	return n.nested
}

func (n *Node) String() string {
	if n.leaf {
		return n.name
	}
	s := n.name + "["
	sep := ""
	for _, n := range n.nested {
		s += sep + n.String()
		sep = ","
	}
	return s + "]"
}

func (n *Node) Add(dir bool, path ...string) *Node {
comp:
	for i, c := range path {
		last := i == len(path)-1
		if last && dir {
			last = false
		}
		if !last {
			if n.leaf {
				return nil
			}
		}
		for _, cand := range n.nested {
			if cand.name == c {
				n = cand
				continue comp
			}
		}
		next := &Node{
			name: c,
			leaf: last,
		}
		n.nested = append(n.nested, next)
		n = next
	}
	return n
}

func MapFSTree(fs vfs.FileSystem, path string) *Node {
	ok, err := vfs.IsDir(fs, path)
	if err != nil {
		return nil
	}
	root := &Node{
		name: ".",
		leaf: !ok,
	}

	_, base, _ := vfs.SplitPath(fs, path)
	vfs.Walk(fs, path, func(path string, fi vfs.FileInfo, err error) error {
		_, next, _ := vfs.SplitPath(fs, path)
		next = next[len(base):]
		root.Add(fi.IsDir(), next...)
		return nil
	})

	return root
}

func MapNodetoASCII(n *Node) string {
	if n == nil {
		return ""
	}
	tree := treeprint.New()
	tree.(*treeprint.Node).Value = n.GetName()
	addNested(tree, n)
	return tree.String()
}

func addNested(tree treeprint.Tree, n *Node) {
	for _, n := range n.nested {
		if n.leaf {
			tree.AddNode(n.GetName())
		} else {
			stree := tree.AddBranch(n.GetName())
			addNested(stree, n)
		}
	}
}
