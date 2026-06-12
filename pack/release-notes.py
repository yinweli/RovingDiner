"""
release 說明生成腳本（由 release workflow 呼叫）

從上一個 tag 到指定 tag 之間的 commit 主旨，依本專案 commit 慣例
「Type | 描述」分組（Feature / Fix / Sheet / Doc / 其他），輸出 Markdown。
不用 GitHub 的 generate-notes：其素材為 PR，本 repo 直接 merge 不走 PR，生成內容會近乎空白。

使用方式：
    python pack/release-notes.py <tag> [--out notes.md]
未給 --out 時印到 stdout。上一個 tag 取 git describe --tags --abbrev=0 <tag>^，
找不到（首個 tag）即涵蓋全部歷史。
"""

from __future__ import annotations

import argparse
import re
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).parent.parent

ORDER = ['Feature', 'Fix', 'Sheet', 'Doc']  # 分組順序（CLAUDE.md commit 慣例的四個 Type）


def capture(cmd: list[str]) -> str:
    done = subprocess.run(cmd, cwd=ROOT, check=True, capture_output=True, text=True, encoding='utf-8')
    return done.stdout.strip()


def previous(tag: str) -> str | None:
    try:
        return capture(['git', 'describe', '--tags', '--abbrev=0', f'{tag}^']) or None
    except subprocess.CalledProcessError:
        return None  # 首個 tag，無前驅


def slug() -> str:
    # 自 remote 解出 owner/repo（支援 ssh 與 https 兩種形式），組 changelog 連結用
    url = capture(['git', 'config', '--get', 'remote.origin.url'])
    found = re.search(r'[:/]([^/:]+/[^/:]+?)(?:\.git)?$', url)
    return found.group(1) if found else 'yinweli/RovingDiner'


def render(tag: str) -> str:
    prev = previous(tag)
    scope = f'{prev}..{tag}' if prev else tag
    subject = capture(['git', 'log', scope, '--pretty=%s', '--no-merges']).splitlines()
    group: dict[str, list[str]] = {k: [] for k in ORDER}
    other: list[str] = []

    for itor in subject:
        head, sep, rest = itor.partition('|')
        kind, desc = head.strip(), rest.strip()

        if sep and kind in group:
            group[kind].append(desc)
        else:
            other.append(itor)

    line: list[str] = []

    for k in ORDER:
        if group[k]:
            line.append(f'### {k}')
            line.extend(f'- {itor}' for itor in group[k])
            line.append('')

    if other:
        line.append('### 其他')
        line.extend(f'- {itor}' for itor in other)
        line.append('')

    repo = slug()

    if prev:
        line.append(f'**Full Changelog**: https://github.com/{repo}/compare/{prev}...{tag}')
    else:
        line.append(f'**Full Changelog**: https://github.com/{repo}/commits/{tag}')

    return '\n'.join(line) + '\n'


def main() -> None:
    parser = argparse.ArgumentParser(description='自 git log 生成 release 說明')
    parser.add_argument('tag', help='本次發版的 tag')
    parser.add_argument('--out', default='', help='輸出檔案；未給時印到 stdout')
    arg = parser.parse_args()
    text = render(arg.tag)

    if arg.out:
        Path(arg.out).write_text(text, encoding='utf-8')
    else:
        sys.stdout.reconfigure(encoding='utf-8')  # 避免 Windows 重導向時非 UTF-8 編碼失敗
        print(text, end='')


if __name__ == '__main__':
    main()
