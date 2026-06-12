"""
遊戲安裝壓縮包打包腳本（由 task build 呼叫）

前提：執行機器有 Go 工具鏈；壓縮包的目標機器沒有，
因此 rodi / roditool / sheeter 三個 exe 直接建置入包（Windows amd64）。

流程：
1. go build cmd/rodi、cmd/roditool（GOOS=windows）
2. go install sheeter@釘住版本（版本由 --sheeter 傳入，單一定義點在 Taskfile 的 SHEETER_VERSION）
3. 收集 gamedata（排除 output 暫存與 Excel 鎖檔）、sheetdata、
   營業規格書.md / .html、pack/play.bat、pack/MANUAL.md
4. 壓成 pack/output/RovingDiner-<版本>.zip（zip 內含同名頂層目錄）

使用方式（於 repo 根目錄執行）：
    python pack/build-package.py --sheeter v3.0.11 [--version v.1.0]
版本未指定時取 git describe --tags --always。
"""

from __future__ import annotations

import argparse
import os
import shutil
import subprocess
import sys
import tempfile
import zipfile
from pathlib import Path

HERE = Path(__file__).parent
ROOT = HERE.parent
OUT = HERE / 'output'

# 包內 exe 一律 Windows amd64；純 Go（無 CGO），任何建置平台都可交叉編譯
GO_ENV = {'GOOS': 'windows', 'GOARCH': 'amd64', 'CGO_ENABLED': '0'}


def run(cmd: list[str], env: dict[str, str] | None = None) -> None:
    subprocess.run(cmd, cwd=ROOT, env={**os.environ, **env} if env else None, check=True)


def capture(cmd: list[str]) -> str:
    done = subprocess.run(cmd, cwd=ROOT, check=True, capture_output=True, text=True, encoding='utf-8')
    return done.stdout.strip()


def build_exe(pkg: Path) -> None:
    run(['go', 'build', '-trimpath', '-o', str(pkg / 'rodi.exe'), './cmd/rodi'], env=GO_ENV)
    run(['go', 'build', '-trimpath', '-o', str(pkg / 'roditool.exe'), './cmd/roditool'], env=GO_ENV)


def install_sheeter(pkg: Path, version: str) -> None:
    module = f'github.com/yinweli/Sheeter/v3/cmd/sheeter@{version}'

    if sys.platform == 'win32':
        run(['go', 'install', module], env={**GO_ENV, 'GOBIN': str(pkg)})
    else:
        # 交叉編譯時 go install 不准設 GOBIN，產物落在 GOPATH/bin/windows_amd64
        run(['go', 'install', module], env=GO_ENV)
        gopath = capture(['go', 'env', 'GOPATH'])
        shutil.copy2(Path(gopath) / 'bin' / 'windows_amd64' / 'sheeter.exe', pkg / 'sheeter.exe')


def collect(pkg: Path) -> None:
    shutil.copytree(ROOT / 'gamedata', pkg / 'gamedata', ignore=shutil.ignore_patterns('output', '~$*'))
    shutil.copytree(ROOT / 'sheetdata', pkg / 'sheetdata')

    for name in ('營業規格書.md', '營業規格書.html'):
        shutil.copy2(ROOT / 'doc' / name, pkg / name)

    shutil.copy2(HERE / 'play.bat', pkg / 'play.bat')
    shutil.copy2(HERE / 'MANUAL.md', pkg / 'MANUAL.md')


def pack_zip(pkg: Path) -> Path:
    OUT.mkdir(exist_ok=True)
    path = OUT / f'{pkg.name}.zip'

    with zipfile.ZipFile(path, 'w', zipfile.ZIP_DEFLATED) as z:
        for itor in sorted(pkg.rglob('*')):
            if itor.is_file():
                z.write(itor, itor.relative_to(pkg.parent))  # 以頂層目錄為 zip 內根

    return path


def main() -> None:
    sys.stdout.reconfigure(encoding='utf-8')  # 避免 Windows 重導向時非 UTF-8 編碼失敗

    parser = argparse.ArgumentParser(description='建置遊戲安裝壓縮包')
    parser.add_argument('--sheeter', required=True, help='sheeter 版本（Taskfile 的 SHEETER_VERSION）')
    parser.add_argument('--version', default='', help='包版本；未給時取 git describe --tags --always')
    arg = parser.parse_args()
    version = arg.version or capture(['git', 'describe', '--tags', '--always'])

    with tempfile.TemporaryDirectory() as tmp:
        pkg = Path(tmp) / f'RovingDiner-{version}'
        pkg.mkdir()
        build_exe(pkg)
        install_sheeter(pkg, arg.sheeter)
        collect(pkg)
        total = sum(1 for itor in pkg.rglob('*') if itor.is_file())
        path = pack_zip(pkg)

    print(f'打包完成：{path}（{total} 個檔案，{path.stat().st_size // 1024} KB）')


if __name__ == '__main__':
    main()
