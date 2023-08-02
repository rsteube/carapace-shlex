/*
Copyright 2012 Google Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
Package shlex implements a simple lexer which splits input in to tokens using
shell-style rules for quoting and commenting.

The basic use case uses the default ASCII lexer to split a string into sub-strings:

	shlex.Split("one \"two three\" four") -> []string{"one", "two three", "four"}

To process a stream of strings:

	l := NewLexer(os.Stdin)
	for ; token, err := l.Next(); err != nil {
		// process token
	}

To access the raw token stream (which includes tokens for comments):

	  t := NewTokenizer(os.Stdin)
	  for ; token, err := t.Next(); err != nil {
		// process token
	  }
*/
package shlex

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// TokenType is a top-level token classification: A word, space, comment, unknown.
type TokenType int

// runeTokenClass is the type of a UTF-8 character classification: A quote, space, escape.
type runeTokenClass int

// the internal state used by the lexer state machine
type lexerState int

// Token is a (type, value) pair representing a lexographical token.
type Token struct {
	Type     TokenType
	Value    string
	RawValue string
	Index    int
}

func (t *Token) add(r rune) {
	t.Value += string(r)
}

func (t *Token) removeLastRaw() {
	runes := []rune(t.RawValue)
	t.RawValue = string(runes[:len(runes)-1])
}

// Equal reports whether tokens a, and b, are equal.
// Two tokens are equal if both their types and values are equal. A nil token can
// never be equal to another token.
func (a *Token) Equal(b *Token) bool {
	if a == nil || b == nil {
		return false
	}
	if a.Type != b.Type {
		return false
	}
	if a.RawValue != b.RawValue {
		return false
	}
	return a.Value == b.Value && a.Index == b.Index
}

