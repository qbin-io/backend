package qbin

import (
	"testing"
)

func TestHighlight(t *testing.T) {
	result := SyntaxExists("javascript")
	if result != true {
		t.Errorf("javascript should be a valid syntax, but it isn't.")
		t.FailNow()
	}

	result = SyntaxExists("")
	if result != false {
		t.Errorf("<empty string> shouldn't be a valid syntax, but it is.")
		t.FailNow()
	}

	result2 := Highlight("console.log('Hello <World>');", "javascript")
	if result2 == "console.log('Hello <World>');" || result2 == "" {
		t.Errorf("Not highlighted: %s", result2)
		t.FailNow()
	}
}
