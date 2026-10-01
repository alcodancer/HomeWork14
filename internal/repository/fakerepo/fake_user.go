package fakerepo

import (
	"context"
	"sync"

	"go.mongodb.org/mongo-driver/v2/bson"

	"example/internal/domain"
)

type FakeUserRepo struct {
	mu     sync.Mutex
	Users  map[string]*domain.User
	NextID int

	ForceCreateErr error
	ForceGetErr    error
	ForceUpdateErr error
	ForceDeleteErr error
	ForceListErr   error
}

func NewFakeUserRepo() *FakeUserRepo {
	return &FakeUserRepo{Users: map[string]*domain.User{}, NextID: 1}
}

func (r *FakeUserRepo) nextOID() bson.ObjectID {
	oid := bson.NewObjectID()
	return oid
}

func (r *FakeUserRepo) Create(_ context.Context, u *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.ForceCreateErr != nil {
		return r.ForceCreateErr
	}
	oid := r.nextOID()
	cp := *u
	cp.ID = oid
	r.Users[oid.Hex()] = &cp
	u.ID = oid
	return nil
}

func (r *FakeUserRepo) GetByID(_ context.Context, id string) (*domain.User, error) {
	if r.ForceGetErr != nil {
		return nil, r.ForceGetErr
	}
	u, ok := r.Users[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	cp := *u
	return &cp, nil
}

func (r *FakeUserRepo) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	if r.ForceGetErr != nil {
		return nil, r.ForceGetErr
	}
	for _, u := range r.Users {
		if u.Email == email {
			cp := *u
			return &cp, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func (r *FakeUserRepo) Update(_ context.Context, u *domain.User) error {
	if r.ForceUpdateErr != nil {
		return r.ForceUpdateErr
	}
	if _, ok := r.Users[u.ID.Hex()]; !ok {
		return domain.ErrUserNotFound
	}
	cp := *u
	r.Users[u.ID.Hex()] = &cp
	return nil
}

func (r *FakeUserRepo) Delete(_ context.Context, id string) error {
	if r.ForceDeleteErr != nil {
		return r.ForceDeleteErr
	}
	if _, ok := r.Users[id]; !ok {
		return domain.ErrUserNotFound
	}
	delete(r.Users, id)
	return nil
}

func (r *FakeUserRepo) List(_ context.Context) ([]*domain.User, error) {
	if r.ForceListErr != nil {
		return nil, r.ForceListErr
	}
	out := make([]*domain.User, 0, len(r.Users))
	for _, u := range r.Users {
		cp := *u
		out = append(out, &cp)
	}
	return out, nil
}