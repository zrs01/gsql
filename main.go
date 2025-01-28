package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/yuin/gluare"
	lua "github.com/yuin/gopher-lua"
	"github.com/zrs01/gsql/goluasql"
	"github.com/ztrue/tracerr"
	luar "layeh.com/gopher-luar"
)

var (
	version = "development"
	debug   bool
)

func main() {
	logrus.SetFormatter(&nested.Formatter{})

	cliapp := cli.NewApp()
	cliapp.Name = "gsql"
	cliapp.Usage = "Database Executor"
	cliapp.Version = version
	cliapp.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:        "debug",
			Aliases:     []string{"d"},
			Usage:       "debug mode",
			Required:    false,
			Destination: &debug,
		},
	}
	cliapp.Action = func(ctx *cli.Context) error {
		if debug {
			logrus.SetLevel(logrus.DebugLevel)
		}
		file := ctx.Args().Get(0)
		if file == "" {
			return tracerr.New("missing script file")
		}
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return fmt.Errorf("file '%s' does not exist", file)
		}
		// get the cli args
		if _, err := doLuaFile(file, ctx.Args(), nil); err != nil {
			return tracerr.Wrap(err)
		}
		return nil
	}

	if err := cliapp.Run(os.Args); err != nil {
		tracerr.PrintSourceColor(err, 0)
	}
}

func doLuaFile(file string, args cli.Args, action func(*lua.LState) (interface{}, error)) (interface{}, error) {
	L := lua.NewState()
	defer L.Close()
	L.PreloadModule("regex", gluare.Loader)
	L.PreloadModule("sql", goluasql.Loader)

	// pass the arguments to the lua
	argArrays := []string{}
	for i := 0; i < args.Len(); i++ {
		argArrays = append(argArrays, args.Get(i))
	}
	L.SetGlobal("args", luar.New(L, argArrays))

	// add the path of given file to the lua package, make all lua files in the path can be found and run
	packagePath := filepath.Join(filepath.Dir(file), "?.lua")
	cmd := strings.Replace("package.path = package.path .. ';"+packagePath+"'", "\\", "/", -1)
	if err := L.DoString(cmd); err != nil {
		return nil, tracerr.Wrap(err)
	}
	if err := L.DoFile(file); err != nil {
		return nil, tracerr.Wrap(err)
	}
	if action != nil {
		// execute the action given from caller
		res, err := action(L)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		return res, nil
	}
	return nil, nil
}
