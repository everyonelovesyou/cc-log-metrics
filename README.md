# cc-log-metrics

Claude Code のセッションログ (`~/.claude/projects/*/*.jsonl`) から人間の介入指標を抽出・分類・集計する CLI。

## 背景

AIコーディングエージェントとの協働では、人間が軌道修正を行う場面がしばしば発生する。やり直し (rewind)、発話による方針転換、文脈の切断と再説明などがそれにあたる。これらの「開発摩擦」はセッションログに事象として記録されているが、素の JSONL を眺めるだけでは全体像が見えにくい。

ccmetrics はこのログを構造化し、介入の頻度や傾向を定量的に把握できるようにするツールとして作った。過去数か月分のログからベースラインを算出し、今後の開発プロセス改善の起点とすることが主な目的である。

## アーキテクチャ

extract (ログ → 正規化イベント) と report (イベント → 集計) を分離する二層パイプライン。LLM 分類 (classify) は任意の中間ステージとして挿入できる。

```
jsonl (生ログ)
   │ ccmetrics extract
   ▼
events.jsonl (正規化イベント)
   │ ccmetrics classify (任意・LLM)
   ▼
events.enriched.jsonl
   │ ccmetrics report
   ▼
report.md / metrics.json
```

中間データを JSONL として残すことで、検出ルールの目視検証や集計のやり直しが LLM 再実行なしで行える。

## 検出するイベント

| イベント | 検出方法 | 観測項目 |
|---|---|---|
| permission_deny | 権限プロンプトで No を選んだ痕跡 (tool_result の拒否文言) | ユーザーによる権限拒否の回数 (添付指示は detail に保持) |
| correction | ユーザー発話への語彙パターンマッチ | 発話による軌道修正の回数 |
| rewind | 同一 parentUuid に複数の子がある分岐点 | 人間がAIの進行を巻き戻した回数 (参考指標) |
| clear_boundary / compact | `/clear` コマンド痕跡・セッション開始 | 再説明コスト (境界後の最初のプロンプト文字数) |
| turn | system エントリの turn_duration | ターンあたりの所要時間 |
| tool_error | ツール結果の error フラグ | ツール実行がエラーとして記録された頻度 (ユーザーによる権限拒否として検出したものを除く) |

## 使い方

```sh
go build -o ccmetrics ./cmd/ccmetrics

# 1. 抽出: 生ログ → 正規化イベント
./ccmetrics extract --since 2026-02-01 -o out/events.jsonl

# 2. 分類 (任意): correction を LLM で3分類 (仕様変更/誤り訂正/好みの伝達)
./ccmetrics classify out/events.jsonl -o out/events.enriched.jsonl

# 3. 集計: セッション/プロジェクト/月の3粒度でレポート生成
./ccmetrics report out/events.enriched.jsonl -o out/report.md --json out/metrics.json
```

## 主な指標

- **介入率** = (correction + permission_deny) / user_prompt 数
  - rewind はコンテキスト節約目的の巻き戻しを含むため分子から除外し、参考指標として集計する
- **再説明量** = /clear・セッション開始直後の最初のプロンプト文字数
- correction の分類内訳 (classify 実施時): 仕様変更 / 誤り訂正 / 好みの伝達

## データの取り扱い

- `out/` 配下の中間データ (events.jsonl) は発話本文を含み得るため git 管理外 (`.gitignore` 済み)
- report.md / metrics.json には発話本文を含まない設計

## 技術スタック

- Go 1.24 (標準ライブラリのみ、外部依存なし)
- LLM 分類は `claude -p` (Claude Code headless モード) を `os/exec` で呼び出し
