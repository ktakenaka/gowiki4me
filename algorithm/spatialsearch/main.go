package main

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"os"
	"strings"

	qtree "github.com/Johniel/go-quadtree/src/tree"
	_ "github.com/mattn/go-sqlite3"
)

type Point struct {
	X float64
	Y float64
}

type Node struct {
	Min   *Point
	Max   *Point
	Depth int32
}

func (node *Node) Children() []*Node {
	dx := (node.Max.X - node.Min.X) / 2.0
	dy := (node.Max.Y - node.Min.Y) / 2.0
	children := make([]*Node, 1<<2)
	for idx := range children {
		ch := &Node{
			Min: &Point{
				X: node.Min.X,
				Y: node.Min.Y,
			},
			Max: &Point{
				X: node.Min.X + dx,
				Y: node.Min.Y + dy,
			},
			Depth: node.Depth + 1,
		}
		if (idx & (1 << 0)) != 0 {
			ch.Min.X += dx
			ch.Max.X += dx
		}
		if (idx & (1 << 1)) != 0 {
			ch.Min.Y += dy
			ch.Max.Y += dy
		}
		children[idx] = ch
	}
	return children
}

func (node *Node) IsInside(p *Point) bool {
	return node.Min.X <= p.X && node.Min.Y <= p.Y && p.X < node.Max.X && p.Y < node.Max.Y
}

func (node *Node) Adjacent() []*Node {
	dx := node.Max.X - node.Min.X
	dy := node.Max.Y - node.Min.Y

	dirX := []float64{-1, -1, -1, 0, 0, +1, +1, +1}
	dirY := []float64{-1, 0, +1, -1, +1, -1, 0, +1}

	adjacent := make([]*Node, len(dirX))
	for d := 0; d < len(dirX); d++ {
		m := &Node{
			Min:   &Point{X: node.Min.X, Y: node.Min.Y},
			Max:   &Point{X: node.Max.X, Y: node.Min.Y},
			Depth: node.Depth,
		}
		m.Min.X += dirX[d] * dx
		m.Min.Y += dirY[d] * dx
		m.Max.X += dirX[d] * dy
		m.Max.Y += dirY[d] * dy
		adjacent[d] = m
	}

	return adjacent
}

type Tree struct {
	min *Point
	max *Point
}

func NewTree(min *Point, max *Point) *Tree {
	return &Tree{min: min, max: max}
}

func (t *Tree) Path(p *Point, depth int32) (*Node, string) {
	node := &Node{Min: t.min, Max: t.max}

	builder := &strings.Builder{}
	label := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz{|}"
	for node.Depth < depth {
		for idx, ch := range node.Children() {
			if ch.IsInside(p) {
				node = ch
				builder.WriteByte(label[idx])
				break
			}
		}
	}
	return node, builder.String()
}

type repository struct {
	db    *sql.DB
	tree  *qtree.Tree
	depth int32
}

func (r *repository) init(minPoint, maxPoint *qtree.Point, depth int32) error {
	os.Remove("./demo.db")
	db, err := sql.Open("sqlite3", "./demo.db")
	if err != nil {
		return err
	}

	createTable := `
CREATE TABLE Points (
  id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  x REAL NOT NULL,
  y REAL NOT NULL,
  path TEXT NOT NULL
);`
	_, err = db.Exec(createTable)
	if err != nil {
		return err
	}

	createIndex := `CREATE INDEX indexPath ON Points(path);`
	_, err = db.Exec(createIndex)
	if err != nil {
		return err
	}

	r.db = db
	r.tree = qtree.NewTree(minPoint, maxPoint)
	r.depth = depth
	return nil
}

func (r *repository) finalize() error {
	return r.db.Close()
}

func (r *repository) insert(p *qtree.Point) error {
	_, h := r.tree.Path(p, 10)
	_, err := r.db.Exec("INSERT INTO Points (x, y, path) VALUES(?,?,?)", p.X, p.Y, h)
	return err
}

