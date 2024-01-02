# `gbfy`

A [Brainfuck](https://en.wikipedia.org/wiki/Brainfuck) interpreter that I wrote myself in Go. 

It can run programs in files and read binary input data for those programs to consume.


```
$ ./gbfy --help
Usage of ./gbfy:
  -d string
        Path to file containing input data for program
  -f string
        Path to Brainfuck program to execute
  -i    Launch interactive interpreter before exit
$ cat test_sum.bf 
## Reads 2 bytes of input and adds them together
,>,<[->+<]>.
$ hexdump -C input.bytes 
00000000  01 02                                             |..|
00000002
$ ./gbfy -d input.bytes -f test.bf
2024/01/02 05:22:36 Output from execution: [3]
```

It can also run commands that are piped, and be used as an interactive REPL.

```
$ echo "+>++<[->+<]>." | ./gbfy -i
2024/01/02 05:23:25 Starting interactive REPL ...
Go Brainfuck Yourself!
 :d[ump] to dump interpreter state
 :f[uck] to reset interpreter state
 :q[uit] to exit REPL loop
gbfy> :d
CELLS d:   1, current cell value: 3
CMDS  i:  13, current command: '.'
OUT   [3]
gbfy> ++++.
gbfy> :d
CELLS d:   1, current cell value: 7
CMDS  i:  18, current command: '.'
OUT   [3 7]
gbfy> 
```