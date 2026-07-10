package report

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"time"
)

var granularityTitles = map[string]string{
	"session": "セッション別",
	"project": "プロジェクト別",
	"month":   "月別",
}

// RenderMarkdown は人間が読むレポートを書き出す。発話本文は含めない。
func RenderMarkdown(w io.Writer, groups map[string][]Group) error {
	fmt.Fprintf(w, "# ccmetrics レポート\n\n生成日時: %s\n", time.Now().Format("2006-01-02 15:04"))
	for _, gran := range []string{"month", "project", "session"} {
		gs, ok := groups[gran]
		if !ok {
			continue
		}
		fmt.Fprintf(w, "\n## %s\n\n", granularityTitles[gran])
		fmt.Fprintln(w, "| キー | プロンプト | 軌道修正 | 権限拒否 | 介入率 | 境界 | 平均再説明量 | tool_error | 平均ターン秒 |")
		fmt.Fprintln(w, "|---|---:|---:|---:|---:|---:|---:|---:|---:|")
		for _, g := range gs {
			fmt.Fprintf(w, "| %s | %d | %d | %d | %.2f | %d | %.0f | %d | %.0f |\n",
				g.Key, g.Prompts, g.Corrections, g.PermissionDenies, g.InterventionRate,
				g.Boundaries, g.AvgReexplainChars, g.ToolErrors, g.AvgTurnSeconds)
		}
		if hasClassification(gs) {
			fmt.Fprintln(w, "\n分類内訳 (classify 実施済みの軌道修正のみ):")
			fmt.Fprintln(w, "\n| キー | 仕様変更 | 誤り訂正 | 好みの伝達 |")
			fmt.Fprintln(w, "|---|---:|---:|---:|")
			for _, g := range gs {
				fmt.Fprintf(w, "| %s | %d | %d | %d |\n", g.Key, g.SpecChanges, g.ErrorCorrections, g.Preferences)
			}
		}
		fmt.Fprintln(w, "\n参考指標 (介入率の分子に含めない。rewind はコンテキスト節約目的の巻き戻しを含むため):")
		fmt.Fprintln(w, "\n| キー | rewind |")
		fmt.Fprintln(w, "|---|---:|")
		for _, g := range gs {
			fmt.Fprintf(w, "| %s | %d |\n", g.Key, g.Rewinds)
		}
	}
	renderAnomalies(w, groups["session"])
	return nil
}

func hasClassification(gs []Group) bool {
	for _, g := range gs {
		if g.SpecChanges+g.ErrorCorrections+g.Preferences > 0 {
			return true
		}
	}
	return false
}

// renderAnomalies はプロンプト10件以上で介入率上位5セッションを特異点として示す。
func renderAnomalies(w io.Writer, sessions []Group) {
	var candidates []Group
	for _, g := range sessions {
		if g.Prompts >= 10 {
			candidates = append(candidates, g)
		}
	}
	if len(candidates) == 0 {
		return
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].InterventionRate > candidates[j].InterventionRate })
	if len(candidates) > 5 {
		candidates = candidates[:5]
	}
	fmt.Fprintln(w, "\n## 特異点 (介入率の高いセッション、プロンプト10件以上)")
	fmt.Fprintln(w, "\n| セッション | プロンプト | 介入率 |")
	fmt.Fprintln(w, "|---|---:|---:|")
	for _, g := range candidates {
		fmt.Fprintf(w, "| %s | %d | %.2f |\n", g.Key, g.Prompts, g.InterventionRate)
	}
}

// RenderJSON は機械可読の metrics.json を書き出す。
func RenderJSON(w io.Writer, groups map[string][]Group) error {
	var flat []Group
	for _, gran := range []string{"month", "project", "session"} {
		flat = append(flat, groups[gran]...)
	}
	doc := struct {
		GeneratedAt string  `json:"generatedAt"`
		Groups      []Group `json:"groups"`
	}{time.Now().Format(time.RFC3339), flat}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(doc)
}
