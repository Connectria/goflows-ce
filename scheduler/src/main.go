// main.go - CLI for goflows-scheduler

package main

import (
	"log"
	"os"

	"github.com/urfave/cli"
)

// main
func main() {
	app := &cli.App{
		Name:  "goflows-scheduler",
		Usage: "CLI to manage the GoFlows scheduler daemon",
		Commands: []cli.Command{
			{
				Name:     "daemon",
				Usage:    "Launch (be) the daemon",
				HideHelp: true,
				Hidden:   true,
				Action: func(c *cli.Context) error {
					daemon()
					return nil
				},
			},
			{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "List configuration",
				Action: func(c *cli.Context) error {
					printConfig()
					return nil
				},
			},
			{
				Name:  "status",
				Usage: "Check the status of the scheduler-daemon",
				Action: func(c *cli.Context) error {
					return daemonCmd("status")
				},
			},
			{
				Name:  "start",
				Usage: "Start the scheduler-daemon",
				Action: func(c *cli.Context) error {
					return daemonCmd("start")
				},
			},
			{
				Name:  "stop",
				Usage: "Stop the scheduler-daemon",
				Action: func(c *cli.Context) error {
					return daemonCmd("stop")
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
