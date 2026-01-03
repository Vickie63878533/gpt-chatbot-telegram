package plugin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// FormatInput formats the input string according to the specified type
func FormatInput(input string, inputType TemplateInputType) interface{} {
	switch inputType {
	case InputTypeJSON:
		var result interface{}
		if err := json.Unmarshal([]byte(input), &result); err != nil {
			return input
		}
		return result
	case InputTypeSpaceSeparated:
		parts := strings.Fields(input)
		return parts
	case InputTypeCommaSeparated:
		parts := strings.Split(input, ",")
		result := make([]string, 0, len(parts))
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result
	case InputTypeText:
		fallthrough
	default:
		return input
	}
}

// ExecuteRequest executes a plugin template request
func ExecuteRequest(template *RequestTemplate, data map[string]interface{}) (*ExecuteResult, error) {
	// Interpolate URL
	urlStr := Interpolate(template.URL, data, url.QueryEscape)
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Add query parameters
	if template.Query != nil {
		q := parsedURL.Query()
		for key, value := range template.Query {
			interpolated := Interpolate(value, data, nil)
			if interpolated != "" && interpolated != "null" {
				q.Add(key, interpolated)
			}
		}
		parsedURL.RawQuery = q.Encode()
	}

	// Prepare headers
	headers := make(map[string]string)
	if template.Headers != nil {
		for key, value := range template.Headers {
			interpolated := Interpolate(value, data, nil)
			if interpolated != "" && interpolated != "null" {
				headers[key] = interpolated
			}
		}
	}

	// Prepare body
	var body io.Reader
	if template.Body != nil {
		switch template.Body.Type {
		case BodyTypeJSON:
			interpolated := interpolateObject(template.Body.Content, data)
			jsonData, err := json.Marshal(interpolated)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal JSON body: %w", err)
			}
			body = bytes.NewReader(jsonData)
			if headers["Content-Type"] == "" {
				headers["Content-Type"] = "application/json"
			}
		case BodyTypeForm:
			if contentMap, ok := template.Body.Content.(map[string]interface{}); ok {
				formData := url.Values{}
				for key, value := range contentMap {
					valueStr := fmt.Sprintf("%v", value)
					interpolated := Interpolate(valueStr, data, nil)
					formData.Add(key, interpolated)
				}
				body = strings.NewReader(formData.Encode())
				if headers["Content-Type"] == "" {
					headers["Content-Type"] = "application/x-www-form-urlencoded"
				}
			}
		case BodyTypeText:
			if contentStr, ok := template.Body.Content.(string); ok {
				interpolated := Interpolate(contentStr, data, nil)
				body = strings.NewReader(interpolated)
			}
		}
	}

	// Create request
	req, err := http.NewRequest(template.Method, parsedURL.String(), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Execute request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Handle response
	if resp.StatusCode >= 400 {
		return renderErrorOutput(template, resp)
	}

	return renderSuccessOutput(template, resp)
}

func renderSuccessOutput(template *RequestTemplate, resp *http.Response) (*ExecuteResult, error) {
	// Handle blob response (image)
	if template.Response.Content.InputType == ResponseTypeBlob {
		if template.Response.Content.OutputType != OutputTypeImage {
			return nil, fmt.Errorf("blob input type only supports image output type")
		}
		// For images, we return the URL directly
		return &ExecuteResult{
			Type:    OutputTypeImage,
			Content: resp.Request.URL.String(),
		}, nil
	}

	// Parse response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var responseData interface{}
	switch template.Response.Content.InputType {
	case ResponseTypeJSON:
		if err := json.Unmarshal(bodyBytes, &responseData); err != nil {
			return nil, fmt.Errorf("failed to parse JSON response: %w", err)
		}
	case ResponseTypeText:
		responseData = string(bodyBytes)
	default:
		responseData = string(bodyBytes)
	}

	// Interpolate output template
	output := Interpolate(template.Response.Content.Output, responseData, nil)

	return &ExecuteResult{
		Type:    template.Response.Content.OutputType,
		Content: output,
	}, nil
}

func renderErrorOutput(template *RequestTemplate, resp *http.Response) (*ExecuteResult, error) {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read error response body: %w", err)
	}

	var responseData interface{}
	switch template.Response.Error.InputType {
	case ResponseTypeJSON:
		if err := json.Unmarshal(bodyBytes, &responseData); err != nil {
			responseData = map[string]interface{}{
				"message": string(bodyBytes),
			}
		}
	case ResponseTypeText:
		responseData = string(bodyBytes)
	default:
		responseData = string(bodyBytes)
	}

	// Interpolate error output template
	output := Interpolate(template.Response.Error.Output, responseData, nil)

	return &ExecuteResult{
		Type:    template.Response.Error.OutputType,
		Content: output,
	}, nil
}

func interpolateObject(obj interface{}, data map[string]interface{}) interface{} {
	if obj == nil {
		return nil
	}

	switch v := obj.(type) {
	case string:
		return Interpolate(v, data, nil)
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			result[key] = interpolateObject(value, data)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, value := range v {
			result[i] = interpolateObject(value, data)
		}
		return result
	default:
		return obj
	}
}
