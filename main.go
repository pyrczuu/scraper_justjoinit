package main

import (
	"context"
	"fmt"
)

func main() {
	ctx := context.Background()

	urls := CollectJustJoinIt(ctx)

	for _, url := range urls {
		fmt.Println(url)
	}
}
