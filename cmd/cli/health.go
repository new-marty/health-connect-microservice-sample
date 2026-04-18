package main

import "github.com/spf13/cobra"

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check API health",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := doGet("/api/health", nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}
