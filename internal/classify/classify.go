// Package classify は correction イベントの LLM 分類を担う。
package classify

import (
	"fmt"

	"cc-log-metrics/internal/extract"
)

// 分類ラベル。metrics.json では機械可読スラッグ、report.md では日本語で表示する。
const (
	LabelSpecChange      = "spec_change"      // 仕様変更 (人間の要求自体が変わった)
	LabelErrorCorrection = "error_correction" // 誤り訂正 (AI が間違えた)
	LabelPreference      = "preference"       // 好みの伝達 (どちらでもない調整)
)

// Classifier は発話のバッチを 3 分類する。テストではモックに差し替える。
type Classifier interface {
	ClassifyBatch(utterances []string) ([]string, error)
}

// Apply は Classification が空の correction イベントに分類を書き込む。
// 分類済みはスキップするため、中断後の再実行が安全 (冪等)。
func Apply(events []extract.Event, c Classifier, batchSize int) (int, error) {
	var targets []int
	for i, ev := range events {
		if ev.Kind == extract.KindCorrection && ev.Classification == "" {
			targets = append(targets, i)
		}
	}

	classified := 0
	for start := 0; start < len(targets); start += batchSize {
		end := min(start+batchSize, len(targets))
		batch := targets[start:end]
		utterances := make([]string, len(batch))
		for j, idx := range batch {
			utterances[j], _ = events[idx].Detail["utterance"].(string)
		}
		labels, err := c.ClassifyBatch(utterances)
		if err != nil {
			return classified, fmt.Errorf("分類に失敗 (%d 件は分類済み): %w", classified, err)
		}
		if len(labels) != len(batch) {
			return classified, fmt.Errorf("分類結果の件数不一致: got %d, want %d", len(labels), len(batch))
		}
		for j, idx := range batch {
			events[idx].Classification = labels[j]
			classified++
		}
	}
	return classified, nil
}
