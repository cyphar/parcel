package main

import (
	"fmt"
	"os"

	"github.com/cyphar/parcel/discovery"
)

func main() {
	for _, arg := range os.Args[1:] {
		name, err := discovery.Resolve(arg)
		fmt.Printf("%s -> (%s, %v)\n", arg, name, err)
		_, disco, err := discovery.Fetch(name)
		fmt.Printf("%s -> %v -- %#v\n", name, err, disco)

		// The full monte.
		_, distribution, err := discovery.Discover(arg, nil)
		fmt.Printf("%s -> (%s, %v)\n", arg, distribution, err)
	}
}
