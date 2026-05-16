package auth

import (
	"context"
	"testing"
	"time"

	a "github.com/ix-pay/ixpay-pro/internal/infrastructure/security/auth"
	"github.com/stretchr/testify/assert"
)

type mockCache struct {
	data map[string]string
}

func newMockCache() *mockCache {
	return &mockCache{
		data: make(map[string]string),
	}
}

func (m *mockCache) Get(key string) (string, error) {
	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return "", nil
}

func (m *mockCache) Set(key string, value interface{}, expiration time.Duration) error {
	m.data[key] = value.(string)
	return nil
}

func (m *mockCache) Delete(key string) error {
	delete(m.data, key)
	return nil
}

func (m *mockCache) Exists(key string) (bool, error) {
	_, ok := m.data[key]
	return ok, nil
}

func (m *mockCache) Close() error {
	return nil
}

func newTestPermissionManager(t *testing.T) *a.PermissionManager {
	cache := newMockCache()
	log := &mockLogger{}

	pm := a.SetupPermissionManager(cache, log)
	assert.NotNil(t, pm)

	return pm
}

func TestSetupPermissionManager(t *testing.T) {
	cache := newMockCache()
	log := &mockLogger{}

	pm := a.SetupPermissionManager(cache, log)
	assert.NotNil(t, pm)
}

func TestCachePermissions(t *testing.T) {
	pm := newTestPermissionManager(t)

	permissions := []a.Permission{
		{
			Path:    "/api/users",
			Method:  "GET",
			Roles:   []string{"admin", "user"},
			Buttons: []string{"user:view"},
		},
		{
			Path:    "/api/users",
			Method:  "POST",
			Roles:   []string{"admin"},
			Buttons: []string{"user:create"},
		},
	}

	err := pm.CachePermissions(permissions)
	assert.NoError(t, err)
}

func TestGetPermission_FromCache(t *testing.T) {
	pm := newTestPermissionManager(t)

	permissions := []a.Permission{
		{
			Path:   "/api/roles",
			Method: "GET",
			Roles:  []string{"admin", "user"},
		},
	}

	pm.CachePermissions(permissions)

	perm, err := pm.GetPermission("GET", "/api/roles")
	assert.NoError(t, err)
	assert.NotNil(t, perm)
	assert.Equal(t, "/api/roles", perm.Path)
	assert.Equal(t, "GET", perm.Method)
	assert.Contains(t, perm.Roles, "admin")
	assert.Contains(t, perm.Roles, "user")
}

func TestGetPermission_NotInCache(t *testing.T) {
	pm := newTestPermissionManager(t)

	perm, err := pm.GetPermission("GET", "/api/nonexistent")
	if err != nil {
		// Permission not cached, returns default
		return
	}
	// If no error, verify structure
	assert.NotNil(t, perm)
	assert.Equal(t, "GET", perm.Method)
	assert.Equal(t, "/api/nonexistent", perm.Path)
}

func TestCheckPermission_AdminRole(t *testing.T) {
	pm := newTestPermissionManager(t)

	ctx := context.WithValue(context.Background(), "role", "admin")

	hasPermission := pm.CheckPermission(ctx, "GET", "/api/users")
	assert.True(t, hasPermission)
}

func TestCheckPermission_UserRole(t *testing.T) {
	pm := newTestPermissionManager(t)

	permissions := []a.Permission{
		{
			Path:   "/api/users",
			Method: "GET",
			Roles:  []string{"admin", "user"},
		},
	}

	pm.CachePermissions(permissions)

	ctx := context.WithValue(context.Background(), "role", "user")

	hasPermission := pm.CheckPermission(ctx, "GET", "/api/users")
	assert.True(t, hasPermission)
}

func TestCheckPermission_UserRoleDenied(t *testing.T) {
	pm := newTestPermissionManager(t)

	permissions := []a.Permission{
		{
			Path:   "/api/admin",
			Method: "GET",
			Roles:  []string{"admin"},
		},
	}

	pm.CachePermissions(permissions)

	ctx := context.WithValue(context.Background(), "role", "user")

	hasPermission := pm.CheckPermission(ctx, "GET", "/api/admin")
	assert.False(t, hasPermission)
}

func TestCheckPermission_NoRole(t *testing.T) {
	pm := newTestPermissionManager(t)

	ctx := context.Background()

	hasPermission := pm.CheckPermission(ctx, "GET", "/api/users")
	assert.False(t, hasPermission)
}

func TestCheckRolePermission(t *testing.T) {
	pm := newTestPermissionManager(t)

	permissions := []a.Permission{
		{
			Path:   "/api/data",
			Method: "GET",
			Roles:  []string{"admin", "editor"},
		},
	}

	pm.CachePermissions(permissions)

	hasPermission := pm.CheckRolePermission("admin", "GET", "/api/data")
	assert.True(t, hasPermission)

	hasPermission = pm.CheckRolePermission("editor", "GET", "/api/data")
	assert.True(t, hasPermission)

	hasPermission = pm.CheckRolePermission("viewer", "GET", "/api/data")
	assert.False(t, hasPermission)
}

