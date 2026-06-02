# CLAUDE.md

This file provides guidance for AI CLIs (such as Claude Code or ChatGPT Codex) when working in this repository.

## Language Conventions

- Documentation and comments: Traditional Chinese
- Code naming (variables, functions, classes): English
- Implementation plans: Chinese descriptions with English technical terms
- For canonical Chinese-English naming (events, attributes, commands, fields, identifiers), refer to the glossary in `doc/營業規格書.md` (section 二、英文詞彙對照).

## Context Management

- Manage context usage proactively during long conversations or when working with large files.
- In long conversations, periodically summarize key decisions, current status, and remaining tasks.
- When working with files longer than 500 lines, handle them in focused sections instead of loading everything at once. The spec docs under `doc/` are large (1000+ lines) — read them in focused sections.
- Before complex multi-step tasks, briefly restate the relevant code style rules and project constraints.
- If context becomes overloaded, warn the user and suggest starting a new conversation or consolidating the current state into a file.

## Code Style

- In Go, `false` checks must be explicit. Negation-style checks are forbidden.
  - Example: `if x == false {}`
  - Forbidden: `if !x {}`

- Closing comments are only allowed for control flow blocks: `if`, `for`, `switch`
  - Example: `} // if`
- Closing comments on functions or methods are forbidden.

- When a function returns more than one value, every return value must be named (single-return functions need not be named). Applies to hand-written code; generated code under `sheet/` is exempt.
  - Example: `func Load(dir string) (data *sheeter.Sheeter, err error)`
  - Forbidden: `func Load(dir string) (*sheeter.Sheeter, error)`

- Variable names must always be singular, even for slices, arrays, maps, and other collections.
  - Example: `item := []Item{}`, `guest := []Guest{}`
  - Forbidden: `items := []Item{}`, `guests := []Guest{}`

- Iterator naming:
  - Use `itor` for general iteration
  - Use `k, v` for map iteration

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
