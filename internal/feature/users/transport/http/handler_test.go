package httptransport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	domainusers "github.com/hachisocial/hachisocial/internal/domain/users"
	"github.com/hachisocial/hachisocial/internal/feature/users/application"
	"github.com/hachisocial/hachisocial/internal/platform/httpserver"
)

type fakeUseCase struct {
	current application.User
}

func (f fakeUseCase) GetCurrent(context.Context, domainusers.ID) (application.User, error) {
	return f.current, nil
}

func (f fakeUseCase) GetPublicProfile(
	context.Context,
	string,
) (application.PublicProfile, error) {
	return application.PublicProfile{}, application.ErrUserNotFound
}

func (f fakeUseCase) UpdateProfile(
	context.Context,
	domainusers.ID,
	application.UpdateProfileCommand,
) (application.User, error) {
	return f.current, nil
}

func (f fakeUseCase) SetStatus(
	context.Context,
	domainusers.ID,
	domainusers.ID,
	string,
) error {
	return nil
}

type fixedPrincipal struct {
	principal Principal
}

func (p fixedPrincipal) Principal(context.Context) (Principal, error) {
	return p.principal, nil
}

func TestGetCurrentUser(t *testing.T) {
	t.Parallel()

	id, _ := domainusers.ParseID("550e8400-e29b-41d4-a716-446655440000")
	handler := NewHandler(
		fakeUseCase{current: application.User{ID: id.String(), Username: "alice"}},
		fixedPrincipal{principal: Principal{UserID: id, Role: domainusers.RoleUser}},
	)

	request := httptest.NewRequest(http.MethodGet, "/me", nil)
	response := httptest.NewRecorder()
	httpserver.RequestIDMiddleware(handler.Routes()).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"username":"alice"`) {
		t.Fatalf("unexpected response: %s", response.Body.String())
	}
}

func TestUpdateProfileRejectsUnknownField(t *testing.T) {
	t.Parallel()

	id, _ := domainusers.ParseID("550e8400-e29b-41d4-a716-446655440000")
	handler := NewHandler(
		fakeUseCase{},
		fixedPrincipal{principal: Principal{UserID: id, Role: domainusers.RoleUser}},
	)

	request := httptest.NewRequest(
		http.MethodPatch,
		"/me",
		strings.NewReader(`{"unknown":"value"}`),
	)
	response := httptest.NewRecorder()
	httpserver.RequestIDMiddleware(handler.Routes()).ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", response.Code, response.Body.String())
	}
}

func TestGetCurrentUserRequiresPrincipal(t *testing.T) {
	t.Parallel()

	handler := NewHandler(fakeUseCase{}, DenyPrincipalProvider{})
	request := httptest.NewRequest(http.MethodGet, "/me", nil)
	response := httptest.NewRecorder()
	httpserver.RequestIDMiddleware(handler.Routes()).ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", response.Code, response.Body.String())
	}
}

func TestStaticPrincipalProvider(t *testing.T) {
	t.Parallel()

	id, _ := domainusers.ParseID("11111111-1111-4111-8111-111111111111")
	provider := NewStaticPrincipalProvider(id, domainusers.RoleUser)

	principal, err := provider.Principal(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if principal.UserID != id || principal.Role != domainusers.RoleUser {
		t.Fatalf("unexpected principal: %+v", principal)
	}
}
