package machine

// Dict interns symbols, so that equal symbols are the same Element.
type Dict struct {
	ascii      [256]Element
	characters map[rune]Element
	symbols    map[string]Element
	integers   map[int]Element
}

func NewDict() *Dict {
	return &Dict{
		characters: make(map[rune]Element),
		symbols:    make(map[string]Element),
		integers:   make(map[int]Element),
	}
}

func (d *Dict) GetByString(x string) Element {
	if val, exists := d.symbols[x]; exists {
		return val
	}
	return nil
}

func (d *Dict) UniqueR(x rune) Element {
	if int(x) < len(d.ascii) {
		if y := d.ascii[x]; y != nil {
			return y
		}
		d.ascii[x] = NewChr(x)
		return d.ascii[x]
	}
	if val, exists := d.characters[x]; exists {
		return val
	}
	d.characters[x] = NewChr(x)
	return d.characters[x]
}

func (d *Dict) UniqueE(x Element) Element {
	c := x.ToString()
	if val, exists := d.symbols[c]; exists {
		return val
	}
	d.symbols[c] = x
	return x
}

// Predef holds the predefined symbols that the engine compares by identity.
type Predef struct {
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

func NewPredef() *Predef {
	return &Predef{}
}

// defineSymbols creates the predefined symbols. Loaded rules hold these
// objects and the engine compares them by identity, so a second load (-add)
// must not replace them.
func (e *Engine) defineSymbols() {
	if e.symbolsDefined {
		return
	}
	e.symbolsDefined = true
	e.nonTerminalSymbols.UniqueE(Null())
	e.predefinedSymbols.zlm = e.varSymbols.UniqueE(Null())

	e.nonTerminalSymbols.UniqueE(NewSym("start"))
	e.nonTerminalSymbols.UniqueE(NewSym("eof"))
	e.nonTerminalSymbols.UniqueE(NewSpSym("sp"))
	e.nonTerminalSymbols.UniqueE(NewNlSym("nl"))
	e.nonTerminalSymbols.UniqueE(NewRepnSym("repeatN"))
	e.nonTerminalSymbols.UniqueE(NewAnything("anything"))
	e.nonTerminalSymbols.UniqueE(NewAnySym("nonTerminal"))
	e.nonTerminalSymbols.UniqueE(NewAnyChr("terminal"))
	e.nonTerminalSymbols.UniqueE(NewUriSym(e, "uri"))
	e.nonTerminalSymbols.UniqueE(NewUrdSym(e, "urd"))
	e.nonTerminalSymbols.UniqueE(NewOutSym(e, "out"))
	e.nonTerminalSymbols.UniqueE(NewErrSym(e, "err"))
	e.nonTerminalSymbols.UniqueE(NewLnoSym("lineNo"))
	e.nonTerminalSymbols.UniqueE(NewIfnSym("fileName"))
	e.nonTerminalSymbols.UniqueE(NewFlagSym("flagError"))
	e.nonTerminalSymbols.UniqueE(NewWarnSym("warnError"))

	e.functionSymbols.UniqueE(NewSym("mark"))

	e.predefinedSymbols.appendFn = e.functionSymbols.UniqueE(NewAppendXSym("append"))
	e.nonTerminalSymbols.UniqueE(NewRepSym("repeat"))
	e.nonTerminalSymbols.UniqueE(NewOptSym("option"))

	e.predefinedSymbols.nil = NewZzz("-")
	e.predefinedSymbols.getFn = NewGetF("g")
	e.predefinedSymbols.strFn = NewStrF("s")
	e.predefinedSymbols.actFn = NewActF("a")
	e.predefinedSymbols.bindFn = NewBindF(":")
	e.predefinedSymbols.takeFn = NewTakeF("%")
	e.predefinedSymbols.start = e.nonTerminalSymbols.GetByString("start")
	e.predefinedSymbols.eof = e.nonTerminalSymbols.GetByString("eof")
	e.predefinedSymbols.put = e.nonTerminalSymbols.GetByString("out")
	e.predefinedSymbols.mark = e.functionSymbols.GetByString("mark")

	e.predefinedSymbols.dropFn = e.functionSymbols.UniqueE(NewDropF("drop"))
	e.predefinedSymbols.doneFn = e.functionSymbols.UniqueE(NewDoneF("done"))
	e.predefinedSymbols.injFn = e.functionSymbols.UniqueE(NewInjF("inj"))

	e.nonTerminalSymbols.UniqueE(NewTrueSym("true"))
	e.nonTerminalSymbols.UniqueE(NewFalseSym("false"))
	e.functionSymbols.UniqueE(NewTrueF("true"))
	e.functionSymbols.UniqueE(NewFalseF("false"))
	e.functionSymbols.UniqueE(NewApplyF("apply"))

	for _, c := range converters {
		e.nonTerminalSymbols.UniqueE(NewIOSymbol(c.name, newConverter(e, c.convert)))
	}

	e.functionSymbols.UniqueE(NewTestf("test"))
	e.functionSymbols.UniqueE(NewIff("if"))
	e.functionSymbols.UniqueE(NewLoopf("loop"))
	e.functionSymbols.UniqueE(NewForf("for"))
	e.functionSymbols.UniqueE(NewBreakf("break"))
	e.functionSymbols.UniqueE(NewContinuef("continue"))
	e.functionSymbols.UniqueE(NewRulef("rule"))
	e.functionSymbols.UniqueE(NewForeachf("foreach"))
	e.functionSymbols.UniqueE(NewRetf("ret"))
	e.functionSymbols.UniqueE(NewLamdaf("lamda"))
	e.functionSymbols.UniqueE(NewSpecf("spec"))
	e.functionSymbols.UniqueE(NewArgsf("args"))
	e.functionSymbols.UniqueE(NewCellf("cell"))
	e.functionSymbols.UniqueE(NewArrayf("array"))
	e.functionSymbols.UniqueE(NewFunf("fun"))
	e.functionSymbols.UniqueE(NewSelF("sel"))
	e.functionSymbols.UniqueE(NewOrOrf("orOr"))
	e.functionSymbols.UniqueE(NewOrOrf("||"))
	e.functionSymbols.UniqueE(NewAndAndf("andAnd"))
	e.functionSymbols.UniqueE(NewAndAndf("&&"))
	for _, o := range operators {
		for _, name := range o.names {
			e.functionSymbols.UniqueE(newOperator(name, o.kind, o.unary, o.binary))
		}
	}
}
