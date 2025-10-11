package cmd

import (
	stdhtml "html"
	"io"
	"strings"
	"unicode/utf8"

	htmlparser "golang.org/x/net/html"
)

func sanitizeLabel(raw string) string {
	unescaped := unescapeLabelEscapes(raw)
	if unescaped == "" {
		return ""
	}

	decoded := stdhtml.UnescapeString(unescaped)
	stripped := stripHTMLTags(decoded)
	collapsed := collapseWhitespace(stripped)
	if collapsed == "" {
		return ""
	}

	replacer := strings.NewReplacer("’", "'", "‘", "'")
	return replacer.Replace(collapsed)
}

func unescapeLabelEscapes(raw string) string {
	if raw == "" {
		return ""
	}

	var builder strings.Builder
	builder.Grow(len(raw))

	for i := 0; i < len(raw); {
		r, size := utf8.DecodeRuneInString(raw[i:])
		if r == '\\' {
			if i+size >= len(raw) {
				i += size
				continue
			}

			nextRune, nextSize := utf8.DecodeRuneInString(raw[i+size:])
			switch nextRune {
			case '\\':
				if i+size+nextSize < len(raw) {
					thirdRune, thirdSize := utf8.DecodeRuneInString(raw[i+size+nextSize:])
					if thirdRune == '\'' || thirdRune == '’' || thirdRune == '‘' {
						builder.WriteRune(thirdRune)
						i += size + nextSize + thirdSize
						continue
					}
				}
				builder.WriteRune('\\')
				i += size + nextSize
				continue
			case '"', '\'', '[', ']', '(', ')', '{', '}', '<', '>', '/':
				builder.WriteRune(nextRune)
				i += size + nextSize
				continue
			case 'n', 'r', 't':
				builder.WriteRune(' ')
				i += size + nextSize
				continue
			case '’', '‘':
				builder.WriteRune(nextRune)
				i += size + nextSize
				continue
			default:
				builder.WriteRune(nextRune)
				i += size + nextSize
				continue
			}
		}

		if r == '\n' || r == '\r' {
			builder.WriteRune(' ')
		} else {
			builder.WriteRune(r)
		}
		i += size
	}

	return builder.String()
}

func stripHTMLTags(input string) string {
	tokenizer := htmlparser.NewTokenizer(strings.NewReader(input))
	var builder strings.Builder

	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case htmlparser.ErrorToken:
			if tokenizer.Err() == io.EOF {
				return builder.String()
			}
			return builder.String()
		case htmlparser.TextToken:
			builder.WriteString(string(tokenizer.Text()))
		case htmlparser.StartTagToken, htmlparser.EndTagToken, htmlparser.SelfClosingTagToken:
			name, _ := tokenizer.TagName()
			if strings.EqualFold(string(name), "br") {
				builder.WriteRune(' ')
			}
		}
	}
}

func collapseWhitespace(s string) string {
	if s == "" {
		return ""
	}

	normalized := strings.ReplaceAll(s, "\u00a0", " ")
	fields := strings.Fields(normalized)
	return strings.Join(fields, " ")
}
