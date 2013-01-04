package terp

import (
	"bytes"
	. "fmt"
	"log"
	"strconv"
)

func (fr *Frame) initExpr() {
	Builtins["expr"] = cmdExpr
}

func cmdExpr(fr *Frame, argv []T) T {
	// Just support 1 arg expressions for now.  We'll concat later.
	ex := Arg1(argv)

	return fr.EvalExpr(ex)
}

// Takes a single word that represents an expression and returns the result.
func (fr *Frame) EvalExpr(a T) (result T) {
	return fr.ParseExpression(a.String())
}

func isNumeric(ch uint8) bool {
	return '0' <= ch && ch <= '9'
}

func isOperator(ch uint8) bool {
	return ch == '+' || ch == '-' || ch == '~' || ch == '!' || ch == '*' ||
		ch == '/' || ch == '%' || ch == '<' || ch == '>' || ch == '=' ||
		ch == '&' || ch == '|' || ch == '^' || ch == '?' || ch == ':'
}

// Takes the string that represents an expression and returns the result.
func (fr *Frame) ParseExpression(s string) (result T) {
	log.Printf("ParseExpression <- %q", s)

	t, _ := fr.ParseExprRel(s)

	return t
}

func (fr *Frame) ParseExprRel(s string) (T, string) {
	log.Printf("ParseExprRel <- %q", s)
	i := 0
	n := len(s)
	var op [2]uint8
	var z T = Empty
	var lookForOp bool = false

Loop:
	for i < n {
		c := s[i]

		if lookForOp == true {
			switch {
			case c == '!', c == '=', c == '<',  c == '>':
				i++
				peek := s[i]
				if peek == '=' {
					op = [2]uint8{c, peek}
					i++
				} else {
					op = [2]uint8{c, 0}
				}
				lookForOp = false
			case White(c):
				i++
			default:
				break Loop
			}
		} else {
			t, rest := fr.ParseExprTerm(s[i:])
			s = rest
			n = len(s)
			i = 0
			lookForOp = true

			if t == Empty {
				break Loop
			}

			if z == Empty {
				z = t
			} else {
				switch op {
				case [2]uint8{'!', '='}:
					z = MkBool(z.Float() != t.Float())

				case [2]uint8{'=', '='}:
					z = MkBool(z.Float() == t.Float())

				case [2]uint8{'<', 0}:
					z = MkBool(z.Float() < t.Float())

				case [2]uint8{'<', '='}:
					z = MkBool(z.Float() <= t.Float())

				case [2]uint8{'>', 0}:
					z = MkBool(z.Float() > t.Float())

				case [2]uint8{'>', '='}:
					z = MkBool(z.Float() >= t.Float())
				}
			}
		}
	}

	return z, s[i:]
}

func (fr *Frame) ParseExprTerm(s string) (T, string) {
	log.Printf("ParseExprTerm <- %q", s)
	i := 0
	n := len(s)
	var op uint8 = 0
	var z T = Empty
	var lookForOp bool = false

Loop:
	for i < n {
		c := s[i]

		if lookForOp == true {
			switch {
			case c == '+', c == '-', c == '|',  c == '^':
				i++
				op = c
				lookForOp = false
			case White(c):
				i++
			default:
				break Loop
			}
		} else {
			t, rest := fr.ParseExprFactor(s[i:])
			s = rest
			n = len(s)
			i = 0
			lookForOp = true

			if t == Empty {
				break Loop
			}

			if z == Empty {
				z = t
			} else {
				switch op {
				case '+':
					z = MkFloat(z.Float() + t.Float())

				case '-':
					z = MkFloat(z.Float() - t.Float())

				case '|':
					z = z // TODO

				case '^':
					z = z // TODO
				}
			}
		}
	}

	return z, s[i:]
}

func (fr *Frame) ParseExprFactor(s string) (T, string) {
	log.Printf("ParseExprFactor <- %q", s)
	i := 0
	n := len(s)
	var op uint8 = 0
	var z T = Empty
	var lookForOp bool = false

Loop:
	for i < n {
		c := s[i]

		if lookForOp == true {
			switch {
			case c == '*', c == '/', c == '%',  c == '&':
				i++
				op = c
				lookForOp = false
			case White(c):
				i++
			default:
				break Loop
			}
		} else {
			t, rest := fr.ParseExprPrim(s[i:])
			s = rest
			n = len(s)
			i = 0
			lookForOp = true

			if t == Empty {
				break Loop
			}

			if z == Empty {
				z = t
			} else {
				switch op {
				case '*':
					z = MkFloat(z.Float() * t.Float())

				case '/':
					z = MkFloat(z.Float() / t.Float())

				case '%':
					z = z // TODO

				case '&':
					z = z // TODO
				}
			}
		}
	}

	return z, s[i:]
}

func (fr *Frame) ParseExprPrim(s string) (T, string) {
	log.Printf("ParseExprPrim <- %q", s)
	i := 0
	n := len(s)

	for i < n {
		c := s[i]

		switch {
		case c == '[':
			return fr.ParseSquare(s[i:])

		case c == '{':
			return fr.ParseCurly(s[i:])

		case c == '$':
			return fr.ParseDollar(s[i:])

		case c == '"':
			return fr.ParseQuote(s[i:])

		case isNumeric(c) || c == '.' && isNumeric(s[i+1]):
			return fr.ParseNumber(s[i:])

		case isOperator(c):
			return Empty, s[i:]
		}

		i++
	}

	panic("ParseExprPrim: No primitive found.")
}

func (fr *Frame) ParseNumber(s string) (T, string) {
	log.Printf("ParseNumber <- %q", s)
	i := 0
	n := len(s)
	decimal := false // only allow one decimal per number
	buf := bytes.NewBuffer(nil)

Loop:
	for i < n {
		c := s[i]

		switch {
		case isNumeric(c) || c == '.' && decimal == false:
			if c == '.' {
				decimal = true
			}

			buf.WriteByte(c)
			i++

		// An operator or whitespace signifies the end of the number.
		case isOperator(c) || White(c):
			break Loop

		default:
			panic(Sprintf("ParseNumber: Unexpected character, '%c', found while parsing number.", c))
		}
	}

	vstr := buf.String()

	if v, ok := strconv.ParseFloat(vstr, 64); ok == nil {
		return MkFloat(v), s[i:]
	}

	panic(Sprintf("ParseNumber: Could not parse float from: %s", vstr))
}
