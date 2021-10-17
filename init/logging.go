package main

import (
	"bytes"
	"fmt"
	"init/quirk"
	"os"
)

const (
	// TOTHINK rename to debug/info/warning
	levelError = iota
	levelWarning
	levelInfo
	levelDebug
)

var (
	verbosityLevel = levelInfo // by default show info messages and errors
	printToConsole bool

	kmsg *os.File
)

func printMessage(format string, requestedLevel, kernelLevel int, v ...interface{}) {
	if verbosityLevel < requestedLevel {
		return
	}

	msg := fmt.Sprintf(format, v...)
	_, _ = fmt.Fprint(kmsg, "<", kernelLevel, ">booster: ", msg, "\n")
	if printToConsole {
		fmt.Println(msg)
	}
}

func debug(format string, v ...interface{}) {
	printMessage(format, levelDebug, 7, v...)
}

func info(format string, v ...interface{}) {
	printMessage(format, levelInfo, 6, v...)
}

func warning(format string, v ...interface{}) {
	printMessage(format, levelWarning, 4, v...)
}

// this is for critical error messages, call this function 'severe' to avoid name clashing with error class
func severe(format string, v ...interface{}) {
	printMessage(format, levelError, 2, v...)
}

const sysKmsgFile = "/proc/sys/kernel/printk_devkmsg"

func disableKmsgThrottling() error {
	data, err := os.ReadFile(sysKmsgFile)
	if err != nil {
		return err
	}
	enable := []byte("on\n")
	if bytes.Equal(data, enable) {
		return nil
	}

	return os.WriteFile(sysKmsgFile, enable, 0644)
}

// console prints message to console
// but if we are compiling the binary with "tets" tag (e.g. for integration tests) then it prints message to kmsg to avoid
// messing log output in qemu console
func console(format string, v ...interface{}) {
	if quirk.TestEnabled {
		msg := fmt.Sprintf(format, v...)
		_, _ = fmt.Fprint(kmsg, "<", 2, ">booster: ", msg, "\n")
	} else {
		fmt.Printf(format, v...)
	}
}
