package patukek_parser

import (
	"patukek/internal/patukek_ast/calc_ops"
	"patukek/internal/patukek_ast/logic_ops"
	"errors"
	"strconv"

	"patukek/internal/patukek_ast"
	"patukek/internal/patukek_err"
	"patukek/internal/patukek_item"
	"patukek/internal/patukek_lexer"
)

type Parser struct {
	items         chan patukek_item.Item
	file          string
	input         string
	prefixParsers map[patukek_item.Type]parsePrefixFn
	infixParsers  map[patukek_item.Type]parseInfixFn
	cur           patukek_item.Item
	peek          patukek_item.Item
	errs          []error
	nestedLoops   uint
}

type (
	parsePrefixFn func() patukek_ast.Node
	parseInfixFn  func(patukek_ast.Node) patukek_ast.Node
)

const (
	Lowest int = iota
	Assignment
	LogicalOr
	LogicalAnd
	Equality
	Relational
	Additive
	Multiplicative
	Call
	Index
)

var precedences = map[patukek_item.Type]int{
	patukek_item.Assign:        Assignment,
	patukek_item.ModulusAssign: Assignment,
	patukek_item.Or:            LogicalOr,
	patukek_item.And:           LogicalAnd,
	patukek_item.Equals:        Equality,
	patukek_item.NotEquals:     Equality,
	patukek_item.LT:            Relational,
	patukek_item.GT:            Relational,
	patukek_item.LTEQ:          Relational,
	patukek_item.GTEQ:          Relational,
	patukek_item.Plus:          Additive,
	patukek_item.Minus:         Additive,
	patukek_item.Modulus:       Multiplicative,
	patukek_item.Slash:         Multiplicative,
	patukek_item.Asterisk:      Multiplicative,
	patukek_item.LParen:        Call,
	patukek_item.LBracket:      Index,
}

func newParser(file, input string, items chan patukek_item.Item) *Parser {
	p := &Parser{
		cur:           <-items,
		peek:          <-items,
		items:         items,
		file:          file,
		input:         input,
		prefixParsers: make(map[patukek_item.Type]parsePrefixFn),
		infixParsers:  make(map[patukek_item.Type]parseInfixFn),
	}
	p.registerPrefix(patukek_item.Ident, p.parseIdentifier)
	p.registerPrefix(patukek_item.Int, p.parseInteger)
	p.registerPrefix(patukek_item.String, p.parseString)
	p.registerPrefix(patukek_item.LParen, p.parseGroupedExpr)
	p.registerPrefix(patukek_item.If, p.parseIfExpr)
	p.registerPrefix(patukek_item.Function, p.parseFunction)
	p.registerPrefix(patukek_item.LBracket, p.parseList)
	p.registerPrefix(patukek_item.Error, p.parseError)

	p.registerInfix(patukek_item.Equals, p.parseEquals)
	p.registerInfix(patukek_item.NotEquals, p.parseNotEquals)
	p.registerInfix(patukek_item.LT, p.parseLess)
	p.registerInfix(patukek_item.GT, p.parseGreater)
	p.registerInfix(patukek_item.LTEQ, p.parseLessEq)
	p.registerInfix(patukek_item.GTEQ, p.parseGreaterEq)
	p.registerInfix(patukek_item.And, p.parseAnd)
	p.registerInfix(patukek_item.Or, p.parseOr)
	p.registerInfix(patukek_item.Plus, p.parsePlus)
	p.registerInfix(patukek_item.Minus, p.parseMinus)
	p.registerInfix(patukek_item.Slash, p.parseSlash)
	p.registerInfix(patukek_item.Asterisk, p.parseAsterisk)
	p.registerInfix(patukek_item.Modulus, p.parseModulus)
	p.registerInfix(patukek_item.Assign, p.parseAssign)
	p.registerInfix(patukek_item.LParen, p.parseCall)

	return p
}

func (p *Parser) next() {
	p.cur = p.peek
	p.peek = <-p.items
}

func (p *Parser) errors() []error {
	return p.errs
}

func (p *Parser) errorf(s string, a ...any) {
	p.errs = append(p.errs, patukek_err.New(p.file, p.input, p.cur.Pos, s, a...))
}

func (p *Parser) parse() patukek_ast.Node {
	var block = patukek_ast.NewBlock()

	for !p.cur.Is(patukek_item.EOF) {
		if s := p.parseStatement(); s != nil {
			block.Add(s)
		}
		p.next()
	}
	return &block
}

func (p *Parser) parseStatement() patukek_ast.Node {
	if p.cur.Is(patukek_item.Return) {
		return p.parseReturn()
	}
	return p.parseExpr(Lowest)
}

func (p *Parser) parseReturn() patukek_ast.Node {
	var ret patukek_ast.Node

	p.next()
	if !p.cur.Is(patukek_item.Semicolon) {
		ret = patukek_ast.NewReturn(p.parseExpr(Lowest), p.cur.Pos)
	}

	if p.peek.Is(patukek_item.Semicolon) {
		p.next()
	}
	return ret
}

