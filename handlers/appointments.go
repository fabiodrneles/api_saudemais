package handlers

import (
	"errors"
	"net/http"
	"saudemais-api/auth"
	"saudemais-api/database"
	"saudemais-api/models"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// AppointmentRequest define a estrutura para o corpo da requisição de agendamento.
type AppointmentRequest struct {
	DateTime string `json:"datetime"`
}

// getPatientIDFromToken extrai o ID do paciente das claims do token JWT.
// O token já foi validado pelo middleware antes de chegar a esta função.
func getPatientIDFromToken(c echo.Context) (uint, error) {
	token, ok := c.Get("user").(*jwt.Token)
	if !ok || token == nil {
		return 0, errors.New("token JWT inválido ou ausente")
	}

	claims, ok := token.Claims.(*auth.JwtCustomClaims)
	if !ok {
		return 0, errors.New("claims JWT inválidas")
	}

	return claims.PatientID, nil
}

// CreateAppointment lida com a criação de uma nova consulta.
func CreateAppointment(c echo.Context) error {
	var req AppointmentRequest
	if err := c.Bind(&req); err != nil {
		return newErrorResponse(c, http.StatusBadRequest, "Corpo da requisição inválido")
	}

	// 1. Validação e conversão do formato da data (RFC3339).
	appointmentTime, err := time.Parse(time.RFC3339, req.DateTime)
	if err != nil {
		return newErrorResponse(c, http.StatusBadRequest, "Formato de data e hora inválido. Use o padrão AAAA-MM-DDTHH:MM:SSZ")
	}

	// 2. Validação da regra de negócio: não permitir agendamentos no passado.
	if appointmentTime.Before(time.Now()) {
		return newErrorResponse(c, http.StatusBadRequest, "Não é permitido agendar consultas no passado")
	}

	patientID, err := getPatientIDFromToken(c)
	if err != nil {
		return newErrorResponse(c, http.StatusUnauthorized, "Acesso não autorizado")
	}

	// 3. Validação da regra de negócio: impedir que um paciente agende duas consultas no mesmo horário.
	var existingAppointment models.Appointment
	err = database.DB.Where("patient_id = ? AND date_time = ?", patientID, appointmentTime).First(&existingAppointment).Error
	if err == nil {
		return newErrorResponse(c, http.StatusConflict, "Já existe uma consulta agendada para este horário")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return newErrorResponse(c, http.StatusInternalServerError, "Erro ao verificar a disponibilidade da consulta")
	}

	// Criação da nova consulta.
	newAppointment := models.Appointment{
		PatientID: patientID,
		DateTime:  appointmentTime,
	}

	if result := database.DB.Create(&newAppointment); result.Error != nil {
		return newErrorResponse(c, http.StatusInternalServerError, "Não foi possível agendar a consulta")
	}

	return c.JSON(http.StatusCreated, newAppointment)
}

// ListAppointments retorna todas as consultas do paciente autenticado, ordenadas por data.
func ListAppointments(c echo.Context) error {
	patientID, err := getPatientIDFromToken(c)
	if err != nil {
		return newErrorResponse(c, http.StatusUnauthorized, "Acesso não autorizado")
	}

	var appointments []models.Appointment
	if err := database.DB.Where("patient_id = ?", patientID).Order("date_time asc").Find(&appointments).Error; err != nil {
		return newErrorResponse(c, http.StatusInternalServerError, "Erro ao buscar as consultas")
	}

	return c.JSON(http.StatusOK, appointments)
}

// CancelAppointment remove uma consulta específica do paciente autenticado.
func CancelAppointment(c echo.Context) error {
	appointmentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return newErrorResponse(c, http.StatusBadRequest, "ID da consulta inválido")
	}

	patientID, err := getPatientIDFromToken(c)
	if err != nil {
		return newErrorResponse(c, http.StatusUnauthorized, "Acesso não autorizado")
	}

	// Busca a consulta, garantindo que ela pertence ao paciente logado.
	// Isso impede que um paciente cancele a consulta de outro.
	var appointment models.Appointment
	result := database.DB.Where("id = ? AND patient_id = ?", appointmentID, patientID).First(&appointment)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return newErrorResponse(c, http.StatusForbidden, "Consulta não encontrada ou acesso não autorizado")
	}
	if result.Error != nil {
		return newErrorResponse(c, http.StatusInternalServerError, "Erro ao buscar a consulta para cancelamento")
	}

	// Deleta a consulta encontrada.
	if result := database.DB.Delete(&appointment); result.Error != nil {
		return newErrorResponse(c, http.StatusInternalServerError, "Falha ao cancelar a consulta")
	}

	return c.NoContent(http.StatusNoContent)
}
