package cursor

import "strings"

// splitSearchTerms splits a string into terms, respecting quoted sections.
// For example: `file.go "some quoted term" other` becomes ["file.go", "some quoted term", "other"]
func splitSearchTerms(input string) []string {
	var terms []string
	var currentTerm strings.Builder
	inQuotes := false

	for _, r := range input {
		switch r {
		case '"':
			inQuotes = !inQuotes
			// Don't include the quotes in the term
		case ' ':
			if inQuotes {
				currentTerm.WriteRune(r)
			} else if currentTerm.Len() > 0 {
				terms = append(terms, currentTerm.String())
				currentTerm.Reset()
			}
		default:
			currentTerm.WriteRune(r)
		}
	}

	if currentTerm.Len() > 0 {
		terms = append(terms, currentTerm.String())
	}

	return terms
}
