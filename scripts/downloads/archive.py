"""Create a deterministic learner archive from an explicitly prepared tree."""
from pathlib import Path
from zipfile import ZipFile, ZipInfo, ZIP_DEFLATED
import sys

source, destination, prefix = Path(sys.argv[1]), Path(sys.argv[2]), sys.argv[3]
destination.parent.mkdir(parents=True, exist_ok=True)
with ZipFile(destination, 'w', compression=ZIP_DEFLATED, compresslevel=9) as archive:
    for path in sorted(source.rglob('*')):
        if path.is_symlink():
            raise ValueError(f'Symlinks are not learner content: {path}')
        if not path.is_file():
            continue
        relative = path.relative_to(source)
        if any(part.startswith('.') for part in relative.parts):
            raise ValueError(f'Unexpected hidden file: {relative}')
        info = ZipInfo(f'{prefix}/{relative.as_posix()}', (2020, 1, 1, 0, 0, 0))
        info.compress_type = ZIP_DEFLATED
        info.create_system = 3
        info.external_attr = 0o100644 << 16
        archive.writestr(info, path.read_bytes())
