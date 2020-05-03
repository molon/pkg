package retry

import (
	"log"
	"testing"

	"context"

	"github.com/molon/pkg/errors"
)

func TestRetry(t *testing.T) {
	// r := New()
	// r.delay = 88 * time.Second
	// log.Printf("%p %#v", r, r)
	// r.Do(context.Background(), "haha",
	// 	func(ctx context.Context) error {
	// 		return nil
	// 	},
	// 	WithDelay(66*time.Second),
	// )
	// log.Printf("%p %#v", r, r)

	ctx := context.Background()
	err := Do(ctx, "MainFlow",
		func(ctx context.Context, idx int) error {
			return errors.New("tmp")
		},
		WithFix(func(ctx context.Context, name string, idx int, err error) error {
			log.Println(err)
			return nil
		}),
		WithN(10),
	)
	if err != nil {
		log.Println("FINAL:", err)
	}
}
