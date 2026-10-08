package machine

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"slices"
)

const (
	DefaultLexPriority int = 1000      // default value for lexical priority (-lexpri)
	DefaultMaxLength   int = 64 * 1024 // default maximum length of the input buffer (-buffer)

	initialBufferLength = 1024      // the input buffer starts this long and grows to maxLength
	outputBufferSize    = 64 * 1024 // size of the buffered output writer
	defaultDisplayWidth = 80
)

// theZlm is the null element. It is stateless, so all engines share it.
var theZlm = newNullSym("null")

// Null is the null value: what an unset variable holds.
func Null() Element { return theZlm }

// Engine loads grammars and applies them to its inputs.
//
// An Engine must be used by one goroutine at a time, and loaded rules belong
// to the engine that loaded them. Engines share no state, so separate engines
// can run concurrently.
type Engine struct {
	contextsCount int           // count of new contexts used to give each a unique identity
	lhsContext    contextHolder // lhs context stack for mismatch events being resolved

	lhsStream          *Stream // lhs registers
	rhsStream          *Stream // rhs registers
	rsLastMatchElement Element // element resulting from last match

	grammars    *selector // table of grammars selected by symbol
	initGrammar *grammar  // initial grammar
	ruleNumbers int       // one more than the highest rule number

	terminalSymbols    *dict   // terminal symbols
	nonTerminalSymbols *dict   // non-terminal symbols
	varSymbols         *dict   // variables
	userSymbols        *dict   // user symbols - guaranteed not to match system symbols
	functionSymbols    *dict   // primitive operator symbols
	predefinedSymbols  *predef // predefined symbols with special significance

	loader *loader // rule loader

	inputs []Input // stack of input sources, the current one last
	input  Input   // current input

	rhsBuffer *rzBuffer // circular buffer at outermost level of rhs

	flagErrors int // incremented by flagSym
	warnErrors int // incremented by warnSym

	externalSystem *External // external interfaces

	tracer *tracer // trace handler - null if no tracing required

	out    *bufio.Writer // output, traces and diagrams; flushed by Start
	errOut io.Writer     // error output (err)

	display      *diagram // to display trace as diagram
	displayWidth int      // width for diagram display

	maxDepth  int // limit on analysis recursion depth - zero means no limit
	maxRepeat int // limit on repetition at repeat     - zero means no limit
	maxLength int // maximum size of the rhs input buffer

	symbolsDefined bool // predefined symbols exist; they must be created only once
}

func NewEngine() *Engine {
	return newEngineWithMaxLength(DefaultMaxLength)
}

// newEngineWithMaxLength makes an engine whose input buffer grows to at most
// maxLength symbols (-buffer).
func newEngineWithMaxLength(maxLength int) *Engine {
	e := &Engine{
		maxLength:          maxLength,
		displayWidth:       defaultDisplayWidth,
		functionSymbols:    newDict(),
		terminalSymbols:    newDict(),
		nonTerminalSymbols: newDict(),
		varSymbols:         newDict(),
		userSymbols:        newDict(),
		predefinedSymbols:  newPredef(),
		externalSystem:     newExternal(),
		grammars:           newSelector(),
		out:                bufio.NewWriterSize(os.Stdout, outputBufferSize),
		errOut:             os.Stderr,
	}
	e.input = NewStdinInput(e) // replaced by the first input in Start
	e.rhsBuffer = newRZBuffer(make([]Element, initialBufferLength), e.maxLength)
	root := newState(e, nil, nil, nil, e.input, 0, 0, 0, e.contextsCount)
	e.contextsCount++
	e.lhsContext = newRootContext(lhContext, root)
	e.lhsStream = newStream(e, "lh", 0)
	e.rhsStream = newStream(e, "rh", 0)
	e.lhsStream.mode = newLZMode(e.lhsContext, e.lhsStream)
	e.rhsStream.mode = newRZMode(newRootContext(rhContext, root), e.rhsStream)
	return e
}

