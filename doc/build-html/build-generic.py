"""
通用規格書 build 腳本（非 SSOT、無 mermaid）

一次產生所有「不需 mermaid」的規格書 HTML：
- 從 template.html 抽出 <style>（配色／版面與 SSOT 同步，不另維護一份）。
- 讀各來源 .md，去掉開頭 H1（標題改由 doc-head 顯示）。
- 注入 template-generic.html 的 placeholder，輸出對應 .html。

目錄欄與 scroll-spy 由 template-generic.html 內的前端腳本自章節標題生成；
CJK 等寬字體覆寫亦在該 template，對齊 mockup / 套件樹的對齊需求。

SSOT（營業規格書.md）走 build.py（含 mermaid），不在此處。

使用方式（於 repo 根目錄執行）：
    python doc/build-html/build-generic.py
"""

from __future__ import annotations

import re
import sys
from pathlib import Path

HERE = Path(__file__).parent
SRC = HERE.parent  # doc/

SSOT_TEMPLATE = HERE / 'template.html'  # 借其 <style>
TEMPLATE = HERE / 'template-generic.html'

# 每份要產 HTML 的文件設定
DOCS = [
    {
        'md': '營業實作規格書.md',
        'out': '營業實作規格書.html',
        'title': '營業實作規格書',
        'eyebrow': 'Roving Diner · Impl Spec',
        'lead': '專案架構與核心引擎決策（套件結構、事件流、決定性、邊界介面、測試策略）。'
                '左側目錄由章節標題自動生成、隨捲動高亮；套件樹以 CJK 等寬字體呈現以保對齊。',
    },
    {
        'md': '營業顯示規格書.md',
        'out': '營業顯示規格書.html',
        'title': '營業顯示規格書',
        'eyebrow': 'Roving Diner · Display Spec',
        'lead': 'TUI 顯示／操作層設計。左側目錄由章節標題自動生成、隨捲動高亮；'
                '版面 mockup 以 CJK 等寬字體呈現以保對齊。',
    },
]


def extract_style(html: str) -> str:
    """擷取 template.html 的 <style>…</style>（含標籤本身）。"""
    m = re.search(r'<style>.*?</style>', html, flags=re.DOTALL)
    if not m:
        raise ValueError('template.html 找不到 <style> 區塊')
    return m.group(0)


def strip_leading_h1(md: str) -> str:
    """移除開頭的 H1（# …）行，避免與 doc-head 標題重複。"""
    return re.sub(r'^﻿?#\s.*\n', '', md, count=1)


def inject(html: str, placeholder: str, body: str) -> str:
    if placeholder not in html:
        raise ValueError(f'template 缺少 placeholder: {placeholder}')
    return html.replace(placeholder, body)


def build_one(template: str, style: str, doc: dict[str, str]) -> int:
    md_path = SRC / doc['md']
    if not md_path.exists():
        print(f'找不到來源: {md_path}', file=sys.stderr)
        return 1
    md = strip_leading_h1(md_path.read_text(encoding='utf-8'))

    html = template
    try:
        html = inject(html, '<!-- INJECT:style -->', style)
        html = inject(html, '<!-- INJECT:title -->', doc['title'])  # 出現多次
        html = inject(html, '<!-- INJECT:eyebrow -->', doc['eyebrow'])
        html = inject(html, '<!-- INJECT:lead -->', doc['lead'])
        html = inject(html, '<!-- INJECT:md -->', md)
    except ValueError as e:
        print(f"注入失敗（{doc['out']}）: {e}", file=sys.stderr)
        return 1

    leftover = re.findall(r'<!-- INJECT:[^>]+ -->', html)
    if leftover:
        print(f"警告：仍有未替換的 placeholder（{doc['out']}）: {leftover}", file=sys.stderr)
        return 1

    out_path = SRC / doc['out']
    out_path.write_bytes(html.encode('utf-8'))
    print(f'已輸出 {doc["out"]} ({len(html)} chars)')
    return 0


def main() -> int:
    if not SSOT_TEMPLATE.exists() or not TEMPLATE.exists():
        print('找不到 template.html 或 template-generic.html', file=sys.stderr)
        return 1

    try:
        style = extract_style(SSOT_TEMPLATE.read_text(encoding='utf-8'))
    except ValueError as e:
        print(f'抽取 style 失敗: {e}', file=sys.stderr)
        return 1

    template = TEMPLATE.read_text(encoding='utf-8')
    for doc in DOCS:
        code = build_one(template, style, doc)
        if code != 0:
            return code
    return 0


if __name__ == '__main__':
    sys.exit(main())
