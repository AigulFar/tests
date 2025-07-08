package unit

import (
	"context"
	_ "database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Troshkins/InnoMoodle/backend/models"
	"github.com/Troshkins/InnoMoodle/backend/repository" // Два уровня вверх от tests/unit
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestUserRepository_CreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := repository.NewUserRepository(sqlxDB)

	user := &models.User{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}

	mock.ExpectQuery(`INSERT INTO "Moodle".users`).
		WithArgs(user.Name, user.Email, user.Password).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err = repo.CreateUser(context.Background(), user)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), user.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}
