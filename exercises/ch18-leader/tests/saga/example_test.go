package saga_test

import (
	"context"
	"errors"
	"fmt"

	"example.com/ch18-leader/saga"
)

func ExampleRun() {
	step := func(name string, fail bool) saga.Step {
		return saga.Step{
			Name: name,
			Do: func(context.Context) error {
				fmt.Println("do", name)
				if fail {
					return errors.New("card declined")
				}
				return nil
			},
			Compensate: func(context.Context) error { fmt.Println("undo", name); return nil },
		}
	}
	err := saga.Run(context.Background(), []saga.Step{
		step("reserve-stock", false),
		step("book-courier", false),
		step("charge-card", true),
	})
	fmt.Println(err)
	// Output:
	// do reserve-stock
	// do book-courier
	// do charge-card
	// undo book-courier
	// undo reserve-stock
	// saga: step "charge-card" failed: card declined; compensated [book-courier reserve-stock]
}
