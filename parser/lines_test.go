package parser

import (
	"testing"
)

func TestEscaped(t *testing.T) {
	texts := []string{
		"\\",
		" \\",
		"    	\\",
		"	  \\",
		"			\\",
	}

	for _, text := range texts {
		line := Line{text: text, num: 0}
		if !line.IsNewlineEscaped() {
			t.Errorf("text \"%s\" is not considered escaped", text)
		}
	}
}

func TestComment(t *testing.T) {
	texts := []string{
		"// a",
		" // dddd",
		"      // ddccd",
		"    	// comment commnet \\",
		"	  // yet another commnet",
		"			// some more more more",
		"//",
	}

	for _, text := range texts {
		line := Line{text: text, num: 0}
		if !line.IsComment() {
			t.Errorf("text \"%s\" is not considered a comment", text)
		}
	}
}

func TestIndented(t *testing.T) {
	texts := []string{
		"  two space \\",
		"	tab",
		"  	   	spaces and tabs",
		"            lots of spaces",
		"				lots of tabs",
	}

	for _, text := range texts {
		line := Line{text: text, num: 0}
		if !line.IsIndented() {
			t.Errorf("text \"%s\" is not considered to be indented", text)
		}
	}
}

func TestHeader(t *testing.T) {
	texts := []string{
		"identifier:",
		"some-other():",
		"yet_another_1: dddd ddd sad dd \\",
		"dis-one-with-args(arg1,arg2,arg3):",
		"with-args_adn_options( arg1, arg2, arg3++ ): option1= option33",
	}

	for _, text := range texts {
		line := Line{text: text, num: 0}
		if !line.IsTargetHeader() {
			t.Errorf("text \"%s\" is not considered target header", text)
		}
	}
}
