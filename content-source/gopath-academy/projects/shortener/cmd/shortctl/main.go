// Command shortctl is a command-line client for the shortener API
// (the Module 14 project).
//
//	shortctl [global flags] create [-code CODE] URL
//	shortctl [global flags] get CODE
//	shortctl [global flags] delete CODE
//	shortctl version
//
// Global flags: -server URL, -timeout DURATION, -output text|json,
// -config FILE. They can also come from SHORTCTL_SERVER,
// SHORTCTL_TIMEOUT, SHORTCTL_OUTPUT and a JSON config file; flags win
// over the environment, which wins over the file (lesson 14.2).
package main

import (
	"context"
	"os"
	"os/signal"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	os.Exit(run(ctx, os.Args[1:], defaultEnv(), os.Stdout, os.Stderr))
}
