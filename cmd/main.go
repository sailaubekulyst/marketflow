package main

import (
	"fmt"

	"marketflow/internal/application"
)

func main() {
	app, err := application.GetApp()
	if err != nil {
		fmt.Println("failed to get app")
		return
	}
	app.Run()
}
