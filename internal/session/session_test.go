package session

import (
	"sync"
	"testing"
)

func TestLoginLogout(t *testing.T) {
	t.Parallel()
	m := New()
	if m.HasUser("alice") {
		t.Fatal("user should not exist before login")
	}
	m.Login("alice")
	if !m.HasUser("alice") {
		t.Fatal("user should exist after login")
	}
	if got := m.State("alice"); got != StateIdle {
		t.Errorf("state = %v, want %v", got, StateIdle)
	}
	m.Logout("alice")
	if m.HasUser("alice") {
		t.Fatal("user should not exist after logout")
	}
}

func TestSetStateRequiresLogin(t *testing.T) {
	t.Parallel()
	m := New()
	if ok := m.SetState("bob", StateWaitURL); ok {
		t.Fatal("SetState should fail for unknown user")
	}
	m.Login("bob")
	if ok := m.SetState("bob", StateWaitURL); !ok {
		t.Fatal("SetState should succeed for known user")
	}
	if got := m.State("bob"); got != StateWaitURL {
		t.Errorf("state = %v, want %v", got, StateWaitURL)
	}
	m.Reset("bob")
	if got := m.State("bob"); got != StateIdle {
		t.Errorf("after reset = %v, want %v", got, StateIdle)
	}
}

func TestConcurrentAccess(t *testing.T) {
	t.Parallel()
	m := New()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			m.Login("u")
			m.SetState("u", StateWaitURL)
		}(i)
		go func() {
			defer wg.Done()
			_ = m.HasUser("u")
			_ = m.State("u")
			_ = m.Count()
		}()
	}
	wg.Wait()
}
