package instance

import (
	"grape-api/pkg"
	"time"
	"uuid"
)

type REGISTRATION_METHOD int

type SETUP_STATUS int

const (
	PUBLIC REGISTRATION_METHOD = iota
	APPROVAL
	CLOSED
)

const (
	INITIALIZED SETUP_STATUS = iota
	UNINITIALIZED
)

type Instance struct {
	pkg.Entity
	Name               string
	Description        string
	SetupStatus        SETUP_STATUS
	RegistrationMethod REGISTRATION_METHOD
	MasterUserID       uuid.UUID
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func NewUninitializedInstance() Instance {
	now := time.Now()

	return Instance{
		Entity:      pkg.NewEntity(),
		SetupStatus: UNINITIALIZED,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func NewInstance(name string, description string, masterUserID uuid.UUID) Instance {
	return Instance{
		Entity:             pkg.NewEntity(),
		Name:               name,
		Description:        description,
		RegistrationMethod: CLOSED,
		MasterUserID:       masterUserID,
		SetupStatus:        UNINITIALIZED,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
}
