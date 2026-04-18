package main

import "github.com/spf13/cobra"

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Data sync management",
}

var syncStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show sync status for all sources",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := doGet("/api/v1/sync/status", nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var syncTriggerCmd = &cobra.Command{
	Use:   "trigger <source>",
	Short: "Trigger sync for a source (oura, strava, hevy, inbody, all)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := doPost("/api/v1/sync/"+args[0], nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

func init() {
	syncCmd.AddCommand(syncStatusCmd, syncTriggerCmd)
}
