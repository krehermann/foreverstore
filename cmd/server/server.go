package main

import (
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

	TCPP2P     TCPP2PServerCmd `cmd:"" help:"start tcp p2p server"`
	FileServer FileServerCmd   `cmd:"" help:"start file server"`
}

func main() {
	logger := zap.NewNop().Sugar()
	ctx := kong.Parse(&CLI, kong.Bind(logger))

	// Call the Run() method of the selected parsed command.
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
