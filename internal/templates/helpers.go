package templates

import "html/template"

// TemplateFuncs returns custom functions for templates
func TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"iterate": iterate,
	}
}

// iterate generates a slice of integers from start to end (inclusive)
// Usage in template: {{range iterate 1 7}}
func iterate(start, end int) []int {
	if start > end {
		return []int{}
	}

	result := make([]int, end-start+1) // +1 because inclusive
	for i := range result {
		result[i] = start + i
	}
	return result
}
