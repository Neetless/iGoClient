package main

import (
	"fmt"
	"log"
	"os"
	"unicode/utf8"

	"github.com/nsf/termbox-go"
)

// TextArea define what TextArea needs.
type TextArea interface {
	GetText(n int) string
	GetMaxLine() int
}

// TextScreen manage what should show in the text area.
type TextScreen struct {
	ta TextArea
}

// SetTextArea set TextArea field.
func (ts *TextScreen) SetTextArea(ta TextArea) {
	ts.ta = ta
}

// GetTextArea return TextArea field.
func (ts *TextScreen) GetTextArea() TextArea {
	return ts.ta
}

// Draw textlog.
func (ts *TextScreen) Draw() {
	for i := 0; i < ts.GetTextArea().GetMaxLine(); i++ {
		setCellLine(0, i+2,
			termbox.ColorDefault, termbox.ColorDefault,
			ts.GetTextArea().GetText(i))
	}
	//termbox.Flush()
}

// TextBox store maxLine number of texts.
type TextBox struct {
	maxLine        int
	oldestPosition int
	textLogs       []string
}

// NewTextBox create TextBox instance.
func NewTextBox(maxLine int) *TextBox {
	return &TextBox{maxLine, 0, make([]string, maxLine)}
}

// GetText return the ordered textLog.
func (tb *TextBox) GetText(n int) string {
	position := (tb.oldestPosition+tb.maxLine-1)%tb.maxLine - n
	for position < 0 {
		position = tb.maxLine + position
	}
	return tb.textLogs[position]
}

// AppendText append text to textLogs.
func (tb *TextBox) AppendText(text string) {
	tb.textLogs[tb.oldestPosition] = text
	// Set position to oldest log.
	tb.oldestPosition = (tb.oldestPosition + 1) % tb.maxLine
}

// GetMaxLine return number of text log's line.
func (tb *TextBox) GetMaxLine() int {
	return tb.maxLine
}

const (
	// NotExist is one of the ID which identify not existing room log.
	NotExist = -1
)

// ChatBox contain a every room's conversation logs.
type ChatBox struct {
	MaxRoom       int
	CurrentRoomID int
	ChatLogs      []ChatLog
}

// NewChatBox create new instance for ChatBox
func NewChatBox(maxRoom int) *ChatBox {
	MaxLogs := 30
	logs := make([]ChatLog, maxRoom)
	for i := range logs {
		logs[i] = ChatLog{NotExist, MaxLogs, NewTextBox(MaxLogs)}
	}
	return &ChatBox{maxRoom, NotExist, logs}
}

// AppendText append chat log
func (cb *ChatBox) AppendText(id int, chatLog string) {
	// When cb has the room's conversation log.
	for i, log := range cb.ChatLogs {
		if log.RoomID == id {
			cb.ChatLogs[i].Logs.AppendText(chatLog)
			return
		}
	}

	// When cb doesn't have the room's conversation log.
	for i, cl := range cb.ChatLogs {
		if cl.RoomID == NotExist {
			cb.ChatLogs[i].RoomID = id
			cb.ChatLogs[i].Logs.AppendText(chatLog)
			return
		}
	}
}

// GetText return given line's text
func (cb *ChatBox) GetText(n int) string {
	if cb.CurrentRoomID == NotExist {
		return " "
	}
	return cb.ChatLogs[cb.CurrentRoomID].Logs.GetText(n)
}

// GetMaxLine return max line of conversation.
func (cb *ChatBox) GetMaxLine() int {
	return cb.ChatLogs[0].MaxLine
}

// ChatLog contain a conversation log.
type ChatLog struct {
	RoomID  int
	MaxLine int
	Logs    *TextBox
}

// RoomBox has room list to show
type RoomBox struct {
	maxRoomNum int
	rooms      []RoomInfo
}

// NewRoomBox is a constructor of RoomBox
func NewRoomBox(max int) *RoomBox {
	return &RoomBox{max, make([]RoomInfo, max)}
}

