package i18n

import (
	"html/template"
	"strings"
)

var translations = make(map[string]map[string]string)

func Set(lang string, data map[string]string) {
	translations[lang] = data

}

func T(key, lang string) string {
	key = strings.TrimSpace(key)
	if l, ok := translations[lang]; ok {
		if val, vok := l[key]; vok {
			return val
		}
	}
	// Fallback to English
	if lang != "en" {
		return T(key, "en")
	}
	return key
}

func GetFuncMap(lang string) template.FuncMap {
	return template.FuncMap{
		"__": func(key string) string {
			return T(key, lang)
		},
	}
}
