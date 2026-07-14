package token

import (
	"time"

	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	"github.com/raozhaizhu/go-estate/internal/util"
)

type Maker interface {
	CreateToken(username string, role role.Role, duration time.Duration, tokenType TokenType) (string, *Payload, error)
	VerifyToken(tokenStr string, tokenType TokenType) (*Payload, error)
	ForgeTokenPair(user *db.User, config *util.Config) (accessToken, refreshToken string, accessPayload, refreshPayload *Payload, err error)
}
