package main

import (
	"bufio"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"unicode"
)

type Symbol string
type List []interface{}

type Env struct {
	vars  map[Symbol]interface{}
	outer *Env
}

func (e *Env) get(s Symbol) interface{} {
	if val, ok := e.vars[s]; ok {
		return val
	}
	if e.outer != nil {
		return e.outer.get(s)
	}
	panic(fmt.Sprintf("Symbol not found: %s", s))
}

func (e *Env) set(s Symbol, val interface{}) {
	e.vars[s] = val
}

func standardEnv() *Env {
	e := &Env{vars: make(map[Symbol]interface{}), outer: nil}

	// Helper for strict arithmetic
	binaryOp := func(name string, opInt func(*big.Int, *big.Int) *big.Int, opFloat func(float64, float64) float64) func([]interface{}) interface{} {
		return func(args []interface{}) interface{} {
			a, b := args[0], args[1]
			switch av := a.(type) {
			case *big.Int:
				bv, ok := b.(*big.Int)
				if !ok {
					panic(fmt.Sprintf("%s: type mismatch, expected integers", name))
				}
				return opInt(av, bv)
			case float64:
				bv, ok := b.(float64)
				if !ok {
					panic(fmt.Sprintf("%s: type mismatch, expected floats", name))
				}
				return opFloat(av, bv)
			default:
				panic(fmt.Sprintf("%s: invalid types", name))
			}
		}
	}

	e.set("+", binaryOp("+", func(a, b *big.Int) *big.Int { return new(big.Int).Add(a, b) }, func(a, b float64) float64 { return a + b }))
	e.set("-", binaryOp("-", func(a, b *big.Int) *big.Int { return new(big.Int).Sub(a, b) }, func(a, b float64) float64 { return a - b }))
	e.set("*", binaryOp("*", func(a, b *big.Int) *big.Int { return new(big.Int).Mul(a, b) }, func(a, b float64) float64 { return a * b }))
	e.set("/", binaryOp("/", func(a, b *big.Int) *big.Int { return new(big.Int).Div(a, b) }, func(a, b float64) float64 { return a / b }))

	// Comparisons
	e.set("<", func(args []interface{}) interface{} {
		a, b := args[0], args[1]
		switch av := a.(type) {
		case *big.Int:
			return av.Cmp(b.(*big.Int)) < 0
		case float64:
			return av < b.(float64)
		default:
			return false
		}
	})
	e.set("=", func(args []interface{}) interface{} {
		a, b := args[0], args[1]
		switch av := a.(type) {
		case *big.Int:
			bv, ok := b.(*big.Int)
			return ok && av.Cmp(bv) == 0
		default:
			return a == b
		}
	})

	// Type Conversions
	e.set("float", func(args []interface{}) interface{} {
		switch v := args[0].(type) {
		case *big.Int:
			f, _ := new(big.Float).SetInt(v).Float64()
			return f
		case float64:
			return v
		case string:
			f, _ := strconv.ParseFloat(v, 64)
			return f
		default:
			panic("cannot convert to float")
		}
	})
	e.set("integer", func(args []interface{}) interface{} {
		switch v := args[0].(type) {
		case float64:
			return big.NewInt(int64(v))
		case *big.Int:
			return v
		case string:
			bi := new(big.Int)
			bi.SetString(v, 10)
			return bi
		default:
			panic("cannot convert to integer")
		}
	})

	// Type Predicates
	e.set("float?", func(args []interface{}) interface{} { _, ok := args[0].(float64); return ok })
	e.set("integer?", func(args []interface{}) interface{} { _, ok := args[0].(*big.Int); return ok })
	e.set("string?", func(args []interface{}) interface{} { _, ok := args[0].(string); return ok })
	e.set("atom?", func(args []interface{}) interface{} { _, ok := args[0].(List); return !ok })

	// List Operations
	e.set("car", func(args []interface{}) interface{} { return args[0].(List)[0] })
	e.set("cdr", func(args []interface{}) interface{} { return args[0].(List)[1:] })
	e.set("cons", func(args []interface{}) interface{} { return append(List{args[0]}, args[1].(List)...) })
	e.set("list", func(args []interface{}) interface{} { return List(args) })
	e.set("null?", func(args []interface{}) interface{} { return len(args[0].(List)) == 0 })
	e.set("length", func(args []interface{}) interface{} { return big.NewInt(int64(len(args[0].(List)))) })

	// String Operations
	e.set("concat", func(args []interface{}) interface{} {
		res := ""
		for _, arg := range args {
			if bi, ok := arg.(*big.Int); ok {
				res += bi.String()
			} else {
				res += fmt.Sprintf("%v", arg)
			}
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

	return e
}

func tokenize(s string) []string {
	var tokens []string
	var builder strings.Builder
	inString := false
	escaped := false

	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]

		if inString {
			if escaped {
				switch r {
				case 'n':
					builder.WriteRune('\n')
				case 't':
					builder.WriteRune('\t')
				case '\\':
					builder.WriteRune('\\')
				case '"':
					builder.WriteRune('"')
				default:
					builder.WriteRune('\\')
					builder.WriteRune(r)
				}
				escaped = false
			} else if r == '\\' {
				escaped = true
			} else if r == '"' {
				inString = false
				tokens = append(tokens, "\""+builder.String()+"\"")
				builder.Reset()
			} else {
				builder.WriteRune(r)
			}
			continue
		}

		if r == '"' {
			inString = true
			continue
		}

		if unicode.IsSpace(r) {
			if builder.Len() > 0 {
				tokens = append(tokens, builder.String())
				builder.Reset()
			}
		} else if r == '(' || r == ')' || r == '\'' {
			if builder.Len() > 0 {
				tokens = append(tokens, builder.String())
				builder.Reset()
			}
			tokens = append(tokens, string(r))
		} else {
			builder.WriteRune(r)
		}
	}
	if builder.Len() > 0 {
		tokens = append(tokens, builder.String())
	}
	return tokens
}

func parse(tokens []string) (interface{}, []string) {
	if len(tokens) == 0 {
		return nil, nil
	}
	token := tokens[0]
	rest := tokens[1:]

	if token == "'" {
		var item interface{}
		item, rest = parse(rest)
		return List{Symbol("quote"), item}, rest
	}
	if token == "(" {
		var list List
		for rest[0] != ")" {
			var item interface{}
			item, rest = parse(rest)
			list = append(list, item)
		}
		return list, rest[1:]
	}

	if strings.HasPrefix(token, "\"") && strings.HasSuffix(token, "\"") {
		return token[1 : len(token)-1], rest
	}

	// Big Integer or Float
	if strings.ContainsAny(token, ".eE") {
		if f, err := strconv.ParseFloat(token, 64); err == nil {
			return f, rest
		}
	} else {
		bi := new(big.Int)
		if _, ok := bi.SetString(token, 10); ok {
			return bi, rest
		}
	}

	return Symbol(token), rest
}

func eval(x interface{}, env *Env) interface{} {
	switch v := x.(type) {
	case string, float64, *big.Int:
		return v
	case Symbol:
		return env.get(v)
	case List:
		if len(v) == 0 {
			return nil
		}
		head := v[0]
		if s, ok := head.(Symbol); ok {
			switch s {
			case "quote":
				return v[1]
			case "define":
				name := v[1].(Symbol)
				val := eval(v[2], env)
				env.set(name, val)
				return nil
			case "if":
				test := eval(v[1], env).(bool)
				if test {
					return eval(v[2], env)
				}
				if len(v) > 3 {
					return eval(v[3], env)
				}
				return nil
			case "begin":
				var result interface{}
				for _, exp := range v[1:] {
					result = eval(exp, env)
				}
				return result
			case "let":
				bindings := v[1].(List)
				body := v[2]
				newEnv := &Env{vars: make(map[Symbol]interface{}), outer: env}
				for _, b := range bindings {
					bind := b.(List)
					name := bind[0].(Symbol)
					val := eval(bind[1], env)
					newEnv.set(name, val)
				}
				return eval(body, newEnv)
			case "lambda":
				params := v[1].(List)
				body := v[2]
				return func(args []interface{}) interface{} {
					newEnv := &Env{vars: make(map[Symbol]interface{}), outer: env}
					for i, p := range params {
						newEnv.set(p.(Symbol), args[i])
					}
					return eval(body, newEnv)
				}
			}
		}
		proc := eval(head, env).(func([]interface{}) interface{})
		var args []interface{}
		for _, arg := range v[1:] {
			args = append(args, eval(arg, env))
		}
		return proc(args)
	}
	return nil
}

func main() {
	env := standardEnv()
	scanner := bufio.NewScanner(os.Stdin)
	var input string
	for scanner.Scan() {
		line := scanner.Text()
		input += " " + line
		if strings.Count(input, "(") == strings.Count(input, ")") && 
		   strings.Count(input, "\"")%2 == 0 {
			tokens := tokenize(input)
			for len(tokens) > 0 {
				var exp interface{}
				exp, tokens = parse(tokens)
				result := eval(exp, env)
				if result != nil {
					if bi, ok := result.(*big.Int); ok {
						fmt.Println(bi.String())
					} else {
						fmt.Printf("%v\n", result)
					}
				}
			}
			input = ""
		}
	}
}
