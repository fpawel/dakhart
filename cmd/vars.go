package main

import (
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	"github.com/powerman/structlog"
	"net"
	"os"
	"path/filepath"
)

var (
	pipeConn     net.Conn
	stendAddr    modbus.Addr
	commModbus, commHart comm.T
	log = structlog.DefaultLogger.
		SetPrefixKeys(
			structlog.KeyApp,
			structlog.KeyPID,
			structlog.KeyLevel,
		).
		SetSuffixKeys(structlog.KeyUnit, structlog.KeySource).
		SetDefaultKeyvals(
			structlog.KeyApp, filepath.Base(os.Args[0]),
			structlog.KeySource, structlog.Auto,
		).
		SetKeysFormat(map[string]string{
			structlog.KeySource: " %6[2]s",
			structlog.KeyUnit:   " %6[2]s",
		})

)

