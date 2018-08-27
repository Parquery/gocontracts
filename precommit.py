#!/usr/bin/env python3

"""Run all pre-commit checks."""

import pathlib
import subprocess
import sys


def main() -> int:
    """"Exectute main routine."""

    here = pathlib.Path(".")

    pths = sorted(here.glob("**/*.go"))

    for pth in pths:
        if "vendor" in here.parents:
            continue

        out = subprocess.check_output(["gofmt", "-s", "-l", pth.as_posix()])
        if len(out) != 0:
            print("Code was not formatted with gofmt -s -l: {}".format(pth))
            return 1

    packages = [pathlib.Path(line)
                for line in subprocess.check_output(["go", "list", "./..."], universal_newlines=True).splitlines()
                if line.strip()]

    packages = [pkg for pkg in packages if 'vendor' not in pkg.parents]

    subprocess.check_call(['go', 'vet'] + [pkg.as_posix() for pkg in packages])

    subprocess.check_call(['golint'] + [pkg.as_posix() for pkg in packages])

    subprocess.check_call(['errcheck'] + [pkg.as_posix() for pkg in packages])

    non_test_pths = [pth for pth in pths if not pth.name.endswith("_test.go" or "testcases" in pth.parents)]
    subprocess.check_call(["gocyclo", "-over", "15"] + [pth.as_posix() for pth in non_test_pths])

    for pkg in packages:
        subprocess.check_call(['go', 'test', '-v', '-covermode=count',
                               '-coverprofile=../../../{}/profile.coverprofile'.format(pkg), pkg.as_posix()])

    subprocess.check_call(['go', 'build', './...'])


if __name__ == "__main__":
    sys.exit(main())
