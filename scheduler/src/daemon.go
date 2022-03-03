// daemon.go - daemon action

package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"
)

// start
func daemon() {
	logger.Info().
		Str("function", "daemon()").
		Msg("Initiated.")

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
		syscall.SIGQUIT, // daemon stop command will send SIGQUIT
	)

	go func() {
		s := <-sigc
		logger.Info().
			Str("function", "watchSignal()").
			Msgf("Received: %v", s)
		scheduler.Stop()
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

	// start cron service
	scheduler = cron.New(cron.WithLocation(time.UTC))
	scheduler.Start()
	logger.Info().
		Str("function", "daemon()").
		Msg("Cron service started")

	// start API listener
	r := router()
	go r.Run(":" + cfg.APIPort)
	logger.Info().
		Str("function", "daemon()").
		Msgf("goflows-scheduler API listing on port %v", cfg.APIPort)

	// just wait...
	for {
		wg := sync.WaitGroup{}
		wg.Add(1)
		wg.Wait()
	}
}
