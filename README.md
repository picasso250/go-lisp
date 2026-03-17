# go-lisp

A robust, modular, and extensible Lisp/Scheme interpreter implemented in Go, featuring arbitrary-precision arithmetic, a Python-inspired built-in library, and advanced capabilities like TCP networking and JSON processing.

## Features

### Core Language
- **Variable Definitions**: `(define x 10)`
- **Functions (Lambdas)**: `(define fact (lambda (n) ...))`
- **Conditionals**: `(if (> x 5) "high" "low")`
- **Local Bindings**: `(let ((a 1) (b 2)) (+ a b))`
- **Sequential Execution**: `(begin (step1) (step2) (result))`
- **Quoting**: `'sym` or `(quote (1 2 3))`
- **Logical Operators**: `and`, `or` (with **short-circuit evaluation**), and `not`.
- **Comments**: Support for single-line comments using `;`.

### Data Types
- **Strict Typing**: Arithmetic and comparison operators (`+`, `-`, `>`, `=`, etc.) require both operands to be of the same type.
- **Integers**: Arbitrary precision "Big Numbers" (via `math/big`).
- **Floats**: 64-bit floating point numbers.
- **Strings**: Support for escape characters (`\n`, `\r`, `\t`, `\"`, `\\`).
- **Booleans**: `true` and `false`.
- **Lists**: Linked-list style operations with robust type-safe sorting and manipulation.
- **Dictionaries (Maps)**: Fast key-value storage using `(dict "key" value)`.

### Advanced Built-ins
- **Networking (TCP)**: `tcp-listen`, `tcp-accept`, `tcp-connect`, `tcp-send`, `tcp-read`, `tcp-close`.
- **Dictionary Ops**: `dict`, `dict-get`, `dict-set!`, `dict-keys`, `dict-has?`.
- **Predicates**: `integer?`, `float?`, `string?`, `list?`, `dict?`, `nil?`, `bool?`.
- **String Ops**: `concat`, `string-split`, `string-trim`, `string-replace`, `string-at`, `string->list`.
- **Math**: `abs`, `pow`, `divmod`, `round`.
- **Conversions**: `int`, `float`, `str`, `bool`, `bin`, `hex`.
- **Iterables**: `range`, `sorted`, `reversed`, `map`, `filter`, `reduce`.
- **System**: `print`, `input`, `type`.
- **Files**: `read-file`, `write-file`.

### Standard Library (`stdlib.lisp`)
The interpreter automatically loads `stdlib.lisp` on startup, providing:
- **JSON**: `json-parse` and `json-stringify` (implemented purely in Lisp!).
- **Aggregates**: `sum`, `all`, `any`.
- **List Utils**: `zip`, `enumerate`, `foldl`.
- **Predicates**: `even?`, `odd?`.
- **String Utils**: `string-join`.

## Usage

### CLI Flags
- `-f <file>`: Execute a Lisp script file.
- `-c "<code>"`: Execute a Lisp code string directly.

### Run Interpreter (REPL)
```bash
go run .
```

### Run a Script
```bash
go run . -f examples/http_server.lisp
```

### Run Code Directly
```bash
go run . -c "(print (+ 1 2))"
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
```

## Examples

### HTTP Server (`examples/http_server.lisp`)
A single-threaded HTTP server implemented in Lisp:
```lisp
(define server (tcp-listen "127.0.0.1:8080"))
(print "Server starting on http://127.0.0.1:8080")
(let ((conn (tcp-accept server)))
  (begin
    (tcp-read conn 1024)
    (tcp-send conn "HTTP/1.1 200 OK\r\n\r\n<h1>Hello!</h1>")
    (tcp-close conn)))
```

### JSON Processing
```lisp
(define data (dict "name" "Lisp" "tags" '("fast" "fun")))
(define json (json-stringify data))
(print json) ; {"name": "Lisp", "tags": ["fast", "fun"]}
```

## Project Structure
- `main.go`: Parser, Tokenizer, and Evaluator core.
- `builtin.go`: Operator factories and native Go built-ins.
- `stdlib.lisp`: Higher-level functions (JSON, list utils) loaded at runtime.
- `tests/`: Extensive E2E test suite.
- `examples/`: Ready-to-run Lisp applications.
