package migration

import (
	"errors"
	"io"
	"strings"
)

type parsedStatement struct {
	Sql string
}

type parsedScript struct {
	Statements []parsedStatement
}

type parser struct {
	lexer lexer
}

func (parser *parser) parseScript() (parsedScript, error) {
	statements := []parsedStatement{}

	makeStatement := func(tokens []string) parsedStatement {
		return parsedStatement{
			Sql: strings.Join(tokens, " "),
		}
	}

	tokens := []string{}
	for {
		nextToken, err := parser.lexer.NextToken()
		if err != nil && !errors.Is(err, io.EOF) {
			return parsedScript{}, err
		}

		if errors.Is(err, io.EOF) {
			break
		} else if nextToken == ";" {
			statements = append(statements, makeStatement(tokens))
			tokens = []string{}
		} else {
			tokens = append(tokens, nextToken)
		}
	}

	if len(tokens) > 0 {
		statements = append(statements, makeStatement(tokens))
	}

	return parsedScript{Statements: statements}, nil
}
