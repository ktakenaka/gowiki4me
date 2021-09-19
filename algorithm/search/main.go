package main

import (
	"errors"
	"fmt"
	"sync"
)

type Edge int

func (e Edge) To() int {
	return int(e)
}

type Graph struct {
	VertexEdges map[int][]Edge
}

func (g *Graph) size() int {
	return len(g.VertexEdges)
}

func (g *Graph) edges(v int) []Edge {
	return g.VertexEdges[v]
}

type Queue struct {
	queue []int
	lock  sync.Mutex
}

func (q *Queue) size() int {
	return len(q.queue)
}

func (q *Queue) push(s int) {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.queue = append(q.queue, s)
}

var (
	ErrEmptyQueue = errors.New("Queue is empty")
)

func (q *Queue) pop() (int, error) {
	if q.size() == 0 {
		return 0, ErrEmptyQueue
	}
	q.lock.Lock()
	defer q.lock.Unlock()
	val := q.queue[0]
	q.queue = q.queue[1:]
	return val, nil
}

func main() {
	g := &Graph{
		VertexEdges: map[int][]Edge{
			1: {2, 3},
			2: {3},
			3: {1},
			4: {},
		},
	}

	resultBFS := bfs(g, 1)
	fmt.Println(resultBFS)

	resultDFS := dfs(g, 1)
	fmt.Println(resultDFS)

}

func bfs(g *Graph, s int) map[int]bool {
	seen := map[int]bool{}
	for i := range g.VertexEdges {
		seen[i] = false
	}
	todo := Queue{}
	todo.push(s)

	for {
		v, err := todo.pop()
		if errors.Is(err, ErrEmptyQueue) {
			break
		}

		if seen[v] {
			continue
		}

		seen[v] = true
		for _, e := range g.edges(v) {
			todo.push(e.To())
		}
	}
	return seen
}

type Stack struct {
	stack []int
	lock  sync.Mutex
}

func (s *Stack) push(n int) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.stack = append(s.stack, n)
}

func (s *Stack) size() int {
	return len(s.stack)
}

var (
	ErrEmptyStack = errors.New("Queue is empty")
)

func (s *Stack) pop() (int, error) {
	if s.size() == 0 {
		return 0, ErrEmptyQueue
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	last := len(s.stack) - 1
	val := s.stack[last]
	s.stack = s.stack[:last]
	return val, nil
}

func dfs(g *Graph, s int) map[int]bool {
	seen := map[int]bool{}
	for i := range g.VertexEdges {
		seen[i] = false
	}
	todo := Stack{}
	todo.push(s)

	for {
		v, err := todo.pop()
		if errors.Is(err, ErrEmptyQueue) {
			break
		}

		if seen[v] {
			continue
		}

		seen[v] = true
		for _, e := range g.edges(v) {
			todo.push(e.To())
		}
	}
	return seen
}
