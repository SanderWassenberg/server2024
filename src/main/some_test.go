package main


import (
    "fmt"
    "testing"
)

// TestHelloName calls greetings.Hello with a name, checking
// for a valid return value.
func TestNothing(t *testing.T) {
    fmt.Println("hello from a test!")
    fmt.Println("Deze test is", t.Name())
}
