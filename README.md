# go-lisp

一个使用 Go 语言实现的健壮、模块化且可扩展的 Lisp/Scheme 解释器。具有任意精度算术运算、受 Python 启发的内置函数库，以及 TCP 网络和 JSON 处理等高级功能。

## 特性

### 核心语言
- **变量定义**: `(define x 10)`
- **匿名函数 (Lambdas)**: `(define fact (lambda (n) ...))`
- **条件判断**: `(if (> x 5) "high" "low")`
- **局部绑定**: `(let ((a 1) (b 2)) (+ a b))`
- **顺序执行**: `(begin (step1) (step2) (result))`
- **引用 (Quoting)**: `'sym` 或 `(quote (1 2 3))`
- **逻辑运算符**: `and`, `or` (支持 **短路求值**), 和 `not`。
- **代码注释**: 支持以 `;` 开头的单行注释。

### 数据类型
- **强类型检查**: 算术和比较运算符（`+`, `-`, `>`, `=` 等）要求操作数类型一致。
- **整数**: 任意精度的“大数”（通过 `math/big` 实现）。
- **浮点数**: 64位浮点数。
- **字符串**: 支持转义字符（`\n`, `\r`, `\t`, `\"`, `\\`）。
- **布尔值**: `true` 和 `false`。
- **列表 (Lists)**: 链表风格的操作，支持类型安全的排序和处理。
- **字典 (Maps)**: 使用 `(dict "key" value)` 实现快速键值对存储。

### 高级内置函数
- **网络编程 (TCP)**: `tcp-listen`, `tcp-accept`, `tcp-connect`, `tcp-send`, `tcp-read`, `tcp-close`。
- **字典操作**: `dict`, `dict-get`, `dict-set!`, `dict-keys`, `dict-has?`。
- **类型断言**: `integer?`, `float?`, `string?`, `list?`, `dict?`, `nil?`, `bool?`。
- **字符串处理**: `concat`, `string-split`, `string-trim`, `string-replace`, `string-at`, `string->list`。
- **数学运算**: `abs`, `pow`, `divmod`, `round`。
- **类型转换**: `int`, `float`, `str`, `bool`, `bin`, `hex`。
- **迭代器函数**: `range`, `sorted`, `reversed`, `map`, `filter`, `reduce`。
- **系统交互**: `print`, `input`, `type`。
- **文件操作**: `read-file`, `write-file`。

### 标准库 (`stdlib.lisp`)
解释器启动时会自动加载 `stdlib.lisp`，提供：
- **JSON 支持**: `json-parse` 和 `json-stringify`（完全使用 Lisp 编写！）。
- **集合运算**: `sum`, `all`, `any`。
- **列表工具**: `zip`, `enumerate`, `foldl`。
- **谓词函数**: `even?`, `odd?`。
- **字符串工具**: `string-join`。

## 使用方法

### 命令行参数
- `-f <file>`: 执行一个 Lisp 脚本文件。
- `-c "<code>"`: 直接执行一段 Lisp 代码字符串。

### 运行解释器 (REPL)
```bash
go run .
```

### 运行脚本
```bash
go run . -f examples/http_server.lisp
```

### 直接运行代码
```bash
go run . -c "(print (+ 1 2))"
```

### 测试与覆盖率
项目使用文件驱动的 E2E 测试套件。`tests/` 目录下的每个 `.lisp` 文件都在最后一行注释中定义了预期输出。

**运行所有测试:**
```bash
go test -v .
```

**检查代码覆盖率:**
```bash
go test -coverprofile=coverage.out .
go tool cover -func=coverage.out
# 生成可读性强的报告:
python parse_coverage.py
```

## 示例

### HTTP 服务器 (`examples/http_server.lisp`)
使用 Lisp 实现的单线程 HTTP 服务器：
```lisp
(define server (tcp-listen "127.0.0.1:8080"))
(print "Server starting on http://127.0.0.1:8080")
(let ((conn (tcp-accept server)))
  (begin
    (tcp-read conn 1024)
    (tcp-send conn "HTTP/1.1 200 OK\r\n\r\n<h1>Hello!</h1>")
    (tcp-close conn)))
```

### JSON 处理
```lisp
(define data (dict "name" "Lisp" "tags" '("fast" "fun")))
(define json (json-stringify data))
(print json) ; {"name": "Lisp", "tags": ["fast", "fun"]}
```

## 项目结构
- `main.go`: 解析器、分词器和求值器核心。
- `builtin.go`: 运算符工厂和原生 Go 内置函数。
- `stdlib.lisp`: 运行时加载的高级函数（JSON、列表工具等）。
- `tests/`: 详尽的 E2E 测试套件。
- `examples/`: 即插即用的 Lisp 应用程序示例。
