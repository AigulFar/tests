package integration

import (
	"context"
	_ "database/sql"
	_ "fmt"
	"github.com/jmoiron/sqlx"
	"testing"
	"time"

	"github.com/Troshkins/InnoMoodle/backend/models"
	"github.com/Troshkins/InnoMoodle/backend/repository"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestQuizRepository_Integration(t *testing.T) {
	ctx := context.Background()

	// Запуск PostgreSQL контейнера (аналогично предыдущему тесту)
	// ... (код запуска контейнера из предыдущего теста)

	// Инициализация схемы
	_, err := db.Exec(`
		CREATE SCHEMA "Moodle";
		CREATE TABLE "Moodle".quizzes (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			start TIMESTAMP,
			"end" TIMESTAMP,
			returnable BOOLEAN,
			random BOOLEAN,
			"time" INTEGER,
			results_shown BOOLEAN,
			try_count INTEGER,
			filling_id INTEGER
		);
		
		CREATE TABLE "Moodle".tasks (
			id SERIAL PRIMARY KEY,
			type INTEGER NOT NULL,
			quiz_id INTEGER NOT NULL
		);
		
		CREATE TABLE "Moodle".one_ans_task (
			id SERIAL PRIMARY KEY,
			task_id INTEGER NOT NULL,
			question TEXT NOT NULL
		);
	`)
	require.NoError(t, err)

	// Создание репозитория
	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := repository.NewQuizRepository(sqlxDB)

	t.Run("Create Quiz and Tasks", func(t *testing.T) {
		// Создание квиза
		start := time.Now()
		end := start.Add(time.Hour * 24)
		quiz := &models.Quiz{
			Name:         "Integration Quiz",
			Start:        &start,
			End:          &end,
			Returnable:   true,
			Random:       false,
			Time:         30,
			ResultsShown: true,
			TryCount:     3,
			FillingID:    1,
		}

		err := repo.CreateQuiz(ctx, quiz)
		require.NoError(t, err)
		require.NotZero(t, quiz.ID)

		// Создание задания
		task := &models.Task{
			Type:   models.TaskTypeOneAnswer,
			QuizID: quiz.ID,
		}
		err = repo.CreateTask(ctx, task)
		require.NoError(t, err)
		require.NotZero(t, task.ID)

		// Создание задания с одним ответом
		oneAnsTask := &models.OneAnsTask{
			TaskID:   task.ID,
			Question: "What is 2+2?",
		}
		err = repo.CreateOneAnsTask(ctx, oneAnsTask)
		require.NoError(t, err)
		require.NotZero(t, oneAnsTask.ID)
	})
}
