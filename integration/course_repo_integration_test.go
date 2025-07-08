package integration

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/Troshkins/InnoMoodle/backend/models"
	"github.com/Troshkins/InnoMoodle/backend/repository"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type CourseRepositoryIntegrationSuite struct {
	suite.Suite
	container  testcontainers.Container
	db         *sqlx.DB
	courseRepo *repository.CourseRepository
	userRepo   *repository.UserRepository
	ctx        context.Context
}

func (s *CourseRepositoryIntegrationSuite) SetupSuite() {
	s.ctx = context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
		},
		WaitingFor: wait.ForLog("database system is ready").WithStartupTimeout(30 * time.Second),
	}

	var err error
	s.container, err = testcontainers.GenericContainer(s.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(s.T(), err)

	endpoint, err := s.container.Endpoint(s.ctx, "")
	require.NoError(s.T(), err)

	connStr := fmt.Sprintf("postgres://testuser:testpass@%s/testdb?sslmode=disable", endpoint)
	db, err := sqlx.Connect("postgres", connStr)
	require.NoError(s.T(), err)
	require.NoError(s.T(), db.Ping())

	s.db = db
	s.createSchema()

	s.courseRepo = repository.NewCourseRepository(s.db)
	s.userRepo = repository.NewUserRepository(s.db)
}

func (s *CourseRepositoryIntegrationSuite) TearDownSuite() {
	if s.container != nil {
		require.NoError(s.T(), s.container.Terminate(s.ctx))
	}
}

func (s *CourseRepositoryIntegrationSuite) SetupTest() {
	s.truncateTables()
}

func (s *CourseRepositoryIntegrationSuite) createSchema() {
	queries := []string{
		`CREATE SCHEMA IF NOT EXISTS "Moodle"`,
		`CREATE TABLE "Moodle".users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) NOT NULL UNIQUE,
			password VARCHAR(255) NOT NULL,
			role VARCHAR(50) NOT NULL
		)`,
		`CREATE TABLE "Moodle".courses (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			completeness INTEGER NOT NULL
		)`,
		`CREATE TABLE "Moodle".course_student (
			course_id INTEGER NOT NULL,
			student_id INTEGER NOT NULL,
			PRIMARY KEY (course_id, student_id)
		)`,
		`CREATE TABLE "Moodle".course_teacher (
			course_id INTEGER NOT NULL,
			teacher_id INTEGER NOT NULL,
			PRIMARY KEY (course_id, teacher_id)
		)`,
		`CREATE TABLE "Moodle".course_blocks (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			course_id INTEGER NOT NULL
		)`,
		`CREATE TABLE "Moodle".announcements (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			info TEXT NOT NULL,
			block_id INTEGER NOT NULL
		)`,
	}

	for _, query := range queries {
		_, err := s.db.Exec(query)
		require.NoError(s.T(), err)
	}
}

func (s *CourseRepositoryIntegrationSuite) truncateTables() {
	tables := []string{
		`"Moodle".announcements`,
		`"Moodle".course_blocks`,
		`"Moodle".course_student`,
		`"Moodle".course_teacher`,
		`"Moodle".courses`,
		`"Moodle".users`,
	}

	for _, table := range tables {
		_, err := s.db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
		require.NoError(s.T(), err)
	}
}

func (s *CourseRepositoryIntegrationSuite) TestCourseEnrollment() {
	// Create course
	course := &models.Course{
		Name:         "Integration Course",
		Completeness: false,
	}
	err := s.courseRepo.CreateCourse(s.ctx, course)
	require.NoError(s.T(), err)
	require.NotZero(s.T(), course.ID)

	// Create student
	student := &models.User{
		Name:     "Test Student",
		Email:    "student@test.com",
		Password: "password",
		Role:     "student",
	}
	err = s.userRepo.CreateUser(s.ctx, student)
	require.NoError(s.T(), err)
	require.NotZero(s.T(), student.ID)

	// Enroll student
	err = s.courseRepo.EnrollStudent(s.ctx, course.ID, student.ID)
	require.NoError(s.T(), err)

	// Verify enrollment
	var count int
	err = s.db.Get(&count,
		`SELECT COUNT(*) FROM "Moodle".course_student 
		WHERE course_id = $1 AND student_id = $2`,
		course.ID, student.ID)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 1, count)
}

