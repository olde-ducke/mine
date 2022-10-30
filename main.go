package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"syscall"
	"time"
	"unicode/utf8"

	"golang.org/x/term"
)

var (
	gameOverMessage = " G A M E   O V E R "
	winMessage      = " Y O U   W O N "
	bombPercentage  = 15
	width           = 10
	height          = 10
	seed            = int64(0)
)

func init() {
	flag.StringVar(&gameOverMessage, "game-over-message", gameOverMessage, "sets game over message")
	flag.StringVar(&gameOverMessage, "g", gameOverMessage, "")

	flag.StringVar(&winMessage, "win-message", winMessage, "sets win message")
	flag.StringVar(&winMessage, "p", winMessage, "")

	flag.IntVar(&bombPercentage, "bomb-percentage", bombPercentage, "sets bomb percentage")
	flag.IntVar(&bombPercentage, "b", bombPercentage, "")

	flag.IntVar(&width, "width", width, "game field width")
	flag.IntVar(&width, "w", width, "")

	flag.IntVar(&height, "height", height, "game field height")
	flag.IntVar(&height, "h", height, "")

	flag.Int64Var(&seed, "seed", seed, "game seed, will use unix-nano if 0")
	flag.Int64Var(&seed, "s", seed, "")

	var shorthands = map[string]string{
		"game-over-message": "-g,",
		"win-message":       "-p,",
		"bomb-percentage":   "-b,",
		"width":             "-w,",
		"height":            "-h,",
		"seed":              "-s,",
	}

	format := "  %-3s --%-19s %s %s\n"
	flag.Usage = func() {
		fmt.Println("simple minesweeper for terminal\nusage:\n$ mine [arguments]")
		fmt.Printf(format, "-h,", "help", "show this message and exit", "")
		flag.VisitAll(func(fn *flag.Flag) {
			if fn.Usage == "" {
				return
			}

			var defVal string
			if fn.DefValue != "" && fn.DefValue != "false" {
				defVal = fmt.Sprintf(" (default: %q)", fn.DefValue)
			}

			fmt.Printf(format, shorthands[fn.Name], fn.Name, fn.Usage, defVal)
		})
	}

	flag.Parse()
}

var keys = map[string][]byte{
	"esc":   []byte{27, 0, 0, 0, 0},
	"q":     []byte{113, 0, 0, 0, 0},
	"Q":     []byte{81, 0, 0, 0, 0},
	"w":     []byte{119, 0, 0, 0, 0},
	"W":     []byte{87, 0, 0, 0, 0},
	"s":     []byte{115, 0, 0, 0, 0},
	"S":     []byte{83, 0, 0, 0, 0},
	"a":     []byte{97, 0, 0, 0, 0},
	"A":     []byte{65, 0, 0, 0, 0},
	"d":     []byte{100, 0, 0, 0, 0},
	"D":     []byte{68, 0, 0, 0, 0},
	"f":     []byte{102, 0, 0, 0, 0},
	"F":     []byte{70, 0, 0, 0, 0},
	"r":     []byte{114, 0, 0, 0, 0},
	"R":     []byte{82, 0, 0, 0, 0},
	"y":     []byte{121, 0, 0, 0, 0},
	"Y":     []byte{89, 0, 0, 0, 0},
	"n":     []byte{110, 0, 0, 0, 0},
	"N":     []byte{78, 0, 0, 0, 0},
	"up":    []byte{27, 91, 65, 0, 0},
	"down":  []byte{27, 91, 66, 0, 0},
	"left":  []byte{27, 91, 68, 0, 0},
	"right": []byte{27, 91, 67, 0, 0},
	"space": []byte{32, 0, 0, 0, 0},
	"enter": []byte{13, 0, 0, 0, 0},
}

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
	cells     [][]cell
	states    [][]state
	generated bool
	rows      int
	cols      int
	cursorRow int
	cursorCol int
}

func (f *field) inBounds(row, col int) bool {
	return 0 <= row && row < f.rows && 0 <= col && col < f.cols
}

