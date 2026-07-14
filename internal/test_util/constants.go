package testUtil

import "github.com/google/uuid"

const (
	UserAgent                = "my-test-user-agent"
	DeviceID                 = "my-test-device-id"
	ClientIp                 = "127.0.0.1"
	ClientIpWithPort         = "127.0.0.1:12345"
	AccessStr                = "access-token-123"
	RefreshStr               = "refresh-token-123"
	AuthorizationAccessToken = "bearer " + AccessStr
)

func NewUUID() uuid.UUID {
	return uuid.New()
}
