// Package report はイベントの集計とレポート生成を担う。
// このパッケージの出力に発話本文を含めてはならない (設計書の制約)。
package report

import (
	"sort"
	"time"

	"cc-log-metrics/internal/classify"
	"cc-log-metrics/internal/extract"
)

// Group は1グループ (セッション / プロジェクト / 月) の集計値。
type Group struct {
	Granularity       string  `json:"granularity"`
	Key               string  `json:"key"`
	Prompts           int     `json:"prompts"`
	Rewinds           int     `json:"rewinds"`
	Corrections       int     `json:"corrections"`
	PermissionDenies  int     `json:"permissionDenies"`
	Boundaries        int     `json:"boundaries"`
	ToolErrors        int     `json:"toolErrors"`
	Compacts          int     `json:"compacts"`
	SpecChanges       int     `json:"specChanges"`
	ErrorCorrections  int     `json:"errorCorrections"`
	Preferences       int     `json:"preferences"`
	InterventionRate  float64 `json:"interventionRate"`
	AvgReexplainChars float64 `json:"avgReexplainChars"`
	AvgTurnSeconds    float64 `json:"avgTurnSeconds"`
}

// Aggregate はイベントを指定粒度で集計し、Key 昇順で返す。
func Aggregate(events []extract.Event, granularity string) []Group {
	type acc struct {
		g            Group
		reexplainSum float64
		reexplainN   int
		turnMsSum    float64
		turnN        int
	}
	byKey := map[string]*acc{}

	keyOf := func(ev extract.Event) string {
		switch granularity {
		case "session":
			return ev.SessionID
		case "project":
			return ev.Project
		default: // month
			t, err := time.Parse(time.RFC3339, ev.Timestamp)
			if err != nil {
				return "unknown"
			}
			return t.Local().Format("2006-01")
		}
	}

	for _, ev := range events {
		key := keyOf(ev)
		a := byKey[key]
		if a == nil {
			a = &acc{g: Group{Granularity: granularity, Key: key}}
			byKey[key] = a
		}
		switch ev.Kind {
		case extract.KindUserPrompt:
			a.g.Prompts++
		case extract.KindRewind:
			a.g.Rewinds++
		case extract.KindCorrection:
			a.g.Corrections++
			switch ev.Classification {
			case classify.LabelSpecChange:
				a.g.SpecChanges++
			case classify.LabelErrorCorrection:
				a.g.ErrorCorrections++
			case classify.LabelPreference:
				a.g.Preferences++
			}
		case extract.KindPermissionDeny:
			a.g.PermissionDenies++
		case extract.KindClearBoundary:
			a.g.Boundaries++
			if n, ok := toFloat(ev.Detail["nextPromptChars"]); ok && n > 0 {
				a.reexplainSum += n
				a.reexplainN++
			}
		case extract.KindToolError:
			a.g.ToolErrors++
		case extract.KindCompact:
			a.g.Compacts++
		case extract.KindTurn:
			if ms, ok := toFloat(ev.Detail["durationMs"]); ok {
				a.turnMsSum += ms
				a.turnN++
			}
		}
	}

	var groups []Group
	for _, a := range byKey {
		if a.g.Prompts > 0 {
			// rewind はコンテキスト節約目的の巻き戻しを含むため分子に含めない (参考指標)
			a.g.InterventionRate = float64(a.g.Corrections+a.g.PermissionDenies) / float64(a.g.Prompts)
		}
		if a.reexplainN > 0 {
			a.g.AvgReexplainChars = a.reexplainSum / float64(a.reexplainN)
		}
		if a.turnN > 0 {
			a.g.AvgTurnSeconds = a.turnMsSum / float64(a.turnN) / 1000
		}
		groups = append(groups, a.g)
	}
	sort.Slice(groups, func(i, j int) bool { return groups[i].Key < groups[j].Key })
	return groups
}

// toFloat は JSON 経由 (float64) とメモリ直渡し (int, int64) の両方を吸収する。
func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	}
	return 0, false
}
