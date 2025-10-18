package main

import (
	"os"

	"kubectl-regex-match/pkg/cmd"

	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

func main() {
	flags := pflag.NewFlagSet("kubectl-regex-match", pflag.ExitOnError)
	pflag.CommandLine = flags

	root := cmd.NewRegExCmd(genericiooptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
