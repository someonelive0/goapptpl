package main

import (
	"fmt"

	"github.com/brianvoe/gofakeit/v7"
)

func main() {
	name := gofakeit.Name()
	email := gofakeit.Email()
	color := gofakeit.Color()
	fmt.Println(name, email, color)
}
