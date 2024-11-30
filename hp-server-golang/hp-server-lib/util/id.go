package util

import (
	"github.com/google/uuid"
)

func NewId() string {
	// 生成 UUID
	id := uuid.New()
	return id.String()
}
