package main

import (
	"fmt"
	"testing"
)

type mockDrawable struct {
	id int
}

func (md *mockDrawable) Draw() {
	fmt.Printf("This mock id is %d\n", md.id)
}

func TestWholeScreen(t *testing.T) {
	m1 := &mockDrawable{id: 1}
	m2 := &mockDrawable{id: 2}
	ws := &WholeScreen{}
	ws.append(m1)
	ws.append(m2)
	ws.drawAll()
}
