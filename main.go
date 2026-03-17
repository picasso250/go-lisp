package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"math/big"
	"os"
	"strconv"
	"strings"
	"unicode"
)

type Symbol string
type List []interface{}
type Dict map[string]interface{}
type TCPListener interface{}
type TCPConn interface{}

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
	return nil
}

func (e *Env) set(s Symbol, val interface{}) {
	e.vars[s] = val
}

func (e *Env) update(s Symbol, val interface{}) bool {
	if _, ok := e.vars[s]; ok {
		e.vars[s] = val
		return true
	}
	if e.outer != nil {
		return e.outer.update(s, val)
	}
	return false
}

// tokenize supports strings like "hello world" and escapes
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
				case 'r':
					builder.WriteRune('\r')
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

		if r == ';' {
			// Skip until newline
			if builder.Len() > 0 {
				tokens = append(tokens, builder.String())
				builder.Reset()
			}
			for i+1 < len(runes) && runes[i+1] != '\n' {
				i++
			}
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

	if token == "def" {
		list := List{Symbol("define")}
		for len(rest) > 0 && rest[0] != "end" {
			var item interface{}
			item, rest = parse(rest)
			list = append(list, item)
		}
		if len(rest) == 0 {
			panic("unexpected EOF: missing end")
		}
		return list, rest[1:]
	}

	if token == "'" {
		var item interface{}
		item, rest = parse(rest)
		return List{Symbol("quote"), item}, rest
	}
	if token == "(" {
		var list List
		for len(rest) > 0 && rest[0] != ")" {
			var item interface{}
			item, rest = parse(rest)
			list = append(list, item)
		}
		if len(rest) == 0 {
			panic("unexpected EOF: missing )")
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
	case string, float64, *big.Int, bool:
		return v
	case Symbol:
		if v == "true" {
			return true
		}
		if v == "false" {
			return false
		}
		return env.get(v)
	case List:
		if len(v) == 0 {
			return nil
		}
		head := v[0]
		if s, ok := head.(Symbol); ok {
			switch s {
			case "set!":
				name := v[1].(Symbol)
				val := eval(v[2], env)
				if !env.update(name, val) {
					panic(fmt.Sprintf("set!: undefined variable %s", name))
				}
				return nil
			case "quote":
				return v[1]
			case "define":
				if head, ok := v[1].(List); ok {
					name := head[0].(Symbol)
					params := head[1:]
					var body interface{}
					if len(v) > 3 {
						body = append(List{Symbol("begin")}, v[2:]...)
					} else {
						body = v[2]
					}
					val := func(args []interface{}) interface{} {
						newEnv := &Env{vars: make(map[Symbol]interface{}), outer: env}
						for i, p := range params {
							newEnv.set(p.(Symbol), args[i])
						}
						return eval(body, newEnv)
					}
					env.set(name, val)
					return nil
				}
				name := v[1].(Symbol)
				val := eval(v[2], env)
				env.set(name, val)
				return nil
			case "if":
				test := eval(v[1], env)
				if test == true {
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
			case "and":
				for _, exp := range v[1:] {
					res := eval(exp, env)
					if res == false {
						return false
					}
				}
				return true
			case "or":
				for _, exp := range v[1:] {
					res := eval(exp, env)
					if res == true {
						return true
					}
				}
				return false
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
		procRaw := eval(head, env)
		if procRaw == nil {
			panic(fmt.Sprintf("Symbol not found: %v", head))
		}
		proc := procRaw.(func([]interface{}) interface{})
		var args []interface{}
		for _, arg := range v[1:] {
			args = append(args, eval(arg, env))
		}
		return proc(args)
	}
	return nil
}

func evalString(w io.Writer, input string, env *Env, quiet bool) interface{} {
	tokens := tokenize(input)
	var lastResult interface{}
	for len(tokens) > 0 {
		var exp interface{}
		exp, tokens = parse(tokens)
		lastResult = eval(exp, env)
		if !quiet && lastResult != nil {
			if bi, ok := lastResult.(*big.Int); ok {
				fmt.Fprintln(w, bi.String())
			} else {
				fmt.Fprintf(w, "%v\n", lastResult)
			}
		}
	}
	return lastResult
}

func main() {
	fileFlag := flag.String("f", "", "Execute file")
	codeFlag := flag.String("c", "", "Execute code string")
	flag.Parse()

	env := standardEnv()

	// Load stdlib if exists
	if content, err := os.ReadFile("stdlib.lisp"); err == nil {
		evalString(io.Discard, string(content), env, true)
	}

	if *codeFlag != "" {
		evalString(os.Stdout, *codeFlag, env, false)
		return
	}

	if *fileFlag != "" {
		content, err := os.ReadFile(*fileFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}
		evalString(os.Stdout, string(content), env, false)
		return
	}

	// Default to REPL
	scanner := bufio.NewScanner(os.Stdin)
	var input string
	for scanner.Scan() {
		line := scanner.Text()
		input += " " + line
		if strings.Count(input, "(") == strings.Count(input, ")") &&
			strings.Count(input, "\"")%2 == 0 {
			evalString(os.Stdout, input, env, false)
			input = ""
		}
	}
}
