package rpa

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	cronMinYear = 1970
	cronMaxYear = 2099
)

var cronMonthAliases = map[string]int{
	"jan": 1,
	"feb": 2,
	"mar": 3,
	"apr": 4,
	"may": 5,
	"jun": 6,
	"jul": 7,
	"aug": 8,
	"sep": 9,
	"oct": 10,
	"nov": 11,
	"dec": 12,
}

var cronWeekdayAliases = map[string]int{
	"sun": 0,
	"mon": 1,
	"tue": 2,
	"wed": 3,
	"thu": 4,
	"fri": 5,
	"sat": 6,
}

type cronField struct {
	wildcard bool
	allowed  map[int]struct{}
}

type cronExpression struct {
	second     cronField
	minute     cronField
	hour       cronField
	dayOfMonth cronField
	month      cronField
	dayOfWeek  cronField
	year       cronField
}

func parseCronExpression(raw string) (*cronExpression, error) {
	parts, err := normalizeCronParts(raw)
	if err != nil {
		return nil, err
	}

	second, err := parseCronField(parts[0], 0, 59, nil, false)
	if err != nil {
		return nil, fmt.Errorf("解析秒字段失败: %w", err)
	}
	minute, err := parseCronField(parts[1], 0, 59, nil, false)
	if err != nil {
		return nil, fmt.Errorf("解析分钟字段失败: %w", err)
	}
	hour, err := parseCronField(parts[2], 0, 23, nil, false)
	if err != nil {
		return nil, fmt.Errorf("解析小时字段失败: %w", err)
	}
	dayOfMonth, err := parseCronField(parts[3], 1, 31, nil, false)
	if err != nil {
		return nil, fmt.Errorf("解析日字段失败: %w", err)
	}
	month, err := parseCronField(parts[4], 1, 12, cronMonthAliases, false)
	if err != nil {
		return nil, fmt.Errorf("解析月字段失败: %w", err)
	}
	dayOfWeek, err := parseCronField(parts[5], 0, 6, cronWeekdayAliases, true)
	if err != nil {
		return nil, fmt.Errorf("解析周字段失败: %w", err)
	}
	year, err := parseCronField(parts[6], cronMinYear, cronMaxYear, nil, false)
	if err != nil {
		return nil, fmt.Errorf("解析年字段失败: %w", err)
	}

	return &cronExpression{
		second:     second,
		minute:     minute,
		hour:       hour,
		dayOfMonth: dayOfMonth,
		month:      month,
		dayOfWeek:  dayOfWeek,
		year:       year,
	}, nil
}

func normalizeCronParts(raw string) ([]string, error) {
	parts := strings.Fields(strings.TrimSpace(raw))
	switch len(parts) {
	case 5:
		return []string{"0", parts[0], parts[1], parts[2], parts[3], parts[4], "*"}, nil
	case 6:
		return []string{parts[0], parts[1], parts[2], parts[3], parts[4], parts[5], "*"}, nil
	case 7:
		return parts, nil
	default:
		return nil, fmt.Errorf("Cron 表达式必须是 5、6 或 7 段")
	}
}

func parseCronField(token string, min int, max int, aliases map[string]int, isWeekday bool) (cronField, error) {
	normalized := strings.TrimSpace(strings.ToLower(token))
	if normalized == "" {
		return cronField{}, fmt.Errorf("字段不能为空")
	}
	if normalized == "*" || normalized == "?" {
		return cronField{wildcard: true}, nil
	}

	field := cronField{allowed: make(map[int]struct{})}
	for _, segment := range strings.Split(normalized, ",") {
		if err := addCronSegment(field.allowed, strings.TrimSpace(segment), min, max, aliases, isWeekday); err != nil {
			return cronField{}, err
		}
	}
	if len(field.allowed) == 0 {
		return cronField{}, fmt.Errorf("字段没有有效取值")
	}
	return field, nil
}

func addCronSegment(target map[int]struct{}, segment string, min int, max int, aliases map[string]int, isWeekday bool) error {
	base, step, stepped, err := splitCronStep(segment)
	if err != nil {
		return err
	}
	start, end, err := parseCronRange(base, min, max, aliases, isWeekday, stepped)
	if err != nil {
		return err
	}
	for value := start; value <= end; value += step {
		target[value] = struct{}{}
	}
	return nil
}

func splitCronStep(segment string) (string, int, bool, error) {
	parts := strings.Split(segment, "/")
	if len(parts) > 2 {
		return "", 0, false, fmt.Errorf("非法步长表达式: %s", segment)
	}
	if len(parts) == 1 {
		return parts[0], 1, false, nil
	}
	step, err := strconv.Atoi(parts[1])
	if err != nil || step <= 0 {
		return "", 0, false, fmt.Errorf("非法步长: %s", segment)
	}
	return parts[0], step, true, nil
}

func parseCronRange(base string, min int, max int, aliases map[string]int, isWeekday bool, stepped bool) (int, int, error) {
	if base == "*" || base == "?" || base == "" {
		return min, max, nil
	}
	if strings.Contains(base, "-") {
		parts := strings.SplitN(base, "-", 2)
		start, err := parseCronValue(parts[0], min, max, aliases, isWeekday)
		if err != nil {
			return 0, 0, err
		}
		end, err := parseCronValue(parts[1], min, max, aliases, isWeekday)
		if err != nil {
			return 0, 0, err
		}
		if start > end {
			return 0, 0, fmt.Errorf("范围起点大于终点: %s", base)
		}
		return start, end, nil
	}
	start, err := parseCronValue(base, min, max, aliases, isWeekday)
	if err != nil {
		return 0, 0, err
	}
	if !stepped {
		return start, start, nil
	}
	return start, max, nil
}

func parseCronValue(raw string, min int, max int, aliases map[string]int, isWeekday bool) (int, error) {
	token := strings.TrimSpace(strings.ToLower(raw))
	if value, ok := aliases[token]; ok {
		return value, nil
	}
	value, err := strconv.Atoi(token)
	if err != nil {
		return 0, fmt.Errorf("非法取值: %s", raw)
	}
	if isWeekday && value == 7 {
		value = 0
	}
	if value < min || value > max {
		return 0, fmt.Errorf("取值超出范围: %s", raw)
	}
	return value, nil
}

func (f cronField) matches(value int) bool {
	if f.wildcard {
		return true
	}
	_, ok := f.allowed[value]
	return ok
}

func (expr *cronExpression) matchesTime(now time.Time) bool {
	if expr == nil {
		return false
	}
	if !expr.second.matches(now.Second()) {
		return false
	}
	if !expr.minute.matches(now.Minute()) {
		return false
	}
	if !expr.hour.matches(now.Hour()) {
		return false
	}
	if !expr.dayOfMonth.matches(now.Day()) {
		return false
	}
	if !expr.month.matches(int(now.Month())) {
		return false
	}
	if !expr.dayOfWeek.matches(int(now.Weekday())) {
		return false
	}
	return expr.year.matches(now.Year())
}
