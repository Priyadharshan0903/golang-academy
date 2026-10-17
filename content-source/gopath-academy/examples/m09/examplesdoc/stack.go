// Package examplesdoc shows testable examples as documentation (lesson 9.3).
package examplesdoc

// Stack is a LIFO stack of strings.
type Stack struct{ items []string }

// Push adds v to the top.
func (s *Stack) Push(v string) { s.items = append(s.items, v) }

// Pop removes and returns the top element, or "" and false when empty.
func (s *Stack) Pop() (string, bool) {
	if len(s.items) == 0 {
		return "", false
	}
	v := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return v, true
}

// Len returns the number of elements.
func (s *Stack) Len() int { return len(s.items) }