func (e *Engine) setMachineElements(args []Element) Element {
	if len(args) < 2 {
		return e.predefinedSymbols.zlm
	}
	k := args[1].ToVal()
	g := e.grammars.get(k.ToString())
	if g != nil {
		e.lhsContext.State().grammar = g
	}
	return k
}

func (e *Engine) addRule(v []Element, i int) {
	if i >= e.ruleNumbers {
		e.ruleNumbers = i + 1
	}
	gr := e.grammars.selectGrammar(v[0])
	if e.initGrammar == nil {
		e.initGrammar = gr
		e.lhsContext.State().grammar = e.initGrammar
	}
	gr.defineRule(v, i)
}

// LoadFromString loads rules, replacing any loaded before.
func (e *Engine) LoadFromString(rules string) error {
	return e.LoadFromStringReset(rules, true)
}

// LoadFromStringReset loads rules; with reset false they are added to the
// rules already loaded (-add).
func (e *Engine) LoadFromStringReset(rules string, reset bool) (err error) {
	e.defineSymbols()
	// if the rules do not load, the engine keeps the rules it had
	grammars, saved, initGrammar, ruleNumbers := e.grammars, e.grammars.save(), e.initGrammar, e.ruleNumbers
	defer func() {
		if err != nil {
			grammars.restore(saved)
			e.grammars, e.initGrammar, e.ruleNumbers = grammars, initGrammar, ruleNumbers
			e.lhsContext.State().grammar = initGrammar
		}
	}()
	if reset {
		// the first grammar of the new rules becomes the initial one; the
		// old initial grammar is not in the new table
		e.grammars = newSelector()
		e.initGrammar = nil
	}
	if e.loader == nil {
		e.loader = newLoader(e)
	}
	return e.loader.load(rules)
}

// SetOutput sends output, traces and diagrams to w (default os.Stdout).
func (e *Engine) SetOutput(w io.Writer) {
	_ = e.Flush()
	e.out.Reset(w)
}

// SetErrOutput sends error output to w (default os.Stderr).
func (e *Engine) SetErrOutput(w io.Writer) {
	e.errOut = w
}

// Flush writes any buffered output. A write error is sticky: once one
// happens, output stops and every later Flush reports it.
func (e *Engine) Flush() error {
	return e.out.Flush()
}

// printf writes trace and diagnostic text to the output. Errors surface in
// Flush.
func (e *Engine) printf(format string, a ...any) {
	_, _ = fmt.Fprintf(e.out, format, a...)
}

func (e *Engine) newline() {
	_ = e.out.WriteByte('\n')
}

// writeErr writes s to the error output; pending output is flushed first so
// the two keep their order on a terminal.
func (e *Engine) writeErr(s string) {
	_ = e.Flush()
	_, _ = io.WriteString(e.errOut, s)
}

// Start runs the initial grammar over the queued inputs (stdin if there are
// none). The status is 0 if the analysis succeeded and no flagError was
// produced, 1 otherwise; err reports a failure such as an exceeded limit or an
// unreadable input, prefixed with the input position.
func (e *Engine) Start() (status int, err error) {
	defer func() {
		if ferr := e.Flush(); err == nil && ferr != nil {
			status, err = 1, fmt.Errorf("writing output: %w", ferr)
			return
		}
		if err != nil {
			status = 1
			err = fmt.Errorf("%s:%d:%d: %w", e.Filename(), e.lineNo(), e.charNo(), err)
		}
	}()
	defer catch(&err) // recover only works in the deferred function itself
	if e.initGrammar != nil {
		if len(e.inputs) == 0 {
			e.inputs = append(e.inputs, NewStdinInput(e))
		}

		e.input = e.inputs[len(e.inputs)-1]

		e.lhsContext.State().grammar = e.initGrammar
		e.tracer.dumpGrammar(e.initGrammar)
		if len(e.initGrammar.get(e.predefinedSymbols.start.token(), e.predefinedSymbols.eof.token())) > 0 {
			e.rhsStream.currentSymbol = e.predefinedSymbols.start
		}
		if e.match() && e.flagErrors == 0 {
			return 0, nil
		}
		return 1, nil
	}
	return 1, nil
}

func (e *Engine) grammar() *grammar {
	return e.lhsContext.State().grammar
}

