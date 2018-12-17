package test

import (
	"testing"
)

func TestQuickSort(t *testing.T) {
	a := []string{"0", "90", "890", "7890", "67890", "567890", "4567890", "690", "34567890", "234567890", "1234567890", "790"}
	quicksort(a)
	t.Log(a)
}

func quicksort(a []string) {
	if len(a) <= 1 {
		return
	}
	i := 0
	j := len(a) - 1
	p := 0
	for i <= j {
		for j >= 0 {
			if len(a[j]) > len(a[p]) {
				a[j], a[p] = a[p], a[j]
				p = j
				j--
				break
			}
			j--
		}
		for i <= j {
			if len(a[i]) < len(a[p]) {
				a[i], a[p] = a[p], a[i]
				p = i
				i++
				break
			}
			i++
		}
	}
	quicksort(a[0:p])
	quicksort(a[p+1:])
}
