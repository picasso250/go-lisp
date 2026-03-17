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

### Type System & Safety
- **Strict Typing**: Arithmetic and comparison operators (`+`, `-`, `>`, `=`, etc.) require both operands to be of the same type.
- **Integers**: Arbitrary precision "Big Numbers" (via `math/big`).
- **Floats**: 64-bit floating point numbers.
- **Strings**: Support for escape characters (`\n`, `\t`, `\"`, `\\`).
- **Booleans**: `true` and `false`.
- **Lists**: Linked-list style operations with robust type-safe sorting and manipulation.

### Python-like Built-ins
Inspired by Python's standard functions for high productivity:
- **Math**: `abs`, `pow`, `divmod`, `round`, `max`, `min`.
- **Conversions**: `int`, `float`, `str`, `bool`, `bin`, `hex`, `oct`, `chr`, `ord`.
- **Iterables**: `range` (single or dual param), `sorted`, `reversed`.
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
Get-Content tests/fact.lisp | go run .
```

### Testing & Coverage
The project uses a file-driven E2E test suite. Each `.lisp` file in `tests/` defines its expected output in the last line comment.

**Run All Tests:**
```bash
go test -v .
```

**Check Code Coverage:**
```bash
go test -coverprofile=coverage.out .
go tool cover -func=coverage.out
# Generate human-readable report:
python parse_coverage.py
# View visual report:
# go tool cover -html=coverage.out
```

## Examples

### Logic & Math
```lisp
(and true (> 10 5))           ; true
(or false (= 1 1))            ; true
(pow 2 100)                   ; 1267650600228229401496703205376 (BigInt!)
(divmod 10 3)                 ; [3 1]
(round 3.6)                   ; 4
```

### Functional Programming
```lisp
(define nums (range 10))      ; [0 1 2 3 4 5 6 7 8 9]
(sum nums)                    ; 45
(map (lambda (x) (* x x)) '(1 2 3)) ; [1 4 9]
(filter even? (range 6))      ; [0 2 4]
(sorted '(3 1 2))             ; [1 2 3] (Type-safe sorting)
```

## Project Structure
- `main.go`: Parser, Tokenizer, and Evaluator core.
- `builtin.go`: Unified operator factories and type-safe built-ins.
- `stdlib.lisp`: Higher-level Lisp functions loaded at runtime.
- `tests/`: Extensive E2E test suite (90%+ coverage).
- `e2e_test.go`: Automated test runner with In-Process evaluation support.
