package templates

import "html/template"

// TemplateFuncs retourne les fonctions personnalisées pour les templates
func TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"iterate": iterate,
	}
}

// iterate génère une slice d'entiers de start à end (inclus)
// Usage dans le template: {{range iterate 1 7}}
func iterate(start, end int) []int {
	if start > end {
		return []int{}
	}

	result := make([]int, end-start+1)
	for i := range result {
		result[i] = start + i
	}
	return result
}
