package main

import "fmt"

func main() {
	newState, action := Transition("active", "idle5m")
	fmt.Printf("newState=%s, action=%s\n", newState, action)
}
