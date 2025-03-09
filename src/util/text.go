package util

import (
	"regexp"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var wordRegex = regexp.MustCompile(`[\p{Lo}\p{Ll}\p{Lu}\p{Nd}\p{Nl}\p{No}]+`)

func NormalizeUnicode(input string) (string, error) {
	normalize := transform.Chain(norm.NFKD, runes.Remove(runes.In(unicode.Mn)), norm.NFKC)
	result, _, err := transform.String(normalize, input)
	return result, err
}

func SplitWords(input string) []string {
	words := wordRegex.FindAllString(input, 99)
	if words == nil {
		words = []string{}
	}

	return words
}
