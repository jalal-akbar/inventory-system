package main

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	rootDir := "web/templates"
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".html") {
			// specific helper functionsmock
			funcMap := template.FuncMap{
				"t":                func(key, lang string) string { return key },
				"formatRupiah":     func(amount interface{}) string { return "" },
				"formatDateTime":   func(t interface{}) string { return "" },
				"formatShortDate":  func(t interface{}) string { return "" },
				"upper":            func(s string) string { return strings.ToUpper(s) },
				"lower":            func(s string) string { return strings.ToLower(s) },
				"firstChar":        func(s string) string { return "" },
				"add":              func(a, b int) int { return a + b },
				"sub":              func(a, b int) int { return a - b },
				"mul":              func(a, b interface{}) float64 { return 0 },
				"abs":              func(a int) int { return a },
				"json":             func(v interface{}) template.JS { return template.JS("[]") },
				"formatPercentage": func(v float64) string { return "" },
			}

			_, err := template.New(filepath.Base(path)).Funcs(funcMap).ParseFiles(path)
			if err != nil {
				fmt.Printf("Error validating %s: %v\n", path, err)
			}
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Walk error: %v\n", err)
	}
	fmt.Println("Template validation complete.")
}
