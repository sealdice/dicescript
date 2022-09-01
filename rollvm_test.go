package dicescript

import "testing"

func TestSimpleRun(t *testing.T) {
	vm := NewVM()
	vm.Run("1+1")
	if !valueEqual(vm.Ret, VMValueNewInt64(2)) {
		t.Errorf("VM Error")
	}
}

func TestUnsupportedOperandType(t *testing.T) {
}
