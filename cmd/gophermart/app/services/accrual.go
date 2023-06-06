package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rs/zerolog"
	"github.com/vtcaregorodtcev/gophermarket/cmd/gophermart/pkg/logger"
	"github.com/vtcaregorodtcev/gophermarket/cmd/gophermart/pkg/models"
)

type AccrualService struct {
	addr string
	log  *zerolog.Logger
}

var accrualServiceInstance *AccrualService

func NewAccrualService(addr string) *AccrualService {
	accrualServiceInstance = &AccrualService{addr: addr, log: logger.NewLogger("ACCRUAL_SERVICE")}

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

	s.log.Info().Msgf("CalcOrderAccrual response status: %d", resp.StatusCode)
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
		s.log.Info().Msgf("CalcOrderAccrual response: %s", r)
	}

	return &response, nil
}
