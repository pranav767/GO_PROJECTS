package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
)

type JobSpec struct {
	Name    string `json:"name"`
	Cron    string `json:"cron"`    // cron expression
	Command string `json:"command"` // shell command to run
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run scheduler to execute configured backup jobs",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath, _ := cmd.Flags().GetString("config")
		if cfgPath == "" {
			return fmt.Errorf("--config is required")
		}
		data, err := ioutil.ReadFile(cfgPath)
		if err != nil {
			return err
		}
		var jobs []JobSpec
		if err := json.Unmarshal(data, &jobs); err != nil {
			return err
		}

		c := cron.New(cron.WithParser(cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)))
		for _, j := range jobs {
			job := j
			_, err := c.AddFunc(job.Cron, func() {
				fmt.Printf("[%s] running command: %s\n", time.Now().Format(time.RFC3339), job.Command)
				// run in background with context and timeout
				ctx, cancel := context.WithTimeout(context.Background(), 24*time.Hour)
				defer cancel()
				sh := exec.CommandContext(ctx, "sh", "-c", job.Command)
				sh.Stdout = os.Stdout
				sh.Stderr = os.Stderr
				if err := sh.Run(); err != nil {
					fmt.Printf("job %s failed: %v\n", job.Name, err)
				}
			})
			if err != nil {
				return fmt.Errorf("failed to schedule job %s: %w", job.Name, err)
			}
			fmt.Printf("scheduled job %s: %s -> %s\n", job.Name, job.Cron, job.Command)
		}
		c.Start()
		fmt.Println("scheduler started")
		// block until terminated
		select {}
	},
}

func init() {
	serveCmd.Flags().String("config", "", "Path to JSON schedule config file")
	rootCmd.AddCommand(serveCmd)
}
