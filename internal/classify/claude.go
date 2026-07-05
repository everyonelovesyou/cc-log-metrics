package classify

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// ClaudeCLI は claude -p (headless モード) で分類する Classifier 実装。
type ClaudeCLI struct {
	Model string
}

const promptTemplate = `あなたは開発ログの分析者です。以下は、コーディングエージェントとの対話でユーザーが軌道修正を行った発話です。
それぞれを次の3つに分類してください:
- "spec_change": 人間の要求自体が変わった (仕様変更・方針転換)
- "error_correction": AI の理解や成果物が間違っていたことの訂正
- "preference": どちらでもない好み・調整の伝達

発話:
%s

回答は JSON 配列のみで、発話と同じ順・同じ件数で出力してください。例: ["preference","error_correction"]`

func (c ClaudeCLI) ClassifyBatch(utterances []string) ([]string, error) {
	var list strings.Builder
	for i, u := range utterances {
		fmt.Fprintf(&list, "%d. %s\n", i+1, u)
	}
	prompt := fmt.Sprintf(promptTemplate, list.String())

	out, err := exec.Command("claude", "-p", prompt, "--model", c.Model).Output()
	if err != nil {
		return nil, fmt.Errorf("claude の実行に失敗: %w", err)
	}
	return parseLabels(string(out), len(utterances))
}

var validLabels = map[string]bool{
	LabelSpecChange:      true,
	LabelErrorCorrection: true,
	LabelPreference:      true,
}

// parseLabels は応答文字列から最初の JSON 配列を取り出し検証する。
func parseLabels(output string, want int) ([]string, error) {
	start := strings.Index(output, "[")
	end := strings.LastIndex(output, "]")
	if start < 0 || end <= start {
		return nil, fmt.Errorf("応答に JSON 配列がありません: %.80s", output)
	}
	var labels []string
	if err := json.Unmarshal([]byte(output[start:end+1]), &labels); err != nil {
		return nil, fmt.Errorf("応答のパースに失敗: %w", err)
	}
	if len(labels) != want {
		return nil, fmt.Errorf("件数不一致: got %d, want %d", len(labels), want)
	}
	for _, l := range labels {
		if !validLabels[l] {
			return nil, fmt.Errorf("不正なラベル: %q", l)
		}
	}
	return labels, nil
}
