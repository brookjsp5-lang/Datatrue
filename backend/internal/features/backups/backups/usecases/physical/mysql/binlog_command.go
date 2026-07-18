package usecases_physical_mysql

import (
	"errors"
	"fmt"
)

type BinlogStreamCommandRequest struct {
	MysqlbinlogBin   string
	DefaultsFilePath string
	Host             string
	Port             int
	Username         string
	ResultFilePrefix string
	StartFile        string
	IsTLS            bool
	ServerID         uint64
}

func BuildBinlogStreamCommand(request BinlogStreamCommandRequest) (CommandPlan, error) {
	if request.MysqlbinlogBin == "" {
		return CommandPlan{}, errors.New("mysqlbinlog binary is required")
	}
	if request.DefaultsFilePath == "" {
		return CommandPlan{}, errors.New("defaults file path is required")
	}
	if request.Host == "" {
		return CommandPlan{}, errors.New("host is required")
	}
	if request.Port <= 0 {
		return CommandPlan{}, errors.New("port is required")
	}
	if request.Username == "" {
		return CommandPlan{}, errors.New("username is required")
	}
	if request.ResultFilePrefix == "" {
		return CommandPlan{}, errors.New("result file prefix is required")
	}
	if request.StartFile == "" {
		return CommandPlan{}, errors.New("start file is required")
	}

	sslMode := "DISABLED"
	if request.IsTLS {
		sslMode = "REQUIRED"
	}

	args := []string{
		"--defaults-file=" + request.DefaultsFilePath,
		"--read-from-remote-server",
		"--raw",
		"--stop-never",
		"--to-last-log",
		"--host=" + request.Host,
		fmt.Sprintf("--port=%d", request.Port),
		"--user=" + request.Username,
		"--result-file=" + request.ResultFilePrefix,
		"--ssl-mode=" + sslMode,
	}

	if request.ServerID > 0 {
		args = append(args, fmt.Sprintf("--connection-server-id=%d", request.ServerID))
	}

	args = append(args, request.StartFile)

	return CommandPlan{
		Executable: request.MysqlbinlogBin,
		Args:       args,
	}, nil
}