func (e *Engine) Filename() string {
	return e.input.Filename()
}

func (e *Engine) lineNo() int {
	return e.input.lineNo()
}

func (e *Engine) charNo() int {
	return e.input.charNo()
}

func (e *Engine) charPos() int {
	return e.input.charPos()
}

func (e *Engine) setExternal(x *External) {
	e.externalSystem = x
}

// External returns the table of functions that rules call by name; Set
// adds to it.
func (e *Engine) External() *External {
	return e.externalSystem
}

// SetLexicalMismatchPriority accepts -lexpri. The original engine sets the
// lexical priority but never applies it (see resolve), and neither does this
// one.
func (e *Engine) SetLexicalMismatchPriority(int) {}

func (e *Engine) SetBuffer(x int) int {
	e.maxLength = x
	e.rhsBuffer.setMax(x)
	return x
}

func (e *Engine) getInput() Element {
	x := e.input.get()
	for x == e.predefinedSymbols.eof && len(e.inputs) > 0 {
		e.inputs[len(e.inputs)-1] = nil
		e.inputs = e.inputs[:len(e.inputs)-1]
		if len(e.inputs) == 0 {
			break
		}
		e.input = e.inputs[len(e.inputs)-1]
		x = e.input.get()
	}
	return x
}

// addInput pushes x on top of the input stack; it is read until its eof and
// then input returns to the previous source (as for include).
func (e *Engine) addInput(x Input) {
	e.inputs = append(e.inputs, x)
	e.input = x
}

// AppendInput queues x after the existing inputs, so sources given on the
// command line are read in the order they were given.
func (e *Engine) AppendInput(x Input) {
	e.inputs = slices.Insert(e.inputs, 0, x)
	e.input = e.inputs[len(e.inputs)-1]
}

func (e *Engine) include(a []Element) Element {
	if len(a) < 2 {
		return e.predefinedSymbols.zlm
	}
	x := a[1].ToVal().ToString()
	if x == "-" {
		e.addInput(NewStdinInput(e))
	} else {
		g, err := NewFileInput(e, x)
		if err != nil {
			fail("include: %v", err)
		}
		e.addInput(g)
	}
	return newNumber(0)
}

func (e *Engine) setTrace(a []Element) Element {
	if len(a) < 2 {
		return e.predefinedSymbols.zlm
	}
	x, ok := a[1].ToVal().(*number)
	if !ok {
		return e.predefinedSymbols.zlm
	}
	return newNumber(LMNumber(e.SetTraceFlag(TraceFlag(x.toInt()))))
}

func (e *Engine) unsetTrace(a []Element) Element {
	if len(a) < 2 {
		return e.predefinedSymbols.zlm
	}
	x, ok := a[1].ToVal().(*number)
	if !ok {
		return e.predefinedSymbols.zlm
	}
	return newNumber(LMNumber(e.UnsetTraceFlag(TraceFlag(x.toInt()))))
}

func (e *Engine) SetMaxDepth(x int) int {
	e.maxDepth = x
	return e.maxDepth
}

func (e *Engine) SetMaxRepeat(x int) int {
	e.maxRepeat = x
	return e.maxRepeat
}

func (e *Engine) SetDiagramWidth(x int) int {
	e.displayWidth = x
	return e.displayWidth
}

func (e *Engine) SetTraceFlag(x TraceFlag) TraceFlag {
	if e.tracer == nil {
		e.tracer = newTracer(e)
	}
	e.tracer.flags |= x
	if x.Has(TraceDiagramText | TraceDiagram) {
		e.display = newDiagram(e, e.displayWidth)
		e.tracer.flags |= TraceDiagram
		e.tracer.flags |= TraceMismatch
		e.tracer.flags |= TraceSymbols
		e.tracer.flags |= TraceContextScope
	}
	return e.tracer.flags
}

func (e *Engine) UnsetTraceFlag(x TraceFlag) TraceFlag {
	if e.tracer == nil {
		e.tracer = newTracer(e)
	}
	e.tracer.flags &= ^x
	return e.tracer.flags
}

