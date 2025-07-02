package handlers

import (
	"errors"
	"net/http"
	"net/mail"
	"saudemais-api/auth"
	"saudemais-api/database"
	"saudemais-api/models"
	"strings"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// RegisterRequest define a estrutura para o corpo da requisição de cadastro.
type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest define a estrutura para o corpo da requisição de login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// newErrorResponse é uma função auxiliar para criar respostas de erro padronizadas.
func newErrorResponse(c echo.Context, code int, message string) error {
	return c.JSON(code, map[string]interface{}{
		"message": message,
		"code":    code,
	})
}

// RegisterPatient lida com a requisição de cadastro de um novo paciente.
func RegisterPatient(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return newErrorResponse(c, http.StatusBadRequest, "Corpo da requisição inválido")
	}

	// Validação dos campos de entrada.
	if strings.TrimSpace(req.Name) == "" {
		return newErrorResponse(c, http.StatusBadRequest, "O nome é obrigatório")
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return newErrorResponse(c, http.StatusBadRequest, "O formato do email informado é inválido")
	}
	if len(req.Password) < 6 { // Exemplo de validação de senha.
		return newErrorResponse(c, http.StatusBadRequest, "A senha deve ter pelo menos 6 caracteres")
	}

	// Criptografa a senha do usuário antes de salvar.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return newErrorResponse(c, http.StatusInternalServerError, "Falha ao processar a senha")
	}

	patient := models.Patient{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	// Tenta criar o paciente no banco de dados.
	if result := database.DB.Create(&patient); result.Error != nil {
		// Verifica se o erro é uma violação de constraint 'unique' (e-mail duplicado).
		if strings.Contains(result.Error.Error(), "unique constraint") {
			return newErrorResponse(c, http.StatusConflict, "O email informado já está cadastrado")
		}
		return newErrorResponse(c, http.StatusInternalServerError, "Não foi possível cadastrar o paciente")
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "Paciente cadastrado com sucesso!"})
}

// Login lida com a requisição de autenticação de um paciente.
func Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return newErrorResponse(c, http.StatusBadRequest, "Corpo da requisição inválido")
	}

	// Validação dos campos de entrada.
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return newErrorResponse(c, http.StatusBadRequest, "O formato do email informado é inválido")
	}
	if strings.TrimSpace(req.Password) == "" {
		return newErrorResponse(c, http.StatusBadRequest, "A senha é obrigatória")
	}

	var patient models.Patient
	// Procura o paciente pelo email no banco de dados.
	err := database.DB.Where("email = ?", req.Email).First(&patient).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Retorna erro genérico para não informar se o email existe ou não (segurança).
		return newErrorResponse(c, http.StatusUnauthorized, "Email ou senha inválidos")
	}
	if err != nil {
		return newErrorResponse(c, http.StatusInternalServerError, "Erro interno no servidor")
	}

	// Compara a senha enviada com a senha criptografada (hash) armazenada no banco.
	if err := bcrypt.CompareHashAndPassword([]byte(patient.Password), []byte(req.Password)); err != nil {
		// A senha não confere, retorna o mesmo erro genérico.
		return newErrorResponse(c, http.StatusUnauthorized, "Email ou senha inválidos")
	}

	// Se as credenciais estiverem corretas, gera um token JWT.
	token, err := auth.GenerateToken(patient.ID, patient.Email)
	if err != nil {
		return newErrorResponse(c, http.StatusInternalServerError, "Falha ao gerar o token de autenticação")
	}

	// Retorna o token para o cliente.
	return c.JSON(http.StatusOK, echo.Map{"token": token})
}
