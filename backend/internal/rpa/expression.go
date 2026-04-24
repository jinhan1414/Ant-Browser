package rpa

import (
	"fmt"
	"strings"
)

func ResolveTemplate(template string, ctx *RuntimeContext) (string, error) {
	result := template
	for {
		start := strings.Index(result, "${")
		if start < 0 {
			return result, nil
		}
		end := strings.Index(result[start:], "}")
		if end < 0 {
			return "", fmt.Errorf("template placeholder not closed")
		}
		end += start
		name := strings.TrimSpace(result[start+2 : end])
		value := contextStringValue(ctx, name)
		result = result[:start] + value + result[end+1:]
	}
}

func EvalBoolExpression(expr string, ctx *RuntimeContext) (bool, error) {
	text := strings.TrimSpace(expr)
	switch {
	case strings.Contains(text, "=="):
		left, right, err := splitBinary(text, "==")
		if err != nil {
			return false, err
		}
		return resolveValue(left, ctx) == trimQuoted(right), nil
	case strings.Contains(text, "!="):
		left, right, err := splitBinary(text, "!=")
		if err != nil {
			return false, err
		}
		return resolveValue(left, ctx) != trimQuoted(right), nil
	case strings.HasPrefix(text, "contains(") && strings.HasSuffix(text, ")"):
		args := strings.TrimSuffix(strings.TrimPrefix(text, "contains("), ")")
		parts := strings.SplitN(args, ",", 2)
		if len(parts) != 2 {
			return false, fmt.Errorf("contains arguments invalid")
		}
		return strings.Contains(resolveValue(parts[0], ctx), trimQuoted(parts[1])), nil
	case strings.HasPrefix(text, "empty(") && strings.HasSuffix(text, ")"):
		name := strings.TrimSuffix(strings.TrimPrefix(text, "empty("), ")")
		return resolveValue(name, ctx) == "", nil
	default:
		return false, fmt.Errorf("unsupported expression: %s", expr)
	}
}

func splitBinary(expr string, token string) (string, string, error) {
	parts := strings.SplitN(expr, token, 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("expression split failed: %s", expr)
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), nil
}

func resolveValue(name string, ctx *RuntimeContext) string {
	return contextStringValue(ctx, strings.TrimSpace(name))
}

func trimQuoted(value string) string {
	text := strings.TrimSpace(value)
	text = strings.TrimPrefix(text, `"`)
	text = strings.TrimSuffix(text, `"`)
	text = strings.TrimPrefix(text, `'`)
	text = strings.TrimSuffix(text, `'`)
	return text
}

func contextStringValue(ctx *RuntimeContext, name string) string {
	value := ctx.Get(strings.TrimSpace(name))
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}
