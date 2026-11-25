package main

import (
	"fmt"
	"os"
	"path/filepath"

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
		if err := doLuaFile(file, ctx.Args()); err != nil {
			return tracerr.Wrap(err)
		}
		return nil
	}

	if err := cliapp.Run(os.Args); err != nil {
		tracerr.PrintSourceColor(err, 0)
	}
}

func doLuaFile(file string, args cli.Args) error {
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
	packagePath := filepath.ToSlash(filepath.Join(filepath.Dir(file), "?.lua"))
	setPackagePath := fmt.Sprintf("package.path = package.path .. '%s'", packagePath)
	if err := L.DoString(setPackagePath); err != nil {
		return tracerr.Wrap(err)
	}
	if err := L.DoFile(file); err != nil {
		return tracerr.Wrap(err)
	}
	return nil
}
