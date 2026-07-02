package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/adapters/modelgateway"
	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/api/runtimeapi"
	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/workspace"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer, stderr io.Writer) error {
	if len(args) == 0 {
		return errors.New("missing command: expected ping or serve")
	}

	switch args[0] {
	case "ping":
		return runPing(args[1:], stdout, stderr)
	case "serve":
		return runServe(args[1:], stdout, stderr)
	default:
		return fmt.Errorf("unknown command %q: expected ping or serve", args[0])
	}
}

func runPing(args []string, stdout io.Writer, stderr io.Writer) error {
	flags := flag.NewFlagSet("ping", flag.ContinueOnError)
	flags.SetOutput(stderr)
	traceID := flags.String("trace-id", "", "trace id for deterministic smoke tests")
	if err := flags.Parse(args); err != nil {
		return err
	}

	return json.NewEncoder(stdout).Encode(runtimeapi.Ping(*traceID))
}

func runServe(args []string, stdout io.Writer, stderr io.Writer) error {
	flags := flag.NewFlagSet("serve", flag.ContinueOnError)
	flags.SetOutput(stderr)
	token := flags.String("token", "", "local engine bearer token")
	if err := flags.Parse(args); err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	store := workspace.NewStore(
		workspace.WithTraceID(runtimeapi.NewTraceID),
		workspace.WithModelGateway(modelgateway.NewGateway()),
		workspace.WithConfigDir(workspace.DefaultConfigDir()),
	)
	return runtimeapi.ServeWithStore(ctx, *token, stdout, store)
}
