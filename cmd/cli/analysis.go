package main

import "github.com/spf13/cobra"

var analysisCmd = &cobra.Command{
	Use:   "analysis",
	Short: "AI-powered health analysis",
}

var analysisRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Generate AI health analysis",
	RunE: func(cmd *cobra.Command, args []string) error {
		date, _ := cmd.Flags().GetString("date")

		var payload interface{}
		if date != "" {
			payload = map[string]string{"date": date}
		}

		data, err := doPost("/api/v1/analysis", payload)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

func init() {
	analysisRunCmd.Flags().String("date", "", "Analysis date (YYYY-MM-DD, defaults to yesterday)")

	analysisCmd.AddCommand(analysisRunCmd)
}
