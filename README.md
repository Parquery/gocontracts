gocontracts
===========

gocontracts generates pre- and post-condition checks from the function descriptions.

The main goal is to introduce 
[design-by-contract](https://en.wikipedia.org/wiki/Design_by_contract) 
into Go such that the contracts are included in the documentation and 
automatically reflected in the code.

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

Installation
============
We provide x86 Linux binaries in the "Releases" section.

To compile from code, run:

```bash
go get -U github.com/Parquery/gocontracts
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