package main

import "github.com/spf13/cobra"

var hevyCmd = &cobra.Command{
	Use:   "hevy",
	Short: "Hevy gym workouts",
}

var hevyWorkoutsCmd = &cobra.Command{
	Use:   "workouts",
	Short: "Gym workouts",
}

var hevyWorkoutsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List workouts",
	RunE: func(cmd *cobra.Command, args []string) error {
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		data, err := doGet("/api/v1/hevy/workouts", map[string]string{"from": from, "to": to})
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var hevyWorkoutsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a workout with sets",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := doGet("/api/v1/hevy/workouts/"+args[0], nil)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

func init() {
	hevyWorkoutsListCmd.Flags().String("from", "", "Start date (YYYY-MM-DD)")
	hevyWorkoutsListCmd.Flags().String("to", "", "End date (YYYY-MM-DD)")

	hevyWorkoutsCmd.AddCommand(hevyWorkoutsListCmd, hevyWorkoutsGetCmd)
	hevyCmd.AddCommand(hevyWorkoutsCmd)
}
