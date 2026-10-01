package mongo

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"example/internal/domain"
)

type UserRepo struct {
	coll *mongo.Collection
}

func NewUserRepo(db *mongo.Database) *UserRepo {
	return &UserRepo{coll: db.Collection("users")}
}

// EnsureIndexes создаёт уникальный индекс на email
func (r *UserRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return err
}

func (r *UserRepo) Create(ctx context.Context, u *domain.User) (string, error) {
	res, err := r.coll.InsertOne(ctx, u)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return "", domain.ErrUserExists
		}
		return "", err
	}
	oid, _ := res.InsertedID.(bson.ObjectID)
	return oid.Hex(), nil
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	var u domain.User
	err = r.coll.FindOne(ctx, bson.M{"_id": oid}).Decode(&u)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	err := r.coll.FindOne(ctx, bson.M{"email": email}).Decode(&u)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) Update(ctx context.Context, u *domain.User) error {
	res, err := r.coll.UpdateOne(ctx,
		bson.M{"_id": u.ID},
		bson.M{"$set": bson.M{"name": u.Name, "email": u.Email}},
	)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return domain.ErrUserExists
		}
		return err
	}
	if res.MatchedCount == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *UserRepo) Delete(ctx context.Context, id string) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return domain.ErrUserNotFound
	}
	res, err := r.coll.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *UserRepo) List(ctx context.Context) ([]*domain.User, error) {
	cursor, err := r.coll.Find(ctx, bson.M{}, options.Find().SetSort(bson.M{"_id": 1}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*domain.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}