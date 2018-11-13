#!/usr/bin/env python3

"""Run all pre-commit checks."""
import os
import pathlib
import subprocess
import sys


def main() -> int:
    """"Exectute main routine."""

    here = pathlib.Path(os.path.abspath(__file__)).parent

    pths = sorted(here.glob("*.go"))

    for subpth in here.iterdir():
        if subpth.is_dir() and not subpth.name.startswith('.'):
            pths.extend(subpth.glob("**/*.go"))

    for pth in pths:
        out = subprocess.check_output(["gofmt", "-s", "-l", pth.as_posix()])
        if len(out) != 0:
            print("Code was not formatted; gofmt -s -l complains: {}".format(pth))
            return 1

    packages = [pathlib.Path(line)
                for line in subprocess.check_output(["go", "list", "./..."], universal_newlines=True).splitlines()
                if line.strip()]

    packages = [pkg for pkg in packages if 'vendor' not in pkg.parents]

    subprocess.check_call(['go', 'vet'] + [pkg.as_posix() for pkg in packages], cwd=here.as_posix())

    subprocess.check_call(['golint'] + [pkg.as_posix() for pkg in packages], cwd=here.as_posix())

    subprocess.check_call(['errcheck'] + [pkg.as_posix() for pkg in packages], cwd=here.as_posix())

    non_test_pths = [pth for pth in pths if not pth.name.endswith("_test.go" or "testcases" in pth.parents)]
    subprocess.check_call(["gocyclo", "-over", "15"] + [pth.as_posix() for pth in non_test_pths])

    for pkg in packages:
        subprocess.check_call(['go', 'test', '-v', '-covermode=count',
                               '-coverprofile=../../../{}/profile.coverprofile'.format(pkg), pkg.as_posix()],
                              cwd=here.as_posix())

    subprocess.check_call(['go', 'build', './...'], cwd=here.as_posix())

    ##
    # Check that the CHANGELOG.md is consistent with -version
    ##
    version = subprocess.check_output(['go', 'run', 'main.go', '-version'], cwd=here.as_posix(),
                                      universal_newlines=True).strip()

    changelog_pth = here / "CHANGELOG.md"
    changelog_lines = changelog_pth.read_text().splitlines()
    if len(changelog_lines) < 1:
        raise RuntimeError("Expected at least a line in {}, but got: {} line(s)".format(
            changelog_pth, len(changelog_lines)))

    # Expect the first line to always refer to the latest version
    changelog_version = changelog_lines[0]

    if version != changelog_version:
        raise AssertionError(("The version in changelog file {} (parsed as its first line), {!r}, "
                              "does not match the output of -version: {!r}").format(
            changelog_pth, changelog_version, version))

if __name__ == "__main__":
    sys.exit(main())
