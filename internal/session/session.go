package session

import "sync"

type State int

const (
	StateIdle State = iota
	StateWaitURL
	StateWaitVideoName
)

type Manager struct {
	mu    sync.RWMutex
	users map[string]State
}

func New() *Manager {
	return &Manager{users: make(map[string]State)}
}

func (m *Manager) Login(user string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[user] = StateIdle
}

func (m *Manager) Logout(user string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.users, user)
}

func (m *Manager) HasUser(user string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.users[user]
	return ok
}

func (m *Manager) State(user string) State {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.users[user]
}

func (m *Manager) SetState(user string, state State) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.users[user]; !ok {
		return false
	}
	m.users[user] = state
	return true
}

func (m *Manager) Reset(user string) {
	m.SetState(user, StateIdle)
}

func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.users)
}
