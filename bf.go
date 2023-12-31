package gbfy

import (
	"errors"
	"fmt"
	"io"
)

type Brainfuck struct {
	cells    [3e4]byte
	cmds     []byte
	i, d     int         // instruction and data pointer
	in       io.Reader   // for reading with ,
	out      io.Writer   // for writing with .
	loops    map[int]int // stores index for matching [ or ]
	parLoops []int       // indices for not-yet-complete loops
}

// New returns a new Brainfuck interpreter.
func New(in io.Reader, out io.Writer) *Brainfuck {
	return &Brainfuck{in: in, out: out, loops: map[int]int{}}
}

// Eval a single Brainfuck command.
func (bf *Brainfuck) Eval(cmd byte) error {
	switch cmd {
	case '>', '<', '+', '-', '.', ',', '[', ']':
		bf.cmds = append(bf.cmds, cmd)
	default:
		return nil
	}
	// Handle loops before calling internal eval.
	if cmd == '[' {
		bf.parLoops = append(bf.parLoops, len(bf.cmds)-1)
	} else if cmd == ']' {
		if len(bf.parLoops) == 0 {
			return errors.New("invalid loop close")
		}
		i, j := len(bf.cmds)-1, bf.parLoops[len(bf.parLoops)-1]
		bf.loops[i] = j
		bf.loops[j] = i
		bf.parLoops = bf.parLoops[:len(bf.parLoops)-1]
	}
	if len(bf.parLoops) > 0 {
		return nil // delay eval if there are partial loops.
	}
	return bf.eval()
}

func (bf *Brainfuck) eval() error {
	if n := len(bf.cmds); n == 0 || n <= bf.i {
		return nil
	}

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
		}
		bf.cells[bf.d] = input[0]
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
	return bf.eval()
}

// Dump interpreter state to caller.
func (bf *Brainfuck) Dump() (int, []byte, int, []byte) {
	return bf.d, bf.cells[:], bf.i, bf.cmds
}
