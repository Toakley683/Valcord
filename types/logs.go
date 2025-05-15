package types

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"time"
)

var (
	CurrentLogName    = ""
	logDirectoryExist = false
	LogFile           *os.File
	MaxLogEntries     = 250
)

func NewLog(log ...any) {

	checkLogDirectory()

	fmt.Println(log...)
	fmt.Fprintln(LogFile, log...)

}

func logCleanup(LogDirectory string) {

	// Remove oldest log file if above max amount of logs

	logEntries, err := os.ReadDir(LogDirectory)
	checkError(err)

	for C, e := range logEntries {

		info, err := e.Info()
		checkError(err)

		if C >= MaxLogEntries {

			fmt.Println(e)

			Dir := LogDirectory + "\\" + info.Name()

			os.Remove(Dir)

		}
	}

}

func checkLogFile(LogDirectory string, LogName string) {

	// Check if File exists, if not create it

	_, err := os.Stat(LogDirectory + "\\" + LogName)

	if errors.Is(err, fs.ErrNotExist) {

		// File doesn't exist

		// Create file
		err = os.WriteFile(LogDirectory+"\\"+LogName, []byte{1}, 0700)
		checkError(err)

		checkLogFile(LogDirectory, LogName)

		return

	}
	checkError(err)

	if err == nil {

		logCleanup(LogDirectory)

		// File exists, read file and output settings

		LogFile, err = os.OpenFile(LogDirectory+"\\"+LogName, os.O_APPEND|os.O_WRONLY, 0644)

		if err != nil {
			LogFile.Close()
			checkError(err)
		}

		return

	}

}

func logDirectoryExists() {

	if logDirectoryExist {
		return
	}
	logDirectoryExist = true

	DateTime := time.Now()

	year := strconv.Itoa(DateTime.Year())
	month := strconv.Itoa(int(DateTime.Local().Month()))
	day := strconv.Itoa(int(DateTime.Local().Day()))

	hour := strconv.Itoa(int(DateTime.Local().Hour()))
	minute := strconv.Itoa(int(DateTime.Local().Minute()))
	second := strconv.Itoa(int(DateTime.Local().Second()))

	CurrentLogName = "Date(" + day + "-" + month + "-" + year + ")_(" + hour + "-" + minute + "-" + second + ").log"

	checkLogFile(Logs_directory, CurrentLogName)

}

func checkLogDirectory() {

	// Check if Directory exists, if not create it

	_, err := os.Stat(Logs_directory)

	if errors.Is(err, fs.ErrNotExist) {
		// File doesn't exist

		err := os.Mkdir(Logs_directory, 0700)
		checkError(err)

		checkLogDirectory()
		return
	}

	if err == nil {

		// Directory exists, continue steps
		logDirectoryExists()

	}

	checkError(err)

}
