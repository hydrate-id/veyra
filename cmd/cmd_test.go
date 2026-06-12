package cmd

import (
	"bytes"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/korfra/veyra/caddy"
	"github.com/korfra/veyra/model"
	"github.com/korfra/veyra/store"
)

func TestCLICommands(t *testing.T) {
	dbPath = "test_cli.db"
	defer os.Remove(dbPath)

	caddyfile := "{\n\tadmin localhost:2019\n\thttp_port 8089\n\thttps_port 8449\n}\n"
	_ = caddy.Start(caddyfile)
	defer func() {
		_ = caddy.Stop()
	}()

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

	output, err = executeCommand("validate")
	if err != nil {
		t.Fatalf("validate command failed: %v", err)
	}
	if !strings.Contains(output, "All rules and Caddyfile configuration are valid.") {
		t.Errorf("unexpected validate output: %q", output)
	}

	output, err = executeCommand("export", "caddy")
	if err != nil {
		t.Fatalf("export caddy failed: %v", err)
	}
	if !strings.Contains(output, "vtest.localhost") {
		t.Errorf("expected exported Caddyfile to contain vtest.localhost: %s", output)
	}

	output, err = executeCommand("export", "json")
	if err != nil {
		t.Fatalf("export json failed: %v", err)
	}
	if !strings.Contains(output, "vtest.localhost") {
		t.Errorf("expected exported JSON to contain vtest.localhost: %s", output)
	}

	jsonFile := "test_rules.json"
	defer os.Remove(jsonFile)

	output, err = executeCommand("export", "json", jsonFile)
	if err != nil {
		t.Fatalf("export json to file failed: %v", err)
	}

	backupFile := "test_backup.json"
	defer os.Remove(backupFile)

	output, err = executeCommand("export", "backup", backupFile)
	if err != nil {
		t.Fatalf("export backup failed: %v", err)
	}

	output, err = executeCommand("remove", "vtest.localhost")
	if err != nil {
		t.Fatalf("remove command failed: %v", err)
	}

	output, err = executeCommand("list")
	if err != nil {
		t.Fatalf("list command failed: %v", err)
	}
	if strings.Contains(output, "vtest.localhost") {
		t.Errorf("expected vtest.localhost to be removed: %s", output)
	}

	output, err = executeCommand("import", "json", jsonFile)
	if err != nil {
		t.Fatalf("import json failed: %v", err)
	}

	output, err = executeCommand("import", "json", jsonFile)
	if err != nil {
		t.Fatalf("import json duplicate failed: %v", err)
	}
	if !strings.Contains(output, "Failed to import rule") {
		t.Errorf("expected duplicate import warning, got: %q", output)
	}

	output, err = executeCommand("import", "backup", backupFile)
	if err != nil {
		t.Fatalf("import backup failed: %v", err)
	}

	output, err = executeCommand("list")
	if err != nil {
		t.Fatalf("list command failed: %v", err)
	}
	if !strings.Contains(output, "vtest.localhost") {
		t.Errorf("expected vtest.localhost to be imported: %s", output)
	}

	output, err = executeCommand("remove", "vtest.localhost")
	if err != nil {
		t.Fatalf("remove command failed: %v", err)
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

func TestCLIExitErrors(t *testing.T) {
	dbPath = "test_cli_errors.db"
	defer os.Remove(dbPath)

	oldExit := osExit
	defer func() { osExit = oldExit }()

	var exitedCode int
	osExit = func(code int) {
		exitedCode = code
		panic("exited")
	}

	runWithExitCheck := func(args ...string) (code int) {
		exitedCode = 0
		defer func() {
			if r := recover(); r != nil {
				t.Logf("recovered panic: %v", r)
				code = exitedCode
			}
		}()
		buf := new(bytes.Buffer)
		RootCmd.SetOut(buf)
		RootCmd.SetErr(buf)
		RootCmd.SetArgs(args)
		err := RootCmd.Execute()
		if err != nil {
			t.Logf("execute error: %v", err)
		}
		t.Logf("output: %q", buf.String())
		return exitedCode
	}

	if code := runWithExitCheck("get", "nonexistent.localhost"); code != 1 {
		t.Errorf("expected exit code 1 for nonexistent get, got: %d", code)
	}

	if code := runWithExitCheck("remove", "nonexistent.localhost"); code != 1 {
		t.Errorf("expected exit code 1 for nonexistent remove, got: %d", code)
	}

	if code := runWithExitCheck("add", "http://invalid.localhost", "localhost:3000"); code != 1 {
		t.Errorf("expected exit code 1 for invalid add, got: %d", code)
	}

	if code := runWithExitCheck("import", "json", "nonexistent.json"); code != 1 {
		t.Errorf("expected exit code 1 for nonexistent import file, got: %d", code)
	}

	if code := runWithExitCheck("export", "invalid_format"); code != 1 {
		t.Errorf("expected exit code 1 for invalid export format, got: %d", code)
	}

	if code := runWithExitCheck("start", "--admin-port", "80"); code != 1 {
		t.Errorf("expected exit code 1 for start command with invalid admin port, got: %d", code)
	}

	invalidJsonFile := "test_invalid.json"
	_ = os.WriteFile(invalidJsonFile, []byte("{invalid-json}"), 0644)
	defer os.Remove(invalidJsonFile)

	if code := runWithExitCheck("import", "json", invalidJsonFile); code != 1 {
		t.Errorf("expected exit code 1 for invalid json import, got: %d", code)
	}

	if code := runWithExitCheck("import", "backup", invalidJsonFile); code != 1 {
		t.Errorf("expected exit code 1 for invalid backup import, got: %d", code)
	}

	if code := runWithExitCheck("import", "invalid_format", invalidJsonFile); code != 1 {
		t.Errorf("expected exit code 1 for invalid import format, got: %d", code)
	}

	if code := runWithExitCheck("export", "json", "/nonexistent-dir/file.json"); code != 1 {
		t.Errorf("expected exit code 1 for invalid export file path, got: %d", code)
	}

	originalDbPath := dbPath
	dbPath = "/nonexistent-dir/veyra.db"
	if code := runWithExitCheck("list"); code != 1 {
		t.Errorf("expected exit code 1 when db path is invalid, got: %d", code)
	}
	dbPath = originalDbPath

	s, err := store.New(dbPath)
	if err == nil {
		_ = s.Save(model.Rule{PublicPattern: "http://invalid.localhost", InternalTarget: "localhost:3000"})
		_ = s.Close()
	}

	if code := runWithExitCheck("validate"); code != 1 {
		t.Errorf("expected exit code 1 for invalid rule validation, got: %d", code)
	}
}

func TestExecuteError(t *testing.T) {
	RootCmd.SetArgs(nil)
	oldExit := osExit
	defer func() { osExit = oldExit }()

	var exitedCode int
	osExit = func(code int) {
		exitedCode = code
		panic("exited")
	}

	defer func() {
		_ = recover()
		if exitedCode != 1 {
			t.Errorf("expected exit code 1 for RootCmd.Execute error, got: %d", exitedCode)
		}
	}()

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"veyra", "nonexistent-command"}

	Execute()
}

func TestCLIVerboseAndHTTPS(t *testing.T) {
	dbPath = "test_verbose_https.db"
	defer os.Remove(dbPath)

	var buf *bytes.Buffer

	executeCommand := func(args ...string) (string, error) {
		verbose = false
		httpsEnabled = false
		buf = new(bytes.Buffer)
		RootCmd.SetOut(buf)
		RootCmd.SetErr(buf)
		RootCmd.SetArgs(args)
		err := RootCmd.Execute()
		return buf.String(), err
	}

	_, err := executeCommand("add", "domain1.test", "localhost:8000")
	if err != nil {
		t.Fatalf("failed to add: %v", err)
	}

	caddyfile, err := executeCommand("export", "caddy")
	if err != nil {
		t.Fatalf("failed to export: %v", err)
	}
	if strings.Contains(caddyfile, "https_port") || strings.Contains(caddyfile, "https://domain1.test") {
		t.Errorf("expected HTTPS to be disabled by default, got: %s", caddyfile)
	}

	_ = caddy.Stop()

	oldExit := osExit
	defer func() { osExit = oldExit }()

	var exitedCode int
	osExit = func(code int) {
		exitedCode = code
		panic("exited")
	}

	runReloadWithPanic := func(args ...string) (string, int) {
		exitedCode = 0
		func() {
			defer func() {
				if r := recover(); r != nil {
				}
			}()
			_, _ = executeCommand(args...)
		}()
		return buf.String(), exitedCode
	}

	outReloadSilent, codeSilent := runReloadWithPanic("reload")
	if codeSilent != 1 {
		t.Errorf("expected exit code 1 on reload failure, got: %d", codeSilent)
	}
	if !strings.Contains(outReloadSilent, "Error: Could not reload Caddy configuration (is 'veyra start' running?)") {
		t.Errorf("expected reload silent output to contain clean message, got: %q", outReloadSilent)
	}
	if strings.Contains(outReloadSilent, "connection refused") {
		t.Errorf("expected reload silent output to NOT contain raw dial error, got: %q", outReloadSilent)
	}

	outReloadVerbose, codeVerbose := runReloadWithPanic("reload", "--verbose")
	if codeVerbose != 1 {
		t.Errorf("expected exit code 1 on reload failure with verbose, got: %d", codeVerbose)
	}
	if !strings.Contains(outReloadVerbose, "Error reloading Caddy:") {
		t.Errorf("expected reload verbose output to contain Error reloading Caddy, got: %q", outReloadVerbose)
	}

	outAddSilent, err := executeCommand("add", "domain2.test", "localhost:8000")
	if err != nil {
		t.Fatalf("add command failed: %v", err)
	}
	if strings.Contains(outAddSilent, "Warning: Could not reload Caddy configuration") {
		t.Errorf("expected add silent reload to suppress warning, got: %q", outAddSilent)
	}

	outAddVerbose, err := executeCommand("add", "domain3.test", "localhost:8000", "--verbose")
	if err != nil {
		t.Fatalf("add command failed with verbose: %v", err)
	}
	if !strings.Contains(outAddVerbose, "Warning: Could not reload Caddy configuration") {
		t.Errorf("expected add verbose reload to show warning, got: %q", outAddVerbose)
	}

	outRemoveSilent, err := executeCommand("remove", "domain2.test")
	if err != nil {
		t.Fatalf("remove command failed: %v", err)
	}
	if strings.Contains(outRemoveSilent, "Warning: Could not reload Caddy configuration") {
		t.Errorf("expected remove silent reload to suppress warning, got: %q", outRemoveSilent)
	}

	outRemoveVerbose, err := executeCommand("remove", "domain3.test", "--verbose")
	if err != nil {
		t.Fatalf("remove command failed with verbose: %v", err)
	}
	if !strings.Contains(outRemoveVerbose, "Warning: Could not reload Caddy configuration") {
		t.Errorf("expected remove verbose reload to show warning, got: %q", outRemoveVerbose)
	}

	go func() {
		bufSub := new(bytes.Buffer)
		RootCmd.SetOut(bufSub)
		RootCmd.SetErr(bufSub)
		RootCmd.SetArgs([]string{"start", "--http-port", "8084", "--https-port", "8447", "--admin-port", "2023", "--https"})
		_ = RootCmd.Execute()
	}()
	time.Sleep(500 * time.Millisecond)
	_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	time.Sleep(200 * time.Millisecond)

	reg, cleanup, err := InitRegistry()
	if err != nil {
		t.Fatalf("failed to init registry: %v", err)
	}
	enabled, err := reg.IsHTTPSEnabled()
	cleanup()
	if err != nil {
		t.Fatalf("failed to check HTTPS: %v", err)
	}
	if !enabled {
		t.Errorf("expected HTTPS to be enabled after start --https")
	}
}
