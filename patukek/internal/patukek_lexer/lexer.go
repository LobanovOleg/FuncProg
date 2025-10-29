package patukek_lexer

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"patukek/internal/patukek_item"
)

type lexer struct {
	items chan patukek_item.Item
	input string
	start int
	pos   int
	width int
}

type stateFn func(*lexer) stateFn

const eof = -1

func (l *lexer) next() rune {
	var r rune
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRun(valid string) bool {
	for strings.IndexRune(valid, l.next()) >= 0 {

	}
	l.backup()
	return true
}

func (l *lexer) emit(t patukek_item.Type) {
	l.items <- patukek_item.Item{
		Typ: t,
		Val: l.input[l.start:l.pos],
		Pos: l.start,
	}
	l.start = l.pos
}

func (l *lexer) current() string {
	return l.input[l.start:l.pos]
}

func (l *lexer) ignoreSpaces() {
	l.acceptRun(" \n\t\r")
	l.ignore()
}

func (l *lexer) errorf(format string, args ...any) {
	l.items <- patukek_item.Item{
		Typ: patukek_item.Error,
		Val: fmt.Sprintf(format, args...),
		Pos: l.start,
	}
	l.start = l.pos
}

func (l *lexer) run() {
	l.ignoreSpaces()
	for state := lexExpression; state != nil; {
		state = state(l)
	}
	close(l.items)
}

func lexIdentifier(l *lexer) stateFn {
	var chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	if l.acceptRun(chars) {
		l.emit(patukek_item.Lookup(l.current()))
	}
	return lexExpression
}

func lexNumber(l *lexer) stateFn {
	var typ = patukek_item.Int
	var digits = "0123456789"

	l.accept("+-")

	l.acceptRun(digits)
	l.emit(typ)
	return lexExpression
}

func lexString(l *lexer) stateFn {
Loop:
	for {
		switch l.next() {

		case eof, '\n':
			l.errorf("unterminated quoted string")
			return nil

		case '"':
			l.backup()
			break Loop
		}
	}
	l.emit(patukek_item.String)
	l.next()
	l.ignore()
	return lexExpression
}

func lexPlus(l *lexer) stateFn {
	l.next()
	l.backup()
	l.emit(patukek_item.Plus)
	return lexExpression
}

func lexMinus(l *lexer) stateFn {
	l.next()
	l.backup()
	l.emit(patukek_item.Minus)
	return lexExpression
}

func lexTimes(l *lexer) stateFn {
	l.next()
	l.backup()
	l.emit(patukek_item.Asterisk)
	return lexExpression
}

func lexSlash(l *lexer) stateFn {
	l.next()
	l.backup()
	l.emit(patukek_item.Slash)
	return lexExpression
}

func lexMod(l *lexer) stateFn {
	l.next()
	l.backup()
	l.emit(patukek_item.Modulus)
	return lexExpression
}

func lexExpression(l *lexer) stateFn {
	switch r := l.next(); {

	case isSpace(r):
		l.ignore()

	case isLetter(r):
		l.backup()
		return lexIdentifier

	case r == '\n':
		l.emit(patukek_item.Semicolon)
		l.ignoreSpaces()

	case r == '"':
		l.ignore()
		return lexString

	case r == '(':
		l.emit(patukek_item.LParen)
		l.ignoreSpaces()

	case r == ')':
		l.emit(patukek_item.RParen)

	case r == '[':
		l.emit(patukek_item.LBracket)
		l.ignoreSpaces()

	case r == ']':
		l.emit(patukek_item.RBracket)

	case r == ',':
		l.emit(patukek_item.Comma)
		l.ignoreSpaces()

	case r == '{':
		l.emit(patukek_item.LBrace)
		l.ignoreSpaces()

	case r == '}':
		l.emit(patukek_item.RBrace)

	case r == '+':
		return lexPlus

	case r == '-':
		return lexMinus

	case r == '*':
		return lexTimes

	case r == '/':
		return lexSlash

	case r == '%':
		return lexMod

	case r == '=':
		if l.next() == '=' {
			l.emit(patukek_item.Equals)
		} else {
			l.backup()
			l.emit(patukek_item.Assign)
		}

	case r == '!':
		if l.next() == '=' {
			l.emit(patukek_item.NotEquals)
		}

	case r == '<':
		next := l.next()
		if next == '=' {
			l.emit(patukek_item.LTEQ)
		} else {
			l.backup()
			l.emit(patukek_item.LT)
		}

	case r == '>':
		next := l.next()
		if next == '=' {
			l.emit(patukek_item.GTEQ)
		} else {
			l.backup()
			l.emit(patukek_item.GT)
		}

	case r == '&':
		next := l.next()
		if next == '&' {
			l.emit(patukek_item.And)
		}

	case r == '|':
		next := l.next()
		if next == '|' {
			l.emit(patukek_item.Or)
		}

	case r == eof:
		l.emit(patukek_item.EOF)
		return nil

	default:
		if isNumber(r) {
			l.backup()
			return lexNumber
		}
		l.errorf("patukek_lexer: invalid patukek_item %q", r)
	}
	return lexExpression
}

func isLetter(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r'
}

func isNumber(r rune) bool {
	return r == '+' || r == '-' || unicode.IsNumber(r)
}

func Lex(in string) chan patukek_item.Item {
	l := &lexer{
		input: in,
		items: make(chan patukek_item.Item),
	}
	go l.run()
	return l.items
}