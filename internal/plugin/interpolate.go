package plugin

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	// {{#each:alias item in array}}...{{/each:alias}} or {{#each item in array}}...{{/each}}
	interpolateLoopRegexp = regexp.MustCompile(`\{\{#each(?::(\w+))?\s+(\w+)\s+in\s+([\w.\[\]]+)\}\}([\s\S]*?)\{\{/each(?::\w+)?\}\}`)

	// {{#if:alias condition}}...{{#else:alias}}...{{/if:alias}} or {{#if condition}}...{{#else}}...{{/if}}
	interpolateConditionRegexp = regexp.MustCompile(`\{\{#if(?::(\w+))?\s+([\w.\[\]]+)\}\}([\s\S]*?)(?:\{\{#else(?::\w+)?\}\}([\s\S]*?))?\{\{/if(?::\w+)?\}\}`)

	// {{variable}}
	interpolateVariableRegexp = regexp.MustCompile(`\{\{([\w.\[\]]+)\}\}`)
)

// Interpolate replaces template variables with actual values
func Interpolate(template string, data interface{}, formatter func(string) string) string {
	return processTemplate(template, data, formatter)
}

func processTemplate(tmpl string, data interface{}, formatter func(string) string) string {
	// Process loops
	tmpl = interpolateLoopRegexp.ReplaceAllStringFunc(tmpl, func(match string) string {
		matches := interpolateLoopRegexp.FindStringSubmatch(match)
		if len(matches) < 5 {
			return match
		}
		// matches[1] = alias (optional)
		// matches[2] = itemName
		// matches[3] = arrayExpr
		// matches[4] = loopContent
		return processLoop(matches[2], matches[3], matches[4], data, formatter)
	})

	// Process conditionals
	tmpl = interpolateConditionRegexp.ReplaceAllStringFunc(tmpl, func(match string) string {
		matches := interpolateConditionRegexp.FindStringSubmatch(match)
		if len(matches) < 4 {
			return match
		}
		// matches[1] = alias (optional)
		// matches[2] = condition
		// matches[3] = trueBlock
		// matches[4] = falseBlock (optional)
		return processConditional(matches[2], matches[3], matches[4], data, formatter)
	})

	// Process variables
	tmpl = interpolateVariableRegexp.ReplaceAllStringFunc(tmpl, func(match string) string {
		matches := interpolateVariableRegexp.FindStringSubmatch(match)
		if len(matches) < 2 {
			return match
		}
		expr := matches[1]
		value := evaluateExpression(expr, data)
		if value == nil {
			return match // Keep original if not found
		}
		valueStr := fmt.Sprintf("%v", value)
		if formatter != nil {
			return formatter(valueStr)
		}
		return valueStr
	})

	return tmpl
}

func processLoop(itemName, arrayExpr, loopContent string, data interface{}, formatter func(string) string) string {
	array := evaluateExpression(arrayExpr, data)
	if array == nil {
		return ""
	}

	// Convert to slice
	var items []interface{}
	switch v := array.(type) {
	case []interface{}:
		items = v
	case []string:
		for _, s := range v {
			items = append(items, s)
		}
	case []int:
		for _, i := range v {
			items = append(items, i)
		}
	default:
		return ""
	}

	var result strings.Builder
	for _, item := range items {
		// Create local data with item
		localData := map[string]interface{}{
			itemName: item,
			".":      item,
		}
		// Merge with parent data
		if dataMap, ok := data.(map[string]interface{}); ok {
			for k, v := range dataMap {
				if k != itemName && k != "." {
					localData[k] = v
				}
			}
		}
		result.WriteString(Interpolate(loopContent, localData, formatter))
	}

	return result.String()
}

func processConditional(condition, trueBlock, falseBlock string, data interface{}, formatter func(string) string) string {
	result := evaluateExpression(condition, data)
	if isTruthy(result) {
		return Interpolate(trueBlock, data, formatter)
	}
	if falseBlock != "" {
		return Interpolate(falseBlock, data, formatter)
	}
	return ""
}

func evaluateExpression(expr string, data interface{}) interface{} {
	if expr == "." {
		if dataMap, ok := data.(map[string]interface{}); ok {
			if val, exists := dataMap["."]; exists {
				return val
			}
		}
		return data
	}

	parts := strings.Split(expr, ".")
	current := data

	for _, part := range parts {
		if current == nil {
			return nil
		}

		// Handle array indexing: key[0] or key[index]
		if strings.Contains(part, "[") && strings.Contains(part, "]") {
			bracketIdx := strings.Index(part, "[")
			arrayKey := part[:bracketIdx]
			indexExpr := part[bracketIdx+1 : len(part)-1]

			// Get the array
			current = getMapValue(current, arrayKey)
			if current == nil {
				return nil
			}

			// Parse index
			index, err := strconv.Atoi(indexExpr)
			if err != nil {
				// Try to evaluate as expression
				indexVal := evaluateExpression(indexExpr, data)
				if indexVal != nil {
					if idx, ok := indexVal.(int); ok {
						index = idx
					} else {
						return nil
					}
				} else {
					return nil
				}
			}

			// Get array element
			switch arr := current.(type) {
			case []interface{}:
				if index >= 0 && index < len(arr) {
					current = arr[index]
				} else {
					return nil
				}
			case []string:
				if index >= 0 && index < len(arr) {
					current = arr[index]
				} else {
					return nil
				}
			case []int:
				if index >= 0 && index < len(arr) {
					current = arr[index]
				} else {
					return nil
				}
			default:
				return nil
			}
		} else {
			current = getMapValue(current, part)
		}
	}

	return current
}

func getMapValue(data interface{}, key string) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		return v[key]
	case map[string]string:
		return v[key]
	default:
		return nil
	}
}

func isTruthy(value interface{}) bool {
	if value == nil {
		return false
	}

	switch v := value.(type) {
	case bool:
		return v
	case string:
		return v != ""
	case int, int64, float64:
		return v != 0
	case []interface{}:
		return len(v) > 0
	case map[string]interface{}:
		return len(v) > 0
	default:
		return true
	}
}
