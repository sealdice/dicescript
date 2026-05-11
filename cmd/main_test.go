package main

import (
	"path/filepath"
	"testing"

	ds "github.com/sealdice/dicescript"
	"github.com/stretchr/testify/assert"
)

func TestRunScriptFile(t *testing.T) {
	scriptPath := filepath.Join("testdata", "init.ds")

	vm := ds.NewVM()
	err := runScriptFile(vm, scriptPath)
	assert.NoError(t, err)

	err = vm.Run("add(base, 1)")
	assert.NoError(t, err)
	assert.Equal(t, "42", vm.Ret.ToString())
}
