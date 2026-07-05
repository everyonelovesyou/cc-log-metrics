package classify

import (
	"testing"

	"cc-log-metrics/internal/extract"
)

// fakeClassifier は呼ばれた発話を記録し、固定ラベルを返す
type fakeClassifier struct {
	calls [][]string
	label string
}

func (f *fakeClassifier) ClassifyBatch(utterances []string) ([]string, error) {
	f.calls = append(f.calls, utterances)
	labels := make([]string, len(utterances))
	for i := range labels {
		labels[i] = f.label
	}
	return labels, nil
}

func corr(utterance, classification string) extract.Event {
	return extract.Event{
		Kind:           extract.KindCorrection,
		Detail:         map[string]any{"utterance": utterance, "pattern": "x"},
		Classification: classification,
	}
}

func TestApplyClassifiesOnlyUnclassifiedCorrections(t *testing.T) {
	events := []extract.Event{
		{Kind: extract.KindUserPrompt},
		corr("そうじゃなくて A", ""),
		corr("やっぱり B", LabelPreference), // 分類済み → スキップ (冪等)
		corr("ではなく C", ""),
	}
	fc := &fakeClassifier{label: LabelErrorCorrection}
	n, err := Apply(events, fc, 10)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if n != 2 {
		t.Errorf("classified = %d, want 2", n)
	}
	if events[1].Classification != LabelErrorCorrection || events[3].Classification != LabelErrorCorrection {
		t.Errorf("未分類の correction に分類が付いていない: %+v", events)
	}
	if events[2].Classification != LabelPreference {
		t.Errorf("分類済みが上書きされた: %v", events[2].Classification)
	}
	if len(fc.calls) != 1 || len(fc.calls[0]) != 2 {
		t.Errorf("呼び出し = %+v, want 1 回・2 件", fc.calls)
	}
}

func TestApplyBatching(t *testing.T) {
	var events []extract.Event
	for i := 0; i < 5; i++ {
		events = append(events, corr("うーん", ""))
	}
	fc := &fakeClassifier{label: LabelPreference}
	if _, err := Apply(events, fc, 2); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	// 5 件をバッチ 2 で → 3 回 (2+2+1)
	if len(fc.calls) != 3 {
		t.Errorf("バッチ回数 = %d, want 3", len(fc.calls))
	}
}
