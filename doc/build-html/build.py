"""
營業規格書 build 腳本

讀取 template.html + doc/營業規格書.md + 14 份 mermaid/*.mmd
依 <!-- INJECT:xxx --> placeholder 替換後輸出 doc/營業規格書.html

使用方式（於 repo 根目錄執行）：
    python doc/build-html/build.py
或於本目錄執行：
    python build.py
"""

from __future__ import annotations

import json
import re
import sys
from pathlib import Path

# 訊息含中文：管線下 stdout/stderr 預設跟系統碼頁（CI runner cp1252、本機 cp950），改 UTF-8 避免編碼失敗
sys.stdout.reconfigure(encoding='utf-8')
sys.stderr.reconfigure(encoding='utf-8')

HERE = Path(__file__).parent
SRC = HERE.parent  # doc/，存放輸出 HTML 與來源 markdown
MD_DIR = SRC

TEMPLATE = HERE / 'template.html'
MERMAID_DIR = HERE / 'mermaid'

SSOT = MD_DIR / '營業規格書.md'
OUTPUT = SRC / '營業規格書.html'

# 14 個流程小節 / 流程圖 placeholder name
FLOW_NAMES: list[str] = [
    'flow-core-1', 'flow-core-2', 'flow-core-3', 'flow-core-4',
    'flow-core-5', 'flow-core-6', 'flow-core-7',
    'flow-sub-skill', 'flow-sub-trigger', 'flow-sub-advance',
    'flow-sub-cleanup', 'flow-sub-exec', 'flow-sub-settle', 'flow-sub-judge',
]


def slice_between(text: str, start_pattern: str, stop_pattern: str | None) -> str:
    """擷取 start_pattern (regex) 之後到 stop_pattern (regex) 之前的內容。

    起始錨點本身不包含於回傳值（從錨點 match 末端起算）。
    stop_pattern 為 None 代表取到文件末端。
    """
    start_match = re.search(start_pattern, text, flags=re.MULTILINE)
    if not start_match:
        raise ValueError(f'找不到起始錨點: {start_pattern!r}')
    body_start = start_match.end()
    rest = text[body_start:]
    if stop_pattern is None:
        return rest.strip('\n')
    stop_match = re.search(stop_pattern, rest, flags=re.MULTILINE)
    body = rest[: stop_match.start()] if stop_match else rest
    return body.strip('\n')


def slice_inclusive(text: str, start_pattern: str, stop_pattern: str | None) -> str:
    """同 slice_between，但包含起始錨點 match 文字本身。"""
    start_match = re.search(start_pattern, text, flags=re.MULTILINE)
    if not start_match:
        raise ValueError(f'找不到起始錨點: {start_pattern!r}')
    body_start = start_match.start()
    rest = text[body_start:]
    if stop_pattern is None:
        return rest.strip('\n')
    stop_match = re.search(stop_pattern, rest, flags=re.MULTILINE)
    body = rest[: stop_match.start()] if stop_match else rest
    return body.strip('\n')


def build_md_payloads(md: str) -> dict[str, str]:
    payload: dict[str, str] = {}

    # md-front: §一 ~ §十八（不含 H1 標題與 §十九 起的流程章節）
    payload['md-front'] = slice_between(
        md, r'^# 營業規格書\s*$', r'^## 十九、核心流程\s*$'
    )

    # §十九 7 個 H3 小節（### N. 標題）
    core_titles = [
        ('md-flow-core-1', r'^### 1\. 營業開始階段\s*$'),
        ('md-flow-core-2', r'^### 2\. 回合開始階段\s*$'),
        ('md-flow-core-3', r'^### 3\. 玩家行動階段\s*$'),
        ('md-flow-core-4', r'^### 4\. 顧客行動階段\s*$'),
        ('md-flow-core-5', r'^### 5\. 回合結束階段\s*$'),
        ('md-flow-core-6', r'^### 6\. 營業成功階段\s*$'),
        ('md-flow-core-7', r'^### 7\. 營業失敗階段\s*$'),
    ]
    for key, pat in core_titles:
        payload[key] = slice_between(md, pat, r'^### |^---\s*$')

    # §二十 8 個 H3 (含 [終止判定])
    sub_titles = [
        ('md-flow-sub-skill', r'^### 啟動技能\s*$'),  # 含 #### [堆疊處理]
        ('md-flow-sub-trigger', r'^### 觸發時機\s*$'),
        ('md-flow-sub-advance', r'^### 推進效果\s*$'),
        ('md-flow-sub-cleanup', r'^### 清理效果\s*$'),
        ('md-flow-sub-exec', r'^### 執行命令\s*$'),
        ('md-flow-sub-freeze', r'^### 凍結語意\s*$'),
        ('md-flow-sub-settle', r'^### 執行結算\s*$'),  # 含 #### 結算補充
        ('md-flow-sub-judge', r'^### \[終止判定\]\s*$'),
    ]
    for key, pat in sub_titles:
        payload[key] = slice_between(md, pat, r'^### |^---\s*$')

    # md-flow-extra：§二十一 流程補充（不含 H2 標題）
    payload['md-flow-extra'] = slice_between(
        md, r'^## 二十一、流程補充\s*$', r'^## |^---\s*$'
    )

    # md-back：§二十二 ~ 文末（含 H2 標題,到 EOF）
    payload['md-back'] = slice_inclusive(
        md, r'^## 二十二、觸發時機清單\s*$', None
    )

    return payload


