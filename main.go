package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"syscall"
	"time"

	"golang.org/x/term"
)

var keys = map[string][]byte{
	"esc":   []byte{27, 0, 0, 0, 0},
	"q":     []byte{113, 0, 0, 0, 0},
	"w":     []byte{119, 0, 0, 0, 0},
	"s":     []byte{115, 0, 0, 0, 0},
	"a":     []byte{97, 0, 0, 0, 0},
	"d":     []byte{100, 0, 0, 0, 0},
	"f":     []byte{102, 0, 0, 0, 0},
	"r":     []byte{114, 0, 0, 0, 0},
	"up":    []byte{27, 91, 65, 0, 0},
	"down":  []byte{27, 91, 66, 0, 0},
	"left":  []byte{27, 91, 68, 0, 0},
	"right": []byte{27, 91, 67, 0, 0},
	"space": []byte{32, 0, 0, 0, 0},
	"enter": []byte{13, 0, 0, 0, 0},
}

var (
	gameOverMessage = "G A M E   O V E R"
	winMessage      = "Y O U   W I N"
	bombPercentage  = 1
	fieldWith       = 1000
	fieldHeight     = 1000
	seed            = time.Now().UnixNano()
)

type cell int

const (
	empty cell = iota
	bomb
)

type state int

const (
	closed state = iota
	opened
	flagged
)

type field struct {
	cells     []cell
	states    []state
	rows      int
	cols      int
	cursorRow int
	cursorCol int
}

func (f *field) fieldGet(row, col int) cell {
	return f.cells[row*f.cols+col]
}

func (f *field) fieldSet(row, col int, cell cell) {
	if f.fieldInBounds(row, col) {
		f.cells[row*f.cols+col] = cell
	}
}

func (f *field) fieldInBounds(row, col int) bool {
	return 0 <= row && row < f.rows && 0 <= col && col < f.cols
}

func (f *field) fieldCheckedGet(row, col int) (cell, bool) {
	if f.fieldInBounds(row, col) {
		return f.fieldGet(row, col), true
	}

	return 0, false
}

func (f *field) fieldGetState(row, col int) state {
	return f.states[row*f.cols+col]
}

func (f *field) fieldResize(rows, cols int) error {
	w, h, err := term.GetSize(int(syscall.Stdin))
	if err != nil {
		return err
	}

	if rows < 7 {
		rows = 7
	}

	if rows > h {
		rows = h - 1
	}

	if cols < 7 {
		cols = 7
	}

	if cols > w/3 {
		cols = w / 3
	}

	f.cells = make([]cell, rows*cols)
	f.states = make([]state, rows*cols)
	f.rows = rows
	f.cols = cols
	f.cursorRow, f.cursorCol = 0, 0
	return nil
}

func (f *field) fieldCountNbors(row, col int) int {
	var count int
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx != 0 || dy != 0 {
				cell, ok := f.fieldCheckedGet(row+dy, col+dx)
				if ok && cell == bomb {
					count++
				}
			}
		}
	}

	return count
}

func (f *field) fieldAtCursor(row, col int) bool {
	return f.cursorRow == row && f.cursorCol == col
}

func (f *field) fieldAroundCursor(row, col int) bool {
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if f.cursorRow+dy == row && f.cursorCol+dx == col {
				return true
			}
		}
	}

	return false
}

func (f *field) fieldPrint() {
	for row := 0; row < f.rows; row++ {
		for col := 0; col < f.cols; col++ {
			if f.fieldAtCursor(row, col) {
				fmt.Print("[")
			} else {
				fmt.Print(" ")
			}
			switch state := f.fieldGetState(row, col); state {
			case opened:
				switch f.fieldGet(row, col) {
				case bomb:
					fmt.Print("@")
				case empty:
					nbors := f.fieldCountNbors(row, col)
					if nbors > 0 {
						fmt.Print(nbors)
					} else {
						fmt.Print(" ")
					}
				}
			case closed:
				fmt.Print(".")
			case flagged:
				fmt.Print("P")
			}

			if f.fieldAtCursor(row, col) {
				fmt.Print("]")
			} else {
				fmt.Print(" ")
			}
		}
		fmt.Print("\n\r")
	}
}

