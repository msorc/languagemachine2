package machine

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"slices"
)

const (
	LEXPRI    int = 1000      // default value for lexical priority (-lexpri)
	MAXLENGTH int = 64 * 1024 // default maximum length of the input buffer (-buffer)

	initialBufferLength = 1024      // the input buffer starts this long and grows to maxLength
	outputBufferSize    = 64 * 1024 // size of the buffered output writer
	defaultDisplayWidth = 80
)

// theZlm is the null element. It is stateless, so all engines share it.
var theZlm = NewZLM("null")

// Null is the null value: what an unset variable holds.
func Null() Element { return theZlm }

// Engine loads grammars and applies them to its inputs.
//
// An Engine must be used by one goroutine at a time, and loaded rules belong
// to the engine that loaded them. Engines share no state, so separate engines
// can run concurrently.
type Engine struct {
	contextsCount int           // count of new contexts used to give each a unique identity
	lhsContext    ContextHolder // lhs context stack for mismatch events being resolved

	lhsStream          *Stream // lhs registers
	rhsStream          *Stream // rhs registers
	rsLastMatchElement Element // element resulting from last match

	grammars    *Selector // table of grammars selected by symbol
	initGrammar *Grammar  // initial grammar
	ruleNumbers int       // one more than the highest rule number

	terminalSymbols    *Dict   // terminal symbols
	nonTerminalSymbols *Dict   // non-terminal symbols
	varSymbols         *Dict   // variables
	userSymbols        *Dict   // user symbols - guaranteed not to match system symbols
	functionSymbols    *Dict   // primitive operator symbols
	predefinedSymbols  *Predef // predefined symbols with special significance

	loader *Loader // rule loader

	inputs []GrammarIO // stack of input sources, the current one last
	input  GrammarIO   // current input

	rhsBuffer *RZBuffer // circular buffer at outermost level of rhs

	flagErrors int // incremented by flagSym
	warnErrors int // incremented by warnSym

	externalSystem *LMExternal // external interfaces

	tracer *Tracer // trace handler - null if no tracing required

	out    *bufio.Writer // output, traces and diagrams; flushed by Start
	errOut io.Writer     // error output (err)

	display      *Diagram // to display trace as diagram
	displayWidth int      // width for diagram display

	maxDepth  int // limit on analysis recursion depth - zero means no limit
	maxRepeat int // limit on repetition at repeat     - zero means no limit
	maxLength int // maximum size of the rhs input buffer

	symbolsDefined bool // predefined symbols exist; they must be created only once
}

func NewEngine() *Engine {
	return NewEngineFromLength(MAXLENGTH)
}

// NewEngineFromLength makes an engine whose input buffer grows to at most
// maxLength symbols (-buffer).
func NewEngineFromLength(maxLength int) *Engine {
	e := &Engine{
		maxLength:          maxLength,
		displayWidth:       defaultDisplayWidth,
		functionSymbols:    NewDict(),
		terminalSymbols:    NewDict(),
		nonTerminalSymbols: NewDict(),
		varSymbols:         NewDict(),
		userSymbols:        NewDict(),
		predefinedSymbols:  NewPredef(),
		externalSystem:     NewLMExternal(),
		grammars:           NewSelector(),
		out:                bufio.NewWriterSize(os.Stdout, outputBufferSize),
		errOut:             os.Stderr,
	}
	e.input = NewGramStdioFromEngine(e) // replaced by the first input in Start
	e.rhsBuffer = NewRZBuffer(make([]Element, initialBufferLength), e.maxLength)
	root := NewState(e, nil, nil, nil, e.input, 0, 0, 0, e.contextsCount)
	e.contextsCount++
	e.lhsContext = NewContextFromState(LHContext, root)
	e.lhsStream = NewStream(e, "lh", 0)
	e.rhsStream = NewStream(e, "rh", 0)
	e.lhsStream.mode = NewLZModeFromContext(e.lhsContext, e.lhsStream)
	e.rhsStream.mode = NewRZModeFromContext(NewContextFromState(RHContext, root), e.rhsStream)
	return e
}

func (e *Engine) SetMachineElements(args []Element) Element {
	if len(args) < 2 {
		return e.predefinedSymbols.zlm
	}
	k := args[1].ToVal()
	g := e.grammars.Get(k.ToString())
	if g != nil {
		e.lhsContext.State().grammar = g
	}
	return k
}