func TestCheckButtonPermission(t *testing.T) {
	pm := newTestPermissionManager(t)

	ctx := context.WithValue(context.Background(), "userButtons", []string{"user:view", "user:edit"})

	hasPermission := pm.CheckButtonPermission(ctx, "user:view")
	assert.True(t, hasPermission)

	hasPermission = pm.CheckButtonPermission(ctx, "user:delete")
	assert.False(t, hasPermission)
}

func TestCheckButtonPermission_NoButtons(t *testing.T) {
	pm := newTestPermissionManager(t)

	ctx := context.Background()

	hasPermission := pm.CheckButtonPermission(ctx, "user:view")
	assert.False(t, hasPermission)
}

func TestCheckAPIPermissionWithButton_Admin(t *testing.T) {
	pm := newTestPermissionManager(t)

	permissions := []a.Permission{
		{
			Path:    "/api/users",
			Method:  "POST",
			Roles:   []string{"admin", "user"},
			Buttons: []string{"user:create"},
		},
	}

	pm.CachePermissions(permissions)

	ctx := context.WithValue(context.Background(), "role", "admin")

	hasPermission := pm.CheckAPIPermissionWithButton(ctx, "POST", "/api/users")
	assert.True(t, hasPermission)
}

func TestCheckAPIPermissionWithButton_UserWithButton(t *testing.T) {
	pm := newTestPermissionManager(t)

	permissions := []a.Permission{
		{
			Path:    "/api/users",
			Method:  "POST",
			Roles:   []string{"admin", "user"},
			Buttons: []string{"user:create"},
		},
	}

	pm.CachePermissions(permissions)

	ctx := context.WithValue(context.Background(), "role", "user")
	ctx = context.WithValue(ctx, "userButtons", []string{"user:create"})

	hasPermission := pm.CheckAPIPermissionWithButton(ctx, "POST", "/api/users")
	assert.True(t, hasPermission)
}

func TestCheckAPIPermissionWithButton_UserWithoutButton(t *testing.T) {
	pm := newTestPermissionManager(t)

	permissions := []a.Permission{
		{
			Path:    "/api/users",
			Method:  "POST",
			Roles:   []string{"admin", "user"},
			Buttons: []string{"user:create"},
		},
	}

	pm.CachePermissions(permissions)

	ctx := context.WithValue(context.Background(), "role", "user")
	ctx = context.WithValue(ctx, "userButtons", []string{"user:view"})

	hasPermission := pm.CheckAPIPermissionWithButton(ctx, "POST", "/api/users")
	assert.False(t, hasPermission)
}

func TestCheckAPIPermissionWithButton_NoButtonsRequired(t *testing.T) {
	pm := newTestPermissionManager(t)

	permissions := []a.Permission{
		{
			Path:    "/api/users",
			Method:  "GET",
			Roles:   []string{"admin", "user"},
			Buttons: []string{},
		},
	}

	pm.CachePermissions(permissions)

	ctx := context.WithValue(context.Background(), "role", "user")
	ctx = context.WithValue(ctx, "userButtons", []string{})

	hasPermission := pm.CheckAPIPermissionWithButton(ctx, "GET", "/api/users")
	assert.True(t, hasPermission)
}

func TestRefreshPermissions(t *testing.T) {
	pm := newTestPermissionManager(t)

	permissions := []a.Permission{
		{
			Path:   "/api/test",
			Method: "GET",
			Roles:  []string{"admin"},
		},
	}

	err := pm.CachePermissions(permissions)
	assert.NoError(t, err)

	newPermissions := []a.Permission{
		{
			Path:   "/api/test",
			Method: "GET",
			Roles:  []string{"admin", "user"},
		},
	}

	err = pm.RefreshPermissions(newPermissions)
	assert.NoError(t, err)

	perm, err := pm.GetPermission("GET", "/api/test")
	assert.NoError(t, err)
	assert.Contains(t, perm.Roles, "admin")
	assert.Contains(t, perm.Roles, "user")
}

func TestPermission_Struct(t *testing.T) {
	perm := a.Permission{
		Path:        "/api/test",
		Method:      "DELETE",
		Roles:       []string{"admin"},
		Buttons:     []string{"test:delete"},
		WechatGrant: false,
	}

	assert.Equal(t, "/api/test", perm.Path)
	assert.Equal(t, "DELETE", perm.Method)
	assert.Equal(t, []string{"admin"}, perm.Roles)
	assert.Equal(t, []string{"test:delete"}, perm.Buttons)
	assert.False(t, perm.WechatGrant)
}
