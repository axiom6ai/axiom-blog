package model

import (
	"github.com/google/uuid"
	"time"
)

type CommentReply struct {
	//自增ID
	Id uint `gorm:"primaryKey; autoIncrement;"`

	//评论cid
	Cid uint `gorm:"column:cid"`

	//回复用户uid
	UserID uuid.UUID `gorm:"column:user_ID"`

	//回复内容
	Content string

	//状态：0-未审核;1-已上线;2-下线(审核拒绝);3-用户删除
	State int

	CreatedAt time.Time
}
