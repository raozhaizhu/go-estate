package token

import (
	"time"

	role "github.com/raozhaizhu/go-estate/internal/domain/user"
)

type Maker interface {
	CreateToken(username string, role role.Role, duration time.Duration, tokenType TokenType) (string, *Payload, error)

	VerifyToken(token string, tokenType TokenType) (*Payload, error)
}
