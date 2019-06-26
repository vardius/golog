/*
Package golog provides simple logger
Basic example:
	package main

	import (
		"context"

		"github.com/vardius/golog"
	)

	func main() {
		ctx := context.Background()
		logger := golog.New()

		logger.Debug(ctx context.Context, "Hello %s!", "you")
	}
*/
package golog
