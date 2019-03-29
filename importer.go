package main

import (
	pathpkg "path"
	"go/types"
	"go/build"
)

// Importer is a types.ImporterFor that uses a "fast" importer, and if that fails,
// falls back to a "slower" implementation.
type Importer struct {
	mainImporter types.ImporterFrom
	cwd string
	packages map[string]*types.Package
	config *types.Config
}

func (i *Importer) Import(path string) (*types.Package, error) {
	// first try the main importer, which should be faster
	pkg, err := i.mainImporter.Import(path)
	return pkg, err
}

func (i *Importer) ImportFrom(path, srcDir string, mode types.ImportMode) (*types.Package, error) {
	if pkg, found := i.packages[path]; found {
		return pkg, nil
	}
	fullDir := pathpkg.Join(i.cwd, srcDir)
	pkg, err := i.mainImporter.ImportFrom(path, fullDir, mode)
	if err != nil {
		buildPkg, err := build.Default.Import(path, fullDir, 0 /* No `AllowBinary` because it messes with modules */)
		if err != nil {
			return nil, err
		}

		// TODO: support xtest and test files here too?
		filenames := append(buildPkg.GoFiles, buildPkg.CgoFiles...)
		files, err := parseFiles(buildPkg.Dir, filenames)
		if err != nil {
			return nil, err
		}
		pkg, err = i.config.Check(path, fset, files, nil)
	}
	i.packages[path] = pkg
	return pkg, nil
}
