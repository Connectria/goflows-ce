// main.go - CLI to daemon processor

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli"
)

// main
func main() {
	app := &cli.App{
		Name:    "goflows-processor",
		Author:  "Connectria Automation & Product Engineering",
		Version: "R1.802",
		Usage:   "CLI to manage the GoFlows processor daemon",
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
				Name:    "listFuncs",
				Aliases: []string{"l"},
				Usage:   "List compiled goflows",
				Action: func(c *cli.Context) error {
					listCliFuncs()
					return nil
				},
			},
			{
				Name:    "process",
				Aliases: []string{"p", "proc"},
				Usage:   "Process OpsGenie evenit with messageID with compiled goflows (live)",
				Action: func(c *cli.Context) error {
					messageID := c.Args().First()
					if len(messageID) == 0 {
						fmt.Println("Error: No messageID specified!")
						return nil
					}

					err := initRabbitMQCLI()
					if err != nil {
						logger.Error().
							Str("function", "main()").
							Msgf("initRabbitMQCLI() returned error: '%v'", err.Error())
					}
					processOpsGenieEvent(messageID)
					closeRabbitMQ()
					return nil
				},
			},
			{
				Name:  "status",
				Usage: "Check the status of the processor-daemon",
				Action: func(c *cli.Context) error {
					return daemonCmd("status", false)
				},
			},
			{
				Name:  "start",
				Usage: "Start the processor-daemon",
				Action: func(c *cli.Context) error {
					return daemonCmd("start", false)
				},
			},
			{
				Name:  "stop",
				Usage: "Stop the processor-daemon",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "force",
						Usage: "force stop even with open call backs",
					},
				},
				Action: func(c *cli.Context) error {
					return daemonCmd("stop", c.Bool("force"))
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
