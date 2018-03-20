// Copyright 2017 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/loader"
	"github.com/open-policy-agent/opa/util"
	"github.com/spf13/cobra"
)

var checkParams = struct {
	format   *util.EnumFlag
	errLimit int
}{
	format: util.NewEnumFlag(checkFormatPretty, []string{
		checkFormatPretty, checkFormatJSON,
	}),
}

const (
	checkFormatPretty = "pretty"
	checkFormatJSON   = "json"
)

var errLimit int

var checkCommand = &cobra.Command{
	Use:   "check <path> [path [...]]",
	Short: "Check Rego source files",
	Long: `Check Rego source files for parse and compilation errors.

If the 'check' command succeeds in parsing and compiling the source file(s), no output
is produced. If the parsing or compiling fails, 'check' will output the errors
and exit with a non-zero exit code.`,

	PreRunE: func(Cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("specify at least one file")
		}
		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(checkModules(args))
	},
}

func checkModules(args []string) int {

	result, err := loader.AllRegos(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	modules := map[string]*ast.Module{}

	for _, m := range result.Modules {
		modules[m.Name] = m.Parsed
	}

	compiler := ast.NewCompiler().SetErrorLimit(checkParams.errLimit)

	compiler.Compile(modules)

	if !compiler.Failed() {
		return 0
	}

	switch checkParams.format.String() {
	case checkFormatJSON:
		result := map[string]error{
			"errors": compiler.Errors,
		}
		bs, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		} else {
			fmt.Fprintln(os.Stdout, string(bs))
		}
	default:
		fmt.Fprintln(os.Stdout, compiler.Errors)
	}

	return 1
}

func init() {
	setMaxErrors(checkCommand.Flags(), &checkParams.errLimit)
	checkCommand.Flags().VarP(checkParams.format, "format", "f", "set output format")
	RootCommand.AddCommand(checkCommand)
}