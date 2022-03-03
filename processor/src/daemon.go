// daemon.go - daemon action

package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// start
func daemon() {
	logger.Info().
		Str("function", "daemon()").
		Msg("Initiated.")

	// redirect stdout to file
	newStdOut, err := os.OpenFile("daemon-std.out", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to create file for newStdOut: %v", err)
	}

	err = syscall.Dup2(int(newStdOut.Fd()), int(os.Stderr.Fd()))
	if err != nil {
		log.Fatalf("Failed to redirect stdout to file: %v", err)
	}

	// redirect stderr to file in case of panic
	newStdErr, err := os.OpenFile("daemon-errors.out", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to create file for newStdErr: %v", err)
	}

	err = syscall.Dup2(int(newStdErr.Fd()), int(os.Stderr.Fd()))
	if err != nil {
		log.Fatalf("Failed to redirect stderr to file: %v", err)
	}

	// listen for signals
	daemonFlag = true
	sigc := make(chan os.Signal, 1)
	signal.Notify(
		sigc,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT, // daemon stop command will send SIGQUIT1
	)

	go func() {
		s := <-sigc
		logger.Info().
			Str("function", "watchSignal()").
			Msgf("Received: %v", s)
		closeRabbitMQ()
		time.Sleep(3 * time.Second)
		os.Exit(0)
	}()

	// start RabbitMQ
	err = initRabbitMQ()
	if err != nil {
		logger.Error().
			Str("function", "daemon()").
			Msgf("initRabbitMQ() returned error: '%v'", err.Error())
	}

	// start API listener
	httpRouter1 := apiRouter()
	go httpRouter1.Run(":" + cfg.APIPort)
	logger.Info().
		Str("function", "daemon()").
		Msgf("goflows-processor API listing on port %v (httpRouter1)", cfg.APIPort)

		// start OpsGenie Alert Action listener
		/** Disabling AlertActions for plug-in devel
			if cfg.AlertActionsEnabled {
				httpRouter2 := alertActionsRouter()
				go httpRouter2.Run(":" + cfg.AlertActionsPort)
				logger.Info().
					Str("function", "daemon()").
					Msgf("goflows-processor alert actions listing on port %v (httpRouter2)", cfg.AlertActionsPort)
			}

		**/
	// monitor and process the queues
	work()
}

// work the queue
func work() {
	for {
		wg := sync.WaitGroup{}
		wg.Add(1)

		// receive OpsGenie "events" via RabbitMQ
		go func() {
			for e := range amqEventsConsumed {
				wg.Add(1)
				go processOpsGenieEvent(string(e.Body))
			}
		}()

		// receive scheduler "tasks" via RabbitMQ
		go func() {
			for t := range amqTasksConsumed {
				wg.Add(1)
				go processSchedulerTask(string(t.Body))
			}
		}()

		wg.Wait()
	}
}