// Named classes of UTF-8 runes
const (
	spaceRunes            = " \t\r\n"
	escapingQuoteRunes    = `"`
	nonEscapingQuoteRunes = "'"
	escapeRunes           = `\`
	commentRunes          = "#"
	pipelineRunes         = "|&;"
)

// Classes of rune token
const (
	unknownRuneClass runeTokenClass = iota
	spaceRuneClass
	escapingQuoteRuneClass
	nonEscapingQuoteRuneClass
	escapeRuneClass
	commentRuneClass
	pipelineRuneClass
	eofRuneClass
)

// Classes of lexographic token
const (
	UnknownToken TokenType = iota
	WordToken
	SpaceToken
	CommentToken
	PipelineToken
)

// Lexer state machine states
const (
	startState           lexerState = iota // no runes have been seen
	inWordState                            // processing regular runes in a word
	escapingState                          // we have just consumed an escape rune; the next rune is literal
	escapingQuotedState                    // we have just consumed an escape rune within a quoted string
	quotingEscapingState                   // we are within a quoted string that supports escaping ("...")
	quotingState                           // we are within a string that does not support escaping ('...')
	commentState                           // we are within a comment (everything following an unquoted or unescaped #
	pipelineState                          // we have just consumed a pipeline delimiter (just consume these until we reach something else)
)

// tokenClassifier is used for classifying rune characters.
type tokenClassifier map[rune]runeTokenClass

func (typeMap tokenClassifier) addRuneClass(runes string, tokenType runeTokenClass) {
	for _, runeChar := range runes {
		typeMap[runeChar] = tokenType
	}
}

// newDefaultClassifier creates a new classifier for ASCII characters.
func newDefaultClassifier() tokenClassifier {
	t := tokenClassifier{}
	t.addRuneClass(spaceRunes, spaceRuneClass)
	t.addRuneClass(escapingQuoteRunes, escapingQuoteRuneClass)
	t.addRuneClass(nonEscapingQuoteRunes, nonEscapingQuoteRuneClass)
	t.addRuneClass(escapeRunes, escapeRuneClass)
	t.addRuneClass(commentRunes, commentRuneClass)
	t.addRuneClass(pipelineRunes, pipelineRuneClass)
	return t
}

// ClassifyRune classifiees a rune
func (t tokenClassifier) ClassifyRune(runeVal rune) runeTokenClass {
	return t[runeVal]
}

// Lexer turns an input stream into a sequence of tokens. Whitespace and comments are skipped.
type Lexer Tokenizer

// NewLexer creates a new lexer from an input stream.
func NewLexer(r io.Reader) *Lexer {

	return (*Lexer)(NewTokenizer(r))
}

// Next returns the next token, or an error. If there are no more tokens,
// the error will be io.EOF.
func (l *Lexer) Next() (*Token, error) {
	for {
		token, err := (*Tokenizer)(l).Next()
		if err != nil {
			return token, err
		}
		switch token.Type {
		case WordToken, PipelineToken:
			return token, nil
		case CommentToken:
			// skip comments
		default:
			return nil, fmt.Errorf("unknown token type: %v", token.Type)
		}
	}
}

// Tokenizer turns an input stream into a sequence of typed tokens
type Tokenizer struct {
	input      bufio.Reader
	classifier tokenClassifier
	index      int
}

func (t *Tokenizer) ReadRune() (r rune, size int, err error) {
	if r, size, err = t.input.ReadRune(); err == nil {
		t.index += 1
	}
	return
}

func (t *Tokenizer) UnreadRune() (err error) {
	if err = t.input.UnreadRune(); err == nil {
		t.index -= 1
	}
	return
}

// NewTokenizer creates a new tokenizer from an input stream.
func NewTokenizer(r io.Reader) *Tokenizer {
	input := bufio.NewReader(r)
	classifier := newDefaultClassifier()
	return &Tokenizer{
		input:      *input,
		classifier: classifier}
}

// scanStream scans the stream for the next token using the internal state machine.
// It will panic if it encounters a rune which it does not know how to handle.
func (t *Tokenizer) scanStream() (*Token, error) {
	state := startState
	token := &Token{}
	var nextRune rune
	var nextRuneType runeTokenClass
	var err error

	for {
		nextRune, _, err = t.ReadRune()
		nextRuneType = t.classifier.ClassifyRune(nextRune)
		token.RawValue += string(nextRune)

		switch {
		case err == io.EOF:
			nextRuneType = eofRuneClass
			err = nil
		case err != nil:
			return nil, err
		}

		switch state {
		case startState: // no runes read yet
			{
				if nextRuneType != spaceRuneClass {
					token.Index = t.index - 1
				}
				switch nextRuneType {
				case eofRuneClass:
					return nil, io.EOF
				case spaceRuneClass:
					token.removeLastRaw()
				case escapingQuoteRuneClass:
					token.Type = WordToken
					state = quotingEscapingState
				case nonEscapingQuoteRuneClass:
					token.Type = WordToken
					state = quotingState
				case escapeRuneClass:
					token.Type = WordToken
					state = escapingState
				case commentRuneClass:
					token.Type = CommentToken
					state = commentState
				case pipelineRuneClass:
					token.Type = PipelineToken
					token.add(nextRune)
					state = pipelineState
				default:
					token.Type = WordToken
					token.add(nextRune)
					state = inWordState
				}
			}
		case pipelineState:
			switch nextRuneType {
			case pipelineRuneClass:
				token.add(nextRune)
			default:
				token.removeLastRaw()
				t.UnreadRune()
				return token, err
			}
		case inWordState: // in a regular word
			switch nextRuneType {
			case pipelineRuneClass:
				token.removeLastRaw()
				t.UnreadRune()
				return token, err
			case eofRuneClass, spaceRuneClass:
				token.removeLastRaw()
				return token, err
			case escapingQuoteRuneClass:
				state = quotingEscapingState
			case nonEscapingQuoteRuneClass:
				state = quotingState
			case escapeRuneClass:
				state = escapingState
			default:
				token.add(nextRune)
			}
		case escapingState: // the rune after an escape character
			switch nextRuneType {
			case eofRuneClass:
				err = fmt.Errorf("EOF found after escape character")
				return token, err
			default:
				state = inWordState
				token.add(nextRune)
			}
		case escapingQuotedState: // the next rune after an escape character, in double quotes
			switch nextRuneType {
			case eofRuneClass:
				err = fmt.Errorf("EOF found after escape character")
				return token, err
			default:
				state = quotingEscapingState
				token.add(nextRune)
			}
		case quotingEscapingState: // in escaping double quotes
			switch nextRuneType {
			case eofRuneClass:
				err = fmt.Errorf("EOF found when expecting closing quote")
				return token, err
			case escapingQuoteRuneClass:
				state = inWordState
			case escapeRuneClass:
				state = escapingQuotedState
			default:
				token.add(nextRune)
			}
		case quotingState: // in non-escaping single quotes
			switch nextRuneType {
			case eofRuneClass:
				err = fmt.Errorf("EOF found when expecting closing quote")
				return token, err
			case nonEscapingQuoteRuneClass:
				state = inWordState
			default:
				token.add(nextRune)
			}
		case commentState: // in a comment
			switch nextRuneType {
			case eofRuneClass:
				return token, err
			case spaceRuneClass:
				if nextRune == '\n' {
					token.removeLastRaw()
					state = startState
					return token, err
				} else {
					token.add(nextRune)
				}
			default:
				token.add(nextRune)
			}
		default:
			return nil, fmt.Errorf("unexpected state: %v", state)
		}
	}
}

// Next returns the next token in the stream.
func (t *Tokenizer) Next() (*Token, error) {
	return t.scanStream()
}

type Tokens []Token

func (t Tokens) Strings() []string {
	s := make([]string, 0, len(t))
	for _, token := range t {
		s = append(s, token.Value)
	}
	return s
}

func (t Tokens) CurrentPipeline() Tokens {
	tokens := make([]Token, 0)
	for _, token := range t {
		switch token.Type {
		case PipelineToken:
			tokens = make([]Token, 0)
		default:
			tokens = append(tokens, token)
		}
	}
	return Tokens(tokens)
}

// Split partitions of a string into tokens.
func Split(s string) (*Tokens, error) {
	l := NewLexer(strings.NewReader(s))
	tokens := make([]Token, 0)
	for {
		token, err := l.Next()
		if err != nil {
			if err == io.EOF {
				t := Tokens(tokens)
				return &t, nil
			}
			return nil, err
		}
		tokens = append(tokens, *token)
	}
}