func (s *CourseRepositoryIntegrationSuite) TestTeacherAssignment() {
	// Create course
	course := &models.Course{
		Name:         "Teacher Assignment Course",
		Completeness: false,
	}
	err := s.courseRepo.CreateCourse(s.ctx, course)
	require.NoError(s.T(), err)
	require.NotZero(s.T(), course.ID)

	// Create teacher
	teacher := &models.User{
		Name:     "Test Teacher",
		Email:    "teacher@test.com",
		Password: "password",
		Role:     "teacher",
	}
	err = s.userRepo.CreateUser(s.ctx, teacher)
	require.NoError(s.T(), err)
	require.NotZero(s.T(), teacher.ID)

	// Assign teacher
	err = s.courseRepo.AssignTeacher(s.ctx, course.ID, teacher.ID)
	require.NoError(s.T(), err)

	// Verify assignment
	var count int
	err = s.db.Get(&count,
		`SELECT COUNT(*) FROM "Moodle".course_teacher 
		WHERE course_id = $1 AND teacher_id = $2`,
		course.ID, teacher.ID)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 1, count)
}

func (s *CourseRepositoryIntegrationSuite) TestCourseBlockAndAnnouncement() {
	// Create course
	course := &models.Course{
		Name:         "Block and Announcement Course",
		Completeness: false,
	}
	err := s.courseRepo.CreateCourse(s.ctx, course)
	require.NoError(s.T(), err)
	require.NotZero(s.T(), course.ID)

	// Create course block
	block := &models.CourseBlock{
		Name:     "Algebra Basics",
		CourseID: course.ID,
	}
	err = s.courseRepo.CreateCourseBlock(s.ctx, block)
	require.NoError(s.T(), err)
	require.NotZero(s.T(), block.ID)

	// Create announcement
	announcement := &models.Announcement{
		Name:    "Important Update",
		Info:    "Midterm exam has been rescheduled",
		BlockID: block.ID,
	}
	err = s.courseRepo.CreateAnnouncement(s.ctx, announcement)
	require.NoError(s.T(), err)
	require.NotZero(s.T(), announcement.ID)

	// Verify announcement
	var dbAnn models.Announcement
	err = s.db.Get(&dbAnn,
		`SELECT id, name, info, block_id 
		FROM "Moodle".announcements 
		WHERE id = $1`,
		announcement.ID)
	require.NoError(s.T(), err)
	require.Equal(s.T(), announcement.Name, dbAnn.Name)
	require.Equal(s.T(), announcement.Info, dbAnn.Info)
	require.Equal(s.T(), announcement.BlockID, dbAnn.BlockID)
}

func (s *CourseRepositoryIntegrationSuite) TestGetCourseByID() {
	course := &models.Course{
		Name:         "Physics 101",
		Completeness: true,
	}
	err := s.courseRepo.CreateCourse(s.ctx, course)
	require.NoError(s.T(), err)

	// Test existing course
	result, err := s.courseRepo.GetCourseByID(s.ctx, course.ID)
	require.NoError(s.T(), err)
	require.Equal(s.T(), course.Name, result.Name)
	require.Equal(s.T(), course.Completeness, result.Completeness)

	// Test non-existing course
	_, err = s.courseRepo.GetCourseByID(s.ctx, 9999)
	require.ErrorIs(s.T(), err, sql.ErrNoRows)
}

func TestCourseRepositoryIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(CourseRepositoryIntegrationSuite))
}
