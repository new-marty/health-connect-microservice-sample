package main

import "github.com/spf13/cobra"

var stravaCmd = &cobra.Command{
	Use:   "strava",
	Short: "Strava activities",
}

var stravaActivitiesCmd = &cobra.Command{
	Use:   "activities",
	Short: "Strava activities",
}

var stravaActivitiesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List activities",
	RunE: func(cmd *cobra.Command, args []string) error {
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		typ, _ := cmd.Flags().GetString("type")
		data, err := doGet("/api/v1/strava/activities", map[string]string{"from": from, "to": to, "type": typ})
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var stravaActivitiesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a specific activity",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := doGet("/api/v1/strava/activities/"+args[0], nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

func init() {
	stravaActivitiesListCmd.Flags().String("from", "", "Start date (YYYY-MM-DD)")
	stravaActivitiesListCmd.Flags().String("to", "", "End date (YYYY-MM-DD)")
	stravaActivitiesListCmd.Flags().String("type", "", "Activity type filter")

	stravaActivitiesCmd.AddCommand(stravaActivitiesListCmd, stravaActivitiesGetCmd)
	stravaCmd.AddCommand(stravaActivitiesCmd)
}
