#!/usr/bin/env python3
"""
Markdown 表格欄寬正規化工具。

用法：
  python doc/build-md/normalize-md-tables.py                # 處理 doc/ 下所有 *.md
  python doc/build-md/normalize-md-tables.py file.md ...    # 處理指定檔案；裸檔名會先從 doc/ 尋找

腳本只調整表格欄位的空白填充，不會改動內容；
CJK 字寬以 East Asian Width 為 W/F/A 者視為 2 計算（A = Ambiguous，
如 →、≤、≥、…，CJK 字型下通常渲染為 2），使表格在等寬字型下可以正確對齊。
"""

import re
import sys
import unicodedata
from pathlib import Path

try:
    sys.stdout.reconfigure(encoding='utf-8')
except AttributeError:
    pass


def char_width(c: str) -> int:
    return 2 if unicodedata.east_asian_width(c) in ('W', 'F', 'A') else 1


def str_width(s: str) -> int:
    return sum(char_width(c) for c in s)


def pad(s: str, width: int, align: str) -> str:
    diff = width - str_width(s)
    if diff <= 0:
        return s
    if align == 'right':
        return ' ' * diff + s
    if align == 'center':
        left = diff // 2
        return ' ' * left + s + ' ' * (diff - left)
    return s + ' ' * diff


SEP_CELL_RE = re.compile(r'^:?-+:?$')


def is_separator_row(cells) -> bool:
    if not cells:
        return False
    for c in cells:
        if not SEP_CELL_RE.match(c.strip()):
            return False
    return True


def split_row(line: str):
    """回傳 (indent, cells)；非表格行回傳 (None, None)。"""
    stripped = line.lstrip()
    if not stripped.startswith('|'):
        return None, None
    indent = line[: len(line) - len(stripped)]
    content = stripped.rstrip()
    if content.startswith('|'):
        content = content[1:]
    if content.endswith('|'):
        content = content[:-1]
    cells = [c.strip() for c in content.split('|')]
    return indent, cells


def parse_alignment(sep_cell: str) -> str:
    s = sep_cell.strip()
    has_l = s.startswith(':')
    has_r = s.endswith(':')
    if has_l and has_r:
        return 'center'
    if has_r:
        return 'right'
    if has_l:
        return 'left'
    return 'default'


def format_table(rows, indent: str):
    """rows: list of cell-lists；第 2 列（index 1）為 separator。"""
    n_cols = max(len(r) for r in rows)
    rows = [r + [''] * (n_cols - len(r)) for r in rows]
    SEP = 1
    alignments = [parse_alignment(c) for c in rows[SEP]]

    widths = []
    for col in range(n_cols):
        w = 0
        for i, r in enumerate(rows):
            if i == SEP:
                continue
            w = max(w, str_width(r[col]))
        widths.append(w)

    # separator 最少需要的 cell 寬度
    for col in range(n_cols):
        a = alignments[col]
        if a == 'center':
            min_cell = 5  # :---:
        elif a in ('left', 'right'):
            min_cell = 4  # :--- or ---:
        else:
            min_cell = 3  # ---
        if widths[col] + 2 < min_cell:
            widths[col] = min_cell - 2

    out = []
    for i, r in enumerate(rows):
        cells = []
        for col in range(n_cols):
            if i == SEP:
                w = widths[col] + 2
                a = alignments[col]
                if a == 'center':
                    cells.append(':' + '-' * (w - 2) + ':')
                elif a == 'right':
                    cells.append('-' * (w - 1) + ':')
                elif a == 'left':
                    cells.append(':' + '-' * (w - 1))
                else:
                    cells.append('-' * w)
            else:
                cells.append(' ' + pad(r[col], widths[col], alignments[col]) + ' ')
        out.append(indent + '|' + '|'.join(cells) + '|')
    return out


def normalize(text: str) -> str:
    lines = text.split('\n')
    out = []
    i = 0
    in_code = False
    while i < len(lines):
        line = lines[i]
        stripped = line.lstrip()
        if stripped.startswith('```') or stripped.startswith('~~~'):
            in_code = not in_code
            out.append(line)
            i += 1
            continue
        if in_code:
            out.append(line)
            i += 1
            continue
        if stripped.startswith('|'):
            block_lines = []
            block_cells = []
            indent = None
            j = i
            while j < len(lines):
                ind, cells = split_row(lines[j])
                if cells is None:
                    break
                if indent is None:
                    indent = ind
                elif ind != indent:
                    break
                block_lines.append(lines[j])
                block_cells.append(cells)
                j += 1
            if len(block_cells) >= 2 and is_separator_row(block_cells[1]):
                out.extend(format_table(block_cells, indent or ''))
            else:
                out.extend(block_lines)
            i = j
            continue
        out.append(line)
        i += 1
    return '\n'.join(out)


def process(path: Path) -> bool:
    text = path.read_text(encoding='utf-8-sig')
    new_text = normalize(text)
    if new_text != text:
        path.write_text(new_text, encoding='utf-8')
        return True
    return False


def main() -> int:
    doc_dir = Path(__file__).resolve().parent.parent  # doc/
    args = sys.argv[1:]
    if args:
        files = []
        for a in args:
            p = Path(a)
            if not p.exists() and not p.is_absolute():
                p = doc_dir / a
            files.append(p)
    else:
        files = sorted(doc_dir.glob('*.md'))
    for f in files:
        if not f.exists():
            print(f'skip (not found): {f}', file=sys.stderr)
            continue
        if process(f):
            print(f'normalized: {f.name}')
        else:
            print(f'unchanged:  {f.name}')
    return 0


if __name__ == '__main__':
    sys.exit(main())
