package main

import (
	"html/template"
	"os"
)

func main() {
	t, err := template.ParseFiles("hello.gohtml")
	if err != nil {
		panic(err)
	}

	m := map[string]string{"Head": "Eyes", "Body": "Tail"}

	data := struct {
		Name  string
		AI    string
		Age   int
		Fishy map[string]string
	}{"<script>alert('Howdy!');</script>", "Bob", 4, m}

	err = t.Execute(os.Stdout, data)
	if err != nil {
		panic(err)
	}
}