func (e *Engine) pushReplacement(s *state, x *rule, l contextHolder, operandsEmpty bool) {
	e.tracer.ruleScope("z=", s, e.lhsContext.Variables(), e.lhsContext.ContextLimitVariable())
	c := newRHContext(s, e.rhsStream.mode.ContextMode(), e.lhsContext)
	e.rhsStream.mode = x.newRHS(e.rhsStream.mode, c, c)
}

func (e *Engine) pushMatched(s *state, x *rule, l contextHolder, operandsEmpty bool) {
	if !operandsEmpty {
		e.lhsContext.makeVar(e.predefinedSymbols.takeFn, newStr(e.lhsStream.toRow()), e.lhsContext, e.lhsStream.variables)
	}
	e.tracer.ruleScope("==", s, e.lhsContext.Variables(), e.lhsContext.ContextLimitVariable())
	c := newRHContext(s, e.rhsStream.mode.ContextMode(), e.lhsContext)
	e.rhsStream.mode = x.newRHS(e.rhsStream.mode, c, l)
}

func (e *Engine) pushRHX(x Element) {
	e.rhsStream.mode = x.newRHX(e.rhsStream.mode, e.rhsStream.mode.ContextMode(), e.lhsContext)
}

func (e *Engine) matchedWith(l, r, x Element) bool {
	if l != nil {
		e.lhsStream.currentSymbol = nil
	}
	if r != nil {
		e.rhsStream.currentSymbol = nil
	}
	e.rsLastMatchElement = x
	return true
}

func (e *Engine) matched(l, r Element) bool {
	if l != nil {
		e.lhsStream.currentSymbol = nil
	}
	if r != nil {
		e.rhsStream.currentSymbol = nil
	}
	e.rsLastMatchElement = nil
	return true
}

func (e *Engine) match() bool {
	for {
		e.lhsStream.modeAdvance()
		e.rhsStream.modeAdvance()
		if e.lhsStream.mode == nil {
			return true // exit from e.lhsStream.mode
		}
		if e.rhsStream.mode == nil {
			return true // no more input
		}
		if e.lhsStream.currentSymbol == e.predefinedSymbols.nil {
			e.lhsStream.currentSymbol = nil
			continue
		}
		if e.rhsStream.currentSymbol == e.predefinedSymbols.nil {
			e.rhsStream.currentSymbol = nil
			continue
		}
		e.tracer.matchSymbols(e.lhsStream.currentSymbol, e.rhsStream.currentSymbol)
		if e.lhsStream.currentSymbol.match(e, e.rhsStream.currentSymbol) {
			continue
		}
		e.tracer.back(e.lhsStream.currentSymbol, e.rhsStream.currentSymbol)
		return false
	}
}

// resolve resolves a mismatch between the goal l and the input symbol r.
// It tries the rule groups (r, l), (r, -), (-, -) and (-, l) in that order;
// the new context's state and the mode snapshots are made once, for the
// first group that has rules.
func (e *Engine) resolve(l, r Element) bool {
	var sta *state
	var cp checkpoint
	// the original sets lexpri (-lexpri) but never applies it here: a
	// terminal goal is resolved at the priority of its context
	pri := e.lhsContext.Priority()

	e.tracer.resolve(l, r, pri)
	if e.lhsContext.Priority().closed() {
		return false
	}

	dontCare := e.predefinedSymbols.nil // the "don't care" initial -
	groups := [4]struct {
		first, goal Element // rule LHS and RHS initials
		v, s        Element // input put back, last match (see resolveGroup)
	}{
		{r.token(), l.token(), nil, r},
		{r.token(), dontCare, nil, r},
		{dontCare, dontCare, r, nil},
		{dontCare, l.token(), r, nil},
	}
	for _, g := range groups {
		x := e.grammar().get(g.first, g.goal)
		if len(x) == 0 {
			continue
		}
		if sta == nil {
			sta = newState(e, e.grammar(), l, r, e.input, e.charPos(), e.lineNo(), e.charNo(), e.contextsCount)
			e.contextsCount++
			cp = e.checkpoint()
		}
		if e.resolveGroup(sta, x, g.v, g.s, pri, cp) {
			return true
		}
	}
	return false
}