func (r *repository) search(p *qtree.Point, depth int32) ([]*qtree.Point, error) {
	_, path := r.tree.Path(p, depth)
	rows, err := r.db.Query("SELECT x, y FROM Points WHERE ? <= path AND path <= ?", path, path+"~")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ps := []*qtree.Point{}
	for rows.Next() {
		var x, y float64
		err := rows.Scan(&x, &y)
		if err != nil {
			return nil, err
		}
		q := &qtree.Point{
			X: x,
			Y: y,
		}
		ps = append(ps, q)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return ps, nil
}

func (r *repository) circleSearch(center *qtree.Point, radius float64) ([]*qtree.Point, error) {
	root, _ := r.tree.Path(center, 0)

	depth := int32(0)
	for ; radius < (root.Max.X-root.Min.X)/math.Pow(2.0, float64(depth)); depth++ {
	}
	depth--
	centerNode, _ := r.tree.Path(center, depth)
	candidates, err := r.search(center, depth)
	if err != nil {
		return nil, err
	}

	for _, adj := range centerNode.Adjacent() {
		matched, err := r.search(adj.Mid(), depth)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, matched...)
	}
	matched := []*qtree.Point{}
	for _, c := range candidates {
		if (c.X-center.X)*(c.X-center.X)+(c.Y-center.Y)*(c.Y-center.Y) <= radius*radius {
			matched = append(matched, c)
		}
	}
	return matched, nil
}

func main() {
	dataset, err := os.Create("./dataset.tsv")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer dataset.Close()

	liner, err := os.Create("./liner.tsv")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer liner.Close()

	circle, err := os.Create("./circle.tsv")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer circle.Close()

	minPoint := &qtree.Point{
		X: 0.0,
		Y: 0.0,
	}
	maxPoint := &qtree.Point{
		X: 32.0,
		Y: 32.0,
	}

	demo := &repository{}
	err = demo.init(minPoint, maxPoint, 10)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer demo.finalize()

	for i := minPoint.X; i < maxPoint.X; i += 0.7 {
		for j := minPoint.Y; j < maxPoint.Y; j += 0.7 {
			p := &qtree.Point{
				X: i,
				Y: j,
			}
			err := demo.insert(p)
			if err != nil {
				log.Fatal(err)
				return
			}
			dataset.Write(([]byte)(fmt.Sprintf("%f\t%f\n", i, j)))
		}
	}

	p := &qtree.Point{
		X: 8.1,
		Y: 8.2,
	}
	ps, err := demo.search(p, 3)
	if err != nil {
		log.Fatal(err)
		return
	}
	for _, p := range ps {
		fmt.Printf("matched: (%f,%f)\n", p.X, p.Y)
	}
	fmt.Println("")

	node, _ := demo.tree.Path(p, 5)
	for _, a := range node.Adjacent() {
		ps, err := demo.search(a.Mid(), node.Depth)
		if err != nil {
			log.Fatal(err)
			return
		}
		for _, p := range ps {
			fmt.Printf("matched: (%f,%f)\n", p.X, p.Y)
		}
	}
	fmt.Println("")

	begin := &qtree.Point{
		X: 4.0,
		Y: 2.0,
	}
	end := &qtree.Point{
		X: 20.0,
		Y: 30.0,
	}
	curr, _ := demo.tree.Path(begin, 5)
	for curr.Min.X <= end.X && curr.Min.Y <= end.Y {
		ps, err := demo.search(curr.Mid(), 5)
		if err != nil {
			log.Fatal(err)
			return
		}
		for _, q := range ps {
			liner.Write(([]byte)(fmt.Sprintf("%f\t%f\n", q.X, q.Y)))
		}
		curr = curr.Adjacent()[7]
	}

	center := &qtree.Point{
		X: 20.0,
		Y: 20.0,
	}
	radius := 5.0
	ps, _ = demo.circleSearch(center, radius)
	for _, p := range ps {
		circle.Write(([]byte)(fmt.Sprintf("%f\t%f\n", p.X, p.Y)))
	}
}
