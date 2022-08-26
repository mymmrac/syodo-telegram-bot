package main

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/BurntSushi/toml"
)

// TextData represents a map of text names and corresponding templates
type TextData map[string]*template.Template

// Temp return text with given data executing template, exits if not found or failed to execute
func (t TextData) Temp(key string, data any) string {
	temp, ok := t[key]
	assert(ok, fmt.Sprintf("template with key %q not found", key))

	buf := &bytes.Buffer{}
	err := temp.Execute(buf, data)
	assert(err == nil, fmt.Errorf("execute template with key %q and data %+v, error: %w", key, data, err))

	return buf.String()
}

// Text return text executing template with no data, exits if not found or failed to execute
func (t TextData) Text(key string) string {
	return t.Temp(key, nil)
}

// LoadTextData loads text templates from specified file
func LoadTextData(filename string) (TextData, error) {
	var textValues map[string]string

	_, err := toml.DecodeFile(filename, &textValues)
	if err != nil {
		return nil, fmt.Errorf("decode text data: %w", err)
	}

	textData := make(map[string]*template.Template, len(textValues))

	fm := template.FuncMap{
		"toPrice": func(amount int) string {
			return fmt.Sprintf("%.2f", float64(amount)/100.0)
		},
	}

	for key, value := range textValues {
		transformedValue := strings.TrimSpace(strings.ReplaceAll(value, "|\n", ""))
		textData[key], err = template.New(key).Funcs(fm).Parse(transformedValue)
		if err != nil {
			return nil, fmt.Errorf("parsing text data of %q with value %q, error: %w", key, transformedValue, err)
		}
	}

	return textData, nil
}
