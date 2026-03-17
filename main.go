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
	return nil
}

func (e *Env) set(s Symbol, val interface{}) {
	e.vars[s] = val
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
	case string, float64, *big.Int, bool:
		return v
	case Symbol:
		if v == "true" { return true }
		if v == "false" { return false }
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
		proc := eval(head, env).(func([]interface{}) interface{})
		var args []interface{}
		for _, arg := range v[1:] {
			args = append(args, eval(arg, env))
		}
		return proc(args)
	}
	return nil
}

func evalString(input string, env *Env, quiet bool) {
	tokens := tokenize(input)
	for len(tokens) > 0 {
		var exp interface{}
		exp, tokens = parse(tokens)
		result := eval(exp, env)
		if !quiet && result != nil {
			if bi, ok := result.(*big.Int); ok {
				fmt.Println(bi.String())
			} else {
				fmt.Printf("%v\n", result)
			}
		}
	}
}

func main() {
	env := standardEnv()
	
	// Load stdlib if exists
	if content, err := os.ReadFile("stdlib.lisp"); err == nil {
		evalString(string(content), env, true)
	}

	scanner := bufio.NewScanner(os.Stdin)
	var input string
	for scanner.Scan() {
		line := scanner.Text()
		input += " " + line
		if strings.Count(input, "(") == strings.Count(input, ")") && 
		   strings.Count(input, "\"")%2 == 0 {
			evalString(input, env, false)
			input = ""
		}
	}
}
