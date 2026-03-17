package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
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

type TailCall struct {
	exp interface{}
	env *Env
}

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

	// Integer or Float
	if strings.ContainsAny(token, ".eE") {
		if f, err := strconv.ParseFloat(token, 64); err == nil {
			return f, rest
		}
	} else {
		if i, err := strconv.ParseInt(token, 10, 64); err == nil {
			return i, rest
		}
	}

	return Symbol(token), rest
}

func eval(x interface{}, env *Env) interface{} {
	for {
		switch v := x.(type) {
		case string, float64, int64, bool:
			return v
		case Symbol:
			if v == "true" {
				return true
			}
			if v == "false" {
				return false
			}
			if v == "nil" {
				return nil
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
					if h, ok := v[1].(List); ok {
						name := h[0].(Symbol)
						params := h[1:]
						var body interface{}
						if len(v) > 3 {
							body = append(List{Symbol("begin")}, v[2:]...)
						} else {
							body = v[2]
						}
						capturedEnv := env
						val := func(args []interface{}) interface{} {
							if len(args) != len(params) {
								panic(fmt.Sprintf("%s: expected %d arguments, got %d", name, len(params), len(args)))
							}
							newEnv := &Env{vars: make(map[Symbol]interface{}), outer: capturedEnv}
							for i, p := range params {
								newEnv.set(p.(Symbol), args[i])
							}
							return TailCall{exp: body, env: newEnv}
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
						x = v[2]
					} else if len(v) > 3 {
						x = v[3]
					} else {
						return nil
					}
					continue
				case "begin":
					for i := 1; i < len(v)-1; i++ {
						eval(v[i], env)
					}
					x = v[len(v)-1]
					continue
				case "and":
					if len(v) == 1 {
						return true
					}
					var res interface{} = true
					for i := 1; i < len(v)-1; i++ {
						res = eval(v[i], env)
						if res == false || res == nil {
							return res
						}
					}
					x = v[len(v)-1]
					continue
				case "or":
					if len(v) == 1 {
						return false
					}
					var res interface{} = false
					for i := 1; i < len(v)-1; i++ {
						res = eval(v[i], env)
						if res != false && res != nil {
							return res
						}
					}
					x = v[len(v)-1]
					continue
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
					x = body
					env = newEnv
					continue
				case "lambda":
					params := v[1].(List)
					body := v[2]
					capturedEnv := env
					return func(args []interface{}) interface{} {
						if len(args) != len(params) {
							panic(fmt.Sprintf("lambda: expected %d arguments, got %d", len(params), len(args)))
						}
						newEnv := &Env{vars: make(map[Symbol]interface{}), outer: capturedEnv}
						for i, p := range params {
							newEnv.set(p.(Symbol), args[i])
						}
						return TailCall{exp: body, env: newEnv}
					}
				}
			}
			// Function application
			procRaw := eval(head, env)
			proc, ok := procRaw.(func([]interface{}) interface{})
			if !ok {
				panic(fmt.Sprintf("Not a function: %v", head))
			}
			var args []interface{}
			for i := 1; i < len(v); i++ {
				argVal := eval(v[i], env)
				// Resolve TailCall from argument evaluation
				for {
					if tc, ok := argVal.(TailCall); ok {
						argVal = eval(tc.exp, tc.env)
					} else {
						break
					}
				}
				args = append(args, argVal)
			}
			result := proc(args)
			if tc, ok := result.(TailCall); ok {
				x = tc.exp
				env = tc.env
				continue
			}
			return result
		default:
			return v
		}
	}
}

func evalString(w io.Writer, input string, env *Env, quiet bool) interface{} {
	tokens := tokenize(input)
	var lastResult interface{}
	for {
		if len(tokens) == 0 {
			break
		}
		var exp interface{}
		exp, tokens = parse(tokens)
		lastResult = eval(exp, env)
		// Resolve final TailCall
		for {
			if tc, ok := lastResult.(TailCall); ok {
				lastResult = eval(tc.exp, tc.env)
			} else {
				break
			}
		}
		if !quiet && lastResult != nil {
			fmt.Fprintf(w, "%v\n", lastResult)
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
