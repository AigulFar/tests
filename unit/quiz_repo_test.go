package unit

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Troshkins/InnoMoodle/backend/models"
	"github.com/Troshkins/InnoMoodle/backend/repository"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestQuizRepository_CreateQuiz(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := repository.NewQuizRepository(sqlxDB)

	quiz := &models.Quiz{
		Name:         "Math Quiz",
		Start:        nil,
		End:          nil,
		Returnable:   true,
		Random:       false,
		TimeLimit:    30,
		ResultsShown: true,
		TryCount:     3,
		FillingID:    1,
	}

	mock.ExpectQuery(`INSERT INTO "Moodle".quizzes`).
		WithArgs(
			quiz.Name, quiz.Start, quiz.End, quiz.Returnable, quiz.Random,
			quiz.TimeLimit, // Fixed argument (was 'quiz.Time')
			quiz.ResultsShown, quiz.TryCount, quiz.FillingID,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err = repo.CreateQuiz(context.Background(), quiz)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), quiz.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}
