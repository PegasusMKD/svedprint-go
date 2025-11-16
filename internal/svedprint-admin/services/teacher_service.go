package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/PegasusMKD/svedprint-go/internal/svedprint-admin/db/sqlc"
	"github.com/PegasusMKD/svedprint-go/internal/svedprint-admin/repositories"
	"github.com/PegasusMKD/svedprint-go/pkg/logger"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/rs/zerolog"
)

type TeacherService struct {
	repository *repositories.TeacherRepository
	userClient *user.Client
	logger     *zerolog.Logger
}

func NewTeacherService(repository *repositories.TeacherRepository, userClient *user.Client) *TeacherService {
	return &TeacherService{
		repository: repository,
		userClient: userClient,
		logger:     logger.Get(),
	}
}

func (service *TeacherService) CreateTeacher(ctx context.Context, username string, password string, schoolUuid string, firstName string, middleName *string, lastName string) (*sqlc.Teacher, error) {
	metadata := json.RawMessage(fmt.Sprintf(`{"role": "teacher", "school_uuid": "%s", "print_allowed": true}`, schoolUuid))
	clerkUser, err := service.userClient.Create(ctx, &user.CreateParams{
		Username:       &username,
		Password:       &password,
		FirstName:      &firstName,
		LastName:       &lastName,
		PublicMetadata: &metadata,
	})

	if err != nil {
		service.logger.Printf("Failed creating the user with username '%s'.", username)
		return nil, err
	}

	teacher, err := service.repository.CreateTeacher(ctx, schoolUuid, firstName, middleName, lastName, clerkUser.ID)
	if err != nil {
		service.logger.Printf("Failed saving teacher details for Clerk ID '%s'.", clerkUser.ID)
		return nil, err
	}

	return teacher, nil
}
