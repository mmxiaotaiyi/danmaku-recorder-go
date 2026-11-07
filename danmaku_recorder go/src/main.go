package main

import (
	"dm_client_go/cli"

	bili "dm_client_go"
	log "github.com/sirupsen/logrus"
)

func init() {
	bili.ChangeLogLevel(log.TraceLevel)
}

func main() {
	cli.Run()
}
