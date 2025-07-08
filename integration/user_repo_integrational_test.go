package integration

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "os"
	"testing"

	"github.com/Troshkins/InnoMoodle/backend/models"
	"github.com/Troshkins/InnoMoodle/backend/repository"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestUserRepository_Integration(t *testing.T) {
	ctx := context.Background()

	// Запуск PostgreSQL контейнера
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "user",
			"POSTGRES_PASSWORD": "password",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	defer pgContainer.Terminate(ctx)

	// Получение порта контейнера
	port, err := pgContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)

	// Формирование строки подключения
	connStr := fmt.Sprintf("host=localhost port=%s user=user password=password dbname=testdb sslmode=disable", port.Port())

	// Подключение к базе данных
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	defer db.Close()

	// Инициализация схемы
	_, err = db.Exec(`
		CREATE SCHEMA "Moodle";
		CREATE TABLE "Moodle".users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL
		);
	`)
	require.NoError(t, err)

	// Создание репозитория
	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := repository.NewUserRepository(sqlxDB)

	t.Run("Create and Get User", func(t *testing.T) {
		// Создание пользователя
		user := &models.User{
			Name:     "Integration User",
			Email:    "integration@test.com",
			Password: "securepassword",
		}

		err := repo.CreateUser(ctx, user)
		require.NoError(t, err)
		require.NotZero(t, user.ID)

		// Получение пользователя по email
		retrievedUser, err := repo.GetUserByEmail(ctx, user.Email)
		require.NoError(t, err)
		require.Equal(t, user.Name, retrievedUser.Name)
		require.Equal(t, user.Email, retrievedUser.Email)
	})
}
