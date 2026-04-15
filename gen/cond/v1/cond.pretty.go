// Package condv1 provides a general perpose represation of condition expression.
package condv1

import (
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
)

func (a *And) Pretty() string {
	if a == nil {
		return `<nil>`
	}

	var sb strings.Builder
	first := true
	for _, p := range a.GetPred() {
		if first {
			first = false
		} else {
			sb.WriteString(` && `)
		}

		switch p := p.GetIs().(type) {
		case *Predicate_Expr:
			sb.WriteString(p.Pretty())
		case *Predicate_Not:
			sb.WriteString(p.Pretty())
		case *Predicate_And:
			sb.WriteString(`(`)
			sb.WriteString(p.Pretty())
			sb.WriteString(`)`)
		case *Predicate_Or:
			sb.WriteString(`(`)
			sb.WriteString(p.Pretty())
			sb.WriteString(`)`)
		}
	}

	return sb.String()
}

func (a *Or) Pretty() string {
	if a == nil {
		return `<nil>`
	}

	var sb strings.Builder
	first := true
	for _, p := range a.GetPred() {
		if first {
			first = false
		} else {
			sb.WriteString(` || `)
		}

		pe, ok := p.GetIs().(*Predicate_Expr)
		if ok {
			sb.WriteString(pe.Pretty())
		} else {
			sb.WriteString(`(`)
			sb.WriteString(p.Pretty())
			sb.WriteString(`)`)
		}
	}

	return sb.String()
}

func (e *Expr) Pretty() string {
	var sb strings.Builder
	sb.WriteString(e.GetField())
	sb.WriteString(e.GetOp().Pretty())
	sb.WriteString(protojson.Format(e.GetValue()))
	return sb.String()
}

func (n *Not) Pretty() string {
	var sb strings.Builder
	sb.WriteString(`!(`)
	sb.WriteString(n.Pred.Pretty())
	sb.WriteString(`)`)
	return sb.String()
}

func (pe *Predicate_Expr) Pretty() string { return pe.Expr.Pretty() }
func (pn *Predicate_Not) Pretty() string  { return pn.Not.Pretty() }
func (pa *Predicate_And) Pretty() string  { return pa.And.Pretty() }
func (po *Predicate_Or) Pretty() string   { return po.Or.Pretty() }

func (p *Predicate) Pretty() string {
	if p == nil {
		return `<nil>`
	}

	return p.GetIs().(interface{ Pretty() string }).Pretty()
}

var opPretty = map[int32]string{
	int32(Op_OP_EQ):       `=`,
	int32(Op_OP_LE):       `≤`,
	int32(Op_OP_LT):       `<`,
	int32(Op_OP_GE):       `≥`,
	int32(Op_OP_GT):       `>`,
	int32(Op_OP_SIM):      `~`,
	int32(Op_OP_IN):       `∈`,
	int32(Op_OP_NI):       `∋`,
	int32(Op_OP_SUBSET):   `⊂`,
	int32(Op_OP_SUBSETEQ): `⊆`,
	int32(Op_OP_SUPSET):   `⊃`,
	int32(Op_OP_SUPSETEQ): `⊇`,
	int32(Op_OP_APPROX):   `≈`,
}

func (o Op) Pretty() string { return opPretty[int32(o.Number())] }
