package main

import (
	"fmt"
	"strconv"
	"strings"
)

//     ******
//    x  ****
//     66****
//    6*****
//  **666**
//  **6**6
// ----------
// ****66****

func main() {
	for n1 := 110_000; n1 <= 669_999; n1++ {
		for n21 := 1; n21 <= 9; n21++ {
			s1 := n1 * n21
			if s1 >= 669_999 || s1 < 660_000 {
				continue
			}
			if !numX(s1, 1, 2) {
				continue
			}

			for n22 := 1; n22 <= 9; n22++ {
				s2 := n1 * n22
				if s2 < 600_000 || s2 > 699_999 {
					continue
				}
				if !numX(s2, 1) {
					continue
				}

				for n23 := 1; n23 <= 9; n23++ {
					s3 := n1 * n23
					if s3 < 1_000_000 {
						continue
					}
					if !numX(s3, 3, 4, 5) {
						continue
					}

					for n24 := 1; n24 <= 9; n24++ {
						s4 := n1 * n24
						if s4 > 996_996 {
							continue
						}
						if !numX(s4, 3, 6) {
							continue
						}

						s := s1 + s2*10 + s3*100 + s4*1_000
						if !numX(s, 5, 6) {
							continue
						}

						fmt.Println("s1: ", s1)
						fmt.Println("s2: ", s2)
						fmt.Println("s3: ", s3)
						fmt.Println("s4: ", s4)
						fmt.Printf("%d * %d%d%d%d = %d\n", n1, n24, n23, n22, n21, s)
						return
					}
				}
			}
		}
	}
}

func numX(s int, p ...int) bool {
	sSlice := strings.Split(strconv.Itoa(s), "")
	for _, num := range p {
		if len(sSlice) < num {
			return false
		}
		if sSlice[num-1] != "6" {
			return false
		}
	}
	return true
}
