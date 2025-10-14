package dicescript

// customDiceMatch captures a regex match during parsing.
type customDiceMatch struct {
	item        *customDiceItem
	groups      []string
	text        string
	byteLen     int
	startOffset int
	payload     any
}

type customDiceCompiled struct {
	item    *customDiceItem
	groups  []string
	text    string
	payload any
}

func (d *ParserCustomData) PrepareCustomDice(p *parser) bool {
	if d == nil || d.ctx == nil || len(d.ctx.CustomDiceInfo) == 0 {
		d.pendingCustomDice = nil
		return false
	}

	match, ok := d.tryMatchCustomDice(p)
	if !ok {
		d.pendingCustomDice = nil
		return false
	}

	d.pendingCustomDice = match
	return true
}

func (d *ParserCustomData) ConsumeCustomDice(p *parser) any {
	match := d.ensurePendingCustomDice(p)
	if match == nil {
		return nil
	}

	if match.byteLen <= 0 {
		// nothing matched; prevent infinite loop by clearing pending state
		d.pendingCustomDice = nil
		return nil
	}

	targetOffset := match.startOffset + match.byteLen
	for p.pt.offset < targetOffset {
		p.read()
	}

	return nil
}

func (d *ParserCustomData) CommitCustomDice() any {
	match := d.pendingCustomDice
	d.pendingCustomDice = nil
	if match == nil {
		return nil
	}

	compiled := &customDiceCompiled{
		item:    match.item,
		groups:  cloneStrings(match.groups),
		text:    match.text,
		payload: match.payload,
	}
	if compiled.text == "" && len(compiled.groups) > 0 {
		compiled.text = compiled.groups[0]
	}

	d.WriteCode(typeCustomDice, compiled)
	return nil
}

func (d *ParserCustomData) ensurePendingCustomDice(p *parser) *customDiceMatch {
	if d == nil {
		return nil
	}

	if d.pendingCustomDice != nil && d.pendingCustomDice.startOffset == p.pt.offset {
		return d.pendingCustomDice
	}

	match, ok := d.tryMatchCustomDice(p)
	if !ok {
		d.pendingCustomDice = nil
		return nil
	}

	d.pendingCustomDice = match
	return match
}

func (d *ParserCustomData) tryMatchCustomDice(p *parser) (*customDiceMatch, bool) {
	if d == nil || d.ctx == nil || len(d.ctx.CustomDiceInfo) == 0 {
		return nil, false
	}

	start := p.pt.offset
	if start < 0 || start >= len(p.data) {
		return nil, false
	}

	input := string(p.data[start:])

	for _, item := range d.ctx.CustomDiceInfo {
		if item == nil {
			continue
		}

		if item.parser != nil {
			stream := &d.stream
			stream.init(p.data, start)
			result, err := item.parser(d.ctx, stream)
			if err != nil {
				if d.ctx != nil {
					d.ctx.Error = err
				}
				p.addErr(err)
				return nil, false
			}
			if result == nil || !result.Matched {
				stream.ResetAttempt()
				continue
			}
			consumed := stream.Consumed()
			if consumed <= 0 {
				stream.ResetAttempt()
				continue
			}
			matchedText := stream.Current()
			groups := result.Groups
			if len(groups) == 0 {
				groups = []string{matchedText}
			} else if groups[0] == "" {
				groups[0] = matchedText
			}
			text := result.Display
			if text == "" {
				text = matchedText
			}
			match := &customDiceMatch{
				item:        item,
				groups:      groups,
				text:        text,
				byteLen:     consumed,
				startOffset: start,
				payload:     result.Payload,
			}
			stream.ResetAttempt()
			return match, true
		}

		if item.re == nil {
			continue
		}

		loc := item.re.FindStringSubmatchIndex(input)
		if loc == nil || loc[0] != 0 {
			continue
		}

		end := loc[1]
		if end <= 0 {
			continue
		}

		groupCount := len(loc) / 2
		if groupCount == 0 {
			continue
		}
		groups := make([]string, groupCount)
		for i := 0; i < groupCount; i++ {
			s := loc[2*i]
			e := loc[2*i+1]
			if s < 0 || e < 0 {
				groups[i] = ""
				continue
			}
			groups[i] = input[s:e]
		}

		match := &customDiceMatch{
			item:        item,
			groups:      groups,
			text:        groups[0],
			byteLen:     end,
			startOffset: start,
		}
		return match, true
	}

	return nil, false
}

func cloneStrings(src []string) []string {
	if len(src) == 0 {
		return nil
	}
	dup := make([]string, len(src))
	copy(dup, src)
	return dup
}
