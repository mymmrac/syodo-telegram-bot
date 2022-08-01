package main

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/BurntSushi/toml"
)

// Text represents a map of text names and corresponding templates
type Text map[string]*template.Template

// Get return text with given data, exits if not found or failed to execute
func (t Text) Get(key string, data any) string {
	temp, ok := t[key]
	assert(ok, fmt.Sprintf("text with key %q not found", key))

	buf := &bytes.Buffer{}
	err := temp.Execute(buf, data)
	assert(err == nil, fmt.Errorf("get text with key %q and data %+v, error: %w", key, data, err))

	return buf.String()
}

// LoadText loads text templates from specified file
func LoadText(filename string) (Text, error) {
	var textData map[string]string

	_, err := toml.DecodeFile(filename, &textData)
	if err != nil {
		return nil, fmt.Errorf("decode text: %w", err)
	}

	text := make(map[string]*template.Template, len(textData))

	for key, value := range textData {
		trimmedValue := strings.TrimSpace(value)
		text[key], err = template.New(key).Parse(trimmedValue)
		if err != nil {
			return nil, fmt.Errorf("parsing text of %q with value %q, error: %w", key, trimmedValue, err)
		}
	}

	return text, nil
}