// resolveGroup tries the rules of one group in order. A rule whose left side
// is a single symbol applies at once; any other is matched, and the engine is
// rolled back to cp when it fails.
func (e *Engine) resolveGroup(sta *state, group []*rule, v, s Element, pri priority, cp checkpoint) bool {
	for _, x := range group {
		if x.priority.allows(pri) {
			if x.lhsLen() == 1 {
				e.rhsStream.currentSymbol = v
				if x.offset < x.rhsLen() {
					e.pushReplacement(sta, x, e.lhsContext, false)
				}
				e.lhsContext = cp.lhsContext
				return true
			}
			e.lhsContext = newLHContext(sta, e.lhsContext, x)
			if e.maxDepth > 0 && e.lhsContext.NestingDepth() >= e.maxDepth {
				fail("maximum depth %d exceeded (-max-depth)", e.maxDepth)
			}
			e.rhsStream.currentSymbol = v
			e.rsLastMatchElement = s
			e.lhsStream.mode = x.newLHS(e.lhsStream.mode, e.lhsContext)
			e.lhsStream.clearX()
			if x.match(e) {
				if x.offset < x.rhsLen() {
					e.pushMatched(sta, x, e.lhsContext, e.lhsStream.emptyX())
				}
				e.commit(cp)
				return true
			}
			e.rollback(cp)
		}
	}
	return false
}

// checkpoint is the engine state that resolving a mismatch returns to: the
// registers and mode of both streams and the left context, taken when the
// first candidate rule is found.
//
// Deliberately not restored, as in the original: the last match
// (rsLastMatchElement), which the next candidate sets before it is read; the
// right stream's current symbol, which each candidate sets from the input it
// puts back; the count of contexts, which only numbers them.
type checkpoint struct {
	lhs, rhs   modeSnap
	lhsContext contextHolder
}

func (e *Engine) checkpoint() checkpoint {
	return checkpoint{snapshot(e.lhsStream.mode), snapshot(e.rhsStream.mode), e.lhsContext}
}

// rollback undoes a candidate rule that failed. The left stream's variables
// are those of the restored context as it is now, not as it was at the
// checkpoint.
func (e *Engine) rollback(cp checkpoint) {
	e.lhsStream.variables = cp.lhsContext.Variables()
	e.lhsContext = cp.lhsContext
	e.lhsStream.mode = cp.lhs.restore()
	e.rhsStream.mode = cp.rhs.restore()
}

// commit ends a resolution that a rule matched: the left side goes back to
// the goal that mismatched, while the right side keeps the rule's
// replacement, which pushRHX has pushed.
func (e *Engine) commit(cp checkpoint) {
	e.lhsStream.mode = cp.lhs.restore()
	e.lhsContext = cp.lhsContext
}

// lhsMark records what an iteration of repeat/option may add on the left
// side: grabbed operands and variable bindings. It is a smaller checkpoint
// than the one for resolving: the modes are put back by Repeat itself.
type lhsMark struct {
	operands    opStack
	streamVars  varElement
	contextVars varElement
}

func (e *Engine) markLhs() lhsMark {
	return lhsMark{e.lhsStream.operands, e.lhsStream.variables, e.lhsContext.Variables()}
}

// releaseLhs undoes a failed iteration: its input is given back (the rhs is
// restored), so what it grabbed or bound must be dropped too.
func (e *Engine) releaseLhs(m lhsMark) {
	e.lhsStream.operands = m.operands
	e.lhsStream.variables = m.streamVars
	e.lhsContext.setVariables(m.contextVars)
}

func (e *Engine) repeat(limit int) bool {
	var w GenMode

	e.lhsStream.currentSymbol = nil
	w = e.lhsStream.mode
	x := snapshot(e.rhsStream.mode)

	for i := 0; limit == 0 || i < limit; i++ {
		if e.maxRepeat == 0 || i < e.maxRepeat {
			e.lhsStream.mode = newLHModeFrom(w)
			m := e.markLhs()
			if !e.match() {
				e.releaseLhs(m)
				break
			}
			x = snapshot(e.rhsStream.mode)
			e.tracer.repeat(i)
		} else {
			fail("maximum repeat count %d exceeded (-max-repeat)", e.maxRepeat)
		}
	}

	e.rhsStream.mode = x.restore()
	e.lhsStream.mode = w.exit()
	return true
}

