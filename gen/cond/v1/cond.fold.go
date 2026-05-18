package condv1

import "errors"

// Folder combines each subtree into a value.
type Folder[T any] struct {
	Expr func(*Expr) (T, error)
	And  func([]T) (T, error)
	Or   func([]T) (T, error)
	Not  func(T) (T, error)
}

// Fold reduces a predicate tree from the leaves up.
func Fold[T any](p *Predicate, folder Folder[T]) (T, error) {
	return fold(p, folder)
}

func fold[T any](p *Predicate, folder Folder[T]) (T, error) {
	var zero T
	if p == nil {
		return zero, errors.New("condv1: cannot fold nil predicate")
	}

	switch node := p.GetIs().(type) {
	case *Predicate_Expr:
		if folder.Expr == nil {
			return zero, errors.New("condv1: missing Expr folder")
		}
		return folder.Expr(node.Expr)

	case *Predicate_And:
		if folder.And == nil {
			return zero, errors.New("condv1: missing And folder")
		}

		parts, err := foldPredicates(node.And.GetPred(), folder)
		if err != nil {
			return zero, err
		}
		return folder.And(parts)

	case *Predicate_Or:
		if folder.Or == nil {
			return zero, errors.New("condv1: missing Or folder")
		}

		parts, err := foldPredicates(node.Or.GetPred(), folder)
		if err != nil {
			return zero, err
		}
		return folder.Or(parts)

	case *Predicate_Not:
		if folder.Not == nil {
			return zero, errors.New("condv1: missing Not folder")
		}

		part, err := fold(node.Not.GetPred(), folder)
		if err != nil {
			return zero, err
		}
		return folder.Not(part)

	default:
		return zero, errors.New("condv1: cannot fold invalid predicate")
	}
}

func foldPredicates[T any](predicates []*Predicate, folder Folder[T]) ([]T, error) {
	parts := make([]T, 0, len(predicates))
	for _, predicate := range predicates {
		part, err := fold(predicate, folder)
		if err != nil {
			return nil, err
		}
		parts = append(parts, part)
	}
	return parts, nil
}
