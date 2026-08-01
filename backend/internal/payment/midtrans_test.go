package payment

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
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
		customer := body["customer_details"].(map[string]any)
		if _, ok := customer["phone"]; ok {
			t.Fatalf("empty optional phone must be omitted: %#v", customer)
		}
		if _, ok := customer["email"]; ok {
			t.Fatalf("empty optional email must be omitted: %#v", customer)
		}
		if body["callbacks"].(map[string]any)["finish"] != "https://landing.test/payments/donation-1" {
			t.Fatalf("unexpected finish callback: %#v", body["callbacks"])
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
		FinishURL:   "https://landing.test/payments/donation-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Token != "snap-token" || res.RedirectURL != "https://snap.test" {
		t.Fatalf("unexpected response: %#v", res)
	}
}

func TestVerifyMidtransSignature(t *testing.T) {
	signature := "ab3e0fa6b1e70ba84dc3be153ef77390bdab28edb8af7a752077af945eec9655d762b5dc4d70bf9db4d7479f2418279c66f940e01b1de266f0e584eeb468e535"
	if !VerifyMidtransSignature("order-1", "200", "10000.00", "server-key", signature) {
		t.Fatal("expected valid signature")
	}
	if VerifyMidtransSignature("order-1", "201", "10000.00", "server-key", signature) {
		t.Fatal("expected invalid signature")
	}
}

func TestMidtransStatusRejectsInactiveTransaction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"status_code": "404", "status_message": "Transaction doesn't exist."})
	}))
	defer server.Close()
	client := &MidtransClient{serverKey: "server-key", baseURL: server.URL, httpClient: server.Client()}
	_, err := client.GetTransactionStatus(context.Background(), "order-1")
	if !errors.Is(err, ErrTransactionNotFound) {
		t.Fatalf("inactive transaction error = %v", err)
	}
}
