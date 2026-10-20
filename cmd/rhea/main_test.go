package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The binary builds and answers: no arguments is the usage and a failure,
// "rhea run" lists the markets' commands. A smoke test, so the command
// package is not the one package with no test at all.
func TestBinarySmoke(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "rhea")
	build := exec.Command("go", "build", "-o", bin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	out, err := exec.Command(bin).CombinedOutput()
	if err == nil || !strings.Contains(string(out), "usage: rhea") {
		t.Fatalf("no arguments: %v\n%s", err, out)
	}
	out, err = exec.Command(bin, "run").CombinedOutput()
	if err != nil || !strings.Contains(string(out), "pibr-exam") {
		t.Fatalf("rhea run: %v\n%s", err, out)
	}
	if _, err := os.Stat(bin); err != nil {
		t.Fatal(err)
	}
}
