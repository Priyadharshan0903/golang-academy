// Package clicobra holds the Cobra example for lesson 14.1: the
// commands of jsonq (examples/m14/jsonq), built with
// github.com/spf13/cobra instead of the flag package.
package clicobra

import (
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

// [start root]

// NewRootCmd builds the command tree. Building it in a function, not in
// package variables and init functions, lets every test get a fresh
// tree with its own input and output.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "jsonq",
		Short:         "Answer simple questions about JSON",
		SilenceUsage:  true, // do not print usage for errors from RunE
		SilenceErrors: true, // the caller prints errors once
	}
	root.PersistentFlags().Bool("compact", false, "print JSON without indentation") // inherited by subcommands
	root.AddCommand(newKeysCmd(), newGetCmd())
	return root
}

func newKeysCmd() *cobra.Command {
	var sorted bool
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "Print the top-level keys of an object read from stdin",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			obj, err := readObject(cmd.InOrStdin())
			if err != nil {
				return err
			}
			keys := make([]string, 0, len(obj))
			for k := range obj {
				keys = append(keys, k)
			}
			if sorted {
				slices.Sort(keys)
			}
			fmt.Fprintln(cmd.OutOrStdout(), strings.Join(keys, "\n"))
			return nil
		},
	}
	cmd.Flags().BoolVarP(&sorted, "sort", "s", false, "sort the keys")
	return cmd
}

func newGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "get KEY",
		Short:   "Print the value of a top-level key",
		Example: `  echo '{"name":"Ada"}' | jsonq get name`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			obj, err := readObject(cmd.InOrStdin())
			if err != nil {
				return err
			}
			v, ok := obj[args[0]]
			if !ok {
				return fmt.Errorf("no key %q", args[0])
			}
			compact, _ := cmd.Flags().GetBool("compact")
			var b []byte
			if compact {
				b, err = json.Marshal(v)
			} else {
				b, err = json.MarshalIndent(v, "", "  ")
			}
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(b))
			return nil
		},
	}
}

// [end root]

func readObject(r io.Reader) (map[string]any, error) {
	var obj map[string]any
	if err := json.NewDecoder(r).Decode(&obj); err != nil {
		return nil, fmt.Errorf("stdin: invalid JSON object: %w", err)
	}
	return obj, nil
}
