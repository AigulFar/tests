package integration

import (
	"context"
	"testing"

	"github.com/Troshkins/InnoMoodle/backend/models"
	"github.com/Troshkins/InnoMoodle/backend/repository"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type RepositoriesIntegrationSuite struct {
	suite.Suite
	ctx        context.Context
	db         *sqlx.DB
	userRepo   *repository.UserRepository
	courseRepo *repository.CourseRepository
	groupRepo  *repository.GroupRepository
}

func (s *RepositoriesIntegrationSuite) SetupSuite() {
	s.ctx = context.Background()

	// Инициализация подключения к БД
	db, err := setupDB()
	require.NoError(s.T(), err)
	s.db = db

	// Инициализация репозиториев
	s.userRepo = repository.NewUserRepository(s.db)
	s.courseRepo = repository.NewCourseRepository(s.db)
	s.groupRepo = repository.NewGroupRepository(s.db)

	// Создание минимальной схемы
	s.createMinimalSchema()
}

func (s *RepositoriesIntegrationSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
}

func (s *RepositoriesIntegrationSuite) SetupTest() {
	s.truncateTables()
}

func (s *RepositoriesIntegrationSuite) createMinimalSchema() {
	queries := []string{
		`DROP SCHEMA IF EXISTS "Moodle" CASCADE`,
		`CREATE SCHEMA "Moodle"`,
		`CREATE TABLE "Moodle".users (
            id BIGSERIAL PRIMARY KEY,
            name TEXT NOT NULL,
            email TEXT NOT NULL,
            password TEXT NOT NULL
        )`,
		`CREATE TABLE "Moodle".courses (
            id BIGSERIAL PRIMARY KEY,
            name TEXT NOT NULL,
            completeness INTEGER NOT NULL
        )`,
		`CREATE TABLE "Moodle".study_groups (
            id BIGSERIAL PRIMARY KEY,
            name TEXT NOT NULL
        )`,
		`CREATE TABLE "Moodle".course_student (
            course_id BIGINT NOT NULL REFERENCES "Moodle".courses(id),
            student_id BIGINT NOT NULL REFERENCES "Moodle".users(id),
            PRIMARY KEY (course_id, student_id)
        )`,
		`CREATE TABLE "Moodle".group_student (
            group_id BIGINT NOT NULL REFERENCES "Moodle".study_groups(id),
            student_id BIGINT NOT NULL REFERENCES "Moodle".users(id),
            PRIMARY KEY (group_id, student_id)
        )`,
	}

	for _, query := range queries {
		_, err := s.db.Exec(query)
		if err != nil {
			s.T().Logf("Error executing query: %s\nError: %v", query, err)
		}
		require.NoError(s.T(), err)
	}
}

func (s *RepositoriesIntegrationSuite) truncateTables() {
	tables := []string{
		`"Moodle".group_student`,
		`"Moodle".course_student`,
		`"Moodle".study_groups`,
		`"Moodle".courses`,
		`"Moodle".users`,
	}
	for _, table := range tables {
		_, err := s.db.Exec("TRUNCATE TABLE " + table + " RESTART IDENTITY CASCADE")
		require.NoError(s.T(), err)
	}
}

// Тест 1: Создание пользователя
func (s *RepositoriesIntegrationSuite) TestCreateUser() {
	user := &models.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	err := s.userRepo.CreateUser(s.ctx, user)
	require.NoError(s.T(), err)
	require.NotZero(s.T(), user.ID)
}

// Тест 2: Создание курса
func (s *RepositoriesIntegrationSuite) TestCreateCourse() {
	course := &models.Course{
		Name:         "Introduction to Testing",
		Completeness: 0, // Убедитесь, что это число, а не строка
	}

	err := s.courseRepo.CreateCourse(s.ctx, course)
	require.NoError(s.T(), err)
	require.NotZero(s.T(), course.ID)

	// Проверка создания курса в БД
	var dbCourse models.Course
	err = s.db.Get(&dbCourse, `SELECT * FROM "Moodle".courses WHERE id = $1`, course.ID)
	require.NoError(s.T(), err)
	require.Equal(s.T(), course.Name, dbCourse.Name)
	require.Equal(s.T(), course.Completeness, dbCourse.Completeness)
}

// Тест 3: Создание учебной группы
func (s *RepositoriesIntegrationSuite) TestCreateStudyGroup() {
	group := &models.StudyGroup{
		Name: "Group A",
	}

	err := s.groupRepo.CreateStudyGroup(s.ctx, group)
	require.NoError(s.T(), err)
	require.NotZero(s.T(), group.ID)
}

// Тест 4: Запись студента на курс
func (s *RepositoriesIntegrationSuite) TestEnrollStudent() {
	// Создаем студента
	student := &models.User{
		Name:     "Student",
		Email:    "student@example.com",
		Password: "pass123",
	}
	err := s.userRepo.CreateUser(s.ctx, student)
	require.NoError(s.T(), err)

	// Создаем курс
	course := &models.Course{
		Name:         "Math 101",
		Completeness: 0,
	}
	err = s.courseRepo.CreateCourse(s.ctx, course)
	require.NoError(s.T(), err)

	// Записываем студента на курс
	err = s.courseRepo.EnrollStudent(s.ctx, course.ID, student.ID)
	require.NoError(s.T(), err)

	// Проверка записи в БД
	var count int
	err = s.db.Get(&count, `
        SELECT COUNT(*) 
        FROM "Moodle".course_student 
        WHERE course_id = $1 AND student_id = $2`,
		course.ID, student.ID)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 1, count)
}

// Тест 5: Добавление студента в группу
func (s *RepositoriesIntegrationSuite) TestAddStudentToGroup() {
	// Создаем студента
	student := &models.User{
		Name:     "Group Student",
		Email:    "group@example.com",
		Password: "group123",
	}
	err := s.userRepo.CreateUser(s.ctx, student)
	require.NoError(s.T(), err)

	// Создаем группу
	group := &models.StudyGroup{
		Name: "Study Group 1",
	}
	err = s.groupRepo.CreateStudyGroup(s.ctx, group)
	require.NoError(s.T(), err)

	// Добавляем студента в группу
	err = s.groupRepo.AddStudentToGroup(s.ctx, group.ID, student.ID)
	require.NoError(s.T(), err)
}

func TestRepositoriesIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(RepositoriesIntegrationSuite))
}
