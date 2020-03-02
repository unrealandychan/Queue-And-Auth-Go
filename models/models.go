package models

import (
	"github.com/google/uuid"
	"github.com/dgrijalva/jwt-go"
)

type Token struct{
	Token string `json:"token"`
}

type Ticket struct{
	UUID uuid.UUID
	Event string
	Ticket int
	Access bool
	CreateTime int64
}

type Event struct{
	Event string
	QueueCount int
	CurrentCount int
}

type Claims struct{
	UUID uuid.UUID `json:"uuid"`
	Event string `json:"event"`
	Ticket int `json:"ticket"`
	Access bool `json:"access"`
	jwt.StandardClaims
}