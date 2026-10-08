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

	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toStr", NewToStrFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toLstr", NewToLstrFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toUstr", NewToUstrFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toQuote", NewToQuoteFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toSym", NewToSymFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toLsym", NewToLsymFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toUsym", NewToUsymFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toSys", NewToSysFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toLsys", NewToLsysFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toUsys", NewToUsysFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toVar", NewToVarFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toNum", NewToNumFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toOct", NewToOctFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toHex", NewToHexFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toBin", NewToBinFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toUrn", NewToUrNstrFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toUrd", NewToUrDstrFromEngine(e)))

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
	e.functionSymbols.UniqueE(NewIdxf("idx"))
	e.functionSymbols.UniqueE(NewIdtf("idt"))
	e.functionSymbols.UniqueE(NewSelF("sel"))
	e.functionSymbols.UniqueE(NewStoValf("stoVal"))
	e.functionSymbols.UniqueE(NewStoValf("="))
	e.functionSymbols.UniqueE(NewStoAddf("stoAdd"))
	e.functionSymbols.UniqueE(NewStoAddf("+="))
	e.functionSymbols.UniqueE(NewStoSubf("stoSub"))
	e.functionSymbols.UniqueE(NewStoSubf("-="))
	e.functionSymbols.UniqueE(NewStoMulf("stoMul"))
	e.functionSymbols.UniqueE(NewStoMulf("*="))
	e.functionSymbols.UniqueE(NewStoDivf("stoDiv"))
	e.functionSymbols.UniqueE(NewStoDivf("/="))
	e.functionSymbols.UniqueE(NewStoModf("stoMod"))
	e.functionSymbols.UniqueE(NewStoModf("%="))

	e.functionSymbols.UniqueE(NewEqf("eeq"))
	e.functionSymbols.UniqueE(NewEeqf("==="))
	e.functionSymbols.UniqueE(NewNef("nee"))
	e.functionSymbols.UniqueE(NewNeef("!=="))

	e.functionSymbols.UniqueE(NewInf("in"))
	e.functionSymbols.UniqueE(NewEqf("eq"))
	e.functionSymbols.UniqueE(NewEqf("=="))
	e.functionSymbols.UniqueE(NewNef("ne"))
	e.functionSymbols.UniqueE(NewNef("!="))
	e.functionSymbols.UniqueE(NewLtf("lt"))
	e.functionSymbols.UniqueE(NewLtf("<"))
	e.functionSymbols.UniqueE(NewGtf("gt"))
	e.functionSymbols.UniqueE(NewGtf(">"))
	e.functionSymbols.UniqueE(NewLef("le"))
	e.functionSymbols.UniqueE(NewLef("<="))
	e.functionSymbols.UniqueE(NewGef("ge"))
	e.functionSymbols.UniqueE(NewGef(">="))
	e.functionSymbols.UniqueE(NewOrOrf("orOr"))
	e.functionSymbols.UniqueE(NewOrOrf("||"))
	e.functionSymbols.UniqueE(NewAndAndf("andAnd"))
	e.functionSymbols.UniqueE(NewAndAndf("&&"))
	e.functionSymbols.UniqueE(NewBitOrf("bitOr"))
	e.functionSymbols.UniqueE(NewBitOrf("|"))
	e.functionSymbols.UniqueE(NewBitXorf("bitXor"))
	e.functionSymbols.UniqueE(NewBitXorf("^"))
	e.functionSymbols.UniqueE(NewBitAndf("bitAnd"))
	e.functionSymbols.UniqueE(NewBitAndf("&"))
	e.functionSymbols.UniqueE(NewAddf("add"))
	e.functionSymbols.UniqueE(NewAddf("+"))
	e.functionSymbols.UniqueE(NewSubf("sub"))
	e.functionSymbols.UniqueE(NewSubf("-"))
	e.functionSymbols.UniqueE(NewMulf("mul"))
	e.functionSymbols.UniqueE(NewMulf("*"))
	e.functionSymbols.UniqueE(NewDivf("div"))
	e.functionSymbols.UniqueE(NewDivf("/"))
	e.functionSymbols.UniqueE(NewModf("mod"))
	e.functionSymbols.UniqueE(NewModf("%"))
	e.functionSymbols.UniqueE(NewPreincf("preinc"))
	e.functionSymbols.UniqueE(NewPredecf("predec"))
	e.functionSymbols.UniqueE(NewPostincf("postinc"))
	e.functionSymbols.UniqueE(NewPostdecf("postdec"))
	e.functionSymbols.UniqueE(NewInvf("inv"))
	e.functionSymbols.UniqueE(NewNotf("not"))
	e.functionSymbols.UniqueE(NewNegf("neg"))
}
