package handlers

import (
	"github.com/PegasusMKD/svedprint-go/internal/svedprint-admin/services"
	"github.com/PegasusMKD/svedprint-go/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type RegistrationHandler struct {
	service *services.TeacherService
	logger  *zerolog.Logger
}

type RegistrationDto struct {
	username string
	password string

	firstName  string
	middleName *string
	lastName   string

	schoolUuid string
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
		dto.username,
		dto.password,
		dto.schoolUuid,
		dto.firstName,
		dto.middleName,
		dto.lastName,
	)

	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(201, teacher)
}
