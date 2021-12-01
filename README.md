# lenslocked-itah

This is Eli's implementation of the lenslocked project from calhoun.io.

Note that I'm using the PDF which was last edited in 2016 (?) not the videos which were 2018-ish so some of my docs are out of date.

## Hot Reloading

Note that for hot reloading i run `modd` instead of `go run main.go`. Installing Modd is `brew install modd` and is from the repo https://github.com/cortesi/modd. To enable modd you need a conf file that I have pasted herein. Note the modd.conf file in the project which I have git ignored.

```txt
**/*.go {
    prep: go test @dirmods
}

# Exclude all test files of the form *_test.go
**/*.go !**/*_test.go **/*.gohtml {
    prep: go build -o lenslocked .
    daemon +sigterm: ./lenslocked
}
```

