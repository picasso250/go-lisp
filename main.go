package main

import (
	"bufio"
	"fmt"
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
	// Arithmetic
	e.set("+", func(args []interface{}) interface{} { return args[0].(float64) + args[1].(float64) })
	e.set("-", func(args []interface{}) interface{} { return args[0].(float64) - args[1].(float64) })
	e.set("*", func(args []interface{}) interface{} { return args[0].(float64) * args[1].(float64) })
	e.set("/", func(args []interface{}) interface{} { return args[0].(float64) / args[1].(float64) })
	
	// Comparisons
	e.set("<", func(args []interface{}) interface{} { return args[0].(float64) < args[1].(float64) })
	e.set(">", func(args []interface{}) interface{} { return args[0].(float64) > args[1].(float64) })
	e.set("<=", func(args []interface{}) interface{} { return args[0].(float64) <= args[1].(float64) })
	e.set(">=", func(args []interface{}) interface{} { return args[0].(float64) >= args[1].(float64) })
	e.set("=", func(args []interface{}) interface{} { return args[0] == args[1] })

	// List Operations
	e.set("car", func(args []interface{}) interface{} { return args[0].(List)[0] })
	e.set("cdr", func(args []interface{}) interface{} { return args[0].(List)[1:] })
	e.set("cons", func(args []interface{}) interface{} {
		return append(List{args[0]}, args[1].(List)...)
	})
	e.set("list", func(args []interface{}) interface{} { return List(args) })
	e.set("null?", func(args []interface{}) interface{} { return len(args[0].(List)) == 0 })
	e.set("length", func(args []interface{}) interface{} { return float64(len(args[0].(List))) })
	
	// Atom Check
	e.set("atom?", func(args []interface{}) interface{} {
		_, isList := args[0].(List)
		return !isList
	})

	// String Operations
	e.set("concat", func(args []interface{}) interface{} {
		res := ""
		for _, arg := range args {
			res += fmt.Sprintf("%v", arg)
		}
		return res
	})

	return e
}

// tokenize supports strings like "hello world"
func tokenize(s string) []string {
	var tokens []string
	var builder strings.Builder
	inString := false

	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == '"' {
			inString = !inString
			builder.WriteRune(r)
			if !inString {
				tokens = append(tokens, builder.String())
				builder.Reset()
			}
			continue
		}

		if inString {
			builder.WriteRune(r)
			continue
		}

		if r == '(' || r == ')' || r == '\'' {
			if builder.Len() > 0 {
				tokens = append(tokens, builder.String())
				builder.Reset()
			}
			tokens = append(tokens, string(r))
		} else if unicode.IsSpace(r) {
			if builder.Len() > 0 {
				tokens = append(tokens, builder.String())
				builder.Reset()
			}
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

	// Handle Strings
	if strings.HasPrefix(token, "\"") && strings.HasSuffix(token, "\"") {
		return token[1 : len(token)-1], rest
	}
	
	// Handle Numbers
	if f, err := strconv.ParseFloat(token, 64); err == nil {
		return f, rest
	}
	
	// Handle Symbols
	return Symbol(token), rest
}

func eval(x interface{}, env *Env) interface{} {
	switch v := x.(type) {
	case string: // Literal String
		return v
	case float64: // Number
		return v
	case Symbol: // Symbol Lookup
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
		// Function call
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
			if len(tokens) == 0 { continue }
			for len(tokens) > 0 {
				var exp interface{}
				exp, tokens = parse(tokens)
				result := eval(exp, env)
				if result != nil {
					fmt.Printf("%v\n", result)
				}
			}
			input = ""
		}
	}
}
