package vo

import "github.com/google/uuid"

type LoginVo struct {
	Token string
	ID    uuid.UUID
}
