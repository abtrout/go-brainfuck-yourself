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

	var in, out bytes.Buffer
	if *dataFile != "" {
		if input, err := os.ReadFile(*dataFile); err != nil {
			log.Fatalf("Failed to read input file: %v", err)
		} else {
			in = *bytes.NewBuffer(input)
		}
	}
	bf := gbfy.New(&in, &out)

	if *codeFile != "" {
		if err := runCodeFile(bf); err != nil {
			log.Fatalf("Execution failed: %v", err)
		}
		log.Printf("Output from execution: %v\n", out.Bytes())
	} else {
		if err := runPiped(bf); err != nil {
			log.Fatalf("Execution failed: %v", err)
		}
		log.Printf("Output from execution: %v\n", out.Bytes())
	}

	if *endInteractive {
		log.Println("Starting interactive REPL ...")
		repl(bf, &out)
	}
}

const welcomeMsg = `Go Brainfuck Yourself!
 :d[ump] to dump interpreter state
 :f[uck] to reset interpreter state
 :q[uit] to exit REPL loop`

func repl(bf *gbfy.Brainfuck, out *bytes.Buffer) {
	fmt.Println(welcomeMsg)

	f, err := os.Open("/dev/tty")
	if err != nil {
		log.Fatalf("Failed to open TTY: %v", err)
	}
	stdIn := bufio.NewReader(f)

	// TODO: Switch interpreter input/output here? Continue
	//       reading from original `input` (i.e. from dataFile)
	//       until EOF -- then switch to TTY?

	for {
		fmt.Print("gbfy> ")
		line, err := stdIn.ReadBytes('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("Error! %v", err)
		}

		if len(line) > 0 && line[0] == ':' {
			switch line[1] {
			case 'd':
				formatDump(bf, out.Bytes())
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

func formatDump(bf *gbfy.Brainfuck, out []byte) {
	d, cells, i, cmds := bf.Dump()
	fmt.Printf("CELLS d: %3d, current cell value: %x\n", d, cells[d])
	fmt.Printf("CMDS  i: %3d, current command: %q\n", i, cmds[i-1])
	fmt.Printf("OUT   %v\n", out)
}
