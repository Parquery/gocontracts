gocontracts
===========
![build status](https://travis-ci.com/Parquery/gocontracts.svg?branch=master)
[![Coverage Status](https://coveralls.io/repos/github/Parquery/gocontracts/badge.svg?branch=master)](https://coveralls.io/github/Parquery/gocontracts?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/Parquery/gocontracts)](https://goreportcard.com/report/github.com/Parquery/gocontracts)
[![godoc](https://img.shields.io/badge/godoc-reference-5272B4.svg)](https://godoc.org/github.com/Parquery/gocontracts/gocontracts)

gocontracts is a tool for 
[design-by-contract](https://en.wikipedia.org/wiki/Design_by_contract) in Go. 

It generates pre- and post-condition checks from the function descriptions so 
that the contracts are included in the documentation and automatically 
reflected in the code.

(If you need invariants, please let us know by commenting on this issue: https://github.com/Parquery/gocontracts/issues/25.)

Workflow
--------
You invoke gocontracts on an individual Go file. Gocontracts will parse the file
and examine the descriptions of all the functions for contracts. The existing
contract checks will be overwritten to match the contracts in the description.

A usual workflow includes defining the contracts in the function
description and invoking gocontracts to automatically update them in the code.

Each contract is defined as a one-line item in a bulleted list. Gocontracts
does not validate the correctness of the conditions (_e.g._ undefined variables,
syntax errors _etc._). The code is simply inserted as-is in the header of the
function body.

Since contracts are logical conditions of the function, failing a contract
causes a `panic()`. If you need to validate the input, rather than check
logical pre- and post-conditions of the function, return an `error` and do
not abuse the contracts.

Related Projects
================
At the time of this writing (2018-08-19), we found only a library that
implemented design-by-contract as functions (https://github.com/lpabon/godbc) and
a draft implementation based on decorators (https://github.com/ligurio/go-contracts).

None of them allowed us to synchronize the documentation with the code.

Examples
========
Simple Example
--------------

Given a function with contracts defined in the description, but no checks 
previously written in code:

```go
package somepackage

// SomeFunc does something.
//
// SomeFunc requires:
// * x >= 0
// * x < 100
//
// SomeFunc ensures:
// * !strings.HasSuffix(result, "smth")
func SomeFunc(x int) (result string) {
	// ...
}
```

, gocontracts will generate the following code:

```go
package somepackage

// SomeFunc does something.
//
// SomeFunc requires:
// * x >= 0
// * x < 100
//
// SomeFunc ensures:
// * !strings.HasSuffix(result, "smth")
func SomeFunc(x int) (result string) {
	// Pre-conditions
	switch {
	case !(x >= 0):
		panic("Violated: x >= 0")
	case !(x < 100):
		panic("Violated: x < 100")
	default:
		// Pass
	}

	// Post-condition
	defer func() {
		if strings.HasSuffix(result, "smth") {
			panic("Violated: !strings.HasSuffix(result, \"smth\")")
		}
	}()

	// ...
}
```

Note that you have to manually import the `strings` package since goconracts
is not smart enough to do that for you.

Conditioning on a Variable
--------------------------
Usually, functions either return an error if something went wrong or 
valid results otherwise. To ensure the contracts conditioned on the error, 
use implication and write:

```go
// Range returns a range of the timestamps available in the database.
//
// Range ensures:
// * err != nil || (empty || first < last)
func (t *Txn) Range() (first int64, last int64, empty bool, err error) {
	// ...
}
```

. Gocontracts will generate the following code:

```go
// Range returns a range of the timestamps available in the database.
//
// Range ensures:
// * err != nil || (empty || first <= last)
func (t *Txn) Range() (first int64, last int64, empty bool, err error) {
	// Post-condition
	defer func() {
	    if !(err != nil || (empty || first < last)) {
	    	panic("Violated: err != nil || (empty || first < last)")
	    }	
	}()
	
	// ...
}
```

Note that conditioning can be seen as logical implication 
(`A ⇒ B` can be written as `¬A ∨ B`). In the above example, we chained 
multiple implications as 
`err == nil ⇒ (¬ empty ⇒ first ≤ last)`.

Toggling Contracts
------------------
When developing a library, it is important to give your users a possibility to toggle families of contracts so that they can adapt _your_ contracts to _their_ use case. For example, some contracts of your library should be verified in testing and in production, some should be verified only in testing of _their_ modules and others should be verified only in _your_ unit tests, but not in _theirs_.

To that end, you can use build tags to allow toggling of contracts at compile time (https://golang.org/pkg/go/build/#hdr-Build_Constraints). Define for each of the contract family a separate file which is built dependening on the build tag. In each of these files, define constant booleans (_e.g._, `InTest`, `InUTest`). Depending on the scenario, set these variables appropriately: in production scenario, `InTest = false` and `InUTest = false`, in the test scenario for others `InTest = true` and  `InUTest = false`, while in the scenario of _your_ unit tests `InTest = true` and `InUTest = true`.

The examples of the three files follow.

`contracts_prod.go`:
```go
// +build prod,!test,!utest
package somepackage

const InTest = false
const InUTest = false
```

`contracts_test.go`:
```go
// +build !prod,test,!utest
package somepackage

const InTest = true
const InUTest = false
```

`contracts_utest.go`:
```go
// +build !prod,!test,utest
package somepackage

const InTest = true
const InUTest = true
```

Include each of these boolean constants in your contract conditions and `&&` 
with the condition. 
For example, this is how you extend the postcondition of the function `Range` 
written in the previous section to be verified only in _yours_ and _theirs_ 
tests, but not in production:

```go
// Range returns a range of the timestamps available in the database.
//
// Range ensures:
// * InTest && (err != nil || (empty || first <= last))
func (t *Txn) Range() (first int64, last int64, empty bool, err error) {
	...
}
```

Since constant booleans are placed first in the conjunction, 
the rest of the condition will not be evaluated incurring thus no 
computational overhead in the production at runtime. 

Usage
=====
Gocontracts reads the Go file and outputs the modified source code to standard
output:

```bash
gocontracts /path/to/some/file.go
```

You can modify the file in-place by supplying the `-w` argument:

```bash
gocontracts -w /path/to/some/file.go
```

If you want to remove the contract checks from the code, supply the 
`-r` argument:

```bash
gocontracts -w -r /path/to/some/file.go
```

The remove argument is particularly useful when you have a build system
in place and you want to distinguish between the debug code and the
release (production) code.

Before building the release code, run the gocontracts with `-r` to remove 
the checks from the code.


Usage
=====
Gocontracts reads the Go file and outputs the modified source code to standard
output:

```bash
gocontracts /path/to/some/file.go
```

You can modify the file in-place by supplying the `-w` argument:

```bash
gocontracts -w /path/to/some/file.go
```

If you want to remove the contract checks from the code, supply the 
`-r` argument:

```bash
gocontracts -w -r /path/to/some/file.go
```

The remove argument is particularly useful when you have a build system
in place and you want to distinguish between the debug code and the
release (production) code.

Before building the release code, run the gocontracts with `-r` to remove 
the checks from the code.

Installation
============
We provide x86 Linux binaries in the "Releases" section.

To compile from code, run:

```bash
go get -u github.com/Parquery/gocontracts
```

Development
===========
* Fork the repository to your user on Github.

* Get the original repository:
```bash
go get github.com/Parquery/gocontracts
```

* Indicate that the local repository is a fork:
```bash
cd gocontracts
git remote add fork https://github.com/YOUR-GITHUB-USERNAME/gocontracts.git
```

* Make your changes.

* Push the changes from the local repository to your remote fork:
 ```bash
 git push fork
 ```

* Create a pull request on Github and send it for review.

Versioning
==========
We follow [Semantic Versioning](http://semver.org/spec/v1.0.0.html). 
The version X.Y.Z indicates:

* X is the major version (backward-incompatible),
* Y is the minor version (backward-compatible), and
* Z is the patch version (backward-compatible bug fix).
