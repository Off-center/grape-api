package user

import (
	"grape-api/pkg"
	"time"
	"uuid"
)

type USER_ROLE int

const (
	MEMBER USER_ROLE = iota
	ADMIN
	MASTER
)

type ACCESS_STATUS int

const (
	ACTIVE ACCESS_STATUS = iota
	PENDING
)

type User struct {
	pkg.Entity

	Name         string
	Username     string
	ImageURL     string
	Role         USER_ROLE
	AccessStatus ACCESS_STATUS
	Bans         []UserBan
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (u *User) IsBanned() bool {
	now := time.Now()

	for _, ban := range u.Bans {
		if ban.RevokedAt.IsZero() && (ban.BanType == PERMANENT || (ban.BanType == TEMPORARY && now.Before(ban.ExpiresAt))) {
			return true
		}
	}

	return false
}

type BAN_TYPE int

const (
	PERMANENT BAN_TYPE = iota
	TEMPORARY
)

type UserBan struct {
	pkg.Entity

	UserID    uuid.UUID
	BanReason string
	BanType   BAN_TYPE
	ExpiresAt time.Time
	RevokedAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewUser(name string, username string, imageURL string, role USER_ROLE, accessStatus ACCESS_STATUS) User {
	return User{
		Entity:       pkg.NewEntity(),
		Name:         name,
		Username:     username,
		ImageURL:     imageURL,
		Role:         role,
		AccessStatus: accessStatus,
		Bans:         []UserBan{},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func NewUserBan(userID uuid.UUID, banReason string, banType BAN_TYPE, expiresAt time.Time) UserBan {
	return UserBan{
		Entity:    pkg.NewEntity(),
		UserID:    userID,
		BanReason: banReason,
		BanType:   banType,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
