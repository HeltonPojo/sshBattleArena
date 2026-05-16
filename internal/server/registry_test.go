package server

import (
	"fmt"
	"sync"
	"testing"
)

func TestRegistryCap(t *testing.T) {
	r := NewRegistry()
	if !r.TryRegister("a") {
		t.Fatal("first register should succeed")
	}
	if !r.TryRegister("b") {
		t.Fatal("second register should succeed")
	}
	if r.TryRegister("c") {
		t.Fatal("third register should fail (cap=2)")
	}
	if r.Count() != 2 {
		t.Errorf("expected count=2, got %d", r.Count())
	}

	r.Remove("a")
	if r.Count() != 1 {
		t.Errorf("expected count=1 after remove, got %d", r.Count())
	}
	if !r.TryRegister("c") {
		t.Fatal("register should succeed after remove")
	}
}

func TestRegistryConcurrent(t *testing.T) {
	r := NewRegistry()
	var wg sync.WaitGroup
	results := make(chan bool, 100)

	for i := range 100 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			ok := r.TryRegister(fmt.Sprintf("p%d", id))
			results <- ok
		}(i)
	}
	wg.Wait()
	close(results)

	accepted := 0
	for ok := range results {
		if ok {
			accepted++
		}
	}
	if accepted != MaxPlayers {
		t.Errorf("expected exactly %d accepted, got %d", MaxPlayers, accepted)
	}
}
