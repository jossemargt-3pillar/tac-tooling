/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"github.com/spf13/cobra"
)

// listChartsCmd represents the listCharts command
var listChartsCmd = &cobra.Command{
	Use:   "listCharts",
	Args:  cobra.NoArgs,
	Short: "List available rke2-charts",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ensureRepository("", ""); err != nil {
			return err
		}

		chartDirs, err := listCharts("")
		if err != nil {
			return err
		}

		prettyPrint(cmd.OutOrStdout(), chartDirs)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listChartsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listChartsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listChartsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
