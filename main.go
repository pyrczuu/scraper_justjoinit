package main

import (
	"context"
	"fmt"
	"log"
)

func main() {
	ctx := context.Background()

	urls, err := scrollAndRead(ctx)
	if err != nil {
		log.Println(err)
	}

	for _, url := range urls {
		fmt.Println(url)
	}
}
