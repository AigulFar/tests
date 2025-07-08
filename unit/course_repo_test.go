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

func TestCourseRepository_CreateCourse(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := repository.NewCourseRepository(sqlxDB)

	course := &models.Course{
		Name:         "Math 101",
		Completeness: 0,
	}

	mock.ExpectQuery(`INSERT INTO "Moodle".courses`).
		WithArgs(course.Name, course.Completeness).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err = repo.CreateCourse(context.Background(), course)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), course.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCourseRepository_EnrollStudent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := repository.NewCourseRepository(sqlxDB)

	mock.ExpectExec(`INSERT INTO "Moodle".course_student`).
		WithArgs(1, 2).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.EnrollStudent(context.Background(), 1, 2)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
