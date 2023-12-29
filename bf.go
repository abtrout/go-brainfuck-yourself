package gbfy

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

type Brainfuck struct {
	cells [3e4]byte // cells
	d     int       // data pointer for cells access
	cmds  []byte    // commands evaluated by the interpreter
	i     int       // instruction pointer for cmds access

	in  io.Reader
	out io.Writer

	loops    map[int]int // stores index for matching [ or ]
	parLoops []int       // indices for not-yet-complete loops
}

// New returns a new Brainfuck interpreter.
func New(in io.Reader, out io.Writer) *Brainfuck {
	return &Brainfuck{in: in, out: out, loops: map[int]int{}}
}

// Eval evaluates a single command with the given interpreter.
func (bf *Brainfuck) Eval(cmd byte) error {
	switch cmd {
	case '>', '<', '+', '-', '.', ',', '[', ']':
		bf.cmds = append(bf.cmds, cmd)

		// TODO: Move this special loop handling to internal .eval?
		j := len(bf.cmds) - 1
		// Special handling for loops.
		if cmd == '[' {
			bf.parLoops = append(bf.parLoops, j)
		}
		if cmd == ']' {
			if len(bf.parLoops) == 0 {
				return errors.New("invalid loop close")
			}
			k := bf.parLoops[len(bf.parLoops)-1]
			bf.loops[j] = k
			bf.loops[k] = j
			bf.parLoops = bf.parLoops[:len(bf.parLoops)-1]
		}
		return bf.eval()
	default:
		return nil
	}
}

func (bf *Brainfuck) eval() error {
	if len(bf.parLoops) > 0 {
		// Delay evaluation if there are unclosed loops.
		return nil
	}

	for bf.i < len(bf.cmds) {
		switch bf.cmds[bf.i] {
		case '>':
			bf.d++
			if bf.d >= len(bf.cells) {
				bf.d -= len(bf.cells)
			}
		case '<':
			bf.d--
			if bf.d < 0 {
				bf.d += len(bf.cells)
			}
		case '+':
			bf.cells[bf.d]++
		case '-':
			bf.cells[bf.d]--
		case '.':
			if _, err := bf.out.Write([]byte{bf.cells[bf.d]}); err != nil {
				return fmt.Errorf("failed to Write output: %v", err)
			}
		case ',':
			input := make([]byte, 1)
			if _, err := bf.in.Read(input); err != nil {
				return fmt.Errorf("failed to Read input: %v", err)
			} else {
				bf.cells[bf.d] = input[0]
			}
		case '[':
			if bf.cells[bf.d] == 0 {
				bf.i = bf.loops[bf.i]
			}
		case ']':
			if bf.cells[bf.d] != 0 {
				bf.i = bf.loops[bf.i]
			}
		}
		bf.i++
	}
	return nil
}

// Dump interpreter state to caller.
func (bf *Brainfuck) Dump() (int, []byte, int, []byte) {
	return bf.d, bf.cells[:], bf.i, bf.cmds
}

// Reset interpreter state.
func (bf *Brainfuck) Reset() {
	bf.d = 0
	bf.cells = [len(bf.cells)]byte{}

	bf.i = 0
	bf.cmds = nil

	bf.loops = map[int]int{}
	bf.parLoops = nil
}

// Run runs a Brainfuck program and returns output bytes or error.
func Run(cmds []byte, in io.Reader) ([]byte, error) {
	var out bytes.Buffer
	bf := New(in, &out)
	for _, cmd := range cmds {
		if err := bf.Eval(cmd); err != nil {
			return nil, err
		}
	}
	return out.Bytes(), nil
}
