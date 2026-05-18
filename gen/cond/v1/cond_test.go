package condv1_test

import (
	condv1 "cond/gen/cond/v1"
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/protobuf/types/known/structpb"
)

func ExamplePredicate_Walk() {
	e := condv1.Eq(`Name`, `Tom`)

	predicate := condv1.NewAnd(
		condv1.NewAnd(
			condv1.NewAnd(e, e, e),
			e,
		),
		condv1.NewAnd(e, e),
		e,
		condv1.NewOr(e, e),
	)

	_ = predicate.Walk(func(ctx condv1.WalkContext, p *condv1.Predicate) (condv1.WalkAction, error) {
		if ctx.Event != condv1.WalkEnter {
			return condv1.WalkContinue, nil
		}

		if ctx.Relation != condv1.RelationRoot {
			if ctx.Index > 0 {
				fmt.Print(strings.Repeat(`   `, ctx.Depth-1))
			}
			switch ctx.Relation {
			case condv1.RelationAnd:
				fmt.Print(`/\ `)
			case condv1.RelationOr:
				fmt.Print(`\/ `)
			}
		}

		if expr := p.GetExpr(); expr != nil {
			fmt.Println(expr.Pretty())
		}
		return condv1.WalkContinue, nil
	})
	// Output:
	// /\ /\ /\ Name="Tom"
	//       /\ Name="Tom"
	//       /\ Name="Tom"
	//    /\ Name="Tom"
	// /\ /\ Name="Tom"
	//    /\ Name="Tom"
	// /\ Name="Tom"
	// /\ \/ Name="Tom"
	//    \/ Name="Tom"
}

func ExamplePredicate_Eval() {
	record := struct {
		Name      string
		Age       float64
		Suspended bool
	}{
		Name:      "Tom",
		Age:       18,
		Suspended: false,
	}

	predicate := condv1.NewAnd(
		condv1.Eq("Name", "Tom"),
		condv1.NewOr(
			condv1.Gt("Age", 21),
			condv1.Eq("Age", 18),
		),
		condv1.NewNot(condv1.Eq("Suspended", true)),
	)

	matched := predicate.Eval(func(e *condv1.Expr) bool {
		switch e.GetOp() {
		case condv1.Op_OP_EQ:
			switch e.GetField() {
			case "Name":
				return record.Name == e.GetValue().GetStringValue()
			case "Age":
				return record.Age == e.GetValue().GetNumberValue()
			case "Suspended":
				return record.Suspended == e.GetValue().GetBoolValue()
			}
		case condv1.Op_OP_GT:
			if e.GetField() == "Age" {
				return record.Age > e.GetValue().GetNumberValue()
			}
		}
		return false
	})

	fmt.Println(matched)
	// Output:
	// true
}

func ExampleFold_sql() {
	predicate := condv1.NewAnd(
		condv1.Eq("name", "Tom"),
		condv1.NewOr(
			condv1.Gt("age", 21),
			condv1.Eq("age", 18),
		),
		condv1.NewNot(condv1.Eq("suspended", true)),
	)

	sql, err := condv1.Fold(predicate, condv1.Folder[string]{
		Expr: exprSQL,
		And: func(parts []string) (string, error) {
			return "(" + strings.Join(parts, " AND ") + ")", nil
		},
		Or: func(parts []string) (string, error) {
			return "(" + strings.Join(parts, " OR ") + ")", nil
		},
		Not: func(part string) (string, error) {
			return "NOT (" + part + ")", nil
		},
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(sql)
	// Output:
	// (name = 'Tom' AND (age > 21 OR age = 18) AND NOT (suspended = TRUE))
}

func ExampleFold_sql_prarameters() {
	predicate := condv1.NewAnd(
		condv1.Eq("name", "Tom"),
		condv1.NewOr(
			condv1.Gt("age", 21),
			condv1.Eq("age", 18),
		),
		condv1.NewNot(condv1.Eq("suspended", true)),
	)

	params := []any{}
	sql, err := condv1.Fold(predicate, condv1.Folder[string]{
		Expr: func(e *condv1.Expr) (string, error) {
			params = append(params, sqlValue(e.GetValue()))
			return fmt.Sprintf(`%s %s ?`, e.GetField(), sqlOp(e.GetOp())), nil
		},
		And: func(parts []string) (string, error) {
			return "(" + strings.Join(parts, " AND ") + ")", nil
		},
		Or: func(parts []string) (string, error) {
			return "(" + strings.Join(parts, " OR ") + ")", nil
		},
		Not: func(part string) (string, error) {
			return "NOT (" + part + ")", nil
		},
	})

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(sql)
	fmt.Println(params)
	// Output:
	// (name = ? AND (age > ? OR age = ?) AND NOT (suspended = ?))
	// ['Tom' 21 18 TRUE]
}

func exprSQL(e *condv1.Expr) (string, error) {
	var sb strings.Builder
	sb.WriteString(e.GetField())
	sb.WriteByte(' ')
	sb.WriteString(sqlOp(e.GetOp()))
	sb.WriteByte(' ')
	writeSQLValue(&sb, e.GetValue())
	return sb.String(), nil
}

func sqlOp(op condv1.Op) string {
	switch op {
	case condv1.Op_OP_EQ:
		return "="
	case condv1.Op_OP_GT:
		return ">"
	default:
		return "?"
	}
}

func sqlValue(v *structpb.Value) string {
	switch v := v.Kind.(type) {
	case *structpb.Value_StringValue:
		var sb strings.Builder
		sb.WriteByte('\'')
		sb.WriteString(strings.ReplaceAll(v.StringValue, "'", "''"))
		sb.WriteByte('\'')
		return sb.String()
	case *structpb.Value_NumberValue:
		return strconv.FormatFloat(v.NumberValue, 'f', -1, 64)
	case *structpb.Value_BoolValue:
		if v.BoolValue {
			return `TRUE`
		}
		return `FALSE`
	default:
		return fmt.Sprint(v)
	}
}

func writeSQLValue(sb *strings.Builder, value *structpb.Value) {
	switch v := value.GetKind().(type) {
	case *structpb.Value_StringValue:
		sb.WriteByte('\'')
		sb.WriteString(strings.ReplaceAll(v.StringValue, "'", "''"))
		sb.WriteByte('\'')
	case *structpb.Value_NumberValue:
		sb.WriteString(strconv.FormatFloat(v.NumberValue, 'f', -1, 64))
	case *structpb.Value_BoolValue:
		if v.BoolValue {
			sb.WriteString("TRUE")
			return
		}
		sb.WriteString("FALSE")
	default:
		sb.WriteString("NULL")
	}
}
