package gbfy

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestEval(t *testing.T) {
	tests := []struct {
		// Command to evaluate.
		cmd byte
		// Expected value of Data pointer.
		d int
		// Expected values of *specific* cells; mapping between cell
		// index and value expected there. Since the cell region is
		// large (3e4) and sparse (0 by default) only specific cells
		// are checked.
		cells map[int]byte
		// Expected output data.
		out []byte
	}{
		// Check < and > move data pointer around circular cell region.
		{'<', 29999, nil, nil},
		{'>', 0, nil, nil},
		// Check that + and - modify cell values.
		{'+', 0, map[int]byte{0: 1}, nil},
		{'-', 0, map[int]byte{0: 0}, nil},
		// Check that , and . read input and write output bytes.
		// NB: output is inspected at the very end since the interpreter
		// uses bytes.Buffer.
		{',', 0, map[int]byte{0: 4}, nil},
		{'>', 1, map[int]byte{0: 4}, nil},
		{',', 1, map[int]byte{0: 4, 1: 8}, nil},
		{',', 1, map[int]byte{0: 4, 1: 15}, nil},
		{'>', 2, map[int]byte{0: 4, 1: 15}, nil},
		{',', 2, map[int]byte{0: 4, 1: 15, 2: 16}, nil},
		{'.', 2, map[int]byte{0: 4, 1: 15, 2: 16}, []byte{16}},
		{'<', 1, map[int]byte{0: 4, 1: 15, 2: 16}, []byte{16}},
		{'.', 1, map[int]byte{0: 4, 1: 15, 2: 16}, []byte{16, 15}},
		{'<', 0, map[int]byte{0: 4, 1: 15, 2: 16}, []byte{16, 15}},
		{'.', 0, map[int]byte{0: 4, 1: 15, 2: 16}, []byte{16, 15, 4}},
		// Hmm, we can't really test [ and ] with eval ...
	}

	bf, _ := New("", []byte{4, 8, 15, 16, 23, 41})
	if err := checkInterpreter(bf, 0, 0, nil, nil); err != nil {
		t.Fatalf("[0] Unexpected interpreter state: %v", err)
	}

	for i, test := range tests {
		if err := bf.eval(test.cmd); err != nil {
			t.Fatalf("[%d] Eval(%q) failed with error: %v", i, test.cmd, err)
		}
		// Note since we called `eval` ourselves rather than `Run`, the instruction
		// pointer is not being used, i.e. is always zero. It's tested separately below.
		if err := checkInterpreter(bf, 0, test.d, test.cells, test.out); err != nil {
			t.Errorf("[%d] Unexpected interpreter state: %v", i, err)
		}
	}
}

func checkInterpreter(bf *Brainfuck, i, d int, cells map[int]byte, out []byte) error {
	if bf.i != i {
		return fmt.Errorf("instruction pointer; got %d, want %d", bf.i, i)
	}
	if bf.d != d {
		return fmt.Errorf("data pointer; got %d, want %d", bf.d, d)
	}
	for idx, got := range bf.cells {
		if want := cells[idx]; got != want {
			return fmt.Errorf("cell value at index %d; got %b, want %b", idx, got, want)
		}
	}
	if diff := cmp.Diff(out, bf.out.Bytes()); diff != "" {
		return fmt.Errorf("output diff (-want +got): %s", diff)
	}
	return nil
}

func TestParseProgram(t *testing.T) {
	invalidLoops := []string{"]", "[]]", "[", "[]["}
	for _, test := range invalidLoops {
		_, err := New(test, nil)
		if err == nil {
			t.Errorf("Parsed invalid program %q; expected error", test)
		}
	}
}

func TestLoops(t *testing.T) {
	// Program computes sum of two adjanct cells, read from input.
	program := ",>,[<+>-]<."
	tests := []struct {
		input, want []byte
	}{
		{[]byte{3, 2}, []byte{5}},
		{[]byte{2, 3}, []byte{5}},
		{[]byte{3, 3}, []byte{6}},
		{[]byte{0, 0}, []byte{0}},
	}
	for _, test := range tests {
		bf, err := New(program, test.input)
		if err != nil {
			t.Fatalf("New failed: %v", err)
		}
		out, err := bf.Run()
		if err != nil {
			t.Fatalf("Run failed: %v", err)
		}
		if diff := cmp.Diff(test.want, out); diff != "" {
			t.Errorf("Mismatched output data (-want +got):\n%s", diff)
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

	bf, err := New(program, nil)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	out, err := bf.Run()
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if diff := cmp.Diff("Hello, World!", string(out)); diff != "" {
		t.Errorf("Mismatched output data (-want +got):\n%s", diff)
	}
}