func (f *field) reset(rows, cols int) error {
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

	f.cells = make([][]cell, rows)
	for i := range f.cells {
		f.cells[i] = make([]cell, cols)
	}
	f.states = make([][]state, rows)
	for i := range f.states {
		f.states[i] = make([]state, cols)
	}

	f.rows = rows
	f.cols = cols
	f.generated = false
	f.cursorRow, f.cursorCol = 0, 0
	return nil
}

func (f *field) countBombs(row, col int) int {
	var count int
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx != 0 || dy != 0 {
				y, x := row+dy, col+dx
				if f.inBounds(y, x) && f.cells[y][x] == bomb {
					count++
				}
			}
		}
	}

	return count
}

func (f *field) flagAtCursor() {
	switch f.states[f.cursorRow][f.cursorCol] {
	case closed:
		f.states[f.cursorRow][f.cursorCol] = flagged
	case flagged:
		f.states[f.cursorRow][f.cursorCol] = closed
	}
}

func (f *field) atCursor(row, col int) bool {
	return f.cursorRow == row && f.cursorCol == col
}

func (f *field) aroundCursor(row, col int) bool {
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if f.cursorRow+dy == row && f.cursorCol+dx == col {
				return true
			}
		}
	}

	return false
}

func (f *field) display() {
	for row := 0; row < f.rows; row++ {
		for col := 0; col < f.cols; col++ {
			if f.atCursor(row, col) {
				fmt.Print("[")
			} else {
				fmt.Print(" ")
			}
			switch f.states[row][col] {
			case opened:
				switch f.cells[row][col] {
				case bomb:
					fmt.Print("@")
				case empty:
					nbors := f.countBombs(row, col)
					if nbors > 0 {
						fmt.Print(nbors)
					} else {
						fmt.Print(" ")
					}
				}
			case closed:
				fmt.Print(".")
			case flagged:
				fmt.Print("%")
			}

			if f.atCursor(row, col) {
				fmt.Print("]")
			} else {
				fmt.Print(" ")
			}
		}
		fmt.Print("\n\r")
	}
}

func (f *field) randomCell() (int, int) {
	row, col := rand.Intn(f.rows), rand.Intn(f.cols)
	return row, col
}

func (f *field) randomize(bombPercentage int) {
	for i := 0; i < f.rows; i++ {
		for j := 0; j < f.cols; j++ {
			f.cells[i][j] = empty
		}
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
		row, col := f.randomCell()
		if f.cells[row][col] == bomb || f.aroundCursor(row, col) {
			continue
		}
		f.cells[row][col] = bomb
		i++
	}
}

func (f *field) openAt(row, col int) bool {
	if !f.generated {
		f.randomize(bombPercentage)
		f.generated = true
	}
	f.states[row][col] = opened

	if f.countBombs(row, col) == 0 {
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				y, x := row+dy, col+dx
				if f.inBounds(y, x) {
					if f.states[y][x] == closed && f.states[y][x] != flagged {
						f.openAt(y, x)
					}
				}
			}
		}
	}

	return f.cells[row][col] == bomb
}

func (f *field) openAtCursor() bool {
	return f.openAt(f.cursorRow, f.cursorCol)
}

func (f *field) openBombs() {
	for i := 0; i < f.rows; i++ {
		for j := 0; j < f.cols; j++ {
			if f.cells[i][j] == bomb {
				f.states[i][j] = opened
			}
		}
	}
}

func (f *field) openNonBombs() {
	for i := 0; i < f.rows; i++ {
		for j := 0; j < f.cols; j++ {
			if f.cells[i][j] != bomb {
				f.states[i][j] = opened
			} else {
				f.states[i][j] = flagged
			}
		}
	}
}

func (f *field) victory() bool {
	for i := 0; i < f.rows; i++ {
		for j := 0; j < f.cols; j++ {
			switch f.states[i][j] {
			case opened:
				if f.cells[i][j] != empty {
					return false
				}

			case flagged, closed:
				if f.cells[i][j] != bomb {
					return false
				}
			}
		}
	}
	return true
}

func (f *field) render() {
	fmt.Print("\x1b[", f.rows, "A")
	fmt.Print("\x1b[", f.cols*3, "D")
	f.display()
}