func (e *Engine) pushX() {
	e.lhsStream.pushX(e.rhsStream.popX())
}

func (e *Engine) pushR(x Element) {
	e.lhsStream.pushX(x)
}

func (e *Engine) bindCvar(l, r Element) bool {
	e.lhsContext.makeVar(l, r, e.lhsContext, e.lhsStream.variables)
	e.tracer.bindCvar(l, r)
	return true
}

func (e *Engine) bindLvar(l, r Element) bool {
	e.lhsContext.makeVar(l, r, e.lhsContext, e.lhsStream.variables)
	e.tracer.bindLvar(l, r)
	return true
}

func (e *Engine) bindXvar(r Element) bool {
	l := e.lhsStream.popX()
	e.tracer.bindRvar(l, r)
	e.lhsContext.makeVar(l, r, e.rhsStream.mode, e.lhsStream.variables)
	return true
}

func (e *Engine) bindUvar(l, r Element) bool {
	e.tracer.bindRvar(l, r)
	e.lhsContext.makeVar(l, r, e.rhsStream.mode, e.lhsStream.variables)
	return true
}

func (e *Engine) takeTvar() bool {
	v := e.rhsStream.mode.ScopeVariables()
	if v != nil && v.Key() == e.predefinedSymbols.takeFn {
		if s, ok := v.Value().(*str); ok {
			for _, x := range s.v {
				e.lhsStream.pushX(x)
			}
		}
	}
	return true
}

func (e *Engine) bindTvar() bool {
	l := e.lhsStream.popX()
	v := e.rhsStream.mode.ScopeVariables()
	var r Element
	if v != nil && v.Key() == e.predefinedSymbols.takeFn {
		r = v.Value()
	} else {
		r = newStr([]Element{})
	}
	e.tracer.bindRvar(l, r)
	e.lhsContext.makeVar(l, r, e.rhsStream.mode, e.lhsStream.variables)
	return true
}

func (e *Engine) deref(pk Element, x scopeHolder) varElement {
	pp := x.ScopeVariables()
	pq := x.scopeContextLimitVariables()
	e.tracer.theRefVars(pk, pp, pq)
	for pp != nil && pk != pp.Key() {
		pp = pp.link()
	}
	return pp
}

func (e *Engine) theRef(s GenMode, k Element, x scopeHolder) GenMode {
	v := e.deref(k, x)
	e.tracer.theRefVar(v)
	if v != nil {
		return newRFMode(s, v)
	}
	return s
}

func (e *Engine) eachRef(s GenMode, k Element, x scopeHolder) GenMode {
	pp := x.ScopeVariables()
	pq := x.scopeContextLimitVariables()
	e.tracer.eachRefVars(k, pp, pq)
	for pp != nil && pp != pq {
		if k == pp.Key() {
			e.tracer.eachRefVar(pp)
			s = newRFMode(s, pp)
		}
		pp = pp.link()
	}
	return s
}

// varKey gives the variable that the value of x names, for each (expr) and
// all (expr); a name that no rule uses matches no variable.
func (e *Engine) varKey(x Element) Element {
	k := e.varSymbols.getByString(x.ToVal().ToString())
	if k == nil {
		return x.ToVal()
	}
	return k
}

func (e *Engine) allRef(s GenMode, k Element, x scopeHolder) GenMode {
	pp := x.ScopeVariables()
	pq := x.scopeContextLimitVariables()
	if pq == nil {
		return s
	}
	e.tracer.eachRefVars(k, pp, pq)
	for pp != nil && pp != pq {
		if k == pp.Key() {
			e.tracer.eachRefVar(pp)
			s = newRFMode(s, pp)
		}
		pp = pp.AllVariables()
	}
	return s
}
