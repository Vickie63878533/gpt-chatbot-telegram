package plugin

import (
	"net/url"
	"testing"
)

func TestInterpolateSimpleVariable(t *testing.T) {
	template := "Hello {{name}}"
	data := map[string]interface{}{
		"name": "World",
	}

	result := Interpolate(template, data, nil)
	expected := "Hello World"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestInterpolateNestedVariable(t *testing.T) {
	template := "Hello {{user.name}}"
	data := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "Alice",
		},
	}

	result := Interpolate(template, data, nil)
	expected := "Hello Alice"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestInterpolateArrayIndex(t *testing.T) {
	template := "First: {{items[0]}}, Second: {{items[1]}}"
	data := map[string]interface{}{
		"items": []interface{}{"apple", "banana"},
	}

	result := Interpolate(template, data, nil)
	expected := "First: apple, Second: banana"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestInterpolateLoop(t *testing.T) {
	template := "{{#each item in items}}{{item}},{{/each}}"
	data := map[string]interface{}{
		"items": []interface{}{"a", "b", "c"},
	}

	result := Interpolate(template, data, nil)
	expected := "a,b,c,"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestInterpolateLoopWithDot(t *testing.T) {
	template := "{{#each item in items}}{{.}},{{/each}}"
	data := map[string]interface{}{
		"items": []interface{}{"x", "y", "z"},
	}

	result := Interpolate(template, data, nil)
	expected := "x,y,z,"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestInterpolateLoopWithObject(t *testing.T) {
	template := "{{#each user in users}}{{user.name}},{{/each}}"
	data := map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{"name": "Alice"},
			map[string]interface{}{"name": "Bob"},
		},
	}

	result := Interpolate(template, data, nil)
	expected := "Alice,Bob,"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestInterpolateConditionalTrue(t *testing.T) {
	template := "{{#if active}}Active{{/if}}"
	data := map[string]interface{}{
		"active": true,
	}

	result := Interpolate(template, data, nil)
	expected := "Active"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestInterpolateConditionalFalse(t *testing.T) {
	template := "{{#if active}}Active{{/if}}"
	data := map[string]interface{}{
		"active": false,
	}

	result := Interpolate(template, data, nil)
	expected := ""

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestInterpolateConditionalWithElse(t *testing.T) {
	template := "{{#if active}}Active{{#else}}Inactive{{/if}}"
	data := map[string]interface{}{
		"active": false,
	}

	result := Interpolate(template, data, nil)
	expected := "Inactive"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestInterpolateWithFormatter(t *testing.T) {
	template := "URL: {{query}}"
	data := map[string]interface{}{
		"query": "hello world",
	}

	result := Interpolate(template, data, url.QueryEscape)
	expected := "URL: hello+world"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestInterpolateComplexTemplate(t *testing.T) {
	template := `{{#each word in words}}
<b>{{word.text}}</b>{{#if word.phonetic}}<i>{{word.phonetic}}</i>{{/if}}
{{/each}}`

	data := map[string]interface{}{
		"words": []interface{}{
			map[string]interface{}{
				"text":     "hello",
				"phonetic": "/həˈloʊ/",
			},
			map[string]interface{}{
				"text": "world",
			},
		},
	}

	result := Interpolate(template, data, nil)

	// Check that it contains expected parts
	if !contains(result, "<b>hello</b>") {
		t.Errorf("Result should contain '<b>hello</b>', got: %q", result)
	}
	if !contains(result, "<i>/həˈloʊ/</i>") {
		t.Errorf("Result should contain phonetic, got: %q", result)
	}
	if !contains(result, "<b>world</b>") {
		t.Errorf("Result should contain '<b>world</b>', got: %q", result)
	}
}

func TestInterpolateMissingVariable(t *testing.T) {
	template := "Hello {{missing}}"
	data := map[string]interface{}{
		"name": "World",
	}

	result := Interpolate(template, data, nil)
	expected := "Hello {{missing}}"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			len(s) > len(substr) && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
