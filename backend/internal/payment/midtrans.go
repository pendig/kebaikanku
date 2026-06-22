package payment

import (
	"bytes"
	"context"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type MidtransClient struct {
	serverKey  string
	baseURL    string
	httpClient *http.Client
}

type SnapRequest struct {
	OrderID     string
	GrossAmount int64
	DonorName   string
	DonorEmail  string
	DonorPhone  string
	ItemName    string
}

type SnapResponse struct {
	Token       string `json:"token"`
	RedirectURL string `json:"redirect_url"`
}

func NewMidtransClient(env, serverKey string) *MidtransClient {
	baseURL := "https://app.sandbox.midtrans.com"
	if strings.EqualFold(env, "production") {
		baseURL = "https://app.midtrans.com"
	}
	return &MidtransClient{
		serverKey: serverKey,
		baseURL:   baseURL,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func VerifyMidtransSignature(orderID, statusCode, grossAmount, serverKey, signature string) bool {
	if orderID == "" || statusCode == "" || grossAmount == "" || serverKey == "" || signature == "" {
		return false
	}
	sum := sha512.Sum512([]byte(orderID + statusCode + grossAmount + serverKey))
	expected := hex.EncodeToString(sum[:])
	return subtle.ConstantTimeCompare([]byte(expected), []byte(strings.ToLower(signature))) == 1
}

func (c *MidtransClient) CreateSnapTransaction(ctx context.Context, req SnapRequest) (*SnapResponse, error) {
	if c == nil || c.serverKey == "" {
		return nil, errors.New("midtrans server key is not configured")
	}
	if req.OrderID == "" || req.GrossAmount < 1 {
		return nil, errors.New("order id and gross amount are required")
	}

	body := map[string]any{
		"transaction_details": map[string]any{
			"order_id":     req.OrderID,
			"gross_amount": req.GrossAmount,
		},
		"customer_details": map[string]any{
			"first_name": req.DonorName,
			"email":      req.DonorEmail,
			"phone":      req.DonorPhone,
		},
		"item_details": []map[string]any{{
			"id":       req.OrderID,
			"name":     req.ItemName,
			"price":    req.GrossAmount,
			"quantity": 1,
		}},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/snap/v1/transactions", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(c.serverKey+":")))

	res, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, fmt.Errorf("midtrans snap request failed: status %d", res.StatusCode)
	}

	var snap SnapResponse
	if err := json.NewDecoder(res.Body).Decode(&snap); err != nil {
		return nil, err
	}
	return &snap, nil
}
