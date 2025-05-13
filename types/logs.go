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
)

func NewLog(log ...any) {

	checkLogDirectory()

	fmt.Println(log...)
	fmt.Fprintln(LogFile, log...)

}

func checkLogFile(LogDirectory string) {

	// Check if File exists, if not create it

	_, err := os.Stat(LogDirectory)

	if errors.Is(err, fs.ErrNotExist) {

		// File doesn't exist

		// Create file
		err = os.WriteFile(LogDirectory, []byte{}, 0700)
		checkError(err)

		checkLogFile(LogDirectory)

		return

	}
	checkError(err)

	if err == nil {

		// File exists, read file and output settings

		LogFile, err = os.OpenFile(LogDirectory, os.O_APPEND|os.O_WRONLY, 0644)

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

	checkLogFile(Logs_directory + "\\" + CurrentLogName)

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
