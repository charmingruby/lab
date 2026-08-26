package model

import (
	"errors"
	"fmt"

	"github.com/charmingruby/new/internal/shared/core"
)

type TicketStatus string

const (
	OpenTicketStatus       TicketStatus = "open"
	InProgressTicketStatus TicketStatus = "in_progress"
	ResolvedTicketStatus   TicketStatus = "resolved"
)

type TicketPriority string

const (
	LowPriority    TicketPriority = "low"
	MediumPriority TicketPriority = "medium"
	HighPriority   TicketPriority = "high"
)

var (
	ErrInvalidTicketTransition = errors.New("invalid ticket status transition")
	ErrInvalidPriority         = errors.New("invalid ticket priority")
)

type Ticket struct {
	AssigneeID  *string        `db:"assignee_id" json:"assignee_id"`
	Title       string         `db:"title"       json:"title"`
	Description string         `db:"description" json:"description"`
	Status      TicketStatus   `db:"status"      json:"status"`
	Priority    TicketPriority `db:"priority"    json:"priority"`
	core.Model  `               db:",inline"`
}

type TicketInput struct {
	Title       string
	Description string
	Priority    string
}

func NewTicket(input TicketInput) (*Ticket, error) {
	priority := TicketPriority(input.Priority)

	if !priority.Valid() {
		return nil, ErrInvalidPriority
	}

	return &Ticket{
		Model:       core.NewModel(),
		Title:       input.Title,
		Description: input.Description,
		Status:      OpenTicketStatus,
		Priority:    priority,
	}, nil
}

func (t *Ticket) Assign(assigneeID string) error {
	return t.transitionTo(InProgressTicketStatus, func() {
		t.AssigneeID = &assigneeID
	})
}

func (t *Ticket) Resolve() error {
	return t.transitionTo(ResolvedTicketStatus, nil)
}

func (t *Ticket) transitionTo(status TicketStatus, after func()) error {
	if !t.Status.CanTransitionTo(status) {
		return fmt.Errorf("%w: %s -> %s", ErrInvalidTicketTransition, t.Status, status)
	}

	t.Touch(func(m *core.Model) {
		t.Status = status
		if after != nil {
			after()
		}
	})

	return nil
}

func (s TicketStatus) CanTransitionTo(target TicketStatus) bool {
	switch s {
	case OpenTicketStatus:
		return target == InProgressTicketStatus || target == ResolvedTicketStatus
	case InProgressTicketStatus:
		return target == ResolvedTicketStatus
	case ResolvedTicketStatus:
		return false
	default:
		return false
	}
}

func (p TicketPriority) Valid() bool {
	switch p {
	case LowPriority, MediumPriority, HighPriority:
		return true
	default:
		return false
	}
}

func (p TicketPriority) String() string {
	return string(p)
}