// GetText return
func (r *RoomBox) GetText(n int) string {
	entered := " "
	if r.rooms[n].Entered {
		entered = "entered"
	}
	return fmt.Sprintf("%d %s %s %s",
		r.rooms[n].ID, r.rooms[n].Name, r.rooms[n].Owner, entered)
}

// GetMaxLine return the max number of array
func (r *RoomBox) GetMaxLine() int {
	return r.maxRoomNum
}

// RemoveRoom remove RoomInfo from list
func (r *RoomBox) RemoveRoom(id int) {
	for i, room := range r.rooms {
		if room.ID == id {
			r.rooms[i].ID = 0
			return
		}
	}
	log.Println("Cannot remove room. No such room.")
}

// AppendRoom append new room to the slice
func (r *RoomBox) AppendRoom(ri RoomInfo) {
	// Check whether the room already exist.
	for _, room := range r.rooms {
		if ri.ID == room.ID {
			log.Println("Cannnot append room. The room already exists.")
			return
		}
	}

	// Add room where empty item in slice.
	for i, room := range r.rooms {
		if room.ID == 0 {
			r.rooms[i] = ri
			return
		}
	}

}

// EnterRoom chage flag to detect entered or not.
func (r *RoomBox) EnterRoom(id int) {
	for i, room := range r.rooms {
		if room.ID == id {
			r.rooms[i].Entered = true
		}
	}
}

// QuitRoom change flag to detect entered or not.
func (r *RoomBox) QuitRoom(id int) {
	for i, room := range r.rooms {
		if room.ID == id {
			r.rooms[i].Entered = false
		}
	}
}

// RoomInfo has a Room information
type RoomInfo struct {
	// When ID = 0, it means no RoomInfo
	ID      int
	Name    string
	Owner   string
	Entered bool
}

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
	termbox.Flush()
}

func (ws *WholeScreen) append(box Drawable) {
	ws.screenBoxes = append(ws.screenBoxes, box)
}

// TODO Not tested and not used
func setCellLine(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x++
		// Multibyte character use 2 cell spaces.
		if utf8.RuneLen(c) > 2 {
			x++
		}
	}
}

func setVoffsetAndCoffset(text []byte, boffset int) (voffset, coffset int) {
	text = text[:boffset]
	for len(text) > 0 {
		r, size := utf8.DecodeRune(text)
		text = text[size:]
		coffset++
		voffset += runeAdvanceLen(r, voffset)
	}
	return
}

const (
	// tabstopLength is a TAB character length
	tabstopLength               = 8
	preferredHorizonalThreshold = 5
)

func runeAdvanceLen(r rune, pos int) int {
	if r == '\t' {
		return tabstopLength - pos%tabstopLength
	}
	// Assume r is 1 character. So 10 length is enough.
	b := make([]byte, 10)
	if size := utf8.EncodeRune(b, r); size > 1 {
		// For multibyte character, use 2 cell.
		return 2
	}

	return 1
}

// EditBox position
const ()

// EditBox is a user input part.
type EditBox struct {
	text        []byte
	lineVoffset int

	// cursorBoffset is an offset according to bytes.
	cursorBoffset int

	// cursorVoffset is an offset according to visual.
	// When user input long sentence, the sentence moves
	// left side to show current [position.
	cursorVoffset int

	// cursorCoffset is an offset accoding to unicode code points.
	cursorCoffset int
}

// Draw EditBox part on screen.
func (eb *EditBox) Draw() {
	setCellLine(0, 0,
		termbox.ColorDefault, termbox.ColorDefault,
		string(eb.text[:]))

	setCellLine(0, 1,
		termbox.ColorDefault, termbox.ColorDefault,
		"------------------------------------------")
	// Highlight cursor position.
	termbox.SetCursor(eb.cursorVoffset, 0)
	//termbox.Flush()
}

