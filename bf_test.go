package gbfy

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestEval(t *testing.T) {
	tests := []struct {
		// Command to evaluate.
		cmd byte
		// Expected value of instruction and Data pointer.
		i, d int
		// Expected values of *specific* cells; mapping between cell
		// index and value expected there. Since the cell region is
		// large (3e4) and sparse (0 by default) only specific cells
		// are checked.
		cells map[int]byte
		// Expected output data bytes.
		out []byte
	}{
		// Check < and > move data pointer around circular cell region.
		{'<', 1, 29999, nil, nil},
		{'>', 2, 0, nil, nil},
		// Check that + and - modify cell values.
		{'+', 3, 0, map[int]byte{0: 1}, nil},
		{'-', 4, 0, map[int]byte{0: 0}, nil},
		// Check that , and . read input and write output bytes.
		{',', 5, 0, map[int]byte{0: 4}, nil},
		{'>', 6, 1, map[int]byte{0: 4}, nil},
		{',', 7, 1, map[int]byte{0: 4, 1: 8}, nil},
		{',', 8, 1, map[int]byte{0: 4, 1: 15}, nil},
		{'>', 9, 2, map[int]byte{0: 4, 1: 15}, nil},
		{',', 10, 2, map[int]byte{0: 4, 1: 15, 2: 16}, nil},
		{'.', 11, 2, map[int]byte{0: 4, 1: 15, 2: 16}, []byte{16}},
		{'<', 12, 1, map[int]byte{0: 4, 1: 15, 2: 16}, []byte{16}},
		{'.', 13, 1, map[int]byte{0: 4, 1: 15, 2: 16}, []byte{16, 15}},
		{'<', 14, 0, map[int]byte{0: 4, 1: 15, 2: 16}, []byte{16, 15}},
		{'.', 15, 0, map[int]byte{0: 4, 1: 15, 2: 16}, []byte{16, 15, 4}},
		// Check that [ and ] are handled correctly by adding two adjacent cells.
		// Evaluation is delayed in the presence of partial loops, so all internal
		// state remain the same until the final ] is Eval'd.
		{'[', 15, 0, map[int]byte{0: 4, 1: 15, 2: 16}, []byte{16, 15, 4}},
		{'-', 15, 0, map[int]byte{0: 4, 1: 15, 2: 16}, []byte{16, 15, 4}},
		{'>', 15, 0, map[int]byte{0: 4, 1: 15, 2: 16}, []byte{16, 15, 4}},
		{'+', 15, 0, map[int]byte{0: 4, 1: 15, 2: 16}, []byte{16, 15, 4}},
		{'<', 15, 0, map[int]byte{0: 4, 1: 15, 2: 16}, []byte{16, 15, 4}},
		{']', 21, 0, map[int]byte{0: 0, 1: 19, 2: 16}, []byte{16, 15, 4}},
		{'>', 22, 1, map[int]byte{0: 0, 1: 19, 2: 16}, []byte{16, 15, 4}},
		{'.', 23, 1, map[int]byte{0: 0, 1: 19, 2: 16}, []byte{16, 15, 4, 19}},
	}

	var out bytes.Buffer
	bf := New(bytes.NewBuffer([]byte{4, 8, 15, 16, 23, 41}), &out)
	if err := checkInterpreter(bf, 0, 0, nil); err != nil {
		t.Fatalf("[0] Unexpected interpreter state: %v", err)
	}

	for i, test := range tests {
		if err := bf.Eval(test.cmd); err != nil {
			t.Fatalf("[%d] Eval(%q) failed with error: %v", i, test.cmd, err)
		}
		if err := checkInterpreter(bf, test.i, test.d, test.cells); err != nil {
			t.Fatalf("[%d] Unexpected interpreter state: %v", i, err)
		}
		if diff := cmp.Diff(test.out, out.Bytes()); diff != "" {
			t.Fatalf("[%d] output diff (-want +got): %s", i, diff)
		}
	}
}

func checkInterpreter(bf *Brainfuck, i, d int, cells map[int]byte) error {
	if bf.i != i {
		return fmt.Errorf("instruction pointer; got %d, want %d", bf.i, i)
	}
	if bf.d != d {
		return fmt.Errorf("data pointer; got %d, want %d", bf.d, d)
	}
	for idx, got := range bf.cells {
		if want := cells[idx]; got != want {
			return fmt.Errorf("cell value at index %d; got %d, want %d", idx, got, want)
		}
	}
	return nil
}

func TestInvalidLoopHandling(t *testing.T) {
	tests := []string{"]", "[]]", "[][]]"}
	for _, test := range tests {
		if _, err := Run([]byte(test), nil); err == nil {
			t.Errorf("Parsed invalid program %q; expected error", test)
		}
	}
}

func TestHelloWorld(t *testing.T) {
	program := `
		>++++++++[<+++++++++>-]<.
		>++++[<+++++++>-]<+.
		+++++++..
		+++.
		>>++++++[<+++++++>-]<++.
		------------.
		>++++++[<+++++++++>-]<+.
		<.
		+++.
		------.
		--------.
		>>>++++[<++++++++>-]<+.`

	out, err := Run([]byte(program), nil)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if diff := cmp.Diff("Hello, World!", string(out)); diff != "" {
		t.Errorf("Mismatched output data (-want +got):\n%s", diff)
	}
}

func TestCellWrapping(t *testing.T) {
	bf := New(nil, nil)
	// Move to the right .... 3e4 - 1 times.
	for i := 0; i < len(bf.cells)-1; i++ {
		if err := bf.Eval('>'); err != nil {
			t.Fatalf("Eval(>) failed with error %v", err)
		}
	}
	if err := checkInterpreter(bf, 29999, 29999, nil); err != nil {
		t.Fatalf("Unexpected interpreter state: %v", err)
	}
	// Move to the right once more -- should wrap!
	if err := bf.Eval('>'); err != nil {
		t.Fatalf("Eval(>) failed with error %v", err)
	} else if err := checkInterpreter(bf, 30000, 0, nil); err != nil {
		t.Fatalf("Unexpected interpreter state: %v", err)
	}
	// Move to the left -- should wrap!
	if err := bf.Eval('<'); err != nil {
		t.Fatalf("Eval(<) failed with error %v", err)
	}
	if err := checkInterpreter(bf, 30001, 29999, nil); err != nil {
		t.Fatalf("Unexpected interpreter state: %v", err)
	}
}

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
