# go-lisp

A simple Lisp/Scheme interpreter in Go.

## Features
- Basic arithmetic: `+`, `-`, `*`, `/`
- Comparisons: `<=`, `>`
- `define`, `if`, `lambda`
- Recursion support

## Usage
Run the interpreter:
```bash
go run main.go
```

Run a smoke test:
```bash
Get-Content fact.lisp | go run main.go
```

## Smoke Test (fact.lisp)
```lisp
(define fact (lambda (n) (if (<= n 1) 1 (* n (fact (- n 1))))))
(fact 5)
```
Output:
```
120
```
