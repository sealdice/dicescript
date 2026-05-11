package main

import (
	"os"
	"path/filepath"
	"testing"

	ds "github.com/sealdice/dicescript"
	"github.com/stretchr/testify/assert"
)

func TestRunScriptFile(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "init.ds")
	content := append([]byte{0xEF, 0xBB, 0xBF}, []byte("func add(a, b) {\n    return a + b;\n}\nbase = 41\n")...)

	err := os.WriteFile(scriptPath, content, 0o600)
	assert.NoError(t, err)

	vm := ds.NewVM()
	err = runScriptFile(vm, scriptPath)
	assert.NoError(t, err)

	err = vm.Run("add(base, 1)")
	assert.NoError(t, err)
	assert.Equal(t, "42", vm.Ret.ToString())
}
