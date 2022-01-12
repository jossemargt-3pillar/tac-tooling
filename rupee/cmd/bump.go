/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// bumpCmd represents the bump command
var bumpCmd = &cobra.Command{
	Use:   "bump <chart name> <version field name>=<new version>",
	Short: "Update version value for a given fiel on Charts",
	Args:  cobra.MinimumNArgs(2),
	// UsageTemplate: ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ensureRepository("", ""); err != nil {
			return err
		}

		chart := args[0]
		versions, err := getVersionsFor("", chart)

		if err != nil {
			return err
		}

		args = args[1:]
		for _, a := range args {
			aa := strings.Split(a, "=")

			if len(aa) < 2 {
				continue
			}

			field := aa[0]
			value := aa[1]

			t, ok := versions[field]
			if !ok {
				fmt.Fprintf(cmd.ErrOrStderr(), "[ERR] could not find %s\n", field)
				continue
			}

			var raw []byte
			raw, err = setVersion(t, value)
			if err != nil {
				return err
			}

			fmt.Fprint(cmd.OutOrStdout(), string(raw))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(bumpCmd)
	// bumpCmd.Flags().BoolP("in-place", "i", false, "Write changes on target file")
}
