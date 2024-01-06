# lit

*lit* is a Git clone created in Go and using the well-known *Cobra* library for command-line-argument-parsing. I created it because I wanted to understand Git internals thoroughly, and implementing this allowed me to verify my understanding. Additionally, it allowed me to put Go's command line abilities and ecosystem to good use.
lit uses the same internal structure as Git (blobs, pointers, etc.) which are stored in a .lit file. To get started, run
```
lit init
```
Currently, supported commands are limited to:
```
lit add <file-or-folder>
lit branch <name>
lit checkout <location>
lit commit
lit log
lit status
```
Other functionality may be added in the future.

# Installation

to install the program, download the files and run `go install`. You will need the Go toolchain installed before doing this. Then, the program should be accessible from the command line as `lit`.

