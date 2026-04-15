package condv1_test

import (
	"fmt"
	"testing"

	condv1 "cond/gen/cond/v1"

	"google.golang.org/protobuf/types/known/structpb"
)

func TestPretty(t *testing.T) {
	cases := []struct {
		p      *condv1.Predicate
		expect string
		desc   string
	}{{
		p: &condv1.Predicate{
			Is: &condv1.Predicate_And{&condv1.And{
				Pred: []*condv1.Predicate{{
					Is: &condv1.Predicate_Expr{&condv1.Expr{
						Field: `Name`,
						Value: structpb.NewStringValue(`Tom`),
					}},
				}},
			}},
		},
		expect: `Name="Tom"`,
	}, {
		p: &condv1.Predicate{
			Is: &condv1.Predicate_And{&condv1.And{
				Pred: []*condv1.Predicate{{
					Is: &condv1.Predicate_Expr{&condv1.Expr{
						Field: `Name`,
						Value: structpb.NewStringValue(`Tom*`),
						Op:    condv1.Op_OP_MATCH,
					}},
				}, {
					Is: &condv1.Predicate_Expr{&condv1.Expr{
						Field: `Age`,
						Value: structpb.NewNumberValue(16),
						Op:    condv1.Op_OP_LT,
					}},
				}, {
					Is: &condv1.Predicate_Expr{&condv1.Expr{
						Field: `Gender`,
						Value: structpb.NewStringValue(`Male`),
					}},
				}},
			}},
		},
		expect: `Name~"Tom*" && Age<16 && Gender="Male"`,
		desc:   `multiple expressions`,
	}, {
		p: &condv1.Predicate{
			Is: &condv1.Predicate_And{&condv1.And{
				Pred: []*condv1.Predicate{{
					Is: &condv1.Predicate_Or{&condv1.Or{
						Pred: []*condv1.Predicate{{
							Is: &condv1.Predicate_Expr{Expr: &condv1.Expr{
								Field: `Name`,
								Value: structpb.NewStringValue(`Tom`),
							}},
						}, {
							Is: &condv1.Predicate_Expr{Expr: &condv1.Expr{
								Field: `Name`,
								Value: structpb.NewStringValue(`Marry`),
							}},
						}},
					}},
				}, {
					Is: &condv1.Predicate_Expr{&condv1.Expr{
						Field: `Age`,
						Value: structpb.NewNumberValue(16),
						Op:    condv1.Op_OP_LT,
					}},
				}, {
					Is: &condv1.Predicate_Expr{&condv1.Expr{
						Field: `Gender`,
						Value: structpb.NewStringValue(`Male`),
					}},
				}},
			}},
		},
		expect: `(Name="Tom" || Name="Marry") && Age<16 && Gender="Male"`,
		desc:   `nested expressions`,
	}, {
		p: &condv1.Predicate{
			Is: &condv1.Predicate_And{&condv1.And{
				Pred: []*condv1.Predicate{{
					Is: &condv1.Predicate_Expr{&condv1.Expr{
						Field: `Name`,
						Value: structpb.NewStringValue(`Tom*`),
						Op:    condv1.Op_OP_MATCH,
					}},
				}, {
					Is: &condv1.Predicate_Expr{&condv1.Expr{
						Field: `Age`,
						Value: structpb.NewNumberValue(16),
						Op:    condv1.Op_OP_LT,
					}},
				}, {
					Is: &condv1.Predicate_Not{&condv1.Not{
						Pred: &condv1.Predicate{
							Is: &condv1.Predicate_Or{&condv1.Or{
								Pred: []*condv1.Predicate{{
									Is: &condv1.Predicate_Expr{&condv1.Expr{
										Field: `Profession`,
										Value: structpb.NewStringValue(`Doctor`),
									}},
								}, {
									Is: &condv1.Predicate_Expr{&condv1.Expr{
										Field: `Profession`,
										Value: structpb.NewStringValue(`Firefighter`),
									}},
								}, {
									Is: &condv1.Predicate_Expr{&condv1.Expr{
										Field: `Profession`,
										Value: structpb.NewStringValue(`Singer`),
									}},
								}},
							}},
						},
					}},
				}},
			}},
		},
		expect: `Name~"Tom*" && Age<16 && !(Profession="Doctor" || Profession="Firefighter" || Profession="Singer")`,
		desc:   `not expressions`,
	}}

	for i, tc := range cases {
		t.Run(fmt.Sprintf(`%d_%s`, i, tc.desc), func(t *testing.T) {
			if actual := tc.p.Pretty(); actual != tc.expect {
				t.Errorf("\nexpect: %s\nactual: %s", tc.expect, actual)
			}
		})
	}
}
