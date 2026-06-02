# CLAUDE.md

This file provides guidance for AI CLIs (such as Claude Code or ChatGPT Codex) when working in this repository.

## Language Conventions

- Documentation and comments: Traditional Chinese
- Code naming (variables, functions, classes): English
- Implementation plans: Chinese descriptions with English technical terms
- For canonical Chinese-English naming (events, attributes, commands, fields, identifiers), refer to the glossary in `doc/營業規格書.md` (section 二、英文詞彙對照).

### Spec Citation Format

Two formats — pick by where the citation lives. Tell at a glance: **full-width `｜` and no 檔名 → a Markdown spec doc; spaced half-width `|` with 檔名 → Go code.**

**In Markdown spec docs (`doc/*.md`):**

- Format: `【章編號、章名｜節編號. 節名｜子節】` — levels separated by full-width `｜`; no spaces around `【` `】` (the full-width brackets carry their own spacing).
- 節 mirrors the `### N.` heading: half-width `.` + space (e.g. `【十七、命令｜1. 屬性修改命令】`). A titled (un-numbered) subsection follows `｜` directly (e.g. `【二十、獨立流程｜觸發時機】`). A whole-chapter cite drops 節 (e.g. `【七、效果類型】`).
- Omit 檔名 for an intra-doc cite; when pointing at another spec, name it in prose before the bracket (e.g. 詳見營業實作規格書【四、解耦的關鍵：邊界介面】).
- Verb roles: `詳見【…】` (info pointer) / `依【…】` (follow the rule there) / `使用【…】` (use the command/syntax) / `引用【…】` (reuse the instance/definition). Do not use `見【…】`. When `【…】` is itself the subject/object and the syntax leaves no room, the verb prefix may be dropped.
- Why full-width `｜`: a half-width `|` would break Markdown tables.

**In Go code comments (`.go`):**

- Format: `【檔名 | 章節編號、章節名稱 | 小節】` — segments separated by half-width `|` with one surrounding space; the third segment (`小節`) is optional.
- `檔名` required: one of `營業規格書` / `營業實作規格書` / `營業顯示規格書`. Do not also write the doc name as prose before the bracket.
- `章節編號、章節名稱`: copied from the chapter's `##` heading; separator always `、`. Numbering follows the source doc — Chinese numerals in `營業規格書` / `營業實作規格書`, Arabic in `營業顯示規格書`.
- `小節` (optional): a number (`### 1.` → `1`), a title (`### 觸發時機` → `觸發時機`), or a named entry inside a list/table (e.g. `phaseJump`).
- The 編號 is authoritative for locating; the 名稱 is for readability. If they drift, the 編號 wins.
- Examples: `【營業規格書 | 二十六、內建函式清單】`、`【營業規格書 | 十七、命令 | 1】`、`【營業顯示規格書 | 3、事件流的消費：速率與步進】`.
- Why half-width `|` + 檔名: comments aren't inside Markdown tables (no clash), and code always cites across into the docs.

## Context Management

- Manage context usage proactively during long conversations or when working with large files.
- In long conversations, periodically summarize key decisions, current status, and remaining tasks.
- When working with files longer than 500 lines, handle them in focused sections instead of loading everything at once. The spec docs under `doc/` are large (1000+ lines) — read them in focused sections.
- Before complex multi-step tasks, briefly restate the relevant code style rules and project constraints.
- If context becomes overloaded, warn the user and suggest starting a new conversation or consolidating the current state into a file.

## Session Continuity

- The repo root has `PROGRESS.md` — the cross-session handoff point. It records milestone status, carryover to-dos, and locked-in design decisions that are NOT visible from git history or the `doc/` specs.
- At the START of a work session, read `PROGRESS.md` to pick up where the last session left off.
- At the END of a work session — or whenever a milestone / sub-task finishes — update `PROGRESS.md`: milestone status, any new carryover to-dos, and design decisions just settled with the user.
- Keep it lean: record only what git and the specs don't already show. Rule details defer to `doc/`; code state defers to git. `PROGRESS.md` holds the deferred to-dos and the "why" decisions, so they survive context loss.

## Workflow

- For multi-step, repetitive, or cross-file tasks, prefer a Python script over a long chain of shell commands.
- Reach for Python when the task involves: batch file operations; parsing or transforming structured data; or any loop / conditional / cross-file text processing.

## Code Style

*Applies to hand-written Go code.*

- In Go, `false` checks must be explicit. Negation-style checks are forbidden.
  - Example: `if x == false {}`
  - Forbidden: `if !x {}`