TERM_RE = re.compile(r'^- \*\*(.+?)\*\*：(.+)$', flags=re.MULTILINE)


def build_terms_json(md: str) -> str:
    """從 §三 名詞列表抽出 {name: desc} 並輸出 JSON 字串。"""
    section = slice_between(md, r'^## 三、名詞列表\s*$', r'^## |^---\s*$')
    terms: dict[str, str] = {}
    for m in TERM_RE.finditer(section):
        name = m.group(1).strip()
        desc = m.group(2).strip()
        terms[name] = desc
    if not terms:
        raise ValueError('§三 名詞列表解析為空,請檢查格式')
    return json.dumps(terms, ensure_ascii=False, indent=0)


def main() -> int:
    if not TEMPLATE.exists():
        print(f'找不到 template.html: {TEMPLATE}', file=sys.stderr)
        return 1
    if not SSOT.exists():
        print(f'找不到 SSOT: {SSOT}', file=sys.stderr)
        return 1

    html = TEMPLATE.read_text(encoding='utf-8')
    md = SSOT.read_text(encoding='utf-8')

    # 1) 切片並注入 markdown
    try:
        payloads = build_md_payloads(md)
    except ValueError as e:
        print(f'切片失敗: {e}', file=sys.stderr)
        return 1
    for key, body in payloads.items():
        placeholder = f'<!-- INJECT:{key} -->'
        if placeholder not in html:
            print(f'template 缺少 placeholder: {placeholder}', file=sys.stderr)
            return 1
        html = html.replace(placeholder, body)

    # 2) 注入名詞 JSON
    try:
        terms_json = build_terms_json(md)
    except ValueError as e:
        print(f'名詞解析失敗: {e}', file=sys.stderr)
        return 1
    placeholder = '<!-- INJECT:terms-json -->'
    if placeholder not in html:
        print(f'template 缺少 placeholder: {placeholder}', file=sys.stderr)
        return 1
    html = html.replace(placeholder, terms_json)

    # 3) 注入 mermaid
    for name in FLOW_NAMES:
        mmd_path = MERMAID_DIR / f'{name}.mmd'
        if not mmd_path.exists():
            print(f'找不到 mermaid 檔: {mmd_path}', file=sys.stderr)
            return 1
        mmd_content = mmd_path.read_text(encoding='utf-8').rstrip('\n')
        placeholder = f'<!-- INJECT:mermaid-{name} -->'
        if placeholder not in html:
            print(f'template 缺少 placeholder: {placeholder}', file=sys.stderr)
            return 1
        # mermaid 區塊前需有換行,使 flowchart 從新行開始
        html = html.replace(placeholder, '\n' + mmd_content)

    # 4) 檢查是否還有未處理的 placeholder
    leftover = re.findall(r'<!-- INJECT:[^>]+ -->', html)
    if leftover:
        print(f'警告：仍有未替換的 placeholder: {leftover}', file=sys.stderr)
        return 1

    OUTPUT.write_bytes(html.encode('utf-8'))
    print(f'已輸出 {OUTPUT.name} ({len(html)} chars)')
    return 0


if __name__ == '__main__':
    sys.exit(main())
