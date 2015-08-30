package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
	utf8 "unicode/utf8"

	"github.com/nsf/termbox-go"
)

type mockDrawable struct {
	t  *testing.T
	id int
}

func (md *mockDrawable) Draw() {
	// Run "go test -v ." to show Logf
	md.t.Logf("This mock id is %d\n", md.id)
}

func TestByteSliceInsert(t *testing.T) {
	expected := []byte("abc1de")
	text := []byte("abcde")
	offset := 3
	what := []byte("1")
	result := byteSliceInsert(text, offset, what)
	t.Logf("result: %s\n", result)
	t.Logf("expected: %s\n", expected)
}

func TestInsertRune(t *testing.T) {
	expected := []byte("iInitial")
	eb := &EditBox{text: []byte("initial"), cursorBoffset: 1}
	insert, _ := utf8.DecodeRune([]byte("I"))
	eb.InsertRune(insert)

	t.Logf("result: %s\n", eb.text)
	t.Logf("expected: %s\n", expected)
}

func TestWholeScreen(t *testing.T) {
	m1 := &mockDrawable{t: t, id: 1}
	m2 := &mockDrawable{t: t, id: 2}
	ws := &WholeScreen{}
	ws.append(m1)
	ws.append(m2)
	ws.drawAll()
}

func TestMoveCursor(t *testing.T) {
	var eb EditBox
	t.Logf("cursor boffset: %d\n", eb.cursorBoffset)
	r, _ := utf8.DecodeLastRune([]byte("a"))
	eb.InsertRune(r)
}

func TestRoomList(t *testing.T) {
	roomList := NewRoomBox(20)
	chatLogs := NewChatBox(20, roomList.rooms)
	roomList.AppendRoom(NewRoomInfo(1, "a", "b"))
	roomList.OtherEnterRoom(1, "test")
	t.Logf((*roomList.rooms)[0].Members[0])
	t.Logf((*chatLogs.rooms)[0].Members[0])
	chatLogs.CurrentRoomID = 1
	chatLogs.ShowRoomMember = true
	t.Logf(fmt.Sprintf("%d\n", chatLogs.CurrentRoomID))
	t.Logf(chatLogs.GetText(0))
	chatLogs.ShowRoomMember = false
	t.Logf(chatLogs.GetText(0))

	t.Logf(fmt.Sprintf("max line: %d", chatLogs.GetMaxLine()))

}

func TestAppendOfTextBox(t *testing.T) {
	tb := NewTextBox(5)
	templateMsg := "count %v"
	/*
		tb.AppendText(fmt.Sprintf(templateMsg, 1))
		t.Logf(tb.GetText(0))
		tb.AppendText(fmt.Sprintf(templateMsg, 2))
		t.Logf(tb.GetText(0))
		t.Logf(tb.GetText(1))
	*/
	for i := 0; i < 10; i++ {
		tb.AppendText(fmt.Sprintf(templateMsg, i))
		for j := 0; j < tb.maxLine; j++ {
			var expected string
			if i-j < 0 {
				expected = ""
			} else {
				expected = fmt.Sprintf(templateMsg, i-j)
			}
			result := tb.GetText(j)
			if !strings.EqualFold(expected, result) {
				t.Errorf("Unexpected strings.\nexpected: %v\nresult: %v", expected, result)
			}
		}
	}
}

func TestRoomBox(t *testing.T) {
	rb := NewRoomBox(10)
	for _, r := range *rb.rooms {
		t.Logf("%v", r)
	}
}

func TestEditBoxDraw(t *testing.T) {
	if err := termbox.Init(); err != nil {
		t.Logf("ERROR: Cannot initialize termbox")
		os.Exit(1)
	}
	defer termbox.Close()
	var eb EditBox
	termbox.SetInputMode(termbox.InputEsc)
	eb.Draw()
	termbox.SetCell(1, 1, 't', termbox.ColorDefault, termbox.ColorDefault)
	termbox.Flush()

	done := make(chan struct{})
	keyInput := Input(done)
	for {
		select {
		case k := <-keyInput:
			switch k.Type {
			case termbox.EventKey:
				switch k.Key {
				case termbox.KeyArrowRight, termbox.KeyCtrlF:
					eb.MoveCursorOneRuneForward()
				case termbox.KeyArrowLeft, termbox.KeyCtrlB:
					eb.MoveCursorOneRuneBackward()
				case termbox.KeyBackspace, termbox.KeyBackspace2:
					eb.DeleteRuneBackward()
				case termbox.KeyEnter:
					eb.GetAndDeleteText()
				case termbox.KeyEsc:
					done <- struct{}{}
					return
				default:
					eb.InsertRune(k.Ch)
				}
			case termbox.EventError:
				done <- struct{}{}
				return
			}
		default:
			termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
			eb.Draw()
		}

	}
}
