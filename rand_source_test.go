package dicescript

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

type sequenceDiceSource struct {
	values []uint64
	index  int
}

func (s *sequenceDiceSource) Uint64() uint64 {
	v := s.values[s.index%len(s.values)]
	s.index++
	return v
}

func TestRollAcceptsDiceSource(t *testing.T) {
	src := &sequenceDiceSource{values: []uint64{4}}

	ret := Roll(src, 6, 0)

	assert.Equal(t, IntType(5), ret)
}

func TestGetCurSeedReportsUnsupportedForStatelessSource(t *testing.T) {
	ctx := NewVM()
	ctx.RandSrc = &sequenceDiceSource{values: []uint64{4}}

	seed, err := ctx.GetCurSeed()

	assert.Nil(t, seed)
	assert.ErrorIs(t, err, ErrDiceSourceStateUnsupported)
}

func TestPCGDiceSourceStateRoundTrip(t *testing.T) {
	src := NewPCGDiceSource(123)
	_ = Roll(src, 100, 0)
	seed, err := src.MarshalBinary()
	assert.NoError(t, err)

	restored, err := NewPCGDiceSourceFromState(seed)
	assert.NoError(t, err)

	assert.Equal(t, Roll(src, 100, 0), Roll(restored, 100, 0))
}

func TestCryptoDiceSourceUsesReaderAndIsStateless(t *testing.T) {
	src := &cryptoDiceSource{reader: bytes.NewReader([]byte{0, 0, 0, 0, 0, 0, 0, 4})}

	ret := Roll(src, 6, 0)
	_, ok := any(src).(StatefulDiceSource)

	assert.Equal(t, IntType(5), ret)
	assert.False(t, ok)
}
