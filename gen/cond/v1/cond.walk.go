package condv1

import "errors"

type PredicateKind int

const (
	KindInvalid PredicateKind = iota
	KindAnd
	KindOr
	KindNot
	KindExpr
)

type WalkEvent int

const (
	WalkEnter WalkEvent = iota
	WalkLeave
)

type WalkRelation int

const (
	RelationRoot WalkRelation = iota
	RelationAnd
	RelationOr
	RelationNot
)

type WalkAction int

const (
	WalkContinue WalkAction = iota
	WalkSkipChildren
	WalkStop
)

type WalkContext struct {
	Event    WalkEvent
	Depth    int
	Index    int
	Relation WalkRelation
	Parent   *Predicate
}

type WalkFunc func(WalkContext, *Predicate) (WalkAction, error)

type WalkExprFunc func(WalkContext, *Expr) error

var errWalkStop = errors.New("walk stopped")

func (p *Predicate) Kind() PredicateKind {
	if p == nil {
		return KindInvalid
	}

	switch p.GetIs().(type) {
	case *Predicate_And:
		return KindAnd
	case *Predicate_Or:
		return KindOr
	case *Predicate_Not:
		return KindNot
	case *Predicate_Expr:
		return KindExpr
	default:
		return KindInvalid
	}
}

func (p *Predicate) Walk(fn WalkFunc) error {
	if p == nil || fn == nil {
		return nil
	}

	err := p.walk(WalkContext{
		Event:    WalkEnter,
		Relation: RelationRoot,
	}, fn)
	if errors.Is(err, errWalkStop) {
		return nil
	}
	return err
}

func (p *Predicate) WalkExpr(fn WalkExprFunc) error {
	if fn == nil {
		return nil
	}

	return p.Walk(func(ctx WalkContext, p *Predicate) (WalkAction, error) {
		if ctx.Event != WalkEnter {
			return WalkContinue, nil
		}

		expr := p.GetExpr()
		if expr == nil {
			return WalkContinue, nil
		}

		return WalkContinue, fn(ctx, expr)
	})
}

func (p *Predicate) walk(ctx WalkContext, fn WalkFunc) error {
	if p == nil {
		return nil
	}

	ctx.Event = WalkEnter
	action, err := fn(ctx, p)
	if err != nil {
		return err
	}

	switch action {
	case WalkStop:
		return errWalkStop
	case WalkSkipChildren:
		return nil
	}

	switch node := p.GetIs().(type) {
	case *Predicate_And:
		for i, a := range node.And.GetPred() {
			if err = a.walk(
				WalkContext{
					Depth:    ctx.Depth + 1,
					Index:    i,
					Relation: RelationAnd,
					Parent:   p,
				},
				fn,
			); err != nil {
				return err
			}
		}
	case *Predicate_Or:
		for i, o := range node.Or.GetPred() {
			if err = o.walk(
				WalkContext{
					Depth:    ctx.Depth + 1,
					Index:    i,
					Relation: RelationOr,
					Parent:   p,
				},
				fn,
			); err != nil {
				return err
			}
		}
	case *Predicate_Not:
		if err := node.Not.GetPred().walk(WalkContext{
			Depth:    ctx.Depth + 1,
			Relation: RelationNot,
			Parent:   p,
		}, fn); err != nil {
			return err
		}
	}

	ctx.Event = WalkLeave
	action, err = fn(ctx, p)
	if err != nil {
		return err
	}
	if action == WalkStop {
		return errWalkStop
	}
	return nil
}