func (e *Engine) AddRule(v []Element, i int) {
	if i >= e.ruleNumbers {
		e.ruleNumbers = i + 1
	}
	gr := e.grammars.Select(v[0])
	if e.initGrammar == nil {
		e.initGrammar = gr
		e.lhsContext.State().grammar = e.initGrammar
	}
	gr.DefineRule(v, i)
}

// LoadFromString loads rules, replacing any loaded before.
func (e *Engine) LoadFromString(rules string) error {
	return e.LoadFromStringReset(rules, true)
}

// LoadFromStringReset loads rules; with reset false they are added to the
// rules already loaded (-add).
func (e *Engine) LoadFromStringReset(rules string, reset bool) (err error) {
	e.defineSymbols()
	if reset {
		// the first grammar of the new rules becomes the initial one; the
		// old initial grammar is not in the new table. If the new rules do
		// not load, the old ones stay in force.
		grammars, initGrammar := e.grammars, e.initGrammar
		e.grammars = NewSelector()
		e.initGrammar = nil
		defer func() {
			if err != nil {
				e.grammars, e.initGrammar = grammars, initGrammar
				e.lhsContext.State().grammar = initGrammar
			}
		}()
	}
	if e.loader == nil {
		e.loader = NewLoader(e)
	}
	return e.loader.Load(rules)
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
			err = fmt.Errorf("%s:%d:%d: %w", e.Filename(), e.Lineno(), e.Charno(), err)
		}
	}()
	defer catch(&err) // recover only works in the deferred function itself
	if e.initGrammar != nil {
		if len(e.inputs) == 0 {
			e.inputs = append(e.inputs, NewGramStdioFromEngine(e))
		}

		e.input = e.inputs[len(e.inputs)-1]

		e.lhsContext.State().grammar = e.initGrammar
		if e.tracer != nil {
			e.tracer.Dumpg(e.initGrammar)
		}
		if e.initGrammar.Get(e.predefinedSymbols.start.Token(), e.predefinedSymbols.eof.Token()) != nil {
			e.rhsStream.currentSymbol = e.predefinedSymbols.start
		}
		if e.Match() && e.flagErrors == 0 {
			return 0, nil
		}
		return 1, nil
	}
	return 1, nil
}

func (e *Engine) Grammar() *Grammar {
	return e.lhsContext.State().grammar
}

func (e *Engine) Filename() string {
	return e.input.Filename()
}

func (e *Engine) Lineno() int {
	return e.input.LineNo()
}

func (e *Engine) Charno() int {
	return e.input.CharNo()
}

func (e *Engine) Charpos() int {
	return e.input.CharPos()
}

func (e *Engine) SetExternal(x *LMExternal) {
	e.externalSystem = x
}

// External returns the table of functions that rules call by name; Set
// adds to it.
func (e *Engine) External() *LMExternal {
	return e.externalSystem
}

// SetLexicalMismatchPriority accepts -lexpri. The original engine sets the
// lexical priority but never applies it (see resolve), and neither does this
// one.
func (e *Engine) SetLexicalMismatchPriority(int) {}

func (e *Engine) SetBuffer(x int) int {
	e.maxLength = x
	e.rhsBuffer.SetMax(x)
	return x
}

func (e *Engine) GetInput() Element {
	x := e.input.Get()
	for x == e.predefinedSymbols.eof && len(e.inputs) > 0 {
		e.inputs[len(e.inputs)-1] = nil
		e.inputs = e.inputs[:len(e.inputs)-1]
		if len(e.inputs) == 0 {
			break
		}
		e.input = e.inputs[len(e.inputs)-1]
		x = e.input.Get()
	}
	return x
}

// AddInput pushes x on top of the input stack; it is read until its eof and
// then input returns to the previous source (as for include).
func (e *Engine) AddInput(x GrammarIO) {
	e.inputs = append(e.inputs, x)
	e.input = x
}

// AppendInput queues x after the existing inputs, so sources given on the
// command line are read in the order they were given.
func (e *Engine) AppendInput(x GrammarIO) {
	e.inputs = slices.Insert(e.inputs, 0, x)
	e.input = e.inputs[len(e.inputs)-1]
}

func (e *Engine) Include(a []Element) Element {
	if len(a) < 2 {
		return e.predefinedSymbols.zlm
	}
	x := a[1].ToVal().ToString()
	if x == "-" {
		e.AddInput(NewGramStdioFromEngine(e))
	} else {
		g, err := NewGramInputFile(e, x)
		if err != nil {
			fail("include: %v", err)
		}
		e.AddInput(g)
	}
	return NewNumber(0)
}

