package dicescript

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newCustomDiceParserTestState(input string, items ...*customDiceItem) (*parser, *ParserCustomData) {
	ctx := NewVM()
	ctx.CustomDiceInfo = items
	data := &ParserCustomData{
		ParserData: ParserData{code: make([]ByteCode, 8)},
		ctx:        ctx,
	}
	return newParser("", []byte(input)), data
}

func TestCustomDiceParserSkipsInvalidCandidates(t *testing.T) {
	noResult := &customDiceItem{parser: func(*Context, *CustomDiceStream) (*CustomDiceParseResult, error) {
		return nil, nil
	}}
	noConsumption := &customDiceItem{parser: func(*Context, *CustomDiceStream) (*CustomDiceParseResult, error) {
		return &CustomDiceParseResult{Matched: true}, nil
	}}
	matchInput := &customDiceItem{parser: func(_ *Context, stream *CustomDiceStream) (*CustomDiceParseResult, error) {
		_, _ = stream.Read()
		return &CustomDiceParseResult{Matched: true, Payload: "payload"}, nil
	}}
	p, data := newCustomDiceParserTestState("X", nil, noResult, noConsumption, matchInput)

	match, ok := data.tryMatchCustomDice(p)
	assert.True(t, ok)
	assert.Equal(t, []string{"X"}, match.groups)
	assert.Equal(t, "X", match.text)
	assert.Equal(t, "payload", match.payload)
	assert.Equal(t, 1, match.byteLen)
	assert.Zero(t, data.stream.Consumed(), "matching must not leave the reusable stream advanced")
}

func TestCustomDiceParserFillsEmptyFirstGroup(t *testing.T) {
	item := &customDiceItem{parser: func(_ *Context, stream *CustomDiceStream) (*CustomDiceParseResult, error) {
		_, _ = stream.Read()
		return &CustomDiceParseResult{Matched: true, Groups: []string{"", "capture"}}, nil
	}}
	p, data := newCustomDiceParserTestState("Z", item)

	match, ok := data.tryMatchCustomDice(p)
	assert.True(t, ok)
	assert.Equal(t, []string{"Z", "capture"}, match.groups)
}

func TestCustomDiceRegexOptionalGroupAndZeroLengthFallback(t *testing.T) {
	zeroLength := &customDiceItem{re: regexp.MustCompile(`^`)}
	optionalGroup := &customDiceItem{re: regexp.MustCompile(`^X(Y)?`)}
	p, data := newCustomDiceParserTestState("X", &customDiceItem{}, zeroLength, optionalGroup)

	match, ok := data.tryMatchCustomDice(p)
	assert.True(t, ok)
	assert.Equal(t, []string{"X", ""}, match.groups)
	assert.Equal(t, "X", match.text)

	p.pt.offset = len(p.data)
	match, ok = data.tryMatchCustomDice(p)
	assert.False(t, ok)
	assert.Nil(t, match)
}

func TestCustomDicePendingMatchLifecycle(t *testing.T) {
	p, data := newCustomDiceParserTestState("X")
	assert.Nil(t, data.ConsumeCustomDice(p))
	assert.Nil(t, data.CommitCustomDice())

	data.pendingCustomDice = &customDiceMatch{startOffset: 0}
	assert.Nil(t, data.ConsumeCustomDice(p))
	assert.Nil(t, data.pendingCustomDice)

	groups := []string{"X", "capture"}
	item := &customDiceItem{}
	data.pendingCustomDice = &customDiceMatch{
		item:    item,
		groups:  groups,
		payload: "payload",
	}
	assert.Nil(t, data.CommitCustomDice())
	assert.Equal(t, 1, data.codeIndex)
	assert.Equal(t, typeCustomDice, data.code[0].T)
	compiled := data.code[0].Value.(*customDiceCompiled)
	assert.Same(t, item, compiled.item)
	assert.Equal(t, "X", compiled.text)
	assert.Equal(t, "payload", compiled.payload)
	assert.Equal(t, groups, compiled.groups)
	groups[0] = "changed"
	assert.Equal(t, "X", compiled.groups[0], "committed groups must be an immutable snapshot")

	assert.Nil(t, cloneStrings(nil))
	assert.Nil(t, (*ParserCustomData)(nil).ensurePendingCustomDice(p))
	match, ok := (*ParserCustomData)(nil).tryMatchCustomDice(p)
	assert.False(t, ok)
	assert.Nil(t, match)
}

func TestCustomDiceEnsurePendingRefreshesAtCurrentOffset(t *testing.T) {
	item := &customDiceItem{re: regexp.MustCompile(`^X`)}
	p, data := newCustomDiceParserTestState("X", item)
	data.pendingCustomDice = &customDiceMatch{startOffset: 1}

	match := data.ensurePendingCustomDice(p)
	assert.NotNil(t, match)
	assert.Same(t, match, data.pendingCustomDice)

	p.data = []byte("Z")
	data.pendingCustomDice = &customDiceMatch{startOffset: 1}
	assert.Nil(t, data.ensurePendingCustomDice(p))
	assert.Nil(t, data.pendingCustomDice)
}
