package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadmeCompactCLIExample(t *testing.T) {
	readmeBytes, err := os.ReadFile("../README.md")
	if err != nil {
		t.Fatal(err)
	}
	_, example, found := strings.Cut(string(readmeBytes), "For compact boxes with evenly spaced blocks and a multiword arrow label:")
	if !found {
		t.Fatal("compact example missing")
	}
	_, commandBlock, _ := strings.Cut(example, "```bash\n")
	command, remainder, found := strings.Cut(commandBlock, "\n```")
	if !found {
		t.Fatal("compact command missing")
	}
	_, outputBlock, _ := strings.Cut(remainder, "```text\n")
	expected, _, found := strings.Cut(outputBlock, "\n```")
	if !found {
		t.Fatal("compact output missing")
	}
	directory := t.TempDir()
	binary := filepath.Join(directory, "mermaid-ascii")
	if output, err := exec.Command("go", "build", "-o", binary, "..").CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v\n%s", err, output)
	}
	process := exec.Command("bash", "-c", command)
	process.Dir = directory
	output, err := process.CombinedOutput()
	if err != nil || strings.TrimRight(string(output), "\n") != expected {
		t.Fatalf("documented command: %v\ngot:\n%s\nwant:\n%s", err, output, expected)
	}
}
