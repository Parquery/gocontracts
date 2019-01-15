#!/usr/bin/env python3
"""
Release the code to the given directory as a binary package and a debian package.

The architecture is assumed to be AMD64 (i.e. Linux x64). If you want to release the code for a different architecture,
then please do that manually.
"""

import argparse
import os
import pathlib
import shutil
import subprocess
import sys
import tempfile
import textwrap


def main() -> int:
    """Execute the main routine."""
    parser = argparse.ArgumentParser()
    parser.add_argument("--release_dir", help="directory where to put the release", required=True)
    args = parser.parse_args()

    release_dir = pathlib.Path(args.release_dir)

    release_dir.mkdir(exist_ok=True, parents=True)

    # set the working directory to the script's directory
    script_dir = pathlib.Path(os.path.dirname(os.path.realpath(__file__)))

    subprocess.check_call(["go", "install", "./..."], cwd=script_dir.as_posix())

    if "GOPATH" not in os.environ:
        raise RuntimeError("Expected variable GOPATH in the environment")

    gopaths = os.environ["GOPATH"].split(os.pathsep)

    if not gopaths:
        raise RuntimeError("Expected at least a directory in GOPATH, but got none")

    # Figure out the main gopath
    gopath = pathlib.Path(gopaths[0])
    go_bin_dir = gopath / "bin"

    bin_path = go_bin_dir / "gocontracts"

    # Get gocontracts version
    version = subprocess.check_output([bin_path.as_posix(), "-version"], universal_newlines=True).strip()

    # Release the binary package
    with tempfile.TemporaryDirectory() as tmp_dir:
        bin_package_dir = pathlib.Path(tmp_dir) / "gocontracts-{}-linux-x64".format(version)

        target = bin_package_dir / "bin/gocontracts"
        target.parent.mkdir(parents=True)

        shutil.copy(bin_path.as_posix(), target.as_posix())

        tar_path = bin_package_dir.parent / "gocontracts-{}-linux-x64.tar.gz".format(version)

        subprocess.check_call([
            "tar", "-czf", tar_path.as_posix(), "gocontracts-{}-linux-x64".format(version)],
            cwd=bin_package_dir.parent.as_posix())

        shutil.move(tar_path.as_posix(), (release_dir / tar_path.name).as_posix())

    # Release the debian package
    with tempfile.TemporaryDirectory() as tmp_dir:
        deb_package_dir = pathlib.Path(tmp_dir) / "gocontracts_{}_amd64".format(version)

        target = deb_package_dir / "usr/bin/gocontracts"
        target.parent.mkdir(parents=True)
        shutil.copy(bin_path.as_posix(), target.as_posix())

        control_pth = deb_package_dir / "DEBIAN/control"
        control_pth.parent.mkdir(parents=True)

        control_pth.write_text(textwrap.dedent('''\
            Package: gocontracts
            Version: {version}
            Maintainer: Marko Ristin (marko.ristin@gmail.com)
            Architecture: amd64
            Description: gocontracts is a tool for design-by-contract in Go.
            '''.format(version=version)))

        subprocess.check_call(["dpkg-deb", "--build", deb_package_dir.as_posix()],
                              cwd=deb_package_dir.parent.as_posix(),
                              stdout=subprocess.DEVNULL)

        deb_pth = deb_package_dir.parent / "gocontracts_{}_amd64.deb".format(version)

        shutil.move(deb_pth.as_posix(), (release_dir / deb_pth.name).as_posix())

    print("Released to: {}".format(release_dir))

    return 0


if __name__ == "__main__":
    sys.exit(main())
