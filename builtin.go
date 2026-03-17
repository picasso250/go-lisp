package main

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
)

// mustSameType 统一处理类型一致性检查
func mustSameType(name string, a, b interface{}) {
	if fmt.Sprintf("%T", a) != fmt.Sprintf("%T", b) {
		panic(fmt.Sprintf("%s: type mismatch", name))
	}
}

// resolve evaluates TailCall to final value
func resolve(val interface{}) interface{} {
	for {
		if tc, ok := val.(TailCall); ok {
			val = eval(tc.exp, tc.env)
		} else {
			return val
		}
	}
}

// binaryOp 算术运算工厂
func binaryOp(name string, opInt func(int64, int64) int64, opFloat func(float64, float64) float64) func([]interface{}) interface{} {
	return func(args []interface{}) interface{} {
		a, b := args[0], args[1]
		mustSameType(name, a, b)
		switch av := a.(type) {
		case int64:
			return opInt(av, b.(int64))
		case float64:
			return opFloat(av, b.(float64))
		default:
			panic(fmt.Sprintf("%s: invalid types", name))
		}
	}
}

// compareOp 比较运算工厂
func compareOp(name string, opInt func(int64, int64) bool, opFloat func(float64, float64) bool, opStr func(string, string) bool) func([]interface{}) interface{} {
	return func(args []interface{}) interface{} {
		a, b := args[0], args[1]
		mustSameType(name, a, b)
		switch av := a.(type) {
		case int64:
			return opInt(av, b.(int64))
		case float64:
			return opFloat(av, b.(float64))
		case string:
			return opStr(av, b.(string))
		default:
			panic(fmt.Sprintf("%s: incomparable type", name))
		}
	}
}

