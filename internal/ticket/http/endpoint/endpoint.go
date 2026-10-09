package endpoint

import "github.com/charmingruby/lab/internal/ticket/usecase"

type Endpoint struct {
	uc *usecase.Usecase
}

func New(uc *usecase.Usecase) *Endpoint {
	return &Endpoint{
		uc: uc,
	}
}
