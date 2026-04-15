package condv1

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/types/known/structpb"
)

func NewOr(p ...*Predicate) *Predicate {
	return &Predicate{Is: &Predicate_Or{Or: &Or{Pred: p}}}
}

func NewAnd(p ...*Predicate) *Predicate {
	return &Predicate{Is: &Predicate_And{And: &And{Pred: p}}}
}

func NewNot(p *Predicate) *Predicate {
	return &Predicate{Is: &Predicate_Not{Not: &Not{Pred: p}}}
}

// Eq create a new [Predicate] with [Op_OP_EQ] [Expr] in it.
// value must be a valid type that [structpb.NewValue] allow,
// else will cause a panic.
func Eq(field string, value any) *Predicate {
	return expr(Op_OP_EQ)(field, value)
}

// Gt create a new [Predicate] with [Op_OP_GT] [Expr] in it.
// value must be a valid type that [structpb.NewValue] allow,
// else will cause a panic.
func Gt(field string, value any) *Predicate {
	return expr(Op_OP_GT)(field, value)
}

func expr(op Op) func(field string, value any) *Predicate {
	return func(field string, value any) *Predicate {
		v, err := structpb.NewValue(value)
		if err != nil {
			panic(err)
		}

		return &Predicate{Is: &Predicate_Expr{Expr: &Expr{Field: field, Value: v, Op: op}}}
	}
}

func (p *Predicate) Eval(fn func(*Expr) bool) bool {
	switch p := p.GetIs().(type) {
	case *Predicate_And:
		result := true
		for _, a := range p.And.GetPred() {
			if result = result && a.Eval(fn); !result {
				break
			}
		}
		return result

	case *Predicate_Or:
		result := false
		for _, o := range p.Or.GetPred() {
			if result = result && o.Eval(fn); result {
				break
			}
		}
		return result

	case *Predicate_Not:
		return !fn(p.Not.Pred.GetExpr())

	case *Predicate_Expr:
		return fn(p.Expr)

	default:
		// unreachable
		return false
	}
}

func exprFn(e *Expr, depth, seq int) {
	fmt.Println(e.Pretty())
}

func andFn(p *Predicate, depth, seq int) {
	var sb strings.Builder
	if seq > 0 {
		for range depth - 1 {
			sb.WriteString(`   `)
		}
	}
	sb.WriteString(`/\ `)
	fmt.Print(sb.String())
	p.Walk(andFn, orFn, notFn, exprFn, depth, seq)
}

func orFn(p *Predicate, depth, seq int) {
	var sb strings.Builder
	if seq > 0 {
		for range depth - 1 {
			sb.WriteString(`   `)
		}
	}
	sb.WriteString(`\/ `)
	fmt.Print(sb.String())
	p.Walk(andFn, orFn, notFn, exprFn, depth, seq)
}

func notFn(p *Predicate, depth, seq int) {
	var sb strings.Builder
	if seq > 0 {
		for range depth - 1 {
			sb.WriteString(`   `)
		}
	}
	fmt.Print(` ~ `)
	p.Walk(andFn, orFn, notFn, exprFn, depth, seq)
}

func (p *Predicate) Walk(
	and, or, not func(*Predicate, int, int),
	expr func(*Expr, int, int),
	depth, seq int,
) {
	switch p := p.GetIs().(type) {
	case *Predicate_And:
		for seq, a := range p.And.GetPred() {
			and(a, depth+1, seq)
		}
	case *Predicate_Or:
		for seq, o := range p.Or.GetPred() {
			or(o, depth+1, seq)
		}
	case *Predicate_Not:
		not(p.Not.GetPred(), depth+1, 0)
	case *Predicate_Expr:
		exprFn(p.Expr, depth, seq)
	default:
		// unreachable
	}
}

/*

/\ a=b
/\ \/ /\ age=12
      /\ gender=male
   \/ /\ age=12
      /\ gender=male
/\ c=d

*/
