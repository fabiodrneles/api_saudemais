package middleware

import (
	"log"
	"net/http"
	"os"
	"saudemais-api/auth"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

// jwtSecretKey armazena a chave secreta usada para assinar os tokens JWT.
// É carregada de uma variável de ambiente na inicialização.
var jwtSecretKey []byte

// init é uma função especial do Go que roda antes da main().
// É usada aqui para carregar e validar a chave secreta do JWT.
func init() {
	secret := os.Getenv("JWT_SECRET_KEY")
	if secret == "" {
		// Se a chave secreta não estiver definida, a aplicação não pode assinar tokens de forma segura.
		// Logamos um erro fatal para impedir que a aplicação inicie em um estado inseguro.
		log.Fatal("FATAL: A variável de ambiente JWT_SECRET_KEY não está definida.")
	}
	jwtSecretKey = []byte(secret)
}

// JWTMiddleware retorna uma função de middleware do Echo que protege rotas validando JWTs.
// Ele verifica a presença de um token válido no cabeçalho 'Authorization'.
func JWTMiddleware() echo.MiddlewareFunc {
	// Configuração para o middleware echo-jwt.
	config := echojwt.Config{
		// NewClaimsFunc é obrigatório para especificar o tipo de objeto de claims a ser usado.
		// Isso permite que o middleware analise corretamente as claims do token em nossa struct personalizada.
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(auth.JwtCustomClaims)
		},

		// SigningKey é a chave secreta usada para validar a assinatura do token.
		// Deve ser a mesma chave usada para assinar o token.
		SigningKey: jwtSecretKey,

		// ErrorHandler é uma função personalizada que é executada quando o middleware encontra um erro
		// (ex: token ausente, expirado ou com assinatura inválida).
		ErrorHandler: func(c echo.Context, err error) error {
			// Retornamos uma resposta de erro JSON padronizada para manter a consistência com o resto da API.
			// NOTA: Em uma aplicação maior, essa lógica de resposta de erro poderia ser centralizada
			// em um pacote compartilhado (ex: 'responses') para evitar duplicação.
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"message": "Token inválido ou expirado",
				"code":    http.StatusUnauthorized,
			})
		},
	}

	return echojwt.WithConfig(config)
}
