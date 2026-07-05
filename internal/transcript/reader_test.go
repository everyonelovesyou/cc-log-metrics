package transcript

import "testing"

func TestReadFile(t *testing.T) {
	entries, stats, err := ReadFile("../../testdata/session_small.jsonl")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if stats.Lines != 12 {
		t.Errorf("Lines = %d, want 12", stats.Lines)
	}
	if stats.Skipped != 1 {
		t.Errorf("Skipped = %d, want 1", stats.Skipped)
	}
	if len(entries) != 11 {
		t.Fatalf("len(entries) = %d, want 11", len(entries))
	}
	if entries[0].Type != "user" || entries[0].UUID != "u1" {
		t.Errorf("entries[0] = %+v, want type=user uuid=u1", entries[0])
	}
	if entries[5].Subtype != "turn_duration" || entries[5].DurationMs != 30000 {
		t.Errorf("turn_duration の解釈が不正: %+v", entries[5])
	}
}
