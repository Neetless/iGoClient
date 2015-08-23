package main

import (
	"fmt"
	"os"

	"github.com/nsf/termbox-go"
)

// Drawable is a interface which has screen draw function.
type Drawable interface {
	Draw()
}

// WholeScreen has every parts of screen box and draw all.
type WholeScreen struct {
	screenBoxes []Drawable
}

func (ws *WholeScreen) drawAll() {
	for _, box := range ws.screenBoxes {
		box.Draw()
	}
}

func (ws *WholeScreen) append(box Drawable) {
	ws.screenBoxes = append(ws.screenBoxes, box)
}

func redraw() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	//x, y := termbox.Size()
	ch := []rune("a")
	termbox.SetCell(0, 0, ch[0], termbox.ColorDefault, termbox.ColorDefault)
	termbox.Flush()
}

func main() {
	if err := termbox.Init(); err != nil {
		fmt.Printf("Fatal error: %s\n", err)
		os.Exit(1)
	}
	defer termbox.Close()
	//termbox.SetInputMode(termbox.InputEsc)
	redraw()
mainloop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				break mainloop
			}

		case termbox.EventError:
			os.Exit(1)
		}
		redraw()
	}
}
