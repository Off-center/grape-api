package pkg

import "uuid"

type Identifiable interface {
	ID() uuid.UUID
}

type Entity struct {
	id uuid.UUID
}

func NewEntity() Entity {
	return Entity{
		id: uuid.NewV7(),
	}
}

func Restore(id uuid.UUID) Entity {
	return Entity{
		id: id,
	}
}

func (e Entity) ID() uuid.UUID {
	return e.id
}

func (e Entity) Equal(other Identifiable) bool {
	return e.id == other.ID()
}
