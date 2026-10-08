package machine

// Dict interns symbols, so that equal symbols are the same Element.
type dict struct {
	ascii      [256]Element
	characters map[rune]Element
	symbols    map[string]Element
	integers   map[int]Element
}

func newDict() *dict {
	return &dict{
		characters: make(map[rune]Element),
		symbols:    make(map[string]Element),
		integers:   make(map[int]Element),
	}
}

func (d *dict) getByString(x string) Element {
	if val, exists := d.symbols[x]; exists {
		return val
	}
	return nil
}

func (d *dict) uniqueR(x rune) Element {
	if int(x) < len(d.ascii) {
		if y := d.ascii[x]; y != nil {
			return y
		}
		d.ascii[x] = newChr(x)
		return d.ascii[x]
	}
	if val, exists := d.characters[x]; exists {
		return val
	}
	d.characters[x] = newChr(x)
	return d.characters[x]
}

func (d *dict) uniqueE(x Element) Element {
	c := x.ToString()
	if val, exists := d.symbols[c]; exists {
		return val
	}
	d.symbols[c] = x
	return x
}

// Predef holds the predefined symbols that the engine compares by identity.
type predef struct {
	start    Element
	eof      Element
	nil      Element
	zlm      Element
	put      Element
	mark     Element
	dropFn   Element
	getFn    Element
	strFn    Element
	actFn    Element
	bindFn   Element
	takeFn   Element
	doneFn   Element
	injFn    Element
	appendFn Element
}

func newPredef() *predef {
	return &predef{}
}

// defineSymbols creates the predefined symbols. Loaded rules hold these
// objects and the engine compares them by identity, so a second load (-add)
// must not replace them.
func (e *Engine) defineSymbols() {
	if e.symbolsDefined {
		return
	}
	e.symbolsDefined = true
	e.nonTerminalSymbols.uniqueE(Null())
	e.predefinedSymbols.zlm = e.varSymbols.uniqueE(Null())

	e.nonTerminalSymbols.uniqueE(newSym("start"))
	e.nonTerminalSymbols.uniqueE(newSym("eof"))
	e.nonTerminalSymbols.uniqueE(newSpSym("sp"))
	e.nonTerminalSymbols.uniqueE(newNlSym("nl"))
	e.nonTerminalSymbols.uniqueE(newRepnSym("repeatN"))
	e.nonTerminalSymbols.uniqueE(newAnything("anything"))
	e.nonTerminalSymbols.uniqueE(newAnySym("nonTerminal"))
	e.nonTerminalSymbols.uniqueE(newAnyChr("terminal"))
	e.nonTerminalSymbols.uniqueE(newUriSym(e, "uri"))
	e.nonTerminalSymbols.uniqueE(newUrdSym(e, "urd"))
	e.nonTerminalSymbols.uniqueE(newOutSym(e, "out"))
	e.nonTerminalSymbols.uniqueE(newErrSym(e, "err"))
	e.nonTerminalSymbols.uniqueE(newLnoSym("lineNo"))
	e.nonTerminalSymbols.uniqueE(newIfnSym("fileName"))
	e.nonTerminalSymbols.uniqueE(newFlagSym("flagError"))
	e.nonTerminalSymbols.uniqueE(newWarnSym("warnError"))

	e.functionSymbols.uniqueE(newSym("mark"))

	e.predefinedSymbols.appendFn = e.functionSymbols.uniqueE(newAppendXSym("append"))
	e.nonTerminalSymbols.uniqueE(newRepSym("repeat"))
	e.nonTerminalSymbols.uniqueE(newOptSym("option"))

	e.predefinedSymbols.nil = newDontCare("-")
	e.predefinedSymbols.getFn = newGetF("g")
	e.predefinedSymbols.strFn = newStrF("s")
	e.predefinedSymbols.actFn = newActF("a")
	e.predefinedSymbols.bindFn = newBindF(":")
	e.predefinedSymbols.takeFn = newTakeF("%")
	e.predefinedSymbols.start = e.nonTerminalSymbols.getByString("start")
	e.predefinedSymbols.eof = e.nonTerminalSymbols.getByString("eof")
	e.predefinedSymbols.put = e.nonTerminalSymbols.getByString("out")
	e.predefinedSymbols.mark = e.functionSymbols.getByString("mark")

	e.predefinedSymbols.dropFn = e.functionSymbols.uniqueE(newDropF("drop"))
	e.predefinedSymbols.doneFn = e.functionSymbols.uniqueE(newDoneF("done"))
	e.predefinedSymbols.injFn = e.functionSymbols.uniqueE(newInjF("inj"))

	e.nonTerminalSymbols.uniqueE(newTrueSym("true"))
	e.nonTerminalSymbols.uniqueE(newFalseSym("false"))
	e.functionSymbols.uniqueE(newTrueF("true"))
	e.functionSymbols.uniqueE(newFalseF("false"))
	e.functionSymbols.uniqueE(newApplyF("apply"))

	for _, c := range converters {
		e.nonTerminalSymbols.uniqueE(newIOSymbol(c.name, newConverter(e, c.convert)))
	}

	e.functionSymbols.uniqueE(newTestf("test"))
	e.functionSymbols.uniqueE(newIff("if"))
	e.functionSymbols.uniqueE(newLoopf("loop"))
	e.functionSymbols.uniqueE(newForf("for"))
	e.functionSymbols.uniqueE(newBreakf("break"))
	e.functionSymbols.uniqueE(newContinuef("continue"))
	e.functionSymbols.uniqueE(newRulef("rule"))
	e.functionSymbols.uniqueE(newForeachf("foreach"))
	e.functionSymbols.uniqueE(newRetf("ret"))
	e.functionSymbols.uniqueE(newLamdaf("lamda"))
	e.functionSymbols.uniqueE(newSpecf("spec"))
	e.functionSymbols.uniqueE(newArgsf("args"))
	e.functionSymbols.uniqueE(newCellf("cell"))
	e.functionSymbols.uniqueE(newArrayf("array"))
	e.functionSymbols.uniqueE(newFunf("fun"))
	e.functionSymbols.uniqueE(newSelF("sel"))
	e.functionSymbols.uniqueE(newOrOrf("orOr"))
	e.functionSymbols.uniqueE(newOrOrf("||"))
	e.functionSymbols.uniqueE(newAndAndf("andAnd"))
	e.functionSymbols.uniqueE(newAndAndf("&&"))
	for _, o := range operators {
		for _, name := range o.names {
			e.functionSymbols.uniqueE(newOperator(name, o.kind, o.unary, o.binary))
		}
	}
}
