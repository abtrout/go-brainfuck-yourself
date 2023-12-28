package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/abtrout/gbfy"
)

var (
	codeFile       = flag.String("f", "", "Path to Brainfuck program to execute")
	dataFile       = flag.String("d", "", "Path to file containing input data for program")
	endInteractive = flag.Bool("i", false, "Launch interactive interpreter before exit")
)

func main() {
	flag.Parse()

	var (
		input  []byte
		output bytes.Buffer
	)
	if *dataFile != "" {
		var err error
		if input, err = os.ReadFile(*dataFile); err != nil {
			log.Fatalf("Failed to read input file: %v", err)
		}
	}

	bf := gbfy.New(bytes.NewBuffer(input), &output)
	if *codeFile != "" {
		if err := runCodeFile(bf); err != nil {
			log.Fatalf("Execution failed: %v", err)
		}
	} else if err := runPiped(bf); err != nil {
		log.Fatalf("Execution failed: %v", err)
	}

	if *endInteractive {
		log.Println("Starting interactive REPL ...")
		repl(bf)
	}
}

const welcomeMsg = `Go Brainfuck Yourself!
 :q[uit] to exit REPL loop 
 :d[ump] to dump interpreter state
 :f[uck] to reset interpreter state`

func repl(bf *gbfy.Brainfuck) {
	fmt.Println(welcomeMsg)

	f, err := os.Open("/dev/tty")
	if err != nil {
		log.Fatalf("Failed to open TTY: %v", err)
	}
	in := bufio.NewReader(f)

	for {
		fmt.Print("gbfy> ")
		line, err := in.ReadBytes('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("Error! %v", err)
		}

		if len(line) > 0 && line[0] == ':' {
			switch line[1] {
			case 'd':
				formatDump(bf.Dump())
			case 'f':
				bf.Reset()
				fmt.Println("Reset interpreter!")
			case 'q':
				break
			}
		} else {
			r := bufio.NewReader(bytes.NewReader(line))
			if err := runCommands(bf, r); err != nil {
				log.Fatalf("Execution failed: %v", err)
			}
		}
	}
}

func runCodeFile(bf *gbfy.Brainfuck) error {
	f, err := os.Open(*codeFile)
	if err != nil {
		return err
	}
	in := bufio.NewReader(f)
	if err := runCommands(bf, in); err != nil {
		return err
	}
	logOutput(bf)
	return nil
}

func runPiped(bf *gbfy.Brainfuck) error {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return err
	}
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		in := bufio.NewReader(os.Stdin)
		if err := runCommands(bf, in); err != nil {
			return err
		}
	}
	logOutput(bf)
	return nil
}

func runCommands(bf *gbfy.Brainfuck, r *bufio.Reader) error {
	for {
		cmd, err := r.ReadByte()
		if err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("failed to ReadByte: %v", err)
		}
		if err := bf.Eval(cmd); err != nil {
			return fmt.Errorf("failed to Eval %q: %v", cmd, err)
		}
	}
	return nil
}

func logOutput(bf *gbfy.Brainfuck) {
	_, _, _, _, out := bf.Dump()
	log.Printf("Output from execution: %v\n", out)
}

func formatDump(d int, cells []byte, i int, cmds []byte, out []byte) {
	fmt.Printf("Cells; d: %d, current cell value: %x\n", d, cells[d])
	fmt.Printf("Cmds: i: %d, current command: %q\n", i, cmds[i-1])
	fmt.Printf("Out: %v\n", out)
}
