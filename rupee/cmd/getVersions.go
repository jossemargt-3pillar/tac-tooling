/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"github.com/spf13/cobra"
)

// getVersionsCmd represents the getVersions command
var getVersionsCmd = &cobra.Command{
	Use:   "getVersions",
	Args:  cobra.ExactArgs(1),
	Short: "Get available versions on a given chart",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ensureRepository("", ""); err != nil {
			return err
		}

		chart := args[0]
		vrs, err := getVersionsFor("", chart)

		if err != nil {
			return err
		}

		prettyPrint(cmd.OutOrStdout(), vrs)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(getVersionsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getVersionsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getVersionsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
