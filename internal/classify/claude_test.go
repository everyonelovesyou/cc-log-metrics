package classify

import (
	"strconv"
	"strings"
	"testing"
)

func TestParseLabels(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		want    int
		wantErr bool
	}{
		{"素の JSON 配列", `["spec_change","preference"]`, 2, false},
		{"前後に説明文", "以下です。\n[\"error_correction\"]\nどうぞ", 1, false},
		{"件数不一致", `["preference"]`, 2, true},
		{"不正ラベル", `["banana"]`, 1, true},
		{"配列なし", "わかりません", 1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			labels, err := parseLabels(tt.output, tt.want)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && len(labels) != tt.want {
				t.Errorf("len = %d, want %d", len(labels), tt.want)
			}
		})
	}
}

func TestMainRejectsNonPositiveBatch(t *testing.T) {
	tests := []int{0, -1}
	for _, batch := range tests {
		err := Main([]string{"--batch", strconv.Itoa(batch), "out/does-not-exist.jsonl"})
		if err == nil {
			t.Fatalf("--batch %d: エラーを期待したが nil だった", batch)
		}
		if !strings.Contains(err.Error(), "--batch") {
			t.Errorf("--batch %d: エラーメッセージに --batch が含まれていない: %v", batch, err)
		}
	}
}
