// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// gotype.go is a copy of the original source maintained
// in $GOROOT/src/go/types/gotype.go, but with the call
// to types.SizesFor factored out so we can provide a local
// implementation when compiling against Go 1.8 and earlier.
//
// This code is here for the sole purpose of satisfying historic
// references to this location, and for making gotype accessible
// via 'go get'.
//
// Do NOT make changes to this version as they will not be maintained
// (and possibly overwritten). Any changes should be made to the original
// and then ported to here.

/*
The gotype command, like the front-end of a Go compiler, parses and
type-checks a single Go package. Errors are reported if the analysis
fails; otherwise gotype is quiet (unless -v is set).

Without a list of paths, gotype reads from standard input, which
must provide a single Go source file defining a complete package.

With a single directory argument, gotype checks the Go files in
that directory, comprising a single package. Use -t to include the
(in-package) _test.go files. Use -x to type check only external
test files.

Otherwise, each path must be the filename of a Go file belonging
to the same package.

Imports are processed by importing directly from the source of
imported packages (default), or by importing from compiled and
installed packages (by setting -c to the respective compiler).

The -c flag must be set to a compiler ("gc", "gccgo") when type-
checking packages containing imports with relative import paths
(import "./mypkg") because the source importer cannot know which
files to include for such packages.

Usage:
	gotype [flags] [path...]

The flags are:
	-t
		include local test files in a directory (ignored if -x is provided)
	-x
		consider only external test files in a directory
	-e
		report all errors (not just the first 10)
	-v
		verbose mode
	-c
		compiler used for installed packages (gc, gccgo, or source); default: source
	-pkg-context
		consider the entire package when type checking, but only report errors for the given file; default: true
	-w
		consider the given directory as the working directory

Flags controlling additional output:
	-ast
		print AST (forces -seq)
	-trace
		print parse trace (forces -seq)
	-comments
		parse comments (ignored unless -ast or -trace is provided)

Examples:

To check the files a.go, b.go, and c.go:

	gotype a.go b.go c.go

To check an entire package including (in-package) tests in the directory dir and print the processed files:

	gotype -t -v dir

To check the external test package (if any) in the current directory, based on installed packages compiled with
cmd/compile:

	gotype -c=gc -x .

To verify the output of a pipe:

	echo "package foo" | gotype

*/
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

var (
	// main operation modes
	autoFiles     = flag.Bool("a", true, "if the base file ends in _test.go, include xtest files, otherwise include in-package test files and normal files.")
	usePkgContext = flag.Bool("pkg-context", true, "check the entire package, but restrict errors to the given file")
	workingDir    = flag.String("w", "", "use the given directory as the working directory (defaults to cwd)")
)

var (
	errorCount = 0
)

const usageString = `usage: gotype [flags] [path ...]

The gotype command, like the front-end of a Go compiler, parses and
type-checks a single Go package. Errors are reported if the analysis
fails; otherwise gotype is quiet (unless -v is set).

Without a list of paths, gotype reads from standard input, which
must provide a single Go source file defining a complete package.

With a single directory argument, gotype checks the Go files in
that directory, comprising a single package. Use -t to include the
(in-package) _test.go files. Use -x to type check only external
test files.

Otherwise, each path must be the filename of a Go file belonging
to the same package.

Imports are processed by importing directly from the source of
imported packages (default), or by importing from compiled and
installed packages (by setting -c to the respective compiler).

The -c flag must be set to a compiler ("gc", "gccgo") when type-
checking packages containing imports with relative import paths
(import "./mypkg") because the source importer cannot know which
files to include for such packages.
`

func usage() {
	fmt.Fprintln(os.Stderr, usageString)
	flag.PrintDefaults()
	os.Exit(2)
}

func report(err error, pathRestriction string) {
	if pathRestriction != "" {
		if pkgErr, isPkgErr := err.(packages.Error); isPkgErr {
			errFileParts := strings.Split(pkgErr.Pos, ":")
			errFilePath := errFileParts[0]
			if errFilePath != pathRestriction {
				return
			}
		}
	}
	fmt.Fprintln(os.Stderr, err)
	errorCount++
}

func getPkgFiles(args []string, useContext bool) (string, string, error) {
	if len(args) == 1 {
		// possibly a directory
		path := args[0]
		info, err := os.Stat(path)
		if err != nil {
			return "", "", err
		}
		if info.IsDir() {
			return path, "", nil
		}

		if useContext {
			dirName := filepath.Dir(path)
			if strings.HasPrefix(path, "./") {
				// filepath.Dir (via filepath.Clean) removes the leading ./
				dirName = "./" + dirName
			}
			return dirName, path, nil
		}
	}

	return "", "", fmt.Errorf("cannot specify more than one path")
}

func checkPkgFiles(pkgPath, targetFile string) {
	wd := *workingDir
	if wd == "" {
		var err error
		wd, err = os.Getwd()
		if err != nil {
			panic(err.Error())
		}
	}

	includeTests := *autoFiles && targetFile != "" && strings.HasSuffix(targetFile, "_test.go")

	cfg := &packages.Config{
		Dir: wd,
		Mode: packages.LoadTypes,
		Tests: includeTests,
	}

	var err error
	targetFile, err = filepath.Abs(targetFile)
	if err != nil {
		report(err, targetFile)
		return
	}

	roots, err := packages.Load(cfg, pkgPath)
	if err != nil {
		report(err, targetFile)
		return
	}
	for _, root := range roots {
		for _, err := range root.Errors {
			report(err, targetFile)
		}
	}
}

func main() {
	flag.Usage = usage
	flag.Parse()

	pkgPath, targetFile, err := getPkgFiles(flag.Args(), *usePkgContext)
	if err != nil {
		report(err, "")
		os.Exit(2)
	}

	checkPkgFiles(pkgPath, targetFile)
	if errorCount > 0 {
		os.Exit(2)
	}
}
