"""Inspect the released ZIP bytes, not just the exporter staging directories."""
from pathlib import Path, PurePosixPath
from zipfile import ZipFile
import hashlib, json, re, tempfile

root = Path(__file__).resolve().parents[2]
reports = []
for edition, name, legacy in [('v1', 'go-by-doing-v1-content', 'go-by-doing-examples.zip'), ('v2', 'gopath-v2-content', 'gopath-course-v2.zip')]:
    directory = root / 'apps' / edition / 'public/downloads'
    archive = directory / f'{name}.zip'
    digest = hashlib.sha256(archive.read_bytes()).hexdigest()
    assert (directory / legacy).read_bytes() == archive.read_bytes()
    extracted = Path(tempfile.mkdtemp(prefix=f'verify-{edition}-'))
    with ZipFile(archive) as z:
        assert z.testzip() is None
        assert len(set(z.namelist())) == len(z.namelist())
        for entry in z.infolist():
            path = PurePosixPath(entry.filename)
            assert not path.is_absolute() and '..' not in path.parts
            assert path.parts[0] == name
            relative = path.parts[1:]
            assert not any(p.startswith('.') for p in relative)
            assert not any(p in {'node_modules', 'app', '.github', '.openai', 'plan', 'site', 'dist'} for p in relative)
            assert path.name not in {'package.json', 'package-lock.json', 'hugo.toml', 'wrangler.json'}
        z.extractall(extracted)
    pack = extracted / name
    manifest = json.loads((pack / 'manifest.json').read_text())
    assert manifest['edition'] == edition
    files = {p.relative_to(pack).as_posix() for p in pack.rglob('*') if p.is_file()}
    expected = {e['path'] for e in manifest['files']}
    assert files == expected | {'manifest.json'}
    for entry in manifest['files']:
        content = (pack / entry['path']).read_bytes()
        assert hashlib.sha256(content).hexdigest() == entry['sha256']
        assert len(content) == entry['bytes']
    for path in pack.rglob('*.md'):
        text = path.read_text()
        assert not re.search(r'^```gopath-(example|exercise|callout|disclosure)', text, re.M)
        assert '{{<' not in text
        assert '../../site/content/learn/' not in text
    assert manifest['inventory']['lessons'] == (48 if edition == 'v1' else 116)
    assert (pack / 'LICENSE').exists() and (pack / 'LICENSE-CONTENT.md').exists()
    reports.append({'edition': edition, 'files': len(files), 'archiveSha256': digest, 'extractedDirectory': str(pack), 'integrity': 'passed'})
(root / 'work/separation').mkdir(parents=True, exist_ok=True)
(root / 'work/separation/archive-verification.json').write_text(json.dumps(reports, indent=2)+'\n')
print(json.dumps(reports, indent=2))
