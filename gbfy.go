package gbfy

import (
	"bytes"
	"errors"
	"fmt"
)

// Brainfuck is the interpreter state.
type Brainfuck struct {
	cells   [3e4]byte     // cells
	d       int           // data pointer for cells access
	cmds    []byte        // program being executed
	i       int           // instruction pointer for cmds access
	loops   map[int]int   // stores index for matching [ or ]
	in, out *bytes.Buffer // input and output buffers
}

// New constructs a new Brainfuck interpreter with the given program
// and input buffer.
func New(program string, input []byte) (*Brainfuck, error) {
	cmds, loops, err := parseProgram(program)
	if err != nil {
		return nil, err
	}
	return &Brainfuck{
		cmds:  cmds,
		loops: loops,
		in:    bytes.NewBuffer(input),
		out:   new(bytes.Buffer),
	}, nil
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
			return errors.New("program expects more input")
		} else {
			bf.cells[bf.d] = b
		}
	case '[':
		// Loop start; jump to matching ] if current cell is zero.
		if bf.cells[bf.d] == 0 {
			bf.i = bf.loops[bf.i]
		}
	case ']':
		// Loop end; jump to matching [ if current cell is non-zero.
		if bf.cells[bf.d] != 0 {
			bf.i = bf.loops[bf.i]
		}
	}
	return nil
}

func parseProgram(program string) ([]byte, map[int]int, error) {
	cmds, opens, loops := []byte{}, []int{}, map[int]int{}
	for _, cmd := range program {
		switch cmd {
		case '[', ']', '<', '>', '+', '-', ',', '.':
			cmds = append(cmds, byte(cmd))
		}
	}
	for i, cmd := range cmds {
		if cmd == '[' {
			opens = append(opens, i)
		} else if cmd == ']' {
			if len(opens) == 0 {
				return nil, nil, fmt.Errorf("mismatched [ at index %d", i)
			}
			// Remove matching [.
			j := opens[len(opens)-1]
			opens = opens[:len(opens)-1]
			// Store mapping.
			loops[i] = j
			loops[j] = i
		}
	}
	if len(opens) > 0 {
		return nil, nil, errors.New("unclosed [ at end of program")
	}
	return cmds, loops, nil
}
