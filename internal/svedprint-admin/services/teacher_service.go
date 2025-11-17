package services

import (
	"context"
	"encoding/json"
	"fmt"

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

type TeacherDto struct {
	Uuid     string `json:"uuid"`
	Username string `json:"username"`

	FirstName  string  `json:"first_name"`
	MiddleName *string `json:"middle_name"`
	LastName   string  `json:"last_name"`

	SchoolUuid string `json:"school_uuid"`
}

func NewTeacherService(repository *repositories.TeacherRepository, userClient *user.Client) *TeacherService {
	return &TeacherService{
		repository: repository,
		userClient: userClient,
		logger:     logger.Get(),
	}
}

func (service *TeacherService) CreateTeacher(ctx context.Context, username string, password string, schoolUuid string, firstName string, middleName *string, lastName string) (*TeacherDto, error) {
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

	return &TeacherDto{
		Uuid:     teacher.Uuid.String(),
		Username: username,

		FirstName:  teacher.FirstName,
		MiddleName: &teacher.MiddleName.String,
		LastName:   teacher.LastName,

		SchoolUuid: teacher.SchoolUuid.String(),
	}, nil
}
