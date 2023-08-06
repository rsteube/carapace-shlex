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

package shlex

import (
	"strings"
	"testing"
)

var (
	// one two "three four" "five \"six\"" seven#eight # nine # ten
	// eleven 'twelve\'
	testString = "one two \"three four\" \"five \\\"six\\\"\" seven#eight # nine # ten\n eleven 'twelve\\' thirteen=13 fourteen/14 | || |after before| & ;"
)

func TestClassifier(t *testing.T) {
	classifier := newDefaultClassifier()
	tests := map[rune]runeTokenClass{
		' ':  spaceRuneClass,
		'"':  escapingQuoteRuneClass,
		'\'': nonEscapingQuoteRuneClass,
		'#':  commentRuneClass}
	for runeChar, want := range tests {
		got := classifier.ClassifyRune(runeChar)
		if got != want {
			t.Errorf("ClassifyRune(%v) -> %v. Want: %v", runeChar, got, want)
		}
	}
}

func TestTokenizer(t *testing.T) {
	testInput := strings.NewReader(testString)
	expectedTokens := []*Token{
		{WORD_TOKEN, "one", "one", 0, IN_WORD_STATE},
		{WORD_TOKEN, "two", "two", 4, IN_WORD_STATE},
		{WORD_TOKEN, "three four", "\"three four\"", 8, IN_WORD_STATE},
		{WORD_TOKEN, "five \"six\"", "\"five \\\"six\\\"\"", 21, IN_WORD_STATE},
		{WORD_TOKEN, "seven#eight", "seven#eight", 36, IN_WORD_STATE},
		{COMMENT_TOKEN, " nine # ten", "# nine # ten", 48, START_STATE},
		{WORD_TOKEN, "eleven", "eleven", 62, IN_WORD_STATE},
		{WORD_TOKEN, "twelve\\", "'twelve\\'", 69, IN_WORD_STATE},
		{WORD_TOKEN, "thirteen=13", "thirteen=13", 79, IN_WORD_STATE},
		{WORD_TOKEN, "fourteen/14", "fourteen/14", 91, IN_WORD_STATE},
		{PIPELINE_TOKEN, "|", "|", 103, PIPELINE_STATE},
		{PIPELINE_TOKEN, "||", "||", 105, PIPELINE_STATE},
		{PIPELINE_TOKEN, "|", "|", 108, PIPELINE_STATE},
		{WORD_TOKEN, "after", "after", 109, IN_WORD_STATE},
		{WORD_TOKEN, "before", "before", 115, IN_WORD_STATE},
		{PIPELINE_TOKEN, "|", "|", 121, PIPELINE_STATE},
		{PIPELINE_TOKEN, "&", "&", 123, PIPELINE_STATE},
		{PIPELINE_TOKEN, ";", ";", 125, PIPELINE_STATE},
		{WORD_TOKEN, "", "", 126, START_STATE},
	}

	tokenizer := newTokenizer(testInput)
	for i, want := range expectedTokens {
		got, err := tokenizer.Next()
		if err != nil {
			t.Error(err)
		}
		if !got.Equal(want) {
			t.Errorf("Tokenizer.Next()[%v] of %q \nGot : %#v\nWant: %#v", i, testString, got, want)
		}
	}
}

func TestLexer(t *testing.T) {
	testInput := strings.NewReader(testString)
	expectedStrings := []string{"one", "two", "three four", "five \"six\"", "seven#eight", "eleven", "twelve\\", "thirteen=13", "fourteen/14"}

	lexer := newLexer(testInput)
	for i, want := range expectedStrings {
		got, err := lexer.Next()
		if err != nil {
			t.Error(err)
		}
		if got.Value != want {
			t.Errorf("Lexer.Next()[%v] of %q -> %v. Want: %v", i, testString, got, want)
		}
	}
}

func TestSplit(t *testing.T) {
	want := []string{"one", "two", "three four", "five \"six\"", "seven#eight", "eleven", "twelve\\", "thirteen=13", "fourteen/14", "|", "||", "|", "after", "before", "|", "&", ";", ""}
	got, err := Split(testString)
	if err != nil {
		t.Error(err)
	}
	if len(want) != len(*got) {
		t.Errorf("Split(%q) -> %v. Want: %v", testString, got, want)
	}
	for i, g := range *got {
		if g.Value != want[i] {
			t.Errorf("Split(%q)[%v] -> %v. Want: %v", testString, i, g.Value, want[i])
		}
	}
}
