package main

import (
	"fmt"
	"go/ast"
	"go/types"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const testhelperRunAnalyzerName = "testhelper_run"

type testhelperRunAnalyzerSettings struct {
	IncludedFunctions []string `mapstructure:"included-functions"`
}

var toolPrefixPattern = regexp.MustCompile(`^gitlab.com/gitlab-org/gitaly(/v\d{2})?/tools`)

// newTesthelperRunAnalyzer returns an analyzer to detect if a package that has tests does
// not contain a `TestMain()` function that executes `testhelper.Run()`.
// For more information:
// https://gitlab.com/gitlab-org/gitaly/-/blob/master/STYLE.md?ref_type=heads#common-setup
func newTesthelperRunAnalyzer(settings *testhelperRunAnalyzerSettings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: testhelperRunAnalyzerName,
		Doc:  `TestMain must be present and call testhelper.Run()`,
		Run:  runTesthelperRunAnalyzer(settings.IncludedFunctions),
	}
}

func runTesthelperRunAnalyzer(rules []string) func(*analysis.Pass) (interface{}, error) {
	return func(pass *analysis.Pass) (interface{}, error) {
		var hasTestMain, hasTests bool

		// Blackbox test packages ending with `_test` are considered to be
		// part of the primary package for compilation, but are scanned in a
		// separate pass by the analyzer. The primary and test packages cannot
		// both define `TestMain`. Only require `TestMain` in the primary package.
		if strings.HasSuffix(pass.Pkg.Name(), "_test") {
			return nil, nil
		}

		// Don't lint tools, they can't import `testhelper`.
		if toolPrefixPattern.MatchString(pass.Pkg.Path()) {
			return nil, nil
		}

		for _, file := range pass.Files {
			if hasTestMain {
				break
			}

			ast.Inspect(file, func(node ast.Node) bool {
				if decl, ok := node.(*ast.FuncDecl); ok {
					declName := decl.Name.Name

					if declName == "TestMain" {
						hasTestMain = true

						analyzeTestMain(pass, decl, rules)
						analyzeFilename(pass, file, decl)
					}

					// Actual tests must start with `Test`, helpers could take a `testing.TB`.
					if strings.HasPrefix(declName, "Test") {
						params := decl.Type.Params
						for _, field := range params.List {
							fieldType := pass.TypesInfo.TypeOf(field.Type)

							// Do we have any tests in this package?
							if types.Implements(fieldType, testingTB) {
								hasTests = true
							}
						}
					}
				}
				return true
			})
		}

		// If we have tests but there's no `TestMain`, report.
		if hasTests && !hasTestMain {
			// We don't have a specific location for this failure, so use the location of the package name
			// in its first file and provide the name in the error text. This list is sorted lexically by
			// filename, so the location of `nolint` directives may not be stable when new files are added.
			pass.Report(analysis.Diagnostic{
				Pos:            pass.Files[0].Name.Pos(),
				End:            pass.Files[0].Name.End(),
				Message:        fmt.Sprintf("no TestMain in package %v", pass.Pkg.Path()),
				SuggestedFixes: nil,
			})
		}

		return nil, nil
	}
}

func analyzeFilename(pass *analysis.Pass, file *ast.File, decl *ast.FuncDecl) {
	fullpath := pass.Fset.File(file.Pos()).Name()
	filename := filepath.Base(fullpath)

	if filename != "testhelper_test.go" {
		pass.Report(analysis.Diagnostic{
			Pos:            decl.Pos(),
			End:            decl.End(),
			Message:        "TestMain should be placed in file 'testhelper_test.go'",
			SuggestedFixes: nil,
		})
	}
}

func analyzeTestMain(pass *analysis.Pass, decl *ast.FuncDecl, rules []string) {
	matcher := NewMatcher(pass)
	var hasRun bool

	ast.Inspect(decl, func(node ast.Node) bool {
		if call, ok := node.(*ast.CallExpr); ok {
			if matcher.MatchFunction(call, rules) {
				hasRun = true
			}
		}
		return true
	})

	if !hasRun {
		pass.Report(analysis.Diagnostic{
			Pos:            decl.Pos(),
			End:            decl.End(),
			Message:        "testhelper.Run not called in TestMain",
			SuggestedFixes: nil,
		})
	}
}
