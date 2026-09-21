package auth

import (
	"fmt"
	"sync"
	"time"

	"rift/internal/model"
	"rift/internal/registry"
)

type Store struct {
	mu    sync.RWMutex
	users map[string]*model.User // by ID
	byName map[string]string     // username -> ID
	perms []model.TwinPermission
}

func NewStore() *Store {
	return &Store{users: map[string]*model.User{}, byName: map[string]string{}}
}

func (s *Store) CreateUser(orgID, username, displayName, password string, role model.Role) (*model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.byName[username]; exists {
		return nil, fmt.Errorf("username %s already exists", username)
	}
	hash, salt := HashPassword(password)
	u := &model.User{
		ID: registry.NewID("user"), OrgID: orgID, Username: username, DisplayName: displayName,
		PasswordHash: hash, PasswordSalt: salt, Role: role, CreatedAt: time.Now().UTC(),
	}
	s.users[u.ID] = u
	s.byName[username] = u.ID
	return u, nil
}

func (s *Store) Authenticate(username, password string) (*model.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.byName[username]
	if !ok {
		return nil, fmt.Errorf("invalid credentials")
	}
	u := s.users[id]
	if !VerifyPassword(password, u.PasswordHash, u.PasswordSalt) {
		return nil, fmt.Errorf("invalid credentials")
	}
	return u, nil
}

func (s *Store) GetUser(id string) (*model.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	return u, ok
}

func (s *Store) ListUsers(orgID string) []*model.User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*model.User, 0)
	for _, u := range s.users {
		if u.OrgID == orgID {
			out = append(out, u)
		}
	}
	return out
}

// Grant implements the Twin Permission Model: e.g. a read-only grant for
// Team A on Factory X, scenario-edit rights for Team B, and actuator control
// for the Maintenance team on specific entities only.
func (s *Store) Grant(p model.TwinPermission) model.TwinPermission {
	s.mu.Lock()
	defer s.mu.Unlock()
	p.ID = registry.NewID("perm")
	s.perms = append(s.perms, p)
	return p
}

// Can answers the ABAC-style question "does this user/role hold `scope` on
// this twin (optionally scoped further to one entity)?". Admin role always
// passes; otherwise an explicit grant is required (default-deny).
func (s *Store) Can(userID, twinID, entityID string, scope model.ScopeAction) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[userID]
	if ok && u.Role == model.RoleAdmin {
		return true
	}
	for _, p := range s.perms {
		if p.TwinID != twinID {
			continue
		}
		if p.EntityID != "" && p.EntityID != entityID {
			continue
		}
		if p.UserID != "" && p.UserID != userID {
			continue
		}
		if p.Role != "" && (!ok || p.Role != u.Role) {
			continue
		}
		for _, sc := range p.Scopes {
			if sc == scope || sc == model.ScopeAdmin {
				return true
			}
		}
	}
	return false
}
