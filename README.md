# go-lisp

A robust and extensible Lisp/Scheme interpreter implemented in Go.

## Features

### Core Language
- **Variable Definitions**: `(define x 10)`
- **Functions (Lambdas)**: `(define fact (lambda (n) ...))`
- **Conditionals**: `(if (> x 5) "high" "low")`
- **Local Bindings**: `(let ((a 1) (b 2)) (+ a b))`
- **Sequential Execution**: `(begin (step1) (step2) (result))`
- **Quoting**: `'sym` or `(quote (1 2 3))`

### Data Types
- **Integers**: Arbitrary precision "Big Numbers" (via `math/big`).
- **Floats**: 64-bit floating point numbers.
- **Strings**: Support for escape characters (`\n`, `\t`, `\"`, `\\`).
- **Lists**: Linked-list style operations.
- **Symbols**: For variable names and identifiers.

### Strict Typing & Predicates
- **Strict Arithmetic**: Operators (`+`, `-`, `*`, `/`) require matching types (e.g., Integer + Integer).
- **Type Conversions**: `(integer x)`, `(float x)`.
- **Type Predicates**: `(integer? x)`, `(float? x)`, `(string? x)`, `(atom? x)`, `(null? x)`.

### Built-in Functions
- **Arithmetic**: `+`, `-`, `*`, `/`
- **Comparisons**: `<`, `>`, `<=`, `>=`, `=`, `!=`
- **List Ops**: `car`, `cdr`, `cons`, `list`, `length`
- **String Ops**: `concat`, `string-split`, `string-trim`

## Usage

### Run Interpreter (REPL/Pipe)
```bash
go run main.go
```

### Run a Lisp Script
```powershell
Get-Content fact.lisp | go run main.go
```

### Run E2E Tests
```bash
go test -v e2e_test.go main.go
```

## Example: Factorial with BigInt
```lisp
(define fact (lambda (n) 
  (if (<= n 1) 
      1 
      (* n (fact (- n 1))))))

(fact 50) 
;; Returns: 30414093201713378043612608166064768844377641568960512000000000000
```

## Example: Strings and Lists
```lisp
(define msg "  Hello, Lisp!  ")
(string-trim msg)                    ; "Hello, Lisp!"
(string-split "apple,banana" ",")    ; ["apple" "banana"]
(concat "Result: " (* 10 10))        ; "Result: 100"
```
