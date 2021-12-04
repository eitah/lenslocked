package main

import (
	"fmt"

	"github.com/eitah/lenslocked/src/lenslocked.com/rand"
)

func main() {
	fmt.Println(rand.String(10))
	fmt.Println(rand.RememberToken())
}
