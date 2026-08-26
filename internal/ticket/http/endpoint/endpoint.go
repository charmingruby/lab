package endpoint

import "github.com/charmingruby/new/internal/ticket/usecase"

type Endpoint struct {
	createTicket usecase.CreateTicketUsecase
	assignTicket usecase.AssignTicketUsecase
	getTicket    usecase.GetTicketUsecase
	listTickets  usecase.ListTicketsUsecase
}

func New(
	createTicket usecase.CreateTicketUsecase,
	assignTicket usecase.AssignTicketUsecase,
	getTicket usecase.GetTicketUsecase,
	listTickets usecase.ListTicketsUsecase,
) *Endpoint {
	return &Endpoint{
		createTicket: createTicket,
		assignTicket: assignTicket,
		getTicket:    getTicket,
		listTickets:  listTickets,
	}
}
