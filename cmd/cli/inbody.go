package main

import "github.com/spf13/cobra"

var inbodyCmd = &cobra.Command{
	Use:   "inbody",
	Short: "InBody body composition scans",
}

var inbodyScansCmd = &cobra.Command{
	Use:   "scans",
	Short: "Body composition scans",
}

var inbodyScansListCmd = &cobra.Command{
	Use:   "list",
	Short: "List scans",
	RunE: func(cmd *cobra.Command, args []string) error {
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		limit, _ := cmd.Flags().GetString("limit")
		data, err := doGet("/api/v1/inbody/scans", map[string]string{"from": from, "to": to, "limit": limit})
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var inbodyScansLatestCmd = &cobra.Command{
	Use:   "latest",
	Short: "Get latest scan",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := doGet("/api/v1/inbody/scans/latest", nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

func init() {
	inbodyScansListCmd.Flags().String("from", "", "Start date (YYYY-MM-DD)")
	inbodyScansListCmd.Flags().String("to", "", "End date (YYYY-MM-DD)")
	inbodyScansListCmd.Flags().String("limit", "", "Max results")

	inbodyScansCmd.AddCommand(inbodyScansListCmd, inbodyScansLatestCmd)
	inbodyCmd.AddCommand(inbodyScansCmd)
}
