package rpa

import "testing"

func TestRuntimeContext_ResolveTemplateAndExpression(t *testing.T) {
	ctx := NewRuntimeContext("profile-a")
	ctx.Set("pageTitle", "控制台首页")
	ctx.Set("username", "alice")

	text, err := ResolveTemplate("欢迎 ${username}", ctx)
	if err != nil {
		t.Fatalf("模板解析失败: %v", err)
	}
	if text != "欢迎 alice" {
		t.Fatalf("模板解析错误: %q", text)
	}

	ok, err := EvalBoolExpression(`pageTitle == "控制台首页"`, ctx)
	if err != nil {
		t.Fatalf("表达式执行失败: %v", err)
	}
	if !ok {
		t.Fatal("条件判断应为 true")
	}
}