func (e *Engine) SetTrace(a []Element) Element {
	if len(a) < 2 {
		return e.predefinedSymbols.zlm
	}
	x, ok := a[1].ToVal().(*Number)
	if !ok {
		return e.predefinedSymbols.zlm
	}
	return NewNumber(LMNumber(e.SetTraceFlag(x.ToInt())))
}

func (e *Engine) UnsetTrace(a []Element) Element {
	if len(a) < 2 {
		return e.predefinedSymbols.zlm
	}
	x, ok := a[1].ToVal().(*Number)
	if !ok {
		return e.predefinedSymbols.zlm
	}
	return NewNumber(LMNumber(e.UnsetTraceFlag(x.ToInt())))
}

func (e *Engine) SetMaxDepth(x int) int {
	e.maxDepth = x
	return e.maxDepth
}

func (e *Engine) SetMaxRepeat(x int) int {
	e.maxRepeat = x
	return e.maxRepeat
}

func (e *Engine) SetDisplayW(x int) int {
	e.displayWidth = x
	return e.displayWidth
}

func (e *Engine) SetTraceFlag(x int) int {
	if e.tracer == nil {
		e.tracer = NewTracer(e)
	}
	e.tracer.Flags |= x
	if x&DIAGRAMT != 0 || x&DIAGRAM != 0 {
		e.display = NewDiagram(e, e.displayWidth)
		e.tracer.Flags |= DIAGRAM
		e.tracer.Flags |= MISMATCH
		e.tracer.Flags |= SYMBOLS
		e.tracer.Flags |= CXSCOPE
	}
	return e.tracer.Flags
}

func (e *Engine) UnsetTraceFlag(x int) int {
	if e.tracer == nil {
		e.tracer = NewTracer(e)
	}
	e.tracer.Flags &= ^x
	return e.tracer.Flags
}

func (e *Engine) PushRhx0(s *State, x *Rule, l ContextHolder, operandsEmpty bool) {
	if e.tracer != nil {
		e.tracer.RuleScope("z=", s, e.lhsContext.Variables(), e.lhsContext.ContextLimitVariable())
	}
	c := NewRHContextFromStateContext(s, e.rhsStream.mode.ContextMode(), e.lhsContext)
	e.rhsStream.mode = x.Newrhs(e.rhsStream.mode, c, c)
}

func (e *Engine) PushRhx1(s *State, x *Rule, l ContextHolder, operandsEmpty bool) {
	if !operandsEmpty {
		e.lhsContext.MakeVar(e.predefinedSymbols.takeFn, NewStr(e.lhsStream.ToRow()), e.lhsContext, e.lhsStream.variables)
	}
	if e.tracer != nil {
		e.tracer.RuleScope("==", s, e.lhsContext.Variables(), e.lhsContext.ContextLimitVariable())
	}
	c := NewRHContextFromStateContext(s, e.rhsStream.mode.ContextMode(), e.lhsContext)
	e.rhsStream.mode = x.Newrhs(e.rhsStream.mode, c, l)
}

func (e *Engine) PushRhx(x Element) {
	e.rhsStream.mode = x.NewRHX(e.rhsStream.mode, e.rhsStream.mode.ContextMode(), e.lhsContext)
}

func (e *Engine) Matched3E(l, r, x Element) bool {
	if l != nil {
		e.lhsStream.currentSymbol = nil
	}
	if r != nil {
		e.rhsStream.currentSymbol = nil
	}
	e.rsLastMatchElement = x
	return true
}

func (e *Engine) Matched2E(l, r Element) bool {
	if l != nil {
		e.lhsStream.currentSymbol = nil
	}
	if r != nil {
		e.rhsStream.currentSymbol = nil
	}
	e.rsLastMatchElement = nil
	return true
}

func (e *Engine) Match() bool {
	for {
		e.lhsStream.ModeAdvance()
		e.rhsStream.ModeAdvance()
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
		if e.tracer != nil {
			e.tracer.MatchSymbols(e.lhsStream.currentSymbol, e.rhsStream.currentSymbol)
		}
		if e.lhsStream.currentSymbol.Match(e, e.rhsStream.currentSymbol) {
			continue
		}
		if e.tracer != nil {
			e.tracer.Back(e.lhsStream.currentSymbol, e.rhsStream.currentSymbol)
		}
		return false
	}
}

