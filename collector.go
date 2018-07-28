package main

// Collector interface is implemented by collectors; these are
// what takes a loadtest.Output and writes it to some storage
// somewhere.
type Collector interface {
	Push(OutputMapper) error
}
