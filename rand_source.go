package dicescript

import (
	cryptorand "crypto/rand"
	"encoding"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"reflect"
	"time"

	exprand "golang.org/x/exp/rand"
)

// DiceSource is the minimum capability required by dice rolling.
type DiceSource interface {
	Uint64() uint64
}

// StatefulDiceSource is a dice source whose current state can be saved and restored.
type StatefulDiceSource interface {
	DiceSource
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

var ErrDiceSourceStateUnsupported = errors.New("random source does not support binary state")

var randSource = NewPCGDiceSource(uint64(time.Now().UnixMilli()))

func NewPCGDiceSource(seed uint64) StatefulDiceSource {
	src := &exprand.PCGSource{}
	src.Seed(seed)
	return src
}

func NewPCGDiceSourceFromState(state []byte) (StatefulDiceSource, error) {
	src := &exprand.PCGSource{}
	err := src.UnmarshalBinary(state)
	return src, err
}

func NewCryptoDiceSource() DiceSource {
	return &cryptoDiceSource{reader: cryptorand.Reader}
}

type cryptoDiceSource struct {
	reader io.Reader
}

func (src *cryptoDiceSource) Uint64() uint64 {
	reader := src.reader
	if reader == nil {
		reader = cryptorand.Reader
	}

	var data [8]byte
	if _, err := io.ReadFull(reader, data[:]); err != nil {
		panic(fmt.Errorf("crypto dice source failed: %w", err))
	}
	return binary.BigEndian.Uint64(data[:])
}

func normalizeDiceSource(src DiceSource) DiceSource {
	if isNilDiceSource(src) {
		return randSource
	}
	return src
}

func isNilDiceSource(src DiceSource) bool {
	if src == nil {
		return true
	}

	v := reflect.ValueOf(src)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}
