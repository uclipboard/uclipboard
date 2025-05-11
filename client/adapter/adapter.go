package adapter

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/uclipboard/uclipboard/model"
)

type ClipboardCmdAdapter interface {
	Copy(s string) error
	Paste() (string, error)
}

// an advanced command executor,
// which supports execute the command from our program running path
// please make sure the cmd is the program name only('prog') rather than the path('/path/to/prog' or './prog')
func neighborExec(cmd string) *exec.Cmd {
	parts := strings.Fields(cmd)

	if len(parts) == 0 {
		panic("adapterExec: cmd is empty")
	}

	cmdName := parts[0]
	exDir := model.ExDir()
	guessExPath := filepath.Join(exDir, cmdName)
	if _, err := os.Stat(guessExPath); err == nil {
		// this cmd exists in program running path
		parts[0] = guessExPath
	}
	return exec.Command(parts[0], parts[1:]...)
}

func defaultCopy(cmd string) func(s string) error {
	return func(s string) error {
		copyCmd := neighborExec(cmd)
		copyCmd.Stdin = strings.NewReader(s)
		if err := copyCmd.Run(); err != nil {
			return err
		}
		return nil
	}
}

func defaultPaste(cmd string) func() (string, error) {
	return func() (string, error) {
		pasteCmd := neighborExec(cmd)
		out := bytes.NewBuffer(nil)
		pasteCmd.Stdout = out
		err := pasteCmd.Run()
		return out.String(), err
	}
}

var (
	ErrEmptyClipboard           = errors.New("perhaps system clipboard is empty")
	ErrLockedClipboard          = errors.New("perhaps system clipboard is locked so that adapter can't access it")
	ErrClipboardDataTypeUnknown = errors.New("the content type of clipboard is unrecognized")
)
