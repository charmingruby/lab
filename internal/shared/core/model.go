package core

import "time"

type Model struct {
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" db:"deleted_at"`
	ID        string     `json:"id"         db:"id"`
}

func NewModel() Model {
	return Model{
		ID:        NewID(),
		CreatedAt: time.Now(),
	}
}

func (m *Model) Touch(fn func(*Model)) {
	fn(m)
	m.touch()
}

func (m *Model) Delete() {
	now := time.Now()
	m.UpdatedAt = &now
	m.DeletedAt = &now
}

func (m *Model) touch() {
	now := time.Now()
	m.UpdatedAt = &now
}
