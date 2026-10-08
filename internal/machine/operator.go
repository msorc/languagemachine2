package machine

type Unary struct {
	Primitive
}

func NewUnary(x string) *Unary {
	return ReSelf(&Unary{Primitive: *NewPrimitiveFromString(x)})
}

func (u *Unary) Result1(x Element) Element {
	return nil
}

func (u *Unary) Trace(s *Stream, t *Tracer) {
	t.TraceArithmetic(s, u.Self())
}

func (u *Unary) Act(sr *Stream, s GenMode) GenMode {
	x := sr.Popx()
	sr.Pushx(u.Self().Result1(x.ToVal()))
	return s
}

type Arithmetic struct {
	Primitive
}

func NewArithmetic(x string) *Arithmetic {
	return ReSelf(&Arithmetic{Primitive: *NewPrimitiveFromString(x)})
}

func (a *Arithmetic) Result2(x, y Element) Element {
	return nil
}

func (a *Arithmetic) Trace(s *Stream, t *Tracer) {
	t.TraceArithmetic(s, a.Self())
}

func (a *Arithmetic) Act(sr *Stream, b GenMode) GenMode {
	y := sr.Popx()
	x := sr.Popx()
	sr.Pushx(a.Self().Result2(x.ToVal(), y.ToVal()))
	return b
}

type Relation struct {
	Primitive
}

func NewRelation(x string) *Relation {
	return ReSelf(&Relation{Primitive: *NewPrimitiveFromString(x)})
}

func (r *Relation) Result2(x, y Element) Element {
	return nil
}

func (r *Relation) Trace(s *Stream, t *Tracer) {
	t.TraceRelation(s, r.Self())
}

func (r *Relation) Act(sr *Stream, b GenMode) GenMode {
	y := sr.Popx()
	x := sr.Popx()
	sr.Pushx(r.Self().Result2(x.ToVal(), y.ToVal()))
	return b
}

type Assignment struct {
	Primitive
}

func NewAssignment(x string) *Assignment {
	return ReSelf(&Assignment{Primitive: *NewPrimitiveFromString(x)})
}

func (a *Assignment) Result2(x, y Element) Element {
	return nil
}

func (a *Assignment) Trace(s *Stream, t *Tracer) {
	t.TraceAssignment(s, a.Self())
}

func (a *Assignment) Act(sr *Stream, b GenMode) GenMode {
	y := sr.Popx()
	x := sr.Popx()
	sr.Pushx(a.Self().Result2(x, y.ToVal()))
	return b
}

type IncDec struct {
	Primitive
}

func NewIncDec(x string) *IncDec {
	return ReSelf(&IncDec{Primitive: *NewPrimitiveFromString(x)})
}

func (i *IncDec) Result1(x Element) Element {
	return nil
}

func (i *IncDec) Trace(s *Stream, t *Tracer) {
	t.TraceAssignment(s, i)
}

func (i *IncDec) Act(sr *Stream, b GenMode) GenMode {
	sr.Pushx(i.Self().Result1(sr.Popx()))
	return b
}

type Index struct {
	Primitive
}

func NewIndex(x string) *Index {
	return ReSelf(&Index{Primitive: *NewPrimitiveFromString(x)})
}

func (i *Index) Result2(x, y Element) Element {
	return nil
}

func (i *Index) Trace(s *Stream, t *Tracer) {
	t.TraceIndex(s, i.Self())
}

func (i *Index) Act(sr *Stream, b GenMode) GenMode {
	y := sr.Popx()
	x := sr.Popx()
	sr.Pushx(i.Self().Result2(x, y.ToVal()))
	return b
}

type Idxf struct {
	Index
}

func NewIdxf(x string) *Idxf {
	return ReSelf(&Idxf{Index: *NewIndex(x)})
}
func (i *Idxf) Result2(x Element, y Element) Element {
	return x.Idxf(y)
}

type Idtf struct {
	Index
}