- Closing comments are only allowed for control flow blocks: `if`, `for`, `switch`
  - Example: `} // if`
- Closing comments on functions or methods are forbidden.

- When a function returns more than one value, every return value must be named (single-return functions need not be named). Applies to hand-written code; generated code under `sheet/` is exempt.
  - Example: `func Load(dir string) (data *sheeter.Sheeter, err error)`
  - Forbidden: `func Load(dir string) (*sheeter.Sheeter, error)`

- Names must always be singular — variables, struct fields, parameters, and named return values — even for slices, arrays, maps, and other collections.
  - Example: `item := []Item{}`, `guest := []Guest{}`, `err []error`
  - Forbidden: `items := []Item{}`, `guests := []Guest{}`, `errs []error`

- Iterator naming:
  - Use `itor` for general iteration
  - Use `k, v` for map iteration

## Spec Authoring & Review (doc/*.md)

*Applies when authoring or reviewing the spec docs under `doc/` — parallel to Code Style above, which governs Go code.*

- **Layer first**: mark the current discussion layer (architecture / flow / mechanism / detail); don't mix layers in one reply, and get the user's OK before drilling to the next layer.
- **Defer edge cases**: when an edge case, exception, or conflict surfaces, note it as a bullet — don't expand a solution on the spot; batch them once the main architecture is stable.
- **One issue per reply**: avoid "while we're at it…".
- **Don't auto-fill**: don't proactively add examples, tables, or reserved-word lists unless asked, or unless the omission causes a real semantic ambiguity.
- **Short by default**: a review reply defaults to ≤5 paragraphs; go longer only when the user asks or the issue is genuinely complex.
- **Bullets over prose**: prefer bullets; split long sentences into separate lines.
- **Propose, don't edit**: in the review phase, default to observations and suggestions only — don't Edit the docs unless the user explicitly says to land it.
- **SSOT, don't restate**: write a rule's detail only in its SSOT section; elsewhere point with `詳見【…】` rather than repeating it (even when restating would be convenient).
- **Lock vocabulary**: before editing, check 【三、名詞列表】 / 【二、英文詞彙對照】 and reuse existing terms; list any new term separately and confirm with the user first.
- **Concise first**: omit non-essential edge cases / notes / defensive prose by default; write the main case plus the general rule rather than enumerating every case.

## Commit Message Convention

Format: `Type | Description`

| Type      | Usage                 |
|:----------|:----------------------|
| `Feature` | New feature or update |
| `Fix`     | Bug fix               |
| `Sheet`   | Update sheet data     |
| `Doc`     | Update documentation  |

- Description must be in Traditional Chinese.
- Do not invent new type names (e.g. `Bugfix`, `Update`, `Refactor` are all forbidden).

## Project Overview

RovingDiner (流浪食堂) is the rules engine for the turn-based "營業" (operations) phase of a restaurant-management game, written in Go. A single 營業 is one settleable game run: the player processes customers with hand cards and skills while the system advances rounds through trigger timings, an effect queue, and settlement rules.

The project is in an early stage. It currently contains the Sheeter-based game-data pipeline (`gamedata/` → `sheet/` + `sheetdata/`) and the design specs under `doc/`. Business logic (the actual operations loop, modules, runtime instances) is not yet implemented.

## Design Specs (`doc/`)

The project's design is rooted in three spec docs under `doc/`. They are the source of truth that the code (not yet written) must conform to — consult the relevant one before implementing. Each is large (1000+ lines); read the section you need, not the whole file.

| Doc                 | Role                                                                                                                                                                                                                                                                                                                                                            | Consult when                                                                                                                                       |
|:--------------------|:----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:---------------------------------------------------------------------------------------------------------------------------------------------------|
| `營業規格書.md`     | **SSOT for all game rules.** Full spec of the 營業 phase: core flow, mechanics (cards / guests / skills / effects / commands), trigger timings, settlement, end conditions, and the 英文詞彙對照 glossary (§二、英文詞彙對照).                                                                                                                                 | Any question about *what the rules are* or the expected behavior. Every rule detail defers to this file.                                           |
| `營業實作規格書.md` | **Project architecture & engineering decisions.** Package structure, the core engine (yield-per-unit event stream, single seeded PRNG, determinism), behavior ports (`Operator` / `Presenter` / `Rander`) plus static data injected directly as `*sheeter.Sheeter`, core↔display decoupling, the portable C# subset, testing strategy, and test-data builders. | Deciding *where code goes*, package / interface design, the engine / event-stream architecture, decoupling, or how to structure tests / test data. |
| `營業顯示規格書.md` | **TUI display / operation layer.** The Bubble Tea debug viewer, how the display *consumes* the event stream (fast / slow / step rates), screen layout & panels, event log, interaction & modals, ASCII+CJK rendering policy.                                                                                                                                    | Working on the TUI / display layer, screen rendering, event-log display, or stepping / selection UI.                                               |