func (f *field) printMessage(message string) {
	length := utf8.RuneCountInString(message)
	if length > f.cols*3 {
		message = message[:f.cols*3]
		length = utf8.RuneCountInString(message)
	}
	fmt.Print("\x1b[s")
	fmt.Print("\x1b[", f.rows/2+1, "A")
	fmt.Print("\x1b[", f.cols*3/2-length/2, "C")
	fmt.Print("\x1b[7m", message, "\x1b[0m")
	fmt.Print("\x1b[u")
}

func (f *field) getConfirmation(message string) bool {
	f.printMessage(message)
	for {
		buf := make([]byte, 5)
		os.Stdin.Read(buf)

		switch {
		case isAKey(buf, "y"), isAKey(buf, "Y"):
			return true
		case isAKey(buf, "n"), isAKey(buf, "N"):
			return false
		}
	}
}

func (f *field) finishGame(message string) bool {
	f.render()
	time.Sleep(time.Second)
	f.printMessage(message)
	time.Sleep(time.Second)
	f.render()
	if ok := f.getConfirmation("restart? [y/n]"); ok {
		if err := f.reset(width, height); err != nil {
			fmt.Println(err)
			return false
		}
		return true
	}

	return false
}

func (f *field) moveUp() {
	if f.cursorRow > 0 {
		f.cursorRow--
	}
}

func (f *field) moveDown() {
	if f.cursorRow < f.rows-1 {
		f.cursorRow++
	}
}

func (f *field) moveLeft() {
	if f.cursorCol > 0 {
		f.cursorCol--
	}
}

func (f *field) moveRight() {
	if f.cursorCol < f.cols-1 {
		f.cursorCol++
	}
}

func setTerminal() (*term.State, error) {
	fd := int(syscall.Stdin)
	if !term.IsTerminal(fd) {
		return nil, errors.New("can only draw in terminal")
	}

	prev, err := term.GetState(fd)
	if err != nil {
		return nil, err
	}

	_, err = term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}

	fmt.Print("\x1b[?25l")
	return prev, nil
}

func restoreTerminal(state *term.State) error {
	fmt.Print("\x1b[?25h")
	return term.Restore(int(syscall.Stdin), state)
}

func isAKey(buf []byte, key string) bool {
	return bytes.Equal(buf, keys[key])
}

func main() {
	var mainField field

	if err := mainField.reset(width, height); err != nil {
		fmt.Println(err)
		return
	}

	mainField.display()

	state, err := setTerminal()
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := restoreTerminal(state); err != nil {
			panic(err)
		}
	}()

	if seed == 0 {
		rand.Seed(time.Now().UnixNano())
	} else {
		rand.Seed(seed)
	}

loop:
	for {
		if mainField.victory() {
			mainField.openNonBombs()
			if !mainField.finishGame(winMessage) {
				break loop
			}
		}
		mainField.render()

		buf := make([]byte, 5)
		os.Stdin.Read(buf)

		switch {
		case isAKey(buf, "esc"), isAKey(buf, "q"), isAKey(buf, "Q"):
			if mainField.getConfirmation("are you sure? [y/n]") {
				break loop
			}

		case isAKey(buf, "up"), isAKey(buf, "w"), isAKey(buf, "W"):
			mainField.moveUp()

		case isAKey(buf, "down"), isAKey(buf, "s"), isAKey(buf, "S"):
			mainField.moveDown()

		case isAKey(buf, "left"), isAKey(buf, "a"), isAKey(buf, "A"):
			mainField.moveLeft()

		case isAKey(buf, "right"), isAKey(buf, "d"), isAKey(buf, "D"):
			mainField.moveRight()

		case isAKey(buf, "r"), isAKey(buf, "R"):
			if mainField.getConfirmation("restart? [y/n]") {
				if err := mainField.reset(width, height); err != nil {
					fmt.Println(err)
					return
				}
			}

		case isAKey(buf, "enter"), isAKey(buf, "f"), isAKey(buf, "F"):
			mainField.flagAtCursor()

		case isAKey(buf, "space"):
			if mainField.openAtCursor() {
				mainField.openBombs()
				if !mainField.finishGame(gameOverMessage) {
					break loop
				}
			}
		}
	}
}
