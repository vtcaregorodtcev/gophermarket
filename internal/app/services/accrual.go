package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/vtcaregorodtcev/gophermarket/internal/logger"
	"github.com/vtcaregorodtcev/gophermarket/internal/models"
)

type AccrualService struct {
	addr string
}

func NewAccrualService(addr string) *AccrualService {
	return &AccrualService{addr: addr}
}

type CalcOrderAccrualResponse struct {
	Order   string             `json:"order"`
	Accrual float64            `json:"accrual"`
	Status  models.OrderStatus `json:"status"`
}

func (s *AccrualService) CalcOrderAccrual(ctx context.Context, orderNumber string) (*CalcOrderAccrualResponse, error) {
	url := fmt.Sprintf("%s/api/orders/%s", s.addr, orderNumber)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	logger.Infof("CalcOrderAccrual response status: %d", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response CalcOrderAccrualResponse

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	r, err := json.Marshal(response)
	if err == nil {
		logger.Infof("CalcOrderAccrual response: %s", r)
	}

	return &response, nil
}