// MoveCursorTo move cursor position by given offset
func (eb *EditBox) MoveCursorTo(offset int) {
	eb.cursorBoffset = offset
	eb.cursorVoffset, eb.cursorCoffset = setVoffsetAndCoffset(eb.text, offset)
}

// MoveCursorOneRuneForward move cursor by 1 rune.
func (eb *EditBox) MoveCursorOneRuneForward() {
	if eb.cursorBoffset == len(eb.text) {
		return
	}
	_, size := eb.RuneUnderCursor()
	eb.MoveCursorTo(eb.cursorBoffset + size)
}

// MoveCursorOneRuneBackward move cursor by 1 rune to left side.
func (eb *EditBox) MoveCursorOneRuneBackward() {
	if eb.cursorBoffset == 0 {
		return
	}
	_, size := eb.RuneBeforeCursor()
	eb.MoveCursorTo(eb.cursorBoffset - size)
}

// RuneBeforeCursor return the previous rune's size and rune itself from boffset.
func (eb *EditBox) RuneBeforeCursor() (rune, int) {
	return utf8.DecodeLastRune(eb.text[:eb.cursorBoffset])
}

// RuneUnderCursor return the rune
func (eb *EditBox) RuneUnderCursor() (rune, int) {
	return utf8.DecodeRune(eb.text[eb.cursorBoffset:])
}

// InsertRune insert character to EditBox.text
func (eb *EditBox) InsertRune(r rune) {
	var buf [utf8.UTFMax]byte
	n := utf8.EncodeRune(buf[:], r)
	eb.text = byteSliceInsert(eb.text, eb.cursorBoffset, buf[:n])
	eb.MoveCursorOneRuneForward()
}

// DeleteRuneBackward delete previous character from boffset position
func (eb *EditBox) DeleteRuneBackward() {
	if eb.cursorBoffset == 0 {
		return
	}
	eb.MoveCursorOneRuneBackward()
	_, size := eb.RuneUnderCursor()
	eb.text = byteSliceRemove(eb.text, eb.cursorBoffset, eb.cursorBoffset+size)
}

// AdjustVOffset adjusts line visual offset to a proper value depending on width
func (eb *EditBox) AdjustVOffset(witdh int) {
	// TODO not implemented yet
}

// GetAndDeleteText return current text and delete all the text
func (eb *EditBox) GetAndDeleteText() []byte {
	text := eb.text
	eb.cursorBoffset = 0
	eb.cursorVoffset = 0
	eb.cursorCoffset = 0
	eb.text = []byte("")
	return text

}

func byteSliceResize(s []byte, desiredCap int) []byte {
	if cap(s) < desiredCap {
		ns := make([]byte, len(s), desiredCap)
		copy(ns, s)
		return ns
	}
	return s
}

func byteSliceInsert(text []byte, offset int, what []byte) []byte {
	n := len(text) + len(what)
	text = byteSliceResize(text, n)
	text = text[:n]
	copy(text[offset+len(what):], text[offset:])
	copy(text[offset:], what)
	return text
}

func byteSliceRemove(text []byte, from, to int) []byte {
	size := to - from
	copy(text[from:], text[to:])
	text = text[:len(text)-size]
	return text
}

// Input wait and hundle keyboard input.
// This function interrupt process.
func Input(done <-chan struct{}) <-chan termbox.Event {
	if !termbox.IsInit {
		// TODO use logrus
		fmt.Errorf("ERROR: termbox is not initialized\n")
		os.Exit(1)
	}
	termbox.SetInputMode(termbox.InputEsc)
	out := make(chan termbox.Event)

	go func() {
		defer close(out)
		for {

			select {
			case <-done:
				// Terminaite this goroutine.
				return
			default:
				out <- termbox.PollEvent()
			}
		}
	}()
	return out
}

func redraw() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	//x, y := termbox.Size()
	ch := []rune("a")
	termbox.SetCell(0, 0, ch[0], termbox.ColorDefault, termbox.ColorDefault)
	termbox.Flush()
}