// ResolveE resolves a mismatch between the goal l and the input symbol r.
// It tries the rule groups (r, l), (r, -), (-, -) and (-, l) in that order;
// the new context's state and the mode snapshots are made once, for the
// first group that has rules.
func (e *Engine) ResolveE(l, r Element) bool {
	var sta *State
	var cp checkpoint
	// the original sets lexpri (-lexpri) but never applies it here: a
	// terminal goal is resolved at the priority of its context
	pri := e.lhsContext.Priority()

	if e.tracer != nil {
		e.tracer.Resolve(l, r, pri)
	}
	if e.lhsContext.Priority() == PRIMASK {
		return false
	}

	dontCare := e.predefinedSymbols.nil // the "don't care" initial -
	groups := [4]struct {
		first, goal Element // rule LHS and RHS initials
		v, s        Element // input put back, last match (see ResolveState)
	}{
		{r.Token(), l.Token(), nil, r},
		{r.Token(), dontCare, nil, r},
		{dontCare, dontCare, r, nil},
		{dontCare, l.Token(), r, nil},
	}
	for _, g := range groups {
		x := e.Grammar().Get(g.first, g.goal)
		if x == nil {
			continue
		}
		if sta == nil {
			sta = NewState(e, e.Grammar(), l, r, e.input, e.Charpos(), e.Lineno(), e.Charno(), e.contextsCount)
			e.contextsCount++
			cp = e.checkpoint()
		}
		if e.ResolveState(sta, x, g.v, g.s, pri, cp) {
			return true
		}
	}
	return false
}

