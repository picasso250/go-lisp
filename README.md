# go-lisp

A robust, modular, and extensible Lisp/Scheme interpreter implemented in Go, featuring arbitrary-precision arithmetic and a Python-inspired built-in library.

## Features

### Core Language
- **Variable Definitions**: `(define x 10)`
- **Functions (Lambdas)**: `(define fact (lambda (n) ...))`
- **Conditionals**: `(if (> x 5) "high" "low")`
- **Local Bindings**: `(let ((a 1) (b 2)) (+ a b))`
- **Sequential Execution**: `(begin (step1) (step2) (result))`
- **Quoting**: `'sym` or `(quote (1 2 3))`
- **Logical Operators**: `and`, `or` (with **short-circuit evaluation**), and `not`.

### Data Types
- **Integers**: Arbitrary precision "Big Numbers" (via `math/big`).
- **Floats**: 64-bit floating point numbers.
- **Strings**: Support for escape characters (`\n`, `\t`, `\"`, `\\`).
- **Booleans**: `true` and `false`.
- **Lists**: Linked-list style operations and Go-slice backed performance.

### Python-like Built-ins
Inspired by Python's standard functions for high productivity:
- **Math**: `abs`, `pow`, `divmod`, `round`, `max`, `min`.
- **Conversions**: `int`, `float`, `str`, `bool`, `bin`, `hex`, `oct`, `chr`, `ord`.
- **Iterables**: `range`, `sorted`, `reversed`.
- **System**: `print`, `input`, `type`, `callable`.
- **Files**: `read-file`, `write-file`.

### Standard Library (`stdlib.lisp`)
The interpreter automatically loads `stdlib.lisp` on startup, providing:
- **Aggregates**: `sum`, `all`, `any`.
- **List Utils**: `zip`, `enumerate`, `foldl`, `list-empty?`.
- **Predicates**: `even?`, `odd?`.

## Usage

### Run Interpreter (REPL/Pipe)
```bash
go run .
```

### Run a Lisp Script
```powershell
Get-Content fact.lisp | go run .
```

### Run E2E Tests
```bash
go test -v .
```

## Examples

### Logic & Math
```lisp
(and true (> 10 5))           ; true
(or false (= 1 1))            ; true
(pow 2 100)                   ; 1267650600228229401496703205376 (BigInt!)
(bin 10)                      ; "0b1010"
(divmod 10 3)                 ; [3 1]
```

### Functional Programming (Python-style)
```lisp
(define nums (range 10))      ; [0 1 2 3 4 5 6 7 8 9]
(sum nums)                    ; 45
(map (lambda (x) (* x x)) '(1 2 3)) ; [1 4 9]
(filter even? (range 6))      ; [0 2 4]
(enumerate '(a b))            ; [[0 a] [1 b]]
(zip '(a b) '(1 2))           ; [[a 1] [b 2]]
```

### String & File I/O
```lisp
(print "Result is:" (+ 1 2))  ; Prints: Result is: 3
(write-file "log.txt" "Done") ; Writes "Done" to log.txt
(string-split "a,b,c" ",")    ; ["a" "b" "c"]
```

## Project Structure
- `main.go`: Parser, Tokenizer, and Evaluator core.
- `builtin.go`: High-performance Go-native built-in functions.
- `stdlib.lisp`: Higher-level Lisp functions loaded at runtime.
- `e2e_test.go`: Automated end-to-end regression tests.
