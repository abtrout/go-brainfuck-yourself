package gbfy

import (
	"bytes"
	"errors"
	"fmt"
)

// Brainfuck is the interpreter state.
type Brainfuck struct {
	cells   [3e4]byte     // cells/tape
	cmds    []byte        // program being executed
	d, i    int           // data and instruction pointers
	in, out *bytes.Buffer // input and output buffers
}

// New constructs a new Brainfuck interpreter with the given
// program and input buffer.
func New(program string, input []byte) *Brainfuck {
	return &Brainfuck{
		cmds: []byte(program),
		in:   bytes.NewBuffer(input),
		out:  new(bytes.Buffer),
	}
}

// Run the program loaded into the interpreter.
func (bf *Brainfuck) Run() ([]byte, error) {
	for bf.i < len(bf.cmds) {
		if err := bf.eval(bf.cmds[bf.i]); err != nil {
			return nil, fmt.Errorf("Run failed with error: %v", err)
		}
		bf.i++
	}
	return bf.out.Bytes(), nil
}

// eval evaluates a single Brainfuck command.
func (bf *Brainfuck) eval(cmd byte) error {
	switch cmd {
	case '>':
		// Increment data pointer.
		bf.d++
		if bf.d >= len(bf.cells) {
			bf.d -= len(bf.cells)
		}
	case '<':
		// Decrement data pointer.
		bf.d--
		if bf.d < 0 {
			bf.d += len(bf.cells)
		}
	case '+':
		// Increment value at current cell.
		bf.cells[bf.d]++
	case '-':
		// Decrement value at current cell.
		bf.cells[bf.d]--
	case '.':
		// Write current cell's value to output buffer.
		bf.out.WriteByte(bf.cells[bf.d])
	case ',':
		// Read value from input and store in current cell.
		if b, err := bf.in.ReadByte(); err != nil {
			return errors.New("Program expects more input!")
		} else {
			bf.cells[bf.d] = b
		}
	case '[':
		// Loop start. Continue through loop body if current cell
		// is non-zero, otehrwise jump to matching ].
		if bf.cells[bf.d] == 0 {
			idx, err := findMatchingClose(bf.cmds, bf.i)
			if err != nil {
				return fmt.Errorf("Invalid program! %v", err)
			}
			bf.i = idx
		}
	case ']':
		// Loop end. Jump to matching [ if current cell is non-zero.
		if bf.cells[bf.d] != 0 {
			idx, err := findMatchingOpen(bf.cmds, bf.i)
			if err != nil {
				return fmt.Errorf("Invalid program! %v", err)
			}
			bf.i = idx
		}
	default:
		// TODO: Consider stripping non-brainfuck characters in pre-processing.
		// Technically they are valid in BF programs and should be treated as comments.
		return fmt.Errorf("Invalid program! Unknown command %q", cmd)
	}
	return nil
}

func findMatchingClose(cmds []byte, i int) (int, error) {
	var opens int
	for i < len(cmds) {
		switch cmds[i] {
		case '[':
			opens++
		case ']':
			if opens == 0 {
				return i, nil
			}
			opens--
		}
		i++
	}
	return 0, fmt.Errorf("Missing matching ] for [ at %d", i)
}

func findMatchingOpen(cmds []byte, i int) (int, error) {
	var closes int
	for i >= 0 {
		switch cmds[i] {
		case '[':
			if closes == 0 {
				return i, nil
			}
			closes--
		case ']':
			closes++
		}
		i--
	}
	return 0, fmt.Errorf("Missing matching [ for ] at %d", i)
}
