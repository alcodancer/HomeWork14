package service_test

import (
	"context"
	"errors"
	"testing"

	"example/internal/domain"
	"example/internal/repository/fakerepo"
	"example/internal/service"
)

func newSvc() (*service.UserService, *fakerepo.FakeUserRepo, *fakerepo.FakeHistoryRepo) {
	ur := fakerepo.NewFakeUserRepo()
	hr := fakerepo.NewFakeHistoryRepo()
	return service.NewUserService(ur, hr), ur, hr
}

func TestCreateUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		prep      func(*fakerepo.FakeUserRepo)
		inName    string
		inEmail   string
		wantErr   error
		wantName  string
		wantEmail string
	}{
		{name: "ok", inName: "Alice", inEmail: "Alice@Example.com", wantName: "Alice", wantEmail: "alice@example.com"},
		{name: "empty name", inName: "  ", inEmail: "a@b.com", wantErr: domain.ErrInvalidName},
		{name: "invalid email", inName: "Bob", inEmail: "nope", wantErr: domain.ErrInvalidEmail},
		{
			name: "duplicate email",
			prep: func(r *fakerepo.FakeUserRepo) {
				_ = r.Create(context.Background(), &domain.User{Name: "Old", Email: "dup@x.com"})
			},
			inName: "New", inEmail: "dup@x.com", wantErr: domain.ErrUserExists,
		},
		{
			name: "repo create error",
			prep: func(r *fakerepo.FakeUserRepo) { r.ForceCreateErr = errors.New("db down") },
			inName: "Alice", inEmail: "a@b.com", wantErr: errors.New("db down"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc, repo, hist := newSvc()
			if tt.prep != nil {
				tt.prep(repo)
			}

			got, err := svc.CreateUser(context.Background(), tt.inName, tt.inEmail)

			if tt.wantErr != nil {
				if err == nil || err.Error() != tt.wantErr.Error() {
					t.Fatalf("want %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if got.Name != tt.wantName || got.Email != tt.wantEmail {
				t.Errorf("got (%q,%q), want (%q,%q)", got.Name, got.Email, tt.wantName, tt.wantEmail)
			}
			if len(hist.Entries[got.ID.Hex()]) != 1 {
				t.Errorf("expected 1 history entry, got %d", len(hist.Entries[got.ID.Hex()]))
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		prep    func(*fakerepo.FakeUserRepo) string
		wantErr error
	}{
		{
			name: "ok",
			prep: func(r *fakerepo.FakeUserRepo) string {
				u := &domain.User{Name: "A", Email: "a@x.com"}
				_ = r.Create(context.Background(), u)
				return u.ID.Hex()
			},
		},
		{name: "empty id", prep: func(*fakerepo.FakeUserRepo) string { return "" }, wantErr: domain.ErrUserNotFound},
		{name: "not found", prep: func(*fakerepo.FakeUserRepo) string { return "ffffffffffffffffffffffff" }, wantErr: domain.ErrUserNotFound},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc, repo, _ := newSvc()
			id := tt.prep(repo)

			_, err := svc.GetUser(context.Background(), id)
			if tt.wantErr != nil {
				if err == nil || !errors.Is(err, tt.wantErr) {
					t.Fatalf("want %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
		})
	}
}

func TestUpdateUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		prep    func(*fakerepo.FakeUserRepo) string
		inName  string
		inEmail string
		wantErr error
	}{
		{
			name: "ok",
			prep: func(r *fakerepo.FakeUserRepo) string {
				u := &domain.User{Name: "Old", Email: "old@x.com"}
				_ = r.Create(context.Background(), u)
				return u.ID.Hex()
			},
			inName: "New", inEmail: "new@x.com",
		},
		{
			name: "not found",
			prep: func(*fakerepo.FakeUserRepo) string { return "ffffffffffffffffffffffff" },
			inName: "N", inEmail: "n@x.com", wantErr: domain.ErrUserNotFound,
		},
		{
			name: "invalid name",
			prep: func(r *fakerepo.FakeUserRepo) string {
				u := &domain.User{Name: "A", Email: "a@x.com"}
				_ = r.Create(context.Background(), u)
				return u.ID.Hex()
			},
			inName: "", inEmail: "a@x.com", wantErr: domain.ErrInvalidName,
		},
		{
			name: "email conflict",
			prep: func(r *fakerepo.FakeUserRepo) string {
				u1 := &domain.User{Name: "A", Email: "a@x.com"}
				u2 := &domain.User{Name: "B", Email: "b@x.com"}
				_ = r.Create(context.Background(), u1)
				_ = r.Create(context.Background(), u2)
				return u1.ID.Hex()
			},
			inName: "A", inEmail: "b@x.com", wantErr: domain.ErrUserExists,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc, repo, _ := newSvc()
			id := tt.prep(repo)

			_, err := svc.UpdateUser(context.Background(), id, tt.inName, tt.inEmail)
			if tt.wantErr != nil {
				if err == nil || !errors.Is(err, tt.wantErr) {
					t.Fatalf("want %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		prep    func(*fakerepo.FakeUserRepo) string
		wantErr error
	}{
		{
			name: "ok",
			prep: func(r *fakerepo.FakeUserRepo) string {
				u := &domain.User{Name: "A", Email: "a@x.com"}
				_ = r.Create(context.Background(), u)
				return u.ID.Hex()
			},
		},
		{name: "empty id", prep: func(*fakerepo.FakeUserRepo) string { return "" }, wantErr: domain.ErrUserNotFound},
		{name: "not found", prep: func(*fakerepo.FakeUserRepo) string { return "ffffffffffffffffffffffff" }, wantErr: domain.ErrUserNotFound},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc, repo, _ := newSvc()
			id := tt.prep(repo)

			err := svc.DeleteUser(context.Background(), id)
			if tt.wantErr != nil {
				if err == nil || !errors.Is(err, tt.wantErr) {
					t.Fatalf("want %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
		})
	}
}

func TestListUsers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		prep      func(*fakerepo.FakeUserRepo)
		wantCount int
		wantErr   error
	}{
		{name: "empty", wantCount: 0},
		{
			name: "two",
			prep: func(r *fakerepo.FakeUserRepo) {
				_ = r.Create(context.Background(), &domain.User{Name: "A", Email: "a@x.com"})
				_ = r.Create(context.Background(), &domain.User{Name: "B", Email: "b@x.com"})
			},
			wantCount: 2,
		},
		{
			name: "repo error",
			prep: func(r *fakerepo.FakeUserRepo) { r.ForceListErr = errors.New("boom") },
			wantErr: errors.New("boom"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc, repo, _ := newSvc()
			if tt.prep != nil {
				tt.prep(repo)
			}

			got, err := svc.ListUsers(context.Background())
			if tt.wantErr != nil {
				if err == nil || err.Error() != tt.wantErr.Error() {
					t.Fatalf("want %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if len(got) != tt.wantCount {
				t.Errorf("want %d, got %d", tt.wantCount, len(got))
			}
		})
	}
}