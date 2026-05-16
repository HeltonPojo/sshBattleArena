package server

import (
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

const MaxPlayers = 2

// Registry tracks connected players and their tea.Programs.
type Registry struct {
	mu       sync.Mutex
	programs map[string]*tea.Program
}

func NewRegistry() *Registry {
	return &Registry{
		programs: make(map[string]*tea.Program),
	}
}

func (r *Registry) TryRegister(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.programs) >= MaxPlayers {
		return false
	}
	r.programs[id] = nil // slot reserved, program set later
	return true
}

func (r *Registry) SetProgram(id string, p *tea.Program) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.programs[id]; ok {
		r.programs[id] = p
	}
}

func (r *Registry) Remove(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.programs, id)
}

func (r *Registry) Count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.programs)
}

func (r *Registry) GetProgram(id string) *tea.Program {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.programs[id]
}

func (r *Registry) PlayerIDs() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	ids := make([]string, 0, len(r.programs))
	for id := range r.programs {
		ids = append(ids, id)
	}
	return ids
}
