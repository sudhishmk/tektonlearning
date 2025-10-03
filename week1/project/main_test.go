package main
import "testing"
func TestMessage(t *testing.T) {
    if Message() != "Hello, CI!" {
        t.Errorf("Message() = %s; want Hello, CI!", Message())
    }
}
