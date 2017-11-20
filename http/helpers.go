package qbinHTTP

import (
	"regexp"
	"strings"
)

// escapeHTML removes all special HTML characters (namely, &<>") in a string and replaces them with their entities (e.g. &amp;).
func escapeHTML(content string) string {
	content = strings.Replace(content, "&", "&amp;", -1)
	content = strings.Replace(content, "<", "&lt;", -1)
	content = strings.Replace(content, ">", "&gt;", -1)
	content = strings.Replace(content, "\"", "&quot;", -1)
	return content
}

// replaceGlobal replaces all global frontend variables with their config value.
func replaceGlobal(content *string) {
	replaceVariable(content, "path", config.Path)
	replaceVariable(content, "root", config.Root)
}

// replaceVariable replaces a single frontend variable with its value.
// For example, replaceVariable(x, "id", "12345") replaces "$$id$$" with "12345".
func replaceVariable(content *string, name string, value string) {
	*content = strings.Replace(*content, "$$"+name+"$$", value, -1)
}

// blockVariableExpressionCache contains regular expressions for all block variables to improve rendering speed
var blockVariableExpressionCache = map[string]*regexp.Regexp{}

// replaceBlockVariable removes if-style blocks if they're not matching the value, and cleans the variable remainders otherwise.
func replaceBlockVariable(content *string, name string, value bool) {
	// We will later replace every block that is only shown for the opposite value -> not seems to be inverted
	not := "!"
	if !value {
		not = ""
	}

	// Try loading regular expression from cache
	expression, exists := blockVariableExpressionCache[not+name]
	if exists == false {
		// Compile regular expression to match blocks for the opposite value
		expression = regexp.MustCompile(`\$\$` + not + name + `\$\$(?U:(?:.|\n)*\$\$/` + name + `\$\$)`)
		blockVariableExpressionCache[not+name] = expression
	}

	// Replace blocks for the opposite value
	*content = expression.ReplaceAllString(*content, "")

	// Replace the variables only (not including the block) if the value matches
	if value {
		replaceVariable(content, name, "")
	} else {
		replaceVariable(content, "!"+name, "")
	}
	replaceVariable(content, "/"+name, "")
}