func (p *Parser) parseExpr(precedence int) patukek_ast.Node {
	if prefixFn, ok := p.prefixParsers[p.cur.Typ]; ok {
		leftExp := prefixFn()

		for !p.peek.Is(patukek_item.Semicolon) && precedence < p.peekPrecedence() {
			if infixFn, ok := p.infixParsers[p.peek.Typ]; ok {
				p.next()
				leftExp = infixFn(leftExp)
			} else {
				break
			}
		}

		if p.peek.Is(patukek_item.Semicolon) {
			p.next()
		}
		return leftExp
	}
	p.noParsePrefixFnError(p.cur.Typ)
	return nil
}

func (p *Parser) parseGroupedExpr() patukek_ast.Node {
	p.next()
	exp := p.parseExpr(Lowest)
	if !p.expectPeek(patukek_item.RParen) {
		return nil
	}
	return exp
}

func (p *Parser) parseBlock() patukek_ast.Node {
	var block patukek_ast.Block
	p.next()

	for !p.cur.Is(patukek_item.RBrace) && !p.cur.Is(patukek_item.EOF) {
		if s := p.parseStatement(); s != nil {
			block.Add(s)
		}
		p.next()
	}

	if !p.cur.Is(patukek_item.RBrace) {
		p.peekError(patukek_item.RBrace)
		return nil
	}

	return &block
}

func (p *Parser) parseIfExpr() patukek_ast.Node {
	pos := p.cur.Pos
	p.next()
	cond := p.parseExpr(Lowest)

	if !p.expectPeek(patukek_item.LBrace) {
		return nil
	}

	body := p.parseBlock()

	var alt patukek_ast.Node
	if p.peek.Is(patukek_item.Else) {
		p.next()

		if p.peek.Is(patukek_item.If) {
			p.next()
			alt = p.parseIfExpr()
		} else {
			if !p.expectPeek(patukek_item.LBrace) {
				return nil
			}
			alt = p.parseBlock()
		}
	}

	return patukek_ast.NewIfExpr(cond, body, alt, pos)
}

func (p *Parser) parseList() patukek_ast.Node {
	nodes := p.parseNodeList(patukek_item.RBracket)
	return patukek_ast.NewList(nodes...)
}

func (p *Parser) parseFunction() patukek_ast.Node {
	pos := p.cur.Pos
	if !p.expectPeek(patukek_item.LParen) {
		return nil
	}

	params := p.parseFunctionParams()
	if !p.expectPeek(patukek_item.LBrace) {
		return nil
	}

	return patukek_ast.NewFunction(params, p.parseBlock(), pos)
}

func (p *Parser) parseFunctionParams() []patukek_ast.Identifier {
	var ret []patukek_ast.Identifier

	if p.peek.Is(patukek_item.RParen) {
		p.next()
		return ret
	}

	p.next()
	ret = append(ret, patukek_ast.NewIdentifier(p.cur.Val, p.cur.Pos))

	for p.peek.Is(patukek_item.Comma) {
		p.next()
		p.next()
		ret = append(ret, patukek_ast.NewIdentifier(p.cur.Val, p.cur.Pos))
	}

	if !p.expectPeek(patukek_item.RParen) {
		return nil
	}
	return ret
}

func (p *Parser) parseIdentifier() patukek_ast.Node {
	return patukek_ast.NewIdentifier(p.cur.Val, p.cur.Pos)
}

func (p *Parser) parseError() patukek_ast.Node {
	p.errs = append(p.errs, errors.New(p.cur.Val))
	return nil
}

func (p *Parser) parseInteger() patukek_ast.Node {
	i, err := strconv.ParseInt(p.cur.Val, 0, 64)
	if err != nil {
		p.errorf("unable to parse %q as integer", p.cur.Val)
		return nil
	}
	return patukek_ast.NewInteger(i)
}

func (p *Parser) parseString() patukek_ast.Node {
	s, err := patukek_ast.NewString(p.file, p.cur.Val, Parse, p.cur.Pos)
	if err != nil {
		p.errorf(err.Error())
		return nil
	}
	return s
}

func (p *Parser) parsePlus(left patukek_ast.Node) patukek_ast.Node {
	pos := p.cur.Pos
	prec := p.precedence()
	p.next()
	return calc_ops.NewPlus(left, p.parseExpr(prec), pos)
}

func (p *Parser) parseMinus(left patukek_ast.Node) patukek_ast.Node {
	pos := p.cur.Pos
	prec := p.precedence()
	p.next()
	return calc_ops.NewMinus(left, p.parseExpr(prec), pos)
}

func (p *Parser) parseAsterisk(left patukek_ast.Node) patukek_ast.Node {
	pos := p.cur.Pos
	prec := p.precedence()
	p.next()
	return calc_ops.NewTimes(left, p.parseExpr(prec), pos)
}

func (p *Parser) parseSlash(left patukek_ast.Node) patukek_ast.Node {
	pos := p.cur.Pos
	prec := p.precedence()
	p.next()
	return calc_ops.NewDivide(left, p.parseExpr(prec), pos)
}

