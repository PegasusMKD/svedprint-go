package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/PegasusMKD/svedprint-go/internal/svedprint-admin/db/sqlc"
	"github.com/PegasusMKD/svedprint-go/pkg/logger"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
)

type TeacherRepository struct {
	queries *sqlc.Queries
	logger  *zerolog.Logger
}

func NewTeacherRepository(queries *sqlc.Queries) *TeacherRepository {
	return &TeacherRepository{
		queries: queries,
		logger:  logger.Get(),
	}
}

func (repo *TeacherRepository) CreateTeacher(ctx context.Context, schoolUuid string, firstName string, middleName *string, lastName string, clerkId string) (*sqlc.Teacher, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var schoolUUID pgtype.UUID
	if err := schoolUUID.Scan(schoolUuid); err != nil {
		return nil, fmt.Errorf("invalid school UUID: %w", err)
	}

	params := sqlc.InsertTeacherParams{
		SchoolUuid: schoolUUID,
		FirstName:  firstName,
		LastName:   lastName,
		MiddleName: pgtype.Text{Valid: false},
		ClerkID:    clerkId,
	}

	if middleName != nil {
		params.MiddleName = pgtype.Text{
			String: *middleName,
			Valid:  true,
		}
	}

	teacher, err := repo.queries.InsertTeacher(ctx, params)
	if err != nil {
		repo.logger.Printf("Failed creating a teacher record for '%s' Clerk ID", clerkId)
		return nil, err
	}

	return &teacher, nil
}
