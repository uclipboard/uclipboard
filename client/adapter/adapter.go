package adapter

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/uclipboard/uclipboard/model"
	"github.com/uclipboard/uclipboard/model/nanos"
)

type ClipboardCmdAdapter interface {
	Copy(s string) error
	Paste() (string, error)
	Watch(onChange func(string)) error
}

// warning: the string is split by space
// so the command can't parse the space in the command
// for example: `echo "hello world"` will be split to
// ["echo", "\"hello", "world\""]
// if you want to use the space in the command,
// please use the `neighborExecSplit` function
func neighborExec(cmd string) *exec.Cmd {
	parts := strings.Fields(cmd)

	if len(parts) == 0 {
		panic("adapterExec: cmd is empty")
	}
	return neighborExecSplit(parts)
}

// an advanced command executor,
// which supports execute the command from our program running path
// please make sure the cmd is the program name only('prog') rather than the path('/path/to/prog' or './prog')
func neighborExecSplit(cmd []string) *exec.Cmd {
	cmdName := cmd[0]
	exDir := model.ExDir()
	guessExPath := filepath.Join(exDir, cmdName)
	if _, err := os.Stat(guessExPath); err == nil {
		// this cmd exists in program running path
		cmd[0] = guessExPath
	}
	return exec.Command(cmd[0], cmd[1:]...)

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

// replace the %s with "/path/to/uclipboard --nanos" then neighborExec
// load the content from the command's stdout and parse as nanos message to draw the content
// then call the onChange function with the content
func defaultWatch(cmd string, onChange func(string)) error {
	cmd = fmt.Sprintf(cmd, model.ExPath()+" --nanos")
	execCmd := neighborExec(cmd)
	stdout, err := execCmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := execCmd.Start(); err != nil {
		return err
	}

	go func() {
		logger := model.NewModuleLogger("adapter watch")
		for {
			s, err := nanos.Load(stdout)
			if err != nil {
				return
			}
			content := string(s.Data)
			if len(content) == 0 {
				logger.Debugf("watch: empty content")
				continue
			}
			onChange(content)
		}
	}()

	return execCmd.Wait()
}

var (
	ErrEmptyClipboard           = errors.New("perhaps system clipboard is empty")
	ErrLockedClipboard          = errors.New("perhaps system clipboard is locked so that adapter can't access it")
	ErrClipboardDataTypeUnknown = errors.New("the content type of clipboard is unrecognized")
	ErrUnsupportedWatchMode     = errors.New("the adapter doesn't support watch mode")
)

var factories = make(map[string]func(*model.UContext) ClipboardCmdAdapter)

func RegisterFactory(name string, factory func(*model.UContext) ClipboardCmdAdapter) {
	if _, ok := factories[name]; ok {
		panic("adapter: already registered " + name)
	}
	factories[name] = factory
}

func GetAdapterFactory(name string) func(*model.UContext) ClipboardCmdAdapter {
	factory, ok := factories[name]
	if !ok {
		panic("adapter: unknown adapter " + name)
	}
	return factory
}
