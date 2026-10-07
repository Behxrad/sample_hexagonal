package main

import (
	"code_base/internal/application"
	"fmt"
	"os"
)

func main() {
	app := application.NewCommerceApp()
	err := app.Start()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
