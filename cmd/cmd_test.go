package cmd

import (
	"bytes"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestCLICommands(t *testing.T) {
	dbPath = "test_cli.db"
	defer os.Remove(dbPath)

	executeCommand := func(args ...string) (string, error) {
		buf := new(bytes.Buffer)
		RootCmd.SetOut(buf)
		RootCmd.SetErr(buf)
		RootCmd.SetArgs(args)

		err := RootCmd.Execute()
		return buf.String(), err
	}

	output, err := executeCommand("list")
	if err != nil {
		t.Fatalf("list command failed: %v", err)
	}
	if !strings.Contains(output, "No rules registered.") {
		t.Errorf("expected empty list message, got: %q", output)
	}

	output, err = executeCommand("add", "vtest.localhost", "localhost:3000")
	if err != nil {
		t.Fatalf("add command failed: %v", err)
	}
	if !strings.Contains(output, "Added rule: vtest.localhost -> localhost:3000") {
		t.Errorf("unexpected add output: %q", output)
	}

	output, err = executeCommand("get", "vtest.localhost")
	if err != nil {
		t.Fatalf("get command failed: %v", err)
	}
	if !strings.Contains(output, "vtest.localhost") || !strings.Contains(output, "localhost:3000") {
		t.Errorf("unexpected get output: %q", output)
	}

	output, err = executeCommand("list")
	if err != nil {
		t.Fatalf("list command failed: %v", err)
	}
	if !strings.Contains(output, "vtest.localhost") || !strings.Contains(output, "localhost:3000") {
		t.Errorf("unexpected list output: %q", output)
	}

	output, err = executeCommand("remove", "vtest.localhost")
	if err != nil {
		t.Fatalf("remove command failed: %v", err)
	}
	if !strings.Contains(output, "Removed rule for public pattern: vtest.localhost") {
		t.Errorf("unexpected remove output: %q", output)
	}
}

func TestStartCommand(t *testing.T) {
	dbPath = "test_start_cli.db"
	defer os.Remove(dbPath)

	go func() {
		buf := new(bytes.Buffer)
		RootCmd.SetOut(buf)
		RootCmd.SetErr(buf)
		RootCmd.SetArgs([]string{"start", "--http-port", "8083", "--https-port", "8446", "--admin-port", "2022"})
		_ = RootCmd.Execute()
	}()

	time.Sleep(1 * time.Second)

	err := syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	if err != nil {
		t.Fatalf("failed to send SIGINT: %v", err)
	}

	time.Sleep(500 * time.Millisecond)
}