func (p *Parser) parseModulus(left patukek_ast.Node) patukek_ast.Node {
	pos := p.cur.Pos
	prec := p.precedence()
	p.next()
	return calc_ops.NewMod(left, p.parseExpr(prec), pos)
}

func (p *Parser) parseEquals(left patukek_ast.Node) patukek_ast.Node {
	pos := p.cur.Pos
	prec := p.precedence()
	p.next()
	return logic_ops.NewEquals(left, p.parseExpr(prec), pos)
}

func (p *Parser) parseNotEquals(left patukek_ast.Node) patukek_ast.Node {
	pos := p.cur.Pos
	prec := p.precedence()
	p.next()
	return logic_ops.NewNotEquals(left, p.parseExpr(prec), pos)
}

func (p *Parser) parseLess(left patukek_ast.Node) patukek_ast.Node {
	pos := p.cur.Pos
	prec := p.precedence()
	p.next()
	return logic_ops.NewLess(left, p.parseExpr(prec), pos)
}

func (p *Parser) parseGreater(left patukek_ast.Node) patukek_ast.Node {
	pos := p.cur.Pos
	prec := p.precedence()
	p.next()
	return logic_ops.NewGreater(left, p.parseExpr(prec), pos)
}

func (p *Parser) parseLessEq(left patukek_ast.Node) patukek_ast.Node {
	pos := p.cur.Pos
	prec := p.precedence()
	p.next()
	return logic_ops.NewLessEq(left, p.parseExpr(prec), pos)
}

func (p *Parser) parseGreaterEq(left patukek_ast.Node) patukek_ast.Node {
	pos := p.cur.Pos
	prec := p.precedence()
	p.next()
	return logic_ops.NewGreaterEq(left, p.parseExpr(prec), pos)
}

func (p *Parser) parseAnd(left patukek_ast.Node) patukek_ast.Node {
	pos := p.cur.Pos
	prec := p.precedence()
	p.next()
	return logic_ops.NewAnd(left, p.parseExpr(prec), pos)
}

func (p *Parser) parseOr(left patukek_ast.Node) patukek_ast.Node {
	pos := p.cur.Pos
	prec := p.precedence()
	p.next()
	return logic_ops.NewOr(left, p.parseExpr(prec), pos)
}

func (p *Parser) parseAssign(left patukek_ast.Node) patukek_ast.Node {
	pos := p.cur.Pos
	p.next()
	right := p.parseExpr(Lowest)

	i, leftIsIdentifier := left.(patukek_ast.Identifier)
	fn, rightIsFunction := right.(patukek_ast.Function)

	if leftIsIdentifier && rightIsFunction {
		fn.Name = i.String()
	}

	return patukek_ast.NewAssign(left, right, pos)
}

func (p *Parser) parseCall(fn patukek_ast.Node) patukek_ast.Node {
	pos := p.cur.Pos
	return patukek_ast.NewCall(fn, p.parseNodeList(patukek_item.RParen), pos)
}

func (p *Parser) parsePair() [2]patukek_ast.Node {
	l := p.parseExpr(Lowest)
	p.next()
	r := p.parseExpr(Lowest)

	return [2]patukek_ast.Node{l, r}
}

func (p *Parser) parseNodeList(end patukek_item.Type) []patukek_ast.Node {
	return p.parseNodeSequence(patukek_item.Comma, end)
}

func (p *Parser) parseNodeSequence(sep, end patukek_item.Type) []patukek_ast.Node {
	var seq []patukek_ast.Node

	p.next()
	if p.cur.Is(end) {
		return seq
	}

	seq = append(seq, p.parseExpr(Lowest))

	for p.peek.Is(sep) {
		p.next()
		p.next()
		seq = append(seq, p.parseExpr(Lowest))
	}

	if !p.expectPeek(end) {
		return nil
	}
	return seq
}

func (p *Parser) expectPeek(t patukek_item.Type) bool {
	if p.peek.Is(t) {
		p.next()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) peekError(t patukek_item.Type) {
	p.errorf("expected next patukek_item to be %v, got %v instead", t, p.peek.Typ)
}

func (p *Parser) peekPrecedence() int {
	if prec, ok := precedences[p.peek.Typ]; ok {
		return prec
	}
	return Lowest
}

func (p *Parser) precedence() int {
	if prec, ok := precedences[p.cur.Typ]; ok {
		return prec
	}
	return Lowest
}

func (p *Parser) registerPrefix(typ patukek_item.Type, fn parsePrefixFn) {
	p.prefixParsers[typ] = fn
}

func (p *Parser) registerInfix(typ patukek_item.Type, fn parseInfixFn) {
	p.infixParsers[typ] = fn
}

func (p *Parser) noParsePrefixFnError(t patukek_item.Type) {
	p.errorf("no parse prefix function for %q found", t)
}

func Parse(file, input string) (prog patukek_ast.Node, errs []error) {
	items := patukek_lexer.Lex(input)
	p := newParser(file, input, items)
	return p.parse(), p.errors()
}