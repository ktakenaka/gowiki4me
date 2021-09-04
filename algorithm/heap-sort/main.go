package main

import "fmt"

type minheap struct {
    arr []int
}

func newMinHeap(arr []int) *minheap {
    minheap := &minheap{
        arr: arr,
    }
    return minheap
}

func (m *minheap) leftchildIndex(index int) int {
    return 2*index + 1
}

func (m *minheap) rightchildIndex(index int) int {
    return 2*index + 2
}

// first and second are indices
func (m *minheap) swap(first, second int) {
    temp := m.arr[first]
    m.arr[first] = m.arr[second]
    m.arr[second] = temp
}

// simply judge if the leaf exits or not
func (m *minheap) leaf(index int, size int) bool {
    if index >= (size/2) && index <= size {
        return true
    }
    return false
}

func (m *minheap) downHeapify(current int, size int) {
	if m.leaf(current, size) {
		return
	}

	fmt.Println("current index", current)
	fmt.Println("size", size)

	smallest := current
	leftChildIndex := m.leftchildIndex(current)
	rightChildIndex := m.rightchildIndex(current)
	if leftChildIndex < size && m.arr[leftChildIndex] < m.arr[smallest] {
        smallest = leftChildIndex
    }
    if rightChildIndex < size && m.arr[rightChildIndex] < m.arr[smallest] {
        smallest = rightChildIndex
    }
    if smallest != current {
        m.swap(current, smallest)
        m.downHeapify(smallest, size)
    }
}

func (m *minheap) buildMinHeap(size int) {
	for i := ((size / 2) - 1); i >= 0; i-- {
		m.downHeapify(i, size)
	}
	fmt.Println("finish buildMinHeap")
}

func (m *minheap) Sort(size int) {
	m.buildMinHeap(size)
	for i := size - 1; i > 0; i-- {
		// the minimum value at the moment is located on [0] and the value should be moved to the last in the slice.
		// That's why to swap(0, i)
		m.swap(0, i)
		m.downHeapify(0, i)
	}
}

func (m *minheap) Print() {
	for _, v := range(m.arr) {
		fmt.Println(v)
	}
}

func main() {
	inputArray := []int{6, 5, 3, 7, 2, 8, -1}
    minHeap := newMinHeap(inputArray)
	minHeap.Sort(len(inputArray))
	fmt.Println("----- result -----")
    minHeap.Print()
}