func NewIdtf(x string) *Idtf {
	return ReSelf(&Idtf{Index: *NewIndex(x)})
}
func (i *Idtf) Result2(x Element, y Element) Element {
	return x.Idtf(y)
}

type StoValf struct {
	Assignment
}

func NewStoValf(x string) *StoValf {
	return ReSelf(&StoValf{Assignment: *NewAssignment(x)})
}
func (s *StoValf) Result2(x Element, y Element) Element {
	return x.StoValf(y)
}

type StoAddf struct {
	Assignment
}

func NewStoAddf(x string) *StoAddf {
	return ReSelf(&StoAddf{Assignment: *NewAssignment(x)})
}
func (s *StoAddf) Result2(x Element, y Element) Element {
	return x.StoAddf(y)
}

type StoSubf struct {
	Assignment
}

func NewStoSubf(x string) *StoSubf {
	return ReSelf(&StoSubf{Assignment: *NewAssignment(x)})
}
func (s *StoSubf) Result2(x Element, y Element) Element {
	return x.StoSubf(y)
}

type StoMulf struct {
	Assignment
}

func NewStoMulf(x string) *StoMulf {
	return ReSelf(&StoMulf{Assignment: *NewAssignment(x)})
}
func (s *StoMulf) Result2(x Element, y Element) Element {
	return x.StoMulf(y)
}

type StoDivf struct {
	Assignment
}

func NewStoDivf(x string) *StoDivf {
	return ReSelf(&StoDivf{Assignment: *NewAssignment(x)})
}
func (s *StoDivf) Result2(x Element, y Element) Element {
	return x.StoDivf(y)
}

type StoModf struct {
	Assignment
}

func NewStoModf(x string) *StoModf {
	return ReSelf(&StoModf{Assignment: *NewAssignment(x)})
}
func (s *StoModf) Result2(x Element, y Element) Element {
	return x.StoModf(y)
}

type Eeqf struct {
	Relation
}

func NewEeqf(x string) *Eeqf {
	return ReSelf(&Eeqf{Relation: *NewRelation(x)})
}
func (e *Eeqf) Result2(x Element, y Element) Element {
	return x.Eeqf(y)
}

type Neef struct {
	Relation
}

func NewNeef(x string) *Neef {
	return ReSelf(&Neef{Relation: *NewRelation(x)})
}
func (n *Neef) Result2(x Element, y Element) Element {
	return x.Neef(y)
}

type Inf struct {
	Relation
}

func NewInf(x string) *Inf {
	return ReSelf(&Inf{Relation: *NewRelation(x)})
}
func (i *Inf) Result2(x Element, y Element) Element {
	return x.Inf(y)
}

type Eqf struct {
	Relation
}

func NewEqf(x string) *Eqf {
	return ReSelf(&Eqf{Relation: *NewRelation(x)})
}
func (e Eqf) Result2(x Element, y Element) Element {
	return x.Eqf(y)
}

type Nef struct {
	Relation
}

func NewNef(x string) *Nef {
	return ReSelf(&Nef{Relation: *NewRelation(x)})
}
func (n *Nef) Result2(x Element, y Element) Element {
	return x.Nef(y)
}

type Ltf struct {
	Relation
}

func NewLtf(x string) *Ltf {
	return ReSelf(&Ltf{Relation: *NewRelation(x)})
}
func (l *Ltf) Result2(x Element, y Element) Element {
	return x.Ltf(y)
}

type Gtf struct {
	Relation
}

func NewGtf(x string) *Gtf {
	return ReSelf(&Gtf{Relation: *NewRelation(x)})
}
func (g *Gtf) Result2(x Element, y Element) Element {
	return x.Gtf(y)
}

type Lef struct {
	Relation
}

func NewLef(x string) *Lef {
	return ReSelf(&Lef{Relation: *NewRelation(x)})
}
func (l *Lef) Result2(x Element, y Element) Element {
	return x.Lef(y)
}

type Gef struct {
	Relation
}

func NewGef(x string) *Gef {
	return ReSelf(&Gef{Relation: *NewRelation(x)})
}
func (g *Gef) Result2(x Element, y Element) Element {
	return x.Gef(y)
}

