package services

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/vtcaregorodtcev/gophermarket/cmd/gophermart/pkg/models"
)

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
	Order   string             `json:"order"`
	Accrual float64            `json:"accrual"`
	Status  models.OrderStatus `json:"status"`
}

func (s *AccrualService) CalcOrderAccrual(orderNumber string) (*CalcOrderAccrualResponse, error) {
	url := fmt.Sprintf("%s/api/orders/%s", s.addr, orderNumber)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response CalcOrderAccrualResponse

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
