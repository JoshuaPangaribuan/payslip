package pkguid

import (
	"github.com/google/uuid"
)

type UUID interface {
	Generate() string
	GenerateV4() uuid.UUID
}

type uuidGen struct{}

func NewUUID() UUID {
	return &uuidGen{}
}

func (u *uuidGen) Generate() string {
	return uuid.New().String()
}

func (u *uuidGen) GenerateV4() uuid.UUID {
	return uuid.New()
}
