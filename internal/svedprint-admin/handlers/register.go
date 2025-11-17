package handlers

import (
	"github.com/PegasusMKD/svedprint-go/internal/svedprint-admin/services"
	"github.com/PegasusMKD/svedprint-go/pkg/logger"
	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type RegistrationHandler struct {
	service *services.TeacherService
	logger  *zerolog.Logger
}

type RegistrationDto struct {
	Username string `json:"username"`
	Password string `json:"password"`

	FirstName  string  `json:"first_name"`
	MiddleName *string `json:"middle_name"`
	LastName   string  `json:"last_name"`

	SchoolUuid string `json:"school_uuid"`
}

func NewRegistrationHandler(service *services.TeacherService) *RegistrationHandler {
	return &RegistrationHandler{
		service: service,
		logger:  logger.Get(),
	}
}

func (handler *RegistrationHandler) RegisterUser(ctx *gin.Context) {
	var dto RegistrationDto
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	teacher, err := handler.service.CreateTeacher(
		ctx.Request.Context(),
		dto.Username,
		dto.Password,
		dto.SchoolUuid,
		dto.FirstName,
		dto.MiddleName,
		dto.LastName,
	)

	if err != nil {
		// Check if this is a Clerk API error
		if apiErr, ok := err.(*clerk.APIErrorResponse); ok {
			// Return the Clerk error with its original status code and structure
			ctx.JSON(apiErr.HTTPStatusCode, gin.H{
				"errors":          apiErr.Errors,
				"status":          apiErr.HTTPStatusCode,
				"clerk_trace_id":  apiErr.TraceID,
			})
			return
		}

		// For non-Clerk errors, return generic 500
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(201, teacher)
}
