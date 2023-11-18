# `gbfy`

A [Brainfuck](https://en.wikipedia.org/wiki/Brainfuck) interpreter that I wrote myself in Go. 


```
$ cat test.bf
,>,<[>+<-]>.
$ hexdump -C input.bytes
00000000  01 02                                             |..|
00000002
$ ./gbfy -f test.bf -i input.bytes -repl
2023/12/02 18:41:06 Output from execution: [3]
2023/12/02 18:41:06 Starting interactive REPL
Welcome!
 :q[uit] to exit REPL loop 
 :d[ump] to dump interpreter state
 :f[uck] to reset interpreter state
gbfy> :d
cmds: ",>,<[>+<-]>."
i: 12, current cmd: '.'
cells: [0 3 0 0 0 0 0 0 0 0]
d: 1, current cell: 3
gbfy> +++++
gbfy> :d
cmds: ",>,<[>+<-]>.+++++"
i: 17, current cmd: '+'
cells: [0 8 0 0 0 0 0 0 0 0]
d: 1, current cell: 8
```