Hierarchy: `營業規格書.md` is the SSOT for *rules*; `營業實作規格書.md` owns the *engine & architecture* (incl. the event-stream definition); `營業顯示規格書.md` only describes how the display *consumes* that engine. An implementation doc must never contradict the rules spec — if it does, the rules spec wins; flag the discrepancy rather than following the implementation doc.

## Repository Layout

| Path         | Contents                                                                                                                                                                                                                                                                                                        |
|:-------------|:----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `gamedata/`  | Source xlsx tables and the Sheeter build config/script (`sheeter.yaml`, `build.bat`)                                                                                                                                                                                                                            |
| `sheet/`     | Sheeter-generated Go readers. Generated code — DO NOT EDIT by hand.                                                                                                                                                                                                                                            |
| `sheetdata/` | Sheeter-generated JSON data. Generated — DO NOT EDIT by hand.                                                                                                                                                                                                                                                  |
| `doc/`       | Design specs — see [Design Specs](#design-specs-doc). `營業規格書.md` (rules SSOT) · `營業實作規格書.md` (architecture & engine) · `營業顯示規格書.md` (TUI display layer). Build tooling under `doc/build-md/` + `doc/build-html/`; HTML output `doc/營業規格書.html` — see [Doc Pipeline](#doc-pipeline). |

## Development / Build / Common Commands

```bash
task lint       # Format + lint code, markdownlint --fix all *.md (root + doc/), normalize md tables (doc/*.md + CLAUDE.md + README.md), prettier on yaml
task doc        # Rebuild doc/營業規格書.html from the SSOT
task sheet      # Regenerate sheet code + data from gamedata/*.xlsx, then lint
task install    # Install dev tools (golangci-lint, sheeter, markdownlint, prettier)
```

### Sheet Pipeline

`sheet/` and `sheetdata/` are generated by [Sheeter](https://github.com/yinweli/Sheeter) from the xlsx files in `gamedata/`. To change game data, edit the xlsx files (Card, Effect, Guest, Seat, Setting, Skill) and run `task sheet` — never edit the generated `.go` or `.json` files directly. The generation is driven by `gamedata/sheeter.yaml` and `gamedata/build.bat`.

### Doc Pipeline

Doc work is split across two tasks. The tooling lives under `doc/` (Python 3, stdlib only):

- **Linting + table normalization run in `task lint`** — `doc/build-md/normalize-md-tables.py` aligns Markdown table column widths for monospace CJK (East Asian Ambiguous chars counted as width 2), adjusting cell padding only, never content. It normalizes all `doc/*.md` plus `CLAUDE.md` and `README.md`, and runs before markdownlint `--fix` (which covers all `*.md`, root + `doc/`) so the checks see the final aligned tables. markdownlint does not touch cell padding, so for tables the two are order-independent regardless.
- **HTML build runs in `task doc`**, via `doc/build-html/build.py` — it slices `doc/營業規格書.md` by section anchors and injects it plus `template.html` and the 14 `mermaid/*.mmd` flow diagrams into placeholders, writing `doc/營業規格書.html`. **SSOT-only**: only `營業規格書.md` has a template + mermaid set; the other two specs get no HTML. Normalization doesn't affect the HTML (the build slices by heading, so cell padding is irrelevant), so `task doc` need not lint or normalize first.

So after editing a spec: `task lint` keeps the `.md` tables aligned, `task doc` regenerates the HTML. When editing 【十九、核心流程】or【二十、獨立流程】in the SSOT, check whether the matching `doc/build-html/mermaid/flow-core-*.mmd` / `flow-sub-*.mmd` diagrams need updating before rebuilding. `build.py` matches sections by heading regex; renaming a spec heading (e.g. `## 二十二、觸發時機清單`) without updating the anchor in `build.py` will fail the build with a clear "找不到起始錨點" error — fix the anchor, don't work around it.

When editing 【十九、核心流程】or【二十、獨立流程】in the SSOT, check whether the matching `doc/build-html/mermaid/flow-core-*.mmd` / `flow-sub-*.mmd` diagrams need updating before rebuilding. `build.py` matches sections by heading regex; renaming a spec heading (e.g. `## 二十二、觸發時機清單`) without updating the anchor in `build.py` will fail the build with a clear "找不到起始錨點" error — fix the anchor, don't work around it.
