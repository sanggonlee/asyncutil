# asyncutil

[![Go Reference](https://pkg.go.dev/badge/github.com/sanggonlee/asyncutil.svg)](https://pkg.go.dev/github.com/sanggonlee/asyncutil)
[![Go Report Card](https://goreportcard.com/badge/github.com/sanggonlee/asyncutil)](https://goreportcard.com/report/github.com/sanggonlee/asyncutil)

A collection of utilities for concurrent programming in Go.

## Usage

[godoc](https://pkg.go.dev/github.com/sanggonlee/asyncutil)

### Collect

```go
func someWork(id int) chan error {
    errs := make(chan error)
    go func(result *Result) {
        // ...
        errs <- err
    }(&Result{ID: id})
    return errs
}

for err := range asyncutil.Collect(
    someWork(1),
    someWork(2),
) {
    if err != nil {
        fmt.Println("Error:", err)
    }
}
```

## Benchmarks

```
$ go test -bench . -count 1 .
goos: darwin
goarch: amd64
pkg: github.com/sanggonlee/asyncutil
```

### Collect

5 functions running, each taking 50 milliseconds:

```
BenchmarkCollect_5_functions_of_50_milliseconds_each-8                21          52155152 ns/op
BenchmarkSequential_5_functions_of_50_milliseconds_each-8              4         262524126 ns/op
```

2 functions running, each taking 30 milliseconds:

```
BenchmarkCollect_2_functions_of_30_milliseconds_each-8                38          32424379 ns/op
BenchmarkSequential_2_functions_of_30_milliseconds_each-8             18          64462037 ns/op
```

5 functions running, each taking 200 milliseconds:

```
BenchmarkCollect_5_functions_of_200_milliseconds_each-8                5         202613911 ns/op
BenchmarkSequential_5_functions_of_200_milliseconds_each-8             1        1012167372 ns/op
```
