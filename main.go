package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Symbol string
type List []interface{}

type Env struct {
	vars   map[Symbol]interface{}
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
	e.set("+", func(args []interface{}) interface{} { return args[0].(float64) + args[1].(float64) })
	e.set("-", func(args []interface{}) interface{} { return args[0].(float64) - args[1].(float64) })
	e.set("*", func(args []interface{}) interface{} { return args[0].(float64) * args[1].(float64) })
	e.set("/", func(args []interface{}) interface{} { return args[0].(float64) / args[1].(float64) })
	e.set("<=", func(args []interface{}) interface{} { return args[0].(float64) <= args[1].(float64) })
	e.set(">", func(args []interface{}) interface{} { return args[0].(float64) > args[1].(float64) })
	return e
}

func tokenize(s string) []string {
	s = strings.ReplaceAll(s, "(", " ( ")
	s = strings.ReplaceAll(s, ")", " ) ")
	return strings.Fields(s)
}

func parse(tokens []string) (interface{}, []string) {
	if len(tokens) == 0 {
		return nil, nil
	}
	token := tokens[0]
	rest := tokens[1:]
	if token == "(" {
		var list List
		for rest[0] != ")" {
			var item interface{}
			item, rest = parse(rest)
			list = append(list, item)
		}
		return list, rest[1:]
	}
	if f, err := strconv.ParseFloat(token, 64); err == nil {
		return f, rest
	}
	return Symbol(token), rest
}

func eval(x interface{}, env *Env) interface{} {
	switch v := x.(type) {
	case Symbol:
		return env.get(v)
	case float64:
		return v
	case List:
		if len(v) == 0 {
			return nil
		}
		head := v[0]
		if s, ok := head.(Symbol); ok {
			switch s {
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
				return eval(v[3], env)
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
		if strings.Count(input, "(") == strings.Count(input, ")") && strings.Contains(input, "(") {
			tokens := tokenize(input)
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
