package main

import "github.com/spf13/cobra"

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Cross-source summaries and reports",
}

var summaryDailyCmd = &cobra.Command{
	Use:   "daily",
	Short: "Daily summary across all sources",
	RunE: func(cmd *cobra.Command, args []string) error {
		date, _ := cmd.Flags().GetString("date")
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		data, err := doGet("/api/v1/summary/daily", map[string]string{"date": date, "from": from, "to": to})
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var summaryWeeklyCmd = &cobra.Command{
	Use:   "weekly",
	Short: "Weekly trends with 7-day rolling averages",
	RunE: func(cmd *cobra.Command, args []string) error {
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		data, err := doGet("/api/v1/summary/weekly", map[string]string{"from": from, "to": to})
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var summaryReportCmd = &cobra.Command{
	Use:   "report",
	Short: "Full report data for a date",
	RunE: func(cmd *cobra.Command, args []string) error {
		date, _ := cmd.Flags().GetString("date")
		lookback, _ := cmd.Flags().GetString("lookback")
		data, err := doGet("/api/v1/summary/report", map[string]string{"date": date, "lookback": lookback})
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

func init() {
	summaryDailyCmd.Flags().String("date", "", "Specific date (YYYY-MM-DD)")
	summaryDailyCmd.Flags().String("from", "", "Start date (YYYY-MM-DD)")
	summaryDailyCmd.Flags().String("to", "", "End date (YYYY-MM-DD)")

	summaryWeeklyCmd.Flags().String("from", "", "Start date (YYYY-MM-DD)")
	summaryWeeklyCmd.Flags().String("to", "", "End date (YYYY-MM-DD)")

	summaryReportCmd.Flags().String("date", "", "Report date (YYYY-MM-DD)")
	summaryReportCmd.Flags().String("lookback", "30", "Days of history to include")

	summaryCmd.AddCommand(summaryDailyCmd, summaryWeeklyCmd, summaryReportCmd)
}
