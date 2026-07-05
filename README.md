# cc-log-metrics

Claude Code のセッションログ (`~/.claude/projects/*/*.jsonl`) から人間の介入指標を抽出・分類・集計する CLI。

設計書: [docs/specs/20260705-cc-log-metrics-design.md](docs/specs/20260705-cc-log-metrics-design.md)

## 使い方

    go build -o ccmetrics ./cmd/ccmetrics

    # 1. 抽出: 生ログ → 正規化イベント
    ./ccmetrics extract --since 2026-02-01 -o out/events.jsonl

    # 2. 分類 (任意): correction を LLM で3分類 (仕様変更/誤り訂正/好みの伝達)
    ./ccmetrics classify out/events.jsonl -o out/events.enriched.jsonl

    # 3. 集計: セッション/プロジェクト/月の3粒度でレポート生成
    ./ccmetrics report out/events.enriched.jsonl -o out/report.md --json out/metrics.json

## データの取り扱い

- `out/` (events.jsonl・レポート類) は発話本文を含み得るため git 管理外
- report.md / metrics.json には発話本文は含まれない設計だが、職場リポジトリへ持ち込むのはこの2ファイル (またはそこから転記した数値) のみとすること

## 主な指標

- 介入率 = (rewind + correction) / user_prompt 数
- 再説明量 = /clear・セッション開始直後の最初のプロンプト文字数