type BitXorf struct {
	Arithmetic
}

func NewBitXorf(x string) *BitXorf {
	return ReSelf(&BitXorf{Arithmetic: *NewArithmetic(x)})
}
func (b *BitXorf) Result2(x Element, y Element) Element {
	return x.BitXorf(y)
}

type BitOrf struct {
	Arithmetic
}

func NewBitOrf(x string) *BitOrf {
	return ReSelf(&BitOrf{Arithmetic: *NewArithmetic(x)})
}
func (b *BitOrf) Result2(x Element, y Element) Element {
	return x.BitOrf(y)
}

type BitAndf struct {
	Arithmetic
}

func NewBitAndf(x string) *BitAndf {
	return ReSelf(&BitAndf{Arithmetic: *NewArithmetic(x)})
}
func (b *BitAndf) Result2(x Element, y Element) Element {
	return x.BitAndf(y)
}

type Addf struct {
	Arithmetic
}

func NewAddf(x string) *Addf {
	return ReSelf(&Addf{Arithmetic: *NewArithmetic(x)})
}
func (a *Addf) Result2(x Element, y Element) Element {
	return x.Addf(y)
}

type Subf struct {
	Arithmetic
}

func NewSubf(x string) *Subf {
	return ReSelf(&Subf{Arithmetic: *NewArithmetic(x)})
}
func (s *Subf) Result2(x Element, y Element) Element {
	return x.Subf(y)
}

type Mulf struct {
	Arithmetic
}

func NewMulf(x string) *Mulf {
	return ReSelf(&Mulf{Arithmetic: *NewArithmetic(x)})
}
func (m *Mulf) Result2(x Element, y Element) Element {
	return x.Mulf(y)
}

type Divf struct {
	Arithmetic
}

func NewDivf(x string) *Divf {
	return ReSelf(&Divf{Arithmetic: *NewArithmetic(x)})
}
func (d *Divf) Result2(x Element, y Element) Element {
	return x.Divf(y)
}

type Modf struct {
	Arithmetic
}

func NewModf(x string) *Modf {
	return ReSelf(&Modf{Arithmetic: *NewArithmetic(x)})
}
func (m *Modf) Result2(x Element, y Element) Element {
	return x.Modf(y)
}

type Preincf struct {
	IncDec
}

func NewPreincf(x string) *Preincf {
	return ReSelf(&Preincf{IncDec: *NewIncDec(x)})
}
func (p *Preincf) Result1(x Element) Element {
	return x.Preincf()
}

type Predecf struct {
	IncDec
}

func NewPredecf(x string) *Predecf {
	return ReSelf(&Predecf{IncDec: *NewIncDec(x)})
}
func (p *Predecf) Result1(x Element) Element {
	return x.Predecf()
}

type Postincf struct {
	IncDec
}

func NewPostincf(x string) *Postincf {
	return ReSelf(&Postincf{IncDec: *NewIncDec(x)})
}
func (p *Postincf) Result1(x Element) Element {
	return x.Postincf()
}

type Postdecf struct {
	IncDec
}

func NewPostdecf(x string) *Postdecf {
	return ReSelf(&Postdecf{IncDec: *NewIncDec(x)})
}
func (p *Postdecf) Result1(x Element) Element {
	return x.Postdecf()
}

type Negf struct {
	Unary
}

func NewNegf(x string) *Negf {
	return ReSelf(&Negf{Unary: *NewUnary(x)})
}
func (n *Negf) Result1(x Element) Element {
	return x.Negf()
}

type Notf struct {
	Unary
}

func NewNotf(x string) *Notf {
	return ReSelf(&Notf{Unary: *NewUnary(x)})
}
func (n *Notf) Result1(x Element) Element {
	return x.Notf()
}

type Invf struct {
	Unary
}

func NewInvf(x string) *Invf {
	return ReSelf(&Invf{Unary: *NewUnary(x)})
}
func (i *Invf) Result1(x Element) Element {
	return x.Invf()
}