func (f *field) fieldRandomCell() (int, int, cell) {
	row, col := rand.Intn(f.rows), rand.Intn(f.cols)
	return row, col, f.fieldGet(row, col)
}

func (f *field) fieldRandomize(bombPercentage int) {
	for i := 0; i < len(f.cells); i++ {
		f.cells[i] = empty
		f.states[i] = closed
	}
	if bombPercentage <= 1 {
		bombPercentage = 1
	}
	if bombPercentage > 80 {
		bombPercentage = 80
	}
	bombCount := (f.rows*f.cols*bombPercentage + 99) / 100
	i := 0
	for i < bombCount {
		row, col, cell := f.fieldRandomCell()
		if cell == bomb || f.fieldAroundCursor(row, col) {
			fmt.Print("skip", i, "\n\r")
			continue
		}
		f.fieldSet(row, col, bomb)
		i++
	}
}

func (f *field) fieldOpenAtCursor() cell {
	i := f.cursorRow*f.cols + f.cursorCol
	f.states[i] = opened
	return f.cells[i]
}

func (f *field) fieldFlagAtCursor() {
	i := f.cursorRow*f.cols + f.cursorCol
	switch f.states[i] {
	case closed:
		f.states[i] = flagged
	case flagged:
		f.states[i] = closed
	}
}

func (f *field) fieldOpenBombs() {
	for i := 0; i < len(f.states); i++ {
		if f.cells[i] == bomb {
			f.states[i] = opened
		}
	}
}

func (f *field) render() {
	fmt.Printf("\x1b[%dA", f.rows)
	fmt.Printf("\x1b[%dD", f.cols*3)
	f.fieldPrint()
}

func setTerminal() (*term.State, error) {
	fd := int(syscall.Stdin)
	prev, err := term.GetState(fd)
	if err != nil {
		return nil, err
	}

	_, err = term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}

	return prev, nil
}

func isAKey(buf []byte, key string) bool {
	return bytes.Compare(buf, keys[key]) == 0
}

func main() {
	var (
		notFirst  bool
		mainField field
	)
	if err := mainField.fieldResize(fieldWith, fieldHeight); err != nil {
		fmt.Println(err)
		return
	}

	mainField.fieldPrint()

	state, err := setTerminal()
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := term.Restore(int(syscall.Stdin), state); err != nil {
			panic(err)
		}
	}()

	rand.Seed(seed)

loop:
	for {
		buf := make([]byte, 5)
		os.Stdin.Read(buf)

		switch {
		case isAKey(buf, "esc"), isAKey(buf, "q"):
			break loop
		case isAKey(buf, "up"), isAKey(buf, "w"):
			if mainField.cursorRow > 0 {
				mainField.cursorRow--
			}
		case isAKey(buf, "down"), isAKey(buf, "s"):
			if mainField.cursorRow < mainField.rows-1 {
				mainField.cursorRow++
			}
		case isAKey(buf, "left"), isAKey(buf, "a"):
			if mainField.cursorCol > 0 {
				mainField.cursorCol--
			}
		case isAKey(buf, "right"), isAKey(buf, "d"):
			if mainField.cursorCol < mainField.cols-1 {
				mainField.cursorCol++
			}
		case isAKey(buf, "enter"), isAKey(buf, "f"):
			mainField.fieldFlagAtCursor()

		case isAKey(buf, "space"):
			if !notFirst {
				mainField.fieldRandomize(bombPercentage)
				notFirst = true
			}
			if cell := mainField.fieldOpenAtCursor(); cell == bomb {
				mainField.fieldOpenBombs()
				mainField.render()
				time.Sleep(time.Second)
				fmt.Print(gameOverMessage, "\n\r")
				break loop
			}

		}

		mainField.render()
	}

}
