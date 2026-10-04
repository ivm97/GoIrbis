package irbis

import "fmt"

// TreeNode is one node in a TRE hierarchy.
type TreeNode struct {
	Value    string
	Children []TreeNode
	level    int
}

// Add appends a child node.
func (node *TreeNode) Add(value string) *TreeNode {
	child := TreeNode{Value: value}
	node.Children = append(node.Children, child)
	return &node.Children[len(node.Children)-1]
}

func (node *TreeNode) String() string {
	return node.Value
}

// TreeFile is an IRBIS TRE menu/tree.
type TreeFile struct {
	Roots []TreeNode
}

func countIndent(text string) (result int) {
	for i := 0; i < len(text); i++ {
		if text[i] == '\t' {
			result++
			continue
		}
		break
	}
	return
}

// AddRoot appends a top-level node.
func (tree *TreeFile) AddRoot(value string) *TreeNode {
	tree.Roots = append(tree.Roots, TreeNode{Value: value})
	return &tree.Roots[len(tree.Roots)-1]
}

// Parse builds the tree from tab-indented TRE lines.
// Returns an error for empty roots with non-zero indent or skipped levels.
func (tree *TreeFile) Parse(lines []string) error {
	type item struct {
		level int
		value string
	}

	flat := make([]item, 0, len(lines))
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		level := countIndent(line)
		flat = append(flat, item{level: level, value: line[level:]})
	}
	if len(flat) == 0 {
		return nil
	}
	if flat[0].level != 0 {
		return fmt.Errorf("tree: first line must have zero indent")
	}

	nodes := make([]*TreeNode, len(flat))
	for i, it := range flat {
		nodes[i] = &TreeNode{Value: it.value, level: it.level}
	}

	stack := make([]*TreeNode, 0, 8)
	for _, n := range nodes {
		if n.level == 0 {
			stack = stack[:0]
			stack = append(stack, n)
			continue
		}
		if n.level > len(stack) {
			return fmt.Errorf("tree: invalid indent level %d", n.level)
		}
		stack = stack[:n.level]
		parent := stack[len(stack)-1]
		parent.Children = append(parent.Children, TreeNode{Value: n.Value, level: n.level})
		stack = append(stack, &parent.Children[len(parent.Children)-1])
	}

	for _, n := range nodes {
		if n.level == 0 {
			tree.Roots = append(tree.Roots, *n)
		}
	}
	return nil
}
