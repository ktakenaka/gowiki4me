package main

import (
	"errors"
	"fmt"
)

var (
	errTreeNil = errors.New("Tree is nil")
	errValueExist = errors.New("Node value already exists")
)

type TreeNode struct {
	val int
	left *TreeNode
	right *TreeNode
}

func (t *TreeNode) Insert(value int) error {
	if t == nil {
		return errTreeNil
	}

	if t.val == value {
		return errValueExist
	}

	if t.val > value {
		if t.left == nil {
			t.left = &TreeNode{val: value}
			return nil
		}

		return t.left.Insert(value)
	}

	if t.val < value {
		if t.right == nil {
			t.right = &TreeNode{val: value}
			return nil
		}

		return t.right.Insert(value)
	}

	return nil
}

func (t *TreeNode) FindMin() int {
	if t.left == nil {
		return t.val
	}
	return t.left.FindMin()
}

func (t *TreeNode) FindMax() int {
	if t.right == nil {
		return t.val
	}
	return t.right.FindMax()
}

func (t *TreeNode) PrintInOrder() {
	if t == nil {
		return
	}

	t.left.PrintInOrder()
	fmt.Print(t.val)
	t.right.PrintInOrder()
}

func main() {
	tree := &TreeNode{val: 50}
	
	tree.Insert(100)
	tree.Insert(10)
	tree.Insert(30)
	tree.PrintInOrder()
}
