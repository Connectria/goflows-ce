// daemoncmds.go - start/stop/status goflows-processor

package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"syscall"
)

// PIDFile is the file used to store the process id
var PIDFile = "/var/tmp/goflows-processor.pid"

// save PID
func savePID(pid int) error {
	file, err := os.Create(PIDFile)
	if err != nil {
		logger.Error().
			Str("function", "savePID()").
			Msgf("os.Create(PIDFile = '%v') returned: '%v'", PIDFile, err)
		return err
	}

	defer file.Close()

	_, err = file.WriteString(strconv.Itoa(pid))
	if err != nil {
		logger.Error().
			Str("function", "savePID()").
			Msgf("file.WriteString(pid = '%v') returned: '%v'", pid, err)
		return err
	}

	logMsg := fmt.Sprintf("Saved process ID to '%v'", PIDFile)
	logger.Info().
		Str("function", "savePID()").
		Msg(logMsg)
	fmt.Println(logMsg)
	file.Sync()
	return nil
}

// process the command
func daemonCmd(cmd string, flag bool) error {
	logger.Info().
		Str("function", "daemonCmd()").
		Msgf("Received '%v' command", cmd)

	// start the daemon
	if cmd == "start" {
		_, err := os.Stat(PIDFile)
		if err == nil {
			logMsg := fmt.Sprintf("daemon is already running. Check process ID in '%v'", PIDFile)
			logger.Warn().
				Str("function", "daemonCmd()").
				Msg(logMsg)
			return errors.New(logMsg)
		}

		// check for open call backs - print and delete
		/** Disabling Callbacks
				pipeline := []bson.D{}
				callBacks := goflows.GetCallBacks(mongoClient, pipeline)
				if len(callBacks) > 0 {
					daemonCmdDeleteCallBacks(callBacks)
					logMsg := fmt.Sprintf("start removed the %v call backs pending!", len(callBacks))
					logger.Warn().
						Str("function", "daemonCmd()").
						Msg(logMsg)
					fmt.Println(logMsg)
				}
		**/

		daemon := exec.Command(os.Args[0], "daemon")
		daemon.Start()

		logMsg := fmt.Sprintf("daemon process ID: %v", daemon.Process.Pid)
		logger.Info().
			Str("function", "daemonCmd()").
			Msg(logMsg)
		fmt.Println(logMsg)

		err = savePID(daemon.Process.Pid)
		if err != nil {
			logMsg = fmt.Sprintf("savePID(%v) returned: '%v'", daemon.Process.Pid, err)
			logger.Error().
				Str("function", "daemonCmd()").
				Msg(logMsg)
			return errors.New(logMsg)
		}

		return nil
	}

	// status or stop
	_, err := os.Stat(PIDFile)
	if err == nil {
		data, err := ioutil.ReadFile(PIDFile)
		if err != nil {
			logMsg := "daemon is not running"
			logger.Warn().
				Str("function", "daemonCmd()").
				Msg(logMsg)
			return errors.New(logMsg)
		}

		ProcessID, err := strconv.Atoi(string(data))
		if err != nil {
			logMsg := fmt.Sprintf("unable to read and parse process ID found in '%v'", PIDFile)
			logger.Error().
				Str("function", "daemonCmd()").
				Msg(logMsg)
			return errors.New(logMsg)
		}

		process, err := os.FindProcess(ProcessID)
		if err != nil {
			logMsg := fmt.Sprintf("unable to find process ID '%v'; os.Find returned:  '%v'", ProcessID, err)
			logger.Warn().
				Str("function", "daemonCmd()").
				Msg(logMsg)
			return errors.New(logMsg)
		}

		if cmd == "status" {
			logMsg := fmt.Sprintf("daemon process ID: %v", ProcessID)
			logger.Info().
				Str("function", "daemonCmd()").
				Msg(logMsg)
			fmt.Println(logMsg)

			/** Disabling Callbacks
						pipeline := []bson.D{}
						callBacks := goflows.GetCallBacks(mongoClient, pipeline)
						if len(callBacks) > 0 {
							daemonCmdPrintCallBacks(callBacks)
							fmt.Printf("number of call backs pending: %v\n", len(callBacks))
						}
			**/
			return nil
		}

		// check for open call backs - print and exit unless "--force" has been specified
		/** Disable Callbacks
				pipeline := []bson.D{}
				callBacks := goflows.GetCallBacks(mongoClient, pipeline)
				if len(callBacks) > 0 {
					daemonCmdPrintCallBacks(callBacks)
					if !flag {
						logMsg := fmt.Sprintf("There are %v call backs pending! goflows-processor was not stopped! Use '--force' to override.", len(callBacks))
						logger.Warn().
							Str("function", "daemonCmd()").
							Msg(logMsg)
						fmt.Println(logMsg)
						return nil
					}

					logMsg := fmt.Sprintf("There are %v call backs pending! Stopping goflows-processor because '--force' was specified.", len(callBacks))
					logger.Warn().
						Str("function", "daemonCmd()").
						Msg(logMsg)
					fmt.Println(logMsg)
				}
		**/
		// continue with stop
		os.Remove(PIDFile)

		logMsg := fmt.Sprintf("terminating process ID: %v", ProcessID)
		logger.Info().
			Str("function", "daemonCmd()").
			Msg(logMsg)
		fmt.Println(logMsg)

		err = process.Signal(syscall.SIGQUIT)
		if err != nil {
			logMsg := fmt.Sprintf("unable to terminate process ID '%v'; process.Signal(syscall.SIGQUIT) returned:  '%v'", ProcessID, err)
			logger.Error().
				Str("function", "daemonCmd()").
				Msg(logMsg)
			return errors.New(logMsg)
		}

		logMsg = fmt.Sprintf("terminated process ID: %v", ProcessID)
		logger.Info().
			Str("function", "daemonCmd()").
			Msg(logMsg)
		fmt.Println(logMsg)
		return nil
	}

	logMsg := "daemon is not running."
	logger.Warn().
		Str("function", "daemonCmd()").
		Msg(logMsg)
	fmt.Println(logMsg)
	return nil
}

// print pending call backs
/** Disable Callbacks
func daemonCmdPrintCallBacks(cbs []goflows.CallBackType) {
	for i, cb := range cbs {
		fmt.Printf("* Pending %v - JobID: %v - Reference: %v - CallBackID: %v\n", i, cb.JobID, cb.Reference, cb.ID)
	}
}

// delete call backs
func daemonCmdDeleteCallBacks(cbs []goflows.CallBackType) {
	for i, cb := range cbs {
		fmt.Printf("* Deleting %v - JobID: %v - Reference: %v - CallBackID: %v\n", i, cb.JobID, cb.Reference, cb.ID)
		err := goflows.DeleteCallBackByJobID(mongoClient, cb.JobID)
		if err != nil {
			fmt.Printf("*** Error received when attempting to delete: %v\n", err.Error())
		}
	}
}
**/
