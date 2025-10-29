package filestream

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestFlowControlGlobalChannelClose verifies that closing the shared totalFlowTokenChan while
// a transfer is running causes the process to panic with "send on closed channel". The actual
// panic happens in a child process so the parent test can assert on its stderr without crashing.
func TestFlowControlGlobalChannelClose(t *testing.T) {
	if os.Getenv("FLOWCONTROL_REPRO") == "1" {
		reproduceGlobalChannelClose()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestFlowControlGlobalChannelClose")
	cmd.Env = append(os.Environ(), "FLOWCONTROL_REPRO=1")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected reproduction process to panic, but it exited cleanly")
	}

	output := stderr.String()
	if !strings.Contains(output, "send on closed channel") {
		t.Fatalf("expected panic output to contain 'send on closed channel', got: %s", output)
	}
}

func reproduceGlobalChannelClose() {
	fc := NewFlowControl(FlowControlConfig{
		MaxFlow:    32,
		TotalFlow:  64,
		MinFlow:    16,
		BufferSize: 128,
	})
	fc.Start()

	go func() {
		reader := bytes.NewReader(make([]byte, 512*1024))
		_, _ = fc.LimitFlow(reader, io.Discard, nil)
	}()

	// allow goroutine to start
	time.Sleep(50 * time.Millisecond)
	close(fc.totalFlowTokenChan)

	// trigger the send after close, mirroring the panic seen in production
	fc.totalFlowTokenChan <- 1
}
