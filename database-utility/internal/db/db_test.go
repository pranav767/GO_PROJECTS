package db

import (
	"compress/gzip"
	"io"
	"os"
	"os/exec"
	"testing"
)

func TestStreamCmdToGzip(t *testing.T) {
	outPath := "test-output.gz"
	// command that prints a predictable string
	cmd := exec.Command("sh", "-c", "printf 'hello-stream'")
	if err := streamCmdToGzip(cmd, outPath); err != nil {
		t.Fatalf("streamCmdToGzip failed: %v", err)
	}
	defer os.Remove(outPath)

	f, err := os.Open(outPath)
	if err != nil {
		t.Fatalf("open gz: %v", err)
	}
	defer f.Close()
	gr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatalf("gzip reader: %v", err)
	}
	defer gr.Close()
	data, err := io.ReadAll(gr)
	if err != nil {
		t.Fatalf("read gzip: %v", err)
	}
	if string(data) != "hello-stream" {
		t.Fatalf("unexpected content: %s", string(data))
	}
}
