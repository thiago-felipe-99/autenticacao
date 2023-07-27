package model_test

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/thiago-felipe-99/autenticacao/model"
)

func TestID(t *testing.T) {
	t.Parallel()

	t.Run("ParserID", func(t *testing.T) {
		t.Parallel()

		uuidID := uuid.New()

		id, err := model.ParseID(uuidID.String())
		assert.NoError(t, err)
		assert.Equal(t, id, model.ID(uuidID))
		assert.Equal(t, id.String(), uuidID.String())

		idValue, err := id.Value()
		assert.NoError(t, err)
		assert.Equal(t, id.String(), idValue)

		_, err = model.ParseID("invalid-id")
		assert.ErrorContains(t, err, "error parsing ID")
	})

	t.Run("ScanID", func(t *testing.T) {
		t.Parallel()

		uuidID := uuid.New()

		id := model.ID{}
		err := id.Scan(uuidID)
		assert.NoError(t, err)
		assert.Equal(t, id, model.ID(uuidID))
		assert.Equal(t, id.String(), uuidID.String())

		idValue, err := id.Value()
		assert.NoError(t, err)
		assert.Equal(t, id.String(), idValue)

		err = id.Scan("invalid-id")
		assert.ErrorContains(t, err, "error parsing ID")
	})

	t.Run("NewID", func(t *testing.T) {
		t.Parallel()

		id := model.NewID()
		uuidID, err := uuid.Parse(id.String())
		assert.NoError(t, err)
		assert.Equal(t, id, model.ID(uuidID))
		assert.Equal(t, id.String(), uuidID.String())

		idValue, err := id.Value()
		assert.NoError(t, err)
		assert.Equal(t, id.String(), idValue)
	})
}

func TestUser(t *testing.T) {
	t.Parallel()

	user := &model.User{
		ID:        model.NewID(),
		Name:      gofakeit.Name(),
		Username:  gofakeit.Username(),
		Email:     gofakeit.Email(),
		Password:  gofakeit.Password(true, true, true, true, true, gofakeit.Number(1, 255)),
		Roles:     []string{gofakeit.Name(), gofakeit.Name(), gofakeit.Name()},
		IsActive:  true,
		CreatedAt: time.Now(),
		CreatedBy: model.NewID(),
		DeletedAt: gofakeit.FutureDate(),
		DeletedBy: model.NewID(),
	}

	postgres := &model.UserPostgres{
		ID:        user.ID,
		Name:      user.Name,
		Username:  user.Username,
		Email:     user.Email,
		Password:  user.Password,
		Roles:     user.Roles,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
		CreatedBy: user.CreatedBy,
		DeletedAt: user.DeletedAt,
		DeletedBy: user.DeletedBy,
	}

	assert.Equal(t, user.Postgres(), postgres)
	assert.Equal(t, postgres.User(), user)
}