package report

import (
	"strings"
	"testing"
)

func TestRenderMarkdownExcludesUtterances(t *testing.T) {
	groups := map[string][]Group{
		"session": Aggregate(sampleEvents(), "session"),
		"project": Aggregate(sampleEvents(), "project"),
		"month":   Aggregate(sampleEvents(), "month"),
	}
	var sb strings.Builder
	if err := RenderMarkdown(&sb, groups); err != nil {
		t.Fatalf("RenderMarkdown: %v", err)
	}
	md := sb.String()
	// 発話本文が漏れていないこと (Global Constraints)
	if strings.Contains(md, "そうじゃなくて") {
		t.Error("report.md に発話本文が含まれている")
	}
	for _, want := range []string{"s1", "/p/alpha", "2026-06", "介入率"} {
		if !strings.Contains(md, want) {
			t.Errorf("report.md に %q がない", want)
		}
	}
	// 権限拒否の添付メッセージが漏れていないこと (Global Constraints)
	if strings.Contains(md, "私がやります") {
		t.Error("report.md に権限拒否の添付メッセージが含まれている")
	}
	// 主表に権限拒否列があり、rewind は参考指標節に移っていること
	if !strings.Contains(md, "| キー | プロンプト | 軌道修正 | 権限拒否 | 介入率 |") {
		t.Error("主表のヘッダーに権限拒否列がない")
	}
	if !strings.Contains(md, "参考指標") {
		t.Error("rewind の参考指標節がない")
	}
}

func TestRenderJSON(t *testing.T) {
	groups := map[string][]Group{"session": Aggregate(sampleEvents(), "session")}
	var sb strings.Builder
	if err := RenderJSON(&sb, groups); err != nil {
		t.Fatalf("RenderJSON: %v", err)
	}
	out := sb.String()
	if !strings.Contains(out, `"generatedAt"`) || !strings.Contains(out, `"interventionRate"`) {
		t.Errorf("metrics.json の構造が不正: %s", out)
	}
	if strings.Contains(out, "そうじゃなくて") {
		t.Error("metrics.json に発話本文が含まれている")
	}
	if !strings.Contains(out, `"permissionDenies"`) {
		t.Error("metrics.json に permissionDenies がない")
	}
	if strings.Contains(out, "私がやります") {
		t.Error("metrics.json に権限拒否の添付メッセージが含まれている")
	}
}
