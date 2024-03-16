package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"

	"go.nhat.io/aferocopy/v2"
)

var testedAppPath string

func init() {
	cwd, _ := os.Getwd()
	testedAppPath = cwd + "/test/example"

	tlogger := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(tlogger)
	slog.SetDefault(logger)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "depsapp:", err.Error())
		os.Exit(1)
	}
}

func run() error {
	log("hello")
	workingDirectory, err := os.MkdirTemp("", "*")
	if err != nil {
		return fmt.Errorf("mktemp -d: %w", err)
	}
	log("working dir = %q", workingDirectory)
	// defer os.RemoveAll(workingDirectory)

	if err := aferocopy.Copy(testedAppPath, workingDirectory, aferocopy.Options{
		Sync: true,
	}); err != nil {
		return fmt.Errorf("copy dirs: %w", err)
	}

	// goListCmd := exec.CommandContext(context.Background(),
	// 	"go", "env",
	// )

	goListCmd := exec.CommandContext(context.Background(),
		"go", "list", "-u", "-m",
		// "-f", "{{ if not .Indirect | and .Update }}{{ .Path }}@{{ .Update.Version }}{{end}}",
		"-f", "{{ if .Update | and (not .Indirect) }}{{ .Path }}@{{ .Update.Version }}{{end}}",
		"all",
	)
	goListCmd.Dir = workingDirectory

	var out bytes.Buffer
	goListCmd.Stdout = &out

	goListCmd.Stderr = os.Stderr

	// scn := bufio.NewScanner(&out)

	if err := goListCmd.Start(); err != nil {
		return fmt.Errorf("starting go list command: %w", err)
	}
	log("started command %s", goListCmd.String())
	if err := goListCmd.Wait(); err != nil {
		return fmt.Errorf("go list command: %w", err)
	}
	log("after wait")
	var lines []string
	var l string
	for l, err = out.ReadString('\n'); err == nil; l, err = out.ReadString('\n') {
		log("read line> %s", l)
		lines = append(lines, l)
	}
	if !errors.Is(err, io.EOF) {
		return fmt.Errorf("reading output: %w", err)
	}

	if len(lines) > 0 {
		for _, l := range lines {
			fmt.Print("go get", l)
		}
		fmt.Println()
		fmt.Println("# run the commands to update!!")
	}

	return nil
}

func log(format string, args ...any) {
	slog.Log(context.Background(), slog.LevelDebug, fmt.Sprintf(format, args...))
}
