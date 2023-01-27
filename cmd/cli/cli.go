package main

import (
	"bufio"
	"net"
	"os"

	"github.com/alecthomas/kong"
	"go.uber.org/zap"
)

type debugFlag bool

func (d debugFlag) BeforeApply(logger *zap.SugaredLogger) error {
	l, err := zap.NewDevelopment()
	if err != nil {
		return err
	}
	logger = l.Sugar()
	return nil
}

var CLI struct {
	Debug debugFlag `help:"Enable debug logging."`

	Tcp TCPCmd `cmd:"" help:"connect to tcp server"`
}

var logger *zap.SugaredLogger

func main() {
	l, _ := zap.NewDevelopment() //zap.NewNop().Sugar()
	logger = l.Sugar()
	ctx := kong.Parse(&CLI, kong.Bind(logger))

	// Call the Run() method of the selected parsed command.
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}

type Unix struct {
	Socket string `help:"socket to connect to"`
}

type TCPCmd struct {
	Addr string `help:"address to connect to"`
	//logger *zap.Logger
}

func (t *TCPCmd) Run() error {
	conn, err := net.Dial("tcp", t.Addr)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(os.Stdin)

	for {

		str, err := reader.ReadString('\n')
		logger.Debugf("read: %s", str)
		if err != nil {
			return err
		}
		_, err = conn.Write([]byte(str))
		if err != nil {
			return err
		}
	}
}
