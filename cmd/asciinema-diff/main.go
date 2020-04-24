package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	diff "github.com/willabides/asciinema-diff"
)

var cli struct {
	File1         string   `kong:"arg,type=existingfile,name=file,help='file to compare'"`
	File2         string   `kong:"arg,type=existingfile,name=file,help='file to compare'"`
	TimeTolerance int      `kong:"short=t,help='amount of time drift allowed between each event, in milliseconds'"`
	Quiet         bool     `kong:"short=q,help='no output on stdout'"`
	Header        []string `kong:"short=h,help='header fields to compare'"`
}

func fatalIfErr(ctx *kong.Context, err error, msg string, args ...interface{}) {
	if err == nil {
		return
	}
	output := fmt.Sprintf(msg, args...)
	if !strings.HasSuffix(output, "\n") {
		output += "\n"
	}
	fmt.Fprintln(ctx.Stderr, output)
	ctx.Exit(1)
}

func main() {
	ctx := kong.Parse(&cli)
	file1, err := os.Open(cli.File1)
	fatalIfErr(ctx, err, "error opening %s", cli.File1)
	defer file1.Close() //nolint:errcheck
	file2, err := os.Open(cli.File2)
	fatalIfErr(ctx, err, "error opening %s", cli.File2)
	defer file2.Close() //nolint:errcheck
	tolerance := time.Duration(cli.TimeTolerance) * time.Millisecond
	got, err := diff.Equal(file1, file2, diff.TimeTolerance(tolerance), diff.CompareHeaderFields(cli.Header...))
	fatalIfErr(ctx, err, "error comparing casts")
	if !got {
		if !cli.Quiet {
			fmt.Fprintf(ctx.Stdout, "casts are not equal\n")
		}
		ctx.Exit(2)
	}
	if !cli.Quiet {
		fmt.Fprintf(ctx.Stdout, "casts are equal\n")
	}
}
