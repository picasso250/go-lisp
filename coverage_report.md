# Coverage Report

## Summary

| File | Coverage | Stmts | Covered | Missed |
| :--- | :--- | :--- | :--- | :--- |
| builtin.go | 98.3% | 232 | 228 | 4 |
| main.go | 92.9% | 155 | 144 | 11 |
| **Total** | **96.1%** | **387** | **372** | **15** |

## Missed Details

### builtin.go

| Lines | Uncovered Code |
| :--- | :--- |
| L30-L31 | `panic(fmt.Sprintf("%s: invalid types", name))` |
| L48-L49 | `panic(fmt.Sprintf("%s: incomparable type", name))` |
| L199-L201 | `{ return l }` |
| L217-L218 | `return false` |

### main.go

| Lines | Uncovered Code |
| :--- | :--- |
| L250-L254 | `{ env := standardEnv() // Load stdlib if exists if content, err := os.ReadFile("stdlib.lisp"); err == nil` |
| L254-L256 | `{ evalString(io.Discard, string(content), env, true) }` |
| L258-L260 | `scanner := bufio.NewScanner(os.Stdin) var input string for scanner.Scan()` |
| L260-L264 | `{ line := scanner.Text() input += " " + line if strings.Count(input, "(") == strings.Count(input, ")") && strings.Count(input, "\"")%2 == 0` |
| L264-L267 | `{ evalString(os.Stdout, input, env, false) input = "" }` |

