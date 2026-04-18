package main

import "github.com/spf13/cobra"

var ouraCmd = &cobra.Command{
	Use:   "oura",
	Short: "Oura Ring data (sleep, scores)",
}

var ouraSleepCmd = &cobra.Command{
	Use:   "sleep",
	Short: "Sleep sessions",
}

var ouraSleepListCmd = &cobra.Command{
	Use:   "list",
	Short: "List sleep sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		typ, _ := cmd.Flags().GetString("type")
		data, err := doGet("/api/v1/oura/sleep", map[string]string{"from": from, "to": to, "type": typ})
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var ouraSleepGetCmd = &cobra.Command{
	Use:   "get <day>",
	Short: "Get sleep for a specific day",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := doGet("/api/v1/oura/sleep/"+args[0], nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var ouraScoresCmd = &cobra.Command{
	Use:   "scores",
	Short: "Daily scores (sleep, readiness, activity)",
}

var ouraScoresListCmd = &cobra.Command{
	Use:   "list",
	Short: "List daily scores",
	RunE: func(cmd *cobra.Command, args []string) error {
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		data, err := doGet("/api/v1/oura/scores", map[string]string{"from": from, "to": to})
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var ouraScoresGetCmd = &cobra.Command{
	Use:   "get <day>",
	Short: "Get scores for a specific day",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := doGet("/api/v1/oura/scores/"+args[0], nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

func init() {
	ouraSleepListCmd.Flags().String("from", "", "Start date (YYYY-MM-DD)")
	ouraSleepListCmd.Flags().String("to", "", "End date (YYYY-MM-DD)")
	ouraSleepListCmd.Flags().String("type", "", "Sleep type filter")

	ouraScoresListCmd.Flags().String("from", "", "Start date (YYYY-MM-DD)")
	ouraScoresListCmd.Flags().String("to", "", "End date (YYYY-MM-DD)")

	ouraSleepCmd.AddCommand(ouraSleepListCmd, ouraSleepGetCmd)
	ouraScoresCmd.AddCommand(ouraScoresListCmd, ouraScoresGetCmd)
	ouraCmd.AddCommand(ouraSleepCmd, ouraScoresCmd)
}
