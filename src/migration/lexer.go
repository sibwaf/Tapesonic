package migration

import (
	"fmt"
	"io"
	"unicode"
)

// why do i have to do this...

var symbolicOperators = map[string]bool{
	"->>": true,
	"||":  true,
	"->":  true,
	"!=":  true,
	"<>":  true,
	"==":  true,
	"<=":  true,
	">=":  true,
	"=":   true,
	">":   true,
	"<":   true,
	"~":   true,
	"+":   true,
	"-":   true,
	"*":   true,
	"/":   true,
	"%":   true,
	"&":   true,
	"|":   true,
	"<<":  true,
	">>":  true,
	";":   true,
	".":   true,
	"(":   true,
	")":   true,
	"[":   true,
	"]":   true,
}

type predicate[T any] func(T) bool

func not[T any](p predicate[T]) predicate[T] {
	return func(t T) bool {
		return !p(t)
	}
}

func eq[T comparable](expected T) predicate[T] {
	return func(actual T) bool {
		return actual == expected
	}
}

type lexer struct {
	runes    []rune
	position int
}

func (lexer *lexer) consumeRune() *rune {
	if lexer.position < len(lexer.runes) {
		result := &lexer.runes[lexer.position]
		lexer.position += 1
		return result
	} else {
		return nil
	}
}

func (lexer *lexer) peekRune() *rune {
	if lexer.position < len(lexer.runes) {
		return &lexer.runes[lexer.position]
	} else {
		return nil
	}
}

func (lexer *lexer) tryConsume(token string) string {
	consumeCount := 0
	for _, expectedRune := range token {
		actualRune := lexer.consumeRune()

		if actualRune != nil {
			consumeCount += 1
		}

		if actualRune == nil || *actualRune != expectedRune {
			lexer.position -= consumeCount
			return ""
		}
	}

	return token
}

func (lexer *lexer) consumeUntil(stop func(rune) bool) []rune {
	result := []rune{}
	for {
		nextRune := lexer.consumeRune()
		if nextRune == nil {
			break
		}

		if stop(*nextRune) {
			lexer.position -= 1
			break
		} else {
			result = append(result, *nextRune)
		}
	}

	return result
}

func (lexer *lexer) NextToken() (string, error) {
	lexer.consumeUntil(not(unicode.IsSpace))

	if lexer.peekRune() == nil {
		return "", io.EOF
	}

	wasComment := lexer.tryConsumeLineComment()
	if wasComment {
		return lexer.NextToken()
	}

	stringLiteral, err := lexer.tryConsumeStringLiteral()
	if err != nil {
		return "", err
	}
	if stringLiteral != "" {
		return stringLiteral, nil
	}

	if word := lexer.tryConsumeWord(); word != "" {
		return word, nil
	}

	if operator := lexer.tryConsumeSymbolicOperator(); operator != "" {
		return operator, nil
	}

	result := lexer.consumeUntil(unicode.IsSpace)
	return string(result), nil
}

func (lexer *lexer) tryConsumeLineComment() bool {
	commentStart := lexer.tryConsume("--")
	if commentStart == "" {
		return false
	} else {
		lexer.consumeUntil(eq('\n'))
		return true
	}
}

func (lexer *lexer) tryConsumeStringLiteral() (string, error) {
	startQuote := lexer.tryConsume("'")
	if startQuote == "" {
		return "", nil
	}

	result := []rune{'\''}

	for {
		content := lexer.consumeUntil(eq('\''))

		closingQuote := lexer.consumeRune()
		if closingQuote == nil {
			return "", fmt.Errorf("unexpected string literal end")
		}

		result = append(result, content...)
		result = append(result, '\'')

		nextRune := lexer.peekRune()
		if nextRune != nil && *nextRune == '\'' { // it was an escaped quote, not a string literal end
			result = append(result, *lexer.consumeRune())
			continue
		} else {
			break
		}
	}

	return string(result), nil
}

func (lexer *lexer) tryConsumeQuotedIdentifier() (string, error) {
	startQuote := lexer.tryConsume("\"")
	if startQuote == "" {
		return "", nil
	}

	result := []rune{'"'}

	content := lexer.consumeUntil(eq('"'))

	closingQuote := lexer.consumeRune()
	if closingQuote == nil {
		return "", fmt.Errorf("unexpected quoted identifier end")
	}

	result = append(result, content...)
	result = append(result, '"')

	return string(result), nil
}

func (lexer *lexer) tryConsumeWord() string {
	nextRune := lexer.peekRune()
	if nextRune == nil || (*nextRune != '_' && !unicode.IsLetter(*nextRune)) {
		return ""
	}

	isValidIdentifierRune := func (x rune) bool {
		return x == '_' || unicode.IsLetter(x) || unicode.IsNumber(x)
	}

	return string(lexer.consumeUntil(not(isValidIdentifierRune)))
}

func (lexer *lexer) tryConsumeSymbolicOperator() string {
	runes := []rune{}
	for range 3 {
		nextRune := lexer.consumeRune()
		if nextRune == nil {
			break
		}
		runes = append(runes, *nextRune)
	}

	lexer.position -= len(runes)

	for length := len(runes); length > 0; length -= 1 {
		maybeOperator := string(runes[:length])
		_, matchFound := symbolicOperators[maybeOperator]
		if matchFound {
			return lexer.tryConsume(maybeOperator)
		}
	}

	return ""
}
