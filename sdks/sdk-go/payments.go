package openpay

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// PaymentsService handles payment operations.
type PaymentsService struct {
	client *Client
}

// Create creates a new payment.
func (s *PaymentsService) Create(ctx context.Context, input CreatePaymentInput) (*Payment, error) {
	var resp struct {
		Data Payment `json:"data"`
	}
	if err := s.client.doRequest(ctx, "POST", "/v1/sdk/payments", input, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// Get retrieves a payment by ID.
func (s *PaymentsService) Get(ctx context.Context, id string) (*Payment, error) {
	var resp struct {
		Data Payment `json:"data"`
	}
	if err := s.client.doRequest(ctx, "GET", fmt.Sprintf("/v1/sdk/payments/%s", id), nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// List lists payments with optional filtering and pagination.
func (s *PaymentsService) List(ctx context.Context, params *ListPaymentsParams) (*PaginatedPayments, error) {
	path := "/v1/sdk/payments"
	if params != nil {
		q := url.Values{}
		if params.Page > 0 {
			q.Set("page", strconv.Itoa(params.Page))
		}
		if params.PerPage > 0 {
			q.Set("perPage", strconv.Itoa(params.PerPage))
		}
		if params.Status != "" {
			q.Set("status", params.Status)
		}
		if params.Search != "" {
			q.Set("search", params.Search)
		}
		if params.BranchID != "" {
			q.Set("branchId", params.BranchID)
		}
		if params.DateFrom != "" {
			q.Set("dateFrom", params.DateFrom)
		}
		if params.DateTo != "" {
			q.Set("dateTo", params.DateTo)
		}
		if qs := q.Encode(); qs != "" {
			path += "?" + qs
		}
	}

	var resp PaginatedPayments
	if err := s.client.doRequest(ctx, "GET", path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
