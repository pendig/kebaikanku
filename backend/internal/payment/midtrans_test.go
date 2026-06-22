package payment

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMidtransCreateSnapTransaction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/snap/v1/transactions" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("server-key:"))
		if r.Header.Get("Authorization") != wantAuth {
			t.Fatalf("unexpected auth header")
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		details := body["transaction_details"].(map[string]any)
		if details["order_id"] != "donation-1" || details["gross_amount"].(float64) != 10000 {
			t.Fatalf("unexpected transaction details: %#v", details)
		}

		_ = json.NewEncoder(w).Encode(SnapResponse{Token: "snap-token", RedirectURL: "https://snap.test"})
	}))
	defer server.Close()

	client := NewMidtransClient("sandbox", "server-key")
	client.baseURL = server.URL
	res, err := client.CreateSnapTransaction(context.Background(), SnapRequest{
		OrderID:     "donation-1",
		GrossAmount: 10000,
		DonorName:   "Budi",
		ItemName:    "Donasi",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Token != "snap-token" || res.RedirectURL != "https://snap.test" {
		t.Fatalf("unexpected response: %#v", res)
	}
}
