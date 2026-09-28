"""Create a deterministic tar.gz with normalized metadata (Python standard library)."""

import gzip
from pathlib import Path
import sys
import tarfile


def package(source, destination):
    with destination.open("wb") as output:
        with gzip.GzipFile(filename="", mode="wb", fileobj=output, mtime=0) as compressed:
            with tarfile.open(fileobj=compressed, mode="w", format=tarfile.USTAR_FORMAT) as archive:
                for path in [source, *sorted(source.rglob("*"))]:
                    info = archive.gettarinfo(str(path), str(path.relative_to(source.parent)))
                    info.uid = info.gid = info.mtime = 0
                    info.uname = info.gname = ""
                    info.mode = 0o755 if path.is_dir() or path.name in ("bitconfig", "bitconfig.exe") else 0o644
                    if path.is_file():
                        with path.open("rb") as contents:
                            archive.addfile(info, contents)
                    else:
                        archive.addfile(info)


if __name__ == "__main__":
    package(Path(sys.argv[1]), Path(sys.argv[2]))
