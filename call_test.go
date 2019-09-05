package analysisutil_test

import (
	"testing"

	"github.com/Matts966/analysisutil"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
	"golang.org/x/tools/go/analysis/passes/buildssa"
)

var Analyzer = &analysis.Analyzer{
	Name: "test_call",
	Run:  run,
	Requires: []*analysis.Analyzer{
		buildssa.Analyzer,
	},
}

func Test(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, Analyzer, "b")
}

func run(pass *analysis.Pass) (interface{}, error) {
	st := analysisutil.TypeOf(pass, "b", "*st")
	open := analysisutil.MethodOf(st, "b.open")
	close := analysisutil.MethodOf(st, "b.close")
	doSomething := analysisutil.MethodOf(st, "b.doSomething")
	doSomethingSpecial := analysisutil.MethodOf(st, "b.doSomethingSpecial")
	errFunc := analysisutil.MethodOf(st, "b.err")
	ie := analysisutil.ObjectOf(pass, "io", "EOF")

	funcs := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA).SrcFuncs
	for _, f := range funcs {
		for _, b := range f.Blocks {
			for i, instr := range b.Instrs {
				recv := analysisutil.ReturnReceiverIfCalled(instr, doSomething)
				if recv == nil {
					continue
				}
				if called, ok := analysisutil.CalledFromAfter(b, i, recv, close); !(called && ok) {
					pass.Reportf(instr.Pos(), "close should be called after calling doSomething")
				}
				if called, ok := analysisutil.CalledFromBefore(b, i, recv, open); !(called && ok) {
					pass.Reportf(instr.Pos(), "open should be called before calling doSomething")
				}
			}

			for i, instr := range b.Instrs {
				recv := analysisutil.ReturnReceiverIfCalled(instr, doSomethingSpecial)
				if recv == nil {
					continue
				}
				if called, ok := analysisutil.CalledFromBefore(b, i, recv, errFunc); !(called && ok) {
					pass.Reportf(instr.Pos(), "err not called")
				}
				if analysisutil.CalledBeforeAndEqualTo(b, recv, errFunc, ie) {
					continue
				}
				pass.Reportf(instr.Pos(), "err should be io.EOF when calling doSomethingSpecial")
			}
		}
	}

	return nil, nil
}
