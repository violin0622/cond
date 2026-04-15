package condv1

func Example_walk() {
	e := Eq(`Name`, `Tom`)

	a3 := NewAnd(
		NewAnd(
			NewAnd(e, e, e),
		),
		NewAnd(e, e),
		e,
		NewOr(e, e),
	)

	a3.Walk(andFn, orFn, notFn, exprFn, 0, 0)
	// Output:
	// /\ /\ /\ Name="Tom"
	//       /\ Name="Tom"
	//       /\ Name="Tom"
	// /\ /\ Name="Tom"
	//    /\ Name="Tom"
	// /\ Name="Tom"
	// /\ \/ Name="Tom"
	//    \/ Name="Tom"
}
