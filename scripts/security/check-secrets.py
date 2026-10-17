"""Scan publishable Git files and learner ZIPs without printing secret values.

Only Git's tracked/unignored file list is read. Environment and private-key
files are rejected by name without opening them. --stage supplies the same
safe file set to Gitleaks; it must name a new directory outside the repository.
"""
import argparse
import io
import json
from pathlib import Path, PurePosixPath
import re
import shutil
import subprocess
import sys
import zipfile

ROOT = Path(__file__).resolve().parents[2]
PATTERNS = {
    "openai-key": r"\bsk-(?:proj-|svcacct-|admin-)?[A-Za-z0-9_-]{20,}",
    "session-token": r"\bsess-[A-Za-z0-9_-]{20,}",
    "github-token": r"\b(?:gh[pousr]_[A-Za-z0-9]{20,}|github_pat_[A-Za-z0-9_]{25,})",
    "aws-access-key": r"\b(?:AKIA|ASIA)[A-Z0-9]{16}\b",
    "google-api-key": r"\bAIza[0-9A-Za-z_-]{30,}",
    "slack-token": r"\bxox[baprs]-[0-9A-Za-z-]{20,}",
    "private-key": r"-----BEGIN (?:RSA |EC |OPENSSH |DSA |ENCRYPTED )?PRIVATE KEY-----",
}
RULES = {name: re.compile(pattern) for name, pattern in PATTERNS.items()}
LIMIT = 20_000_000


def protected(parts):
    return any(part.startswith('.env') or part in {'id_rsa', 'id_ed25519'}
               or part.endswith(('.pem', '.key', '.p12', '.pfx', '.jks', '.keystore'))
               for part in parts)


def scan_text(name, data, findings):
    try:
        text = data.decode('utf-8')
    except UnicodeDecodeError:
        return
    for rule, pattern in RULES.items():
        for match in pattern.finditer(text):
            findings.append({'file': name, 'line': text.count('\n', 0, match.start()) + 1, 'rule': rule})


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--stage', type=Path)
    args = parser.parse_args()
    if args.stage:
        stage = args.stage.resolve()
        if stage == ROOT or ROOT in stage.parents or stage.exists():
            parser.error('--stage must be a new directory outside the repository')
        stage.mkdir(parents=True)
    raw = subprocess.check_output(['git', 'ls-files', '--cached', '--others', '--exclude-standard', '-z'], cwd=ROOT)
    names = sorted(set(name.decode() for name in raw.split(b'\0') if name))
    findings, files, entries = [], 0, 0
    for name in names:
        relative = PurePosixPath(name)
        path = ROOT / name
        if relative.is_absolute() or '..' in relative.parts:
            findings.append({'file': name, 'rule': 'unsafe-path'})
            continue
        if protected(relative.parts):
            findings.append({'file': name, 'rule': 'protected-file-in-publication'})
            continue
        if any((ROOT.joinpath(*relative.parts[:i])).is_symlink() for i in range(1, len(relative.parts) + 1)):
            findings.append({'file': name, 'rule': 'symlink-requires-review'})
            continue
        if not path.is_file():
            continue  # Git may still list a deleted tracked file.
        if path.stat().st_size > LIMIT:
            findings.append({'file': name, 'rule': 'file-exceeds-scan-limit'})
            continue
        data = path.read_bytes()
        files += 1
        if name.endswith('.zip'):
            try:
                with zipfile.ZipFile(io.BytesIO(data)) as archive:
                    members = archive.infolist()
                    if len(members) > 10_000 or sum(m.file_size for m in members) > 100_000_000:
                        raise ValueError('archive limits')
                    for member in members:
                        if member.is_dir():
                            continue
                        location = name + '!' + member.filename
                        parts = PurePosixPath(member.filename).parts
                        if protected(parts):
                            findings.append({'file': location, 'rule': 'protected-file-in-archive'})
                            continue
                        if member.file_size > LIMIT:
                            raise ValueError('entry limit')
                        scan_text(location, archive.read(member), findings)
                        entries += 1
            except (zipfile.BadZipFile, ValueError, RuntimeError, NotImplementedError):
                findings.append({'file': name, 'rule': 'archive-not-fully-scanned'})
        else:
            scan_text(name, data, findings)
        if args.stage:
            target = stage / name
            target.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(path, target)
    print(json.dumps({'files': files, 'archiveEntries': entries, 'findings': findings}, indent=2))
    return bool(findings)


if __name__ == '__main__':
    sys.exit(main())
