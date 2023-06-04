package services

import "github.com/vtcaregorodtcev/gophermarket/cmd/gophermart/pkg/models"

type AccrualService struct {
	addr string
}

var accrualServiceInstance *AccrualService

func NewAccrualService(addr string) *AccrualService {
	accrualServiceInstance = &AccrualService{addr: addr}

	return accrualServiceInstance
}

func GetAccrualServiceInstance() *AccrualService {
	return accrualServiceInstance
}

type CalcOrderAccrualResponse struct {
	Accrual float64
	Status  models.OrderStatus
}

func (s *AccrualService) CalcOrderAccrual(orderID uint) (*CalcOrderAccrualResponse, error) {
	return &CalcOrderAccrualResponse{
		Accrual: 100.0,
		Status:  models.PROCESSED,
	}, nil
}