// ResolveState tries the rules of one group, from a, in order. A rule whose
// left side is a single symbol applies at once; any other is matched, and
// the engine is rolled back to cp when it fails.
func (e *Engine) ResolveState(sta *State, a *Rule, v, s Element, pri int, cp checkpoint) bool {
	x := a
	for {
		if x == nil {
			return false
		}
		if x.Allow(pri) {
			if x.Lhlength() == 1 {
				e.rhsStream.currentSymbol = v
				if x.offset < x.Rhlength() {
					e.PushRhx0(sta, x, e.lhsContext, false)
				}
				e.lhsContext = cp.lhsContext
				return true
			}
			e.lhsContext = NewLHContextFromRule(sta, e.lhsContext, x)
			if e.maxDepth > 0 && e.lhsContext.NestingDepth() >= e.maxDepth {
				fail("maximum depth %d exceeded (-max-depth)", e.maxDepth)
			}
			e.rhsStream.currentSymbol = v
			e.rsLastMatchElement = s
			e.lhsStream.mode = x.Newlhs(e.lhsStream.mode, e.lhsContext)
			e.lhsStream.ClearX()
			if x.Match(e) {
				break
			}
			e.rollback(cp)
		}
		x = x.next
	}
	if x.offset < x.Rhlength() {
		e.PushRhx1(sta, x, e.lhsContext, e.lhsStream.EmptyX())
	}
	e.commit(cp)
	return true
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
	lhsContext ContextHolder
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
// replacement, which PushRhx has pushed.
func (e *Engine) commit(cp checkpoint) {
	e.lhsStream.mode = cp.lhs.restore()
	e.lhsContext = cp.lhsContext
}

// lhsMark records what an iteration of repeat/option may add on the left
// side: grabbed operands and variable bindings. It is a smaller checkpoint
// than the one for resolving: the modes are put back by Repeat itself.
type lhsMark struct {
	operands    OpStack
	streamVars  VarElement
	contextVars VarElement
}

func (e *Engine) markLhs() lhsMark {
	return lhsMark{e.lhsStream.operands, e.lhsStream.variables, e.lhsContext.Variables()}
}

// releaseLhs undoes a failed iteration: its input is given back (the rhs is
// restored), so what it grabbed or bound must be dropped too.
func (e *Engine) releaseLhs(m lhsMark) {
	e.lhsStream.operands = m.operands
	e.lhsStream.variables = m.streamVars
	e.lhsContext.SetVariables(m.contextVars)
}

func (e *Engine) Repeat(limit int) bool {
	var w GenMode

	e.lhsStream.currentSymbol = nil
	w = e.lhsStream.mode
	x := snapshot(e.rhsStream.mode)

	for i := 0; limit == 0 || i < limit; i++ {
		if e.maxRepeat == 0 || i < e.maxRepeat {
			e.lhsStream.mode = NewLHModeFromMode(w)
			m := e.markLhs()
			if !e.Match() {
				e.releaseLhs(m)
				break
			}
			x = snapshot(e.rhsStream.mode)
			if e.tracer != nil {
				e.tracer.Repeat(i)
			}
		} else {
			fail("maximum repeat count %d exceeded (-max-repeat)", e.maxRepeat)
		}
	}

	e.rhsStream.mode = x.restore()
	e.lhsStream.mode = w.Return()
	return true
}

func (e *Engine) PushX() {
	e.lhsStream.Pushx(e.rhsStream.Popx())
}

func (e *Engine) PushR(x Element) {
	e.lhsStream.Pushx(x)
}

func (e *Engine) BindCvar(l, r Element) bool {
	e.lhsContext.MakeVar(l, r, e.lhsContext, e.lhsStream.variables)
	if e.tracer != nil {
		e.tracer.BindCvar(l, r)
	}
	return true
}

func (e *Engine) BindLvar(l, r Element) bool {
	e.lhsContext.MakeVar(l, r, e.lhsContext, e.lhsStream.variables)
	if e.tracer != nil {
		e.tracer.BindLvar(l, r)
	}
	return true
}

func (e *Engine) BindXvarE(r Element) bool {
	l := e.lhsStream.Popx()
	if e.tracer != nil {
		e.tracer.BindRvar(l, r)
	}
	e.lhsContext.MakeVar(l, r, e.rhsStream.mode, e.lhsStream.variables)
	return true
}

func (e *Engine) BindUvar(l, r Element) bool {
	if e.tracer != nil {
		e.tracer.BindRvar(l, r)
	}
	e.lhsContext.MakeVar(l, r, e.rhsStream.mode, e.lhsStream.variables)
	return true
}

func (e *Engine) TakeTvar() bool {
	v := e.rhsStream.mode.ScopeVariables()
	if v != nil && v.Key() == e.predefinedSymbols.takeFn {
		if s, ok := v.Value().(*Str); ok {
			for _, x := range s.V {
				e.lhsStream.Pushx(x)
			}
		}
	}
	return true
}

func (e *Engine) BindTvar() bool {
	l := e.lhsStream.Popx()
	v := e.rhsStream.mode.ScopeVariables()
	var r Element
	if v != nil && v.Key() == e.predefinedSymbols.takeFn {
		r = v.Value()
	} else {
		r = NewStr([]Element{})
	}
	if e.tracer != nil {
		e.tracer.BindRvar(l, r)
	}
	e.lhsContext.MakeVar(l, r, e.rhsStream.mode, e.lhsStream.variables)
	return true
}

func (e *Engine) Deref(pk Element, x ScopeHolder) VarElement {
	pp := x.ScopeVariables()
	pq := x.ScopeContextLimitVariables()
	if e.tracer != nil {
		e.tracer.TheRefVars(pk, pp, pq)
	}
	for pp != nil && pk != pp.Key() {
		pp = pp.Link()
	}
	return pp
}

func (e *Engine) TheRef(s GenMode, k Element, x ScopeHolder) GenMode {
	v := e.Deref(k, x)
	if e.tracer != nil {
		e.tracer.TheRefVar(v)
	}
	if v != nil {
		return NewRFModeFromVar(s, v)
	}
	return s
}

func (e *Engine) EachRef(s GenMode, k Element, x ScopeHolder) GenMode {
	pp := x.ScopeVariables()
	pq := x.ScopeContextLimitVariables()
	if e.tracer != nil {
		e.tracer.EachRefVars(k, pp, pq)
	}
	for pp != nil && pp != pq {
		if k == pp.Key() {
			if e.tracer != nil {
				e.tracer.EachRefVar(pp)
			}
			s = NewRFModeFromVar(s, pp)
		}
		pp = pp.Link()
	}
	return s
}

// varKey gives the variable that the value of x names, for each (expr) and
// all (expr); a name that no rule uses matches no variable.
func (e *Engine) varKey(x Element) Element {
	k := e.varSymbols.GetByString(x.ToVal().ToString())
	if k == nil {
		return x.ToVal()
	}
	return k
}

func (e *Engine) AllRef(s GenMode, k Element, x ScopeHolder) GenMode {
	pp := x.ScopeVariables()
	pq := x.ScopeContextLimitVariables()
	if pq == nil {
		return s
	}
	if e.tracer != nil {
		e.tracer.EachRefVars(k, pp, pq)
	}
	for pp != nil && pp != pq {
		if k == pp.Key() {
			if e.tracer != nil {
				e.tracer.EachRefVar(pp)
			}
			s = NewRFModeFromVar(s, pp)
		}
		pp = pp.AllVariables()
	}
	return s
}