// standardEnv sets up the initial environment with built-in functions
func standardEnv() *Env {
	e := &Env{vars: make(map[Symbol]interface{}), outer: nil}

	// Basic Arithmetic
	e.set("+", func(args []interface{}) interface{} {
		if len(args) == 0 {
			return int64(0)
		}
		res := args[0]
		for _, next := range args[1:] {
			mustSameType("+", res, next)
			switch a := res.(type) {
			case int64:
				res = a + next.(int64)
			case float64:
				res = a + next.(float64)
			}
		}
		return res
	})

	e.set("*", func(args []interface{}) interface{} {
		if len(args) == 0 {
			return int64(1)
		}
		res := args[0]
		for _, next := range args[1:] {
			mustSameType("*", res, next)
			switch a := res.(type) {
			case int64:
				res = a * next.(int64)
			case float64:
				res = a * next.(float64)
			}
		}
		return res
	})

	e.set("-", func(args []interface{}) interface{} {
		if len(args) == 0 {
			panic("- requires at least 1 argument")
		}
		if len(args) == 1 {
			switch v := args[0].(type) {
			case int64:
				return -v
			case float64:
				return -v
			default:
				panic("- requires number")
			}
		}
		res := args[0]
		for _, next := range args[1:] {
			mustSameType("-", res, next)
			switch a := res.(type) {
			case int64:
				res = a - next.(int64)
			case float64:
				res = a - next.(float64)
			}
		}
		return res
	})

	e.set("/", binaryOp("/", func(a, b int64) int64 { return a / b }, func(a, b float64) float64 { return a / b }))
	e.set("%", binaryOp("%", func(a, b int64) int64 { return a % b }, func(a, b float64) float64 { return math.Mod(a, b) }))

	// Comparisons with Chaining Support
	makeCompare := func(name string, opInt func(int64, int64) bool, opFloat func(float64, float64) bool, opStr func(string, string) bool) func([]interface{}) interface{} {
		return func(args []interface{}) interface{} {
			if len(args) < 2 {
				panic(fmt.Sprintf("%s requires at least 2 arguments", name))
			}
			for i := 0; i < len(args)-1; i++ {
				a, b := args[i], args[i+1]
				mustSameType(name, a, b)
				var ok bool
				switch av := a.(type) {
				case int64:
					ok = opInt(av, b.(int64))
				case float64:
					ok = opFloat(av, b.(float64))
				case string:
					ok = opStr(av, b.(string))
				default:
					panic(fmt.Sprintf("%s: incomparable type", name))
				}
				if !ok {
					return false
				}
			}
			return true
		}
	}

	e.set("<", makeCompare("<", func(a, b int64) bool { return a < b }, func(a, b float64) bool { return a < b }, func(a, b string) bool { return a < b }))
	e.set("<=", makeCompare("<=", func(a, b int64) bool { return a <= b }, func(a, b float64) bool { return a <= b }, func(a, b string) bool { return a <= b }))
	e.set(">", makeCompare(">", func(a, b int64) bool { return a > b }, func(a, b float64) bool { return a > b }, func(a, b string) bool { return a > b }))
	e.set(">=", makeCompare(">=", func(a, b int64) bool { return a >= b }, func(a, b float64) bool { return a >= b }, func(a, b string) bool { return a >= b }))

	e.set("=", func(args []interface{}) interface{} {
		if len(args) < 2 {
			panic("= requires at least 2 arguments")
		}
		for i := 0; i < len(args)-1; i++ {
			a, b := args[i], args[i+1]
			mustSameType("=", a, b)
			if a != b {
				return false
			}
		}
		return true
	})
	e.set("!=", func(args []interface{}) interface{} {
		return !e.get("=").(func([]interface{}) interface{})(args).(bool)
	})

	// Python-like Math
	e.set("abs", func(args []interface{}) interface{} {
		switch v := args[0].(type) {
		case int64:
			if v < 0 {
				return -v
			}
			return v
		case float64:
			return math.Abs(v)
		default:
			panic("abs requires number")
		}
	})
	e.set("pow", func(args []interface{}) interface{} {
		a, b := args[0], args[1]
		if ai, ok := a.(int64); ok {
			return int64(math.Pow(float64(ai), float64(b.(int64))))
		}
		return math.Pow(a.(float64), b.(float64))
	})
	e.set("divmod", func(args []interface{}) interface{} {
		a, b := args[0].(int64), args[1].(int64)
		return List{a / b, a % b}
	})
	e.set("round", func(args []interface{}) interface{} { return math.Round(args[0].(float64)) })

	// Conversions & Predicates
	e.set("int", func(args []interface{}) interface{} {
		switch v := args[0].(type) {
		case float64:
			return int64(v)
		case string:
			i, _ := strconv.ParseInt(v, 10, 64)
			return i
		default:
			return v.(int64)
		}
	})
	e.set("integer", e.get("int")) // alias
	e.set("float", func(args []interface{}) interface{} {
		switch v := args[0].(type) {
		case int64:
			return float64(v)
		case string:
			f, _ := strconv.ParseFloat(v, 64)
			return f
		case float64:
			return v
		default:
			panic(fmt.Sprintf("float conversion failed: unsupported type %T", v))
		}
	})
	e.set("str", func(args []interface{}) interface{} {
		return fmt.Sprintf("%v", args[0])
	})
	e.set("bool", func(args []interface{}) interface{} {
		if args[0] == nil || args[0] == false {
			return false
		}
		if i, ok := args[0].(int64); ok && i == 0 {
			return false
		}
		if f, ok := args[0].(float64); ok && f == 0 {
			return false
		}
		if s, ok := args[0].(string); ok && s == "" {
			return false
		}
		if l, ok := args[0].(List); ok && len(l) == 0 {
			return false
		}
		return true
	})
	e.set("integer?", func(args []interface{}) interface{} { _, ok := args[0].(int64); return ok })
	e.set("float?", func(args []interface{}) interface{} { _, ok := args[0].(float64); return ok })
	e.set("string?", func(args []interface{}) interface{} { _, ok := args[0].(string); return ok })
	e.set("list?", func(args []interface{}) interface{} { _, ok := args[0].(List); return ok })
	e.set("dict?", func(args []interface{}) interface{} { _, ok := args[0].(Dict); return ok })
	e.set("nil?", func(args []interface{}) interface{} { return args[0] == nil })
	e.set("bool?", func(args []interface{}) interface{} { _, ok := args[0].(bool); return ok })

	// List Ops
	e.set("car", func(args []interface{}) interface{} { return args[0].(List)[0] })
	e.set("cdr", func(args []interface{}) interface{} { return args[0].(List)[1:] })
	e.set("cons", func(args []interface{}) interface{} { return append(List{args[0]}, args[1].(List)...) })
	e.set("list", func(args []interface{}) interface{} { return List(args) })
	e.set("null?", func(args []interface{}) interface{} { return len(args[0].(List)) == 0 })
	e.set("length", func(args []interface{}) interface{} { return int64(len(args[0].(List))) })
	e.set("append", func(args []interface{}) interface{} {
		res := List{}
		for _, arg := range args {
			res = append(res, arg.(List)...)
		}
		return res
	})
	e.set("nth", func(args []interface{}) interface{} {
		idx := args[1].(int64)
		return args[0].(List)[idx]
	})
	e.set("reverse", func(args []interface{}) interface{} {
		l := args[0].(List)
		res := make(List, len(l))
		for i, v := range l {
			res[len(l)-1-i] = v
		}
		return res
	})
	e.set("reversed", e.get("reverse"))

	// Dictionary Ops
	e.set("dict", func(args []interface{}) interface{} {
		d := make(Dict)
		for i := 0; i < len(args); i += 2 {
			d[args[i].(string)] = args[i+1]
		}
		return d
	})
	e.set("dict-get", func(args []interface{}) interface{} {
		d, key := args[0].(Dict), args[1].(string)
		return d[key]
	})
	e.set("dict-set!", func(args []interface{}) interface{} {
		d, key, val := args[0].(Dict), args[1].(string), args[2]
		d[key] = val
		return val
	})
	e.set("dict-keys", func(args []interface{}) interface{} {
		d := args[0].(Dict)
		keys := make(List, 0, len(d))
		for k := range d {
			keys = append(keys, k)
		}
		return keys
	})
	e.set("dict-has?", func(args []interface{}) interface{} {
		d, key := args[0].(Dict), args[1].(string)
		_, ok := d[key]
		return ok
	})

	// TCP Ops
	e.set("tcp-listen", func(args []interface{}) interface{} {
		addr := args[0].(string)
		l, err := net.Listen("tcp", addr)
		if err != nil {
			panic(fmt.Sprintf("tcp-listen failed: %v", err))
		}
		return l
	})
	e.set("tcp-accept", func(args []interface{}) interface{} {
		l := args[0].(net.Listener)
		conn, err := l.Accept()
		if err != nil {
			panic(fmt.Sprintf("tcp-accept failed: %v", err))
		}
		return conn
	})
	e.set("tcp-connect", func(args []interface{}) interface{} {
		addr := args[0].(string)
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			panic(fmt.Sprintf("tcp-connect failed: %v", err))
		}
		return conn
	})
	e.set("tcp-send", func(args []interface{}) interface{} {
		conn, data := args[0].(net.Conn), args[1].(string)
		n, err := conn.Write([]byte(data))
		if err != nil {
			panic(fmt.Sprintf("tcp-send failed: %v", err))
		}
		return int64(n)
	})
	e.set("tcp-read", func(args []interface{}) interface{} {
		conn := args[0].(net.Conn)
		max := 1024
		if len(args) > 1 {
			max = int(args[1].(int64))
		}
		buf := make([]byte, max)
		n, err := conn.Read(buf)
		if err != nil && err != io.EOF {
			panic(fmt.Sprintf("tcp-read failed: %v", err))
		}
		return string(buf[:n])
	})
	e.set("tcp-close", func(args []interface{}) interface{} {
		if l, ok := args[0].(net.Listener); ok {
			l.Close()
		} else if c, ok := args[0].(net.Conn); ok {
			c.Close()
		}
		return nil
	})

	// String Ops
	e.set("concat", func(args []interface{}) interface{} {
		res := ""
		for _, arg := range args {
			res += fmt.Sprintf("%v", arg)
		}
		return res
	})
	e.set("string-split", func(args []interface{}) interface{} {
		s, sep := args[0].(string), args[1].(string)
		parts := strings.Split(s, sep)
		res := make(List, len(parts))
		for i, p := range parts {
			res[i] = p
		}
		return res
	})
	e.set("string-trim", func(args []interface{}) interface{} { return strings.TrimSpace(args[0].(string)) })
	e.set("string-length", func(args []interface{}) interface{} { return int64(len([]rune(args[0].(string)))) })
	e.set("string-contains?", func(args []interface{}) interface{} { return strings.Contains(args[0].(string), args[1].(string)) })
	e.set("string-replace", func(args []interface{}) interface{} {
		return strings.ReplaceAll(args[0].(string), args[1].(string), args[2].(string))
	})
	e.set("string-at", func(args []interface{}) interface{} {
		s := []rune(args[0].(string))
		i := args[1].(int64)
		return string(s[i])
	})
	e.set("string->list", func(args []interface{}) interface{} {
		s := []rune(args[0].(string))
		res := make(List, len(s))
		for i, r := range s {
			res[i] = string(r)
		}
		return res
	})

	// Files
	e.set("read-file", func(args []interface{}) interface{} {
		content, err := os.ReadFile(args[0].(string))
		if err != nil {
			panic(fmt.Sprintf("read-file failed: %v", err))
		}
		return string(content)
	})
	e.set("write-file", func(args []interface{}) interface{} {
		err := os.WriteFile(args[0].(string), []byte(args[1].(string)), 0644)
		if err != nil {
			panic(fmt.Sprintf("write-file failed: %v", err))
		}
		return true
	})

	// Iterables
	e.set("range", func(args []interface{}) interface{} {
		start := int64(0)
		stop := args[0].(int64)
		if len(args) > 1 {
			start = stop
			stop = args[1].(int64)
		}
		res := List{}
		for i := start; i < stop; i++ {
			res = append(res, i)
		}
		return res
	})
	e.set("sorted", func(args []interface{}) interface{} {
		l := append(List{}, args[0].(List)...)
		if len(l) <= 1 {
			return l
		}
		// Check for mixed types
		firstType := fmt.Sprintf("%T", l[0])
		for _, item := range l[1:] {
			if fmt.Sprintf("%T", item) != firstType {
				panic("sorted: mixed types")
			}
		}
		sort.Slice(l, func(i, j int) bool {
			switch av := l[i].(type) {
			case int64:
				return av < l[j].(int64)
			case float64:
				return av < l[j].(float64)
			case string:
				return av < l[j].(string)
			default:
				return false
			}
		})
		return l
	})

	// System & Debug
	e.set("print", func(args []interface{}) interface{} {
		for i, arg := range args {
			if i > 0 {
				fmt.Print(" ")
			}
			fmt.Print(arg)
		}
		fmt.Println()
		return nil
	})
	e.set("input", func(args []interface{}) interface{} {
		if len(args) > 0 {
			fmt.Print(args[0])
		}
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			return scanner.Text()
		}
		return ""
	})
	e.set("type", func(args []interface{}) interface{} { return fmt.Sprintf("%T", args[0]) })
	e.set("bin", func(args []interface{}) interface{} { return "0b" + strconv.FormatInt(args[0].(int64), 2) })
	e.set("hex", func(args []interface{}) interface{} { return "0x" + strconv.FormatInt(args[0].(int64), 16) })

	return e
}
