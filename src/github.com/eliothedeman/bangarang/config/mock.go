package config

type MockProvider struct {
	snapshots map[string]*Snapshot
	users     map[string]*User
}

func NewMockProvider() Provider {
	return &MockProvider{
		snapshots: make(map[string]*Snapshot),
	}
}

func (m *MockProvider) GetConfig(version string) (*AppConfig, error) {

	s, err := m.getSnapshot(version)

	return s.App, err
}

func (m *MockProvider) GetCurrent() (*AppConfig, error) {
	return m.GetConfig(CurrentVersionHash)
}

func (m *MockProvider) PutConfig(c *AppConfig, u *User) (string, error) {

	// check to see if the given user has write permissions
	if u.Permissions < WRITE {
		return "", InsufficientPermissions(WRITE, u.Permissions)
	}

	old, _ := m.getSnapshot(CurrentVersionHash)
	if old != nil {

		m.snapshots[old.Hash] = old
	}

	// put the current config in as "current"
	s := newSnapshot(c, u)
	m.snapshots[CurrentVersionHash] = s

	return s.Hash, nil
}

func (m *MockProvider) getSnapshot(hash string) (*Snapshot, error) {
	c, ok := m.snapshots[hash]
	if !ok {
		return nil, ConfigVersionNotFound(hash)
	}

	return c, nil
}

func (m *MockProvider) ListSnapshots() []*Snapshot {
	snaps := make([]*Snapshot, 0, len(m.snapshots))
	for _, s := range m.snapshots {
		snaps = append(snaps, s)
	}

	return snaps
}

func (m *MockProvider) GetUser(name string) (*User, error) {
	u, ok := m.users[name]
	if !ok {
		return nil, UserNotFound(name)
	}

	return u, nil
}

func (m *MockProvider) GetUserByUserName(name string) (*User, error) {

	for _, u := range m.users {
		if u.UserName == name {
			return u, nil
		}
	}

	return nil, UserNotFound(name)
}

func (m *MockProvider) DeleteUser(name string) error {
	delete(m.users, name)

	return nil
}

func (m *MockProvider) PutUser(u *User) error {
	m.users[u.UserName] = u
	return nil
}

func (m *MockProvider) ListUsers() ([]*User, error) {
	users := make([]*User, 0, len(m.users))

	for _, u := range m.users {
		users = append(users, u)
	}

	return users, nil
}
