package cli

// import (
// 	"fmt"
// 	"testing"
// )

// func TestLexe(t *testing.T) {
// 	testCases := []struct {
// 		args   []string
// 		tokens []ArgToken
// 	}{
// 		{[]string{}, []ArgToken{}},
// 		{[]string{"hund"}, []ArgToken{{Value: "hund", Kind: ValueArg}}},
// 		{
// 			[]string{"dd", "-s", "--longopt", "value", "-k", "value", "--a"},
// 			[]ArgToken{
// 				{Value: "dd", Kind: ValueArg},
// 				{Value: "s", Kind: ShortOpt},
// 				{Value: "longopt", Kind: LongOpt},
// 				{Value: "value", Kind: ValueArg},
// 				{Value: "k", Kind: ShortOpt},
// 				{Value: "value", Kind: ValueArg},
// 				{Value: "a", Kind: LongOpt},
// 			},
// 		},
// 	}
// 	for _, tc := range testCases {
// 		t.Run(fmt.Sprintf("%v", tc.args), func(t *testing.T) {
// 			tokens, err := Parse(tc.args)
// 			if err != nil {
// 				t.Fatal(err.Error())
// 			}
// 			if len(tokens) != len(tc.tokens) {
// 				t.Errorf("got %v; want %v", tokens, tc.tokens)
// 				return
// 			}
// 			for index, token := range tokens {
// 				if token != tc.tokens[index] {
// 					t.Errorf("got %v; want %v", token, tc.tokens[index])
// 				}
// 			}
// 		})
// 	}
// }

// func TestErrors(t *testing.T) {
// 	testCases := [][]string{
// 		{"", "", ""},
// 		{"-sdfasdfa"},
// 		{"-aa"},
// 		{"---asdfasd"},
// 	}

// 	for _, tc := range testCases {
// 		testName := fmt.Sprintf("%v", tc)
// 		t.Run(testName, func(t *testing.T) {
// 			_, err := Parse(tc)
// 			if err == nil {
// 				t.Fatal("expected error")
// 			}
// 		})
// 	}
// }
