# goquickfix
[![CI Status](https://github.com/motemen/go-quickfix/actions/workflows/ci.yaml/badge.svg?branch=master)](https://github.com/motemen/go-quickfix/actions?query=branch:master)
[![Go Report Card](https://goreportcard.com/badge/github.com/motemen/go-quickfix)](https://goreportcard.com/report/github.com/motemen/go-quickfix)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/motemen/go-quickfix/blob/master/LICENSE)
[![release](https://img.shields.io/github/release/motemen/go-quickfix/all.svg)](https://github.com/motemen/go-quickfix/releases)
[![pkg.go.dev](https://pkg.go.dev/badge/github.com/motemen/go-quickfix)](https://pkg.go.dev/github.com/motemen/go-quickfix)

The `goquickfix` command quickly fixes Go source that is well typed but
Go refuses to compile e.g. `declared and not used: x`.

## Installation

```sh
go install github.com/motemen/go-quickfix/cmd/goquickfix@latest
```

## Usage

```
goquickfix [-w] [-revert] <path>...

Flags:
  -revert: try to revert possible quickfixes introduced by goquickfix
  -w: write result to (source) file instead of stdout
```

### Description

While coding, you may write a Go program that is completely well typed
but `go build` (or `run` or `test`) refuses to build, like this:

```go
package main

import (
	"fmt"
	"log"
)

func main() {
	nums := []int{3, 1, 4, 1, 5}
	for i, n := range nums {
		fmt.Println(n)
	}
}
```

The Go compiler will complain:

```
main.go:5:2: "log" imported and not used
main.go:10:6: declared and not used: i
```

Do we have to bother to comment out the import line or remove
the unused identifier one by one for the Go compiler? Of course no,
`goquickfix` does the work instead of you.

Run

```
goquickfix -w main.go
```

and you will get the source rewritten so that Go compiles it well without
changing the semantics:

```go
package main

import (
	"fmt"
	_ "log"
)

func main() {
	nums := []int{3, 1, 4, 1, 5}
	for i, n := range nums {
		fmt.Println(n)
		_ = i
	}
}
```

Now, you can `go run` or `go test` your code successfully.
