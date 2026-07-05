package main

import "testing"

func TestRun(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"サブコマンドなし", nil, true},
		{"不明なサブコマンド", []string{"bogus"}, true},
		{"extract は未実装エラー", []string{"extract"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := run(tt.args); (err != nil) != tt.wantErr {
				t.Errorf("run(%v) error = %v, wantErr %v", tt.args, err, tt.wantErr)
			}
		})
	}
}
