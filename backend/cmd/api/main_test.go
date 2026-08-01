package main

import (
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kebaikankuid/kebaikanku/backend/internal/config"
	"github.com/kebaikankuid/kebaikanku/backend/internal/database"
	"github.com/kebaikankuid/kebaikanku/backend/internal/domain"
	"github.com/kebaikankuid/kebaikanku/backend/internal/payment"
	"github.com/kebaikankuid/kebaikanku/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func authenticatedRequest(t *testing.T, router http.Handler, method, path string, body *bytes.Buffer, contentType string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, body)
	req.AddCookie(adminCookie(t, router))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	return response
}

func TestAdminUploadPaginationAndEncryptedSettings(t *testing.T) {
	db, router, _, _ := setupAPITest(t, 50)
	uploadDir := t.TempDir()
	appConfig.UploadDir = uploadDir
	appConfig.PublicUploadBaseURL = "https://cdn.example.test/campaigns"
	appConfig.AdminSettingsEncryptionKey = base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{9}, 32))

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "banner.png")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write([]byte("\x89PNG\r\n\x1a\n" + strings.Repeat("x", 32)))
	_ = writer.Close()
	response := authenticatedRequest(t, router, http.MethodPost, "/api/v1/admin/uploads", &body, writer.FormDataContentType())
	if response.Code != http.StatusCreated || !strings.Contains(response.Body.String(), "https://cdn.example.test/campaigns/") {
		t.Fatalf("upload = %d: %s", response.Code, response.Body.String())
	}

	campaign := createCampaign(t, db, "pagination", time.Now().Add(time.Hour))
	donor := domain.Donor{ID: "pagination-donor", Name: "Donor", PhoneNumber: "+628111111"}
	if err := db.Create(&donor).Error; err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		status := "pending"
		if i == 2 {
			status = "success"
		}
		if err := db.Create(&domain.Donation{ID: fmt.Sprintf("page-%d", i), CampaignID: campaign.ID, DonorID: donor.ID, Amount: 5000, Status: status}).Error; err != nil {
			t.Fatal(err)
		}
	}
	response = authenticatedRequest(t, router, http.MethodGet, "/api/v1/admin/donations?page=2&limit=1&status=pending", &bytes.Buffer{}, "")
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"total":2`) || !strings.Contains(response.Body.String(), `"pages":2`) {
		t.Fatalf("pagination = %d: %s", response.Code, response.Body.String())
	}
	response = authenticatedRequest(t, router, http.MethodGet, "/api/v1/admin/donations?campaign_id="+campaign.ID+"&sort=amount_desc", &bytes.Buffer{}, "")
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"total":3`) {
		t.Fatalf("campaign filter = %d: %s", response.Code, response.Body.String())
	}

	settings := `{"mode":"sandbox","server_key":"SB-Mid-server-supersecret","client_key":"SB-Mid-client-publicish"}`
	response = authenticatedRequest(t, router, http.MethodPut, "/api/v1/admin/settings/payment", bytes.NewBufferString(settings), "application/json")
	if response.Code != http.StatusOK || strings.Contains(response.Body.String(), "supersecret") {
		t.Fatalf("settings = %d: %s", response.Code, response.Body.String())
	}
	var stored domain.PaymentSetting
	if err := db.First(&stored, "id = ?", "midtrans").Error; err != nil {
		t.Fatal(err)
	}
	if strings.Contains(stored.ServerKeyCipher, "supersecret") {
		t.Fatal("stored plaintext server key")
	}
	_, effectiveKey, err := effectivePaymentConfig()
	if err != nil || effectiveKey != "SB-Mid-server-supersecret" {
		t.Fatalf("effective override = %q, %v", effectiveKey, err)
	}
	response = authenticatedRequest(t, router, http.MethodGet, "/api/v1/admin/settings/payment", &bytes.Buffer{}, "")
	if response.Code != http.StatusOK || strings.Contains(response.Body.String(), "supersecret") || !strings.Contains(response.Body.String(), `"source":"override"`) || !strings.Contains(response.Body.String(), "SB-Mid-server-") {
		t.Fatalf("redacted settings = %d: %s", response.Code, response.Body.String())
	}
	appConfig.MidtransServerKey = "SB-Mid-server-environment"
	appConfig.MidtransClientKey = "SB-Mid-client-environment"
	response = authenticatedRequest(t, router, http.MethodPut, "/api/v1/admin/settings/payment", bytes.NewBufferString(`{"mode":"sandbox","server_key":"","client_key":""}`), "application/json")
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"source":"environment"`) {
		t.Fatalf("reset settings = %d: %s", response.Code, response.Body.String())
	}
}

func TestAdminUploadRejectsHTML(t *testing.T) {
	_, router, _, _ := setupAPITest(t, 50)
	appConfig.UploadDir = t.TempDir()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, _ := writer.CreateFormFile("file", "fake.png")
	_, _ = part.Write([]byte("<!doctype html><title>not an image</title>"))
	_ = writer.Close()
	response := authenticatedRequest(t, router, http.MethodPost, "/api/v1/admin/uploads", &body, writer.FormDataContentType())
	if response.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("upload = %d: %s", response.Code, response.Body.String())
	}
}

func setupAPITest(t *testing.T, limit int) (*gorm.DB, http.Handler, *httptest.Server, *int32) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "-"))), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.Organization{}, &domain.Campaign{}, &domain.Donor{}, &domain.Donation{}, &domain.Waitlist{}, &domain.PaymentSetting{}); err != nil {
		t.Fatal(err)
	}
	var snapCalls int32
	snapServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v2/") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"order_id":"order","transaction_id":"trx-sync","transaction_status":"settlement","transaction_time":"2026-07-30 12:00:00"}`))
			return
		}
		atomic.AddInt32(&snapCalls, 1)
		if r.URL.Path != "/snap/v1/transactions" {
			t.Fatalf("Snap path = %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"token":"snap-token","redirect_url":"https://pay.example/snap-token"}`))
	}))
	t.Cleanup(snapServer.Close)
	appConfig = &config.Config{
		Env:                "test",
		AdminPassword:      "test-admin-password",
		AdminSessionSecret: "test-admin-session-secret-which-is-long-enough",
		MidtransServerKey:  "server-key",
		PublicRateLimit:    limit,
	}
	appStore = repository.NewStore(db)
	database.DB = db
	appPayment = payment.NewMidtransClientWithBaseURL("server-key", snapServer.URL, snapServer.Client())
	appPaymentServerKey = "server-key"
	publicRequestGuard.Lock()
	publicRequestWindow = map[string]rateLimitEntry{}
	publicRequestCleaned = time.Time{}
	publicRequestGuard.Unlock()
	router, err := newRouter(appConfig)
	if err != nil {
		t.Fatal(err)
	}
	return db, router, snapServer, &snapCalls
}

func TestPaymentStatusRefreshQueriesMidtransAndCountsOnce(t *testing.T) {
	db, router, _, _ := setupAPITest(t, 20)
	campaign := createCampaign(t, db, "campaign-status-refresh", time.Now().Add(time.Hour))
	checkout := requestJSON(t, router, http.MethodPost, "/api/v1/donations", donationPayloadJSON(campaign.ID, 10000), nil)
	donationID := responseDonationID(t, checkout)
	for i := 0; i < 2; i++ {
		response := requestJSON(t, router, http.MethodGet, "/api/v1/donations/"+donationID+"/status", "", nil)
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"status":"success"`) || !strings.Contains(response.Body.String(), `"provider_status":"settlement"`) {
			t.Fatalf("status refresh %d = %d: %s", i, response.Code, response.Body.String())
		}
	}
	var updated domain.Campaign
	if err := db.First(&updated, "id = ?", campaign.ID).Error; err != nil {
		t.Fatal(err)
	}
	if updated.CollectedAmount != 10000 {
		t.Fatalf("collected = %.0f, want 10000", updated.CollectedAmount)
	}
}

func TestMidtransStatusMapping(t *testing.T) {
	for provider, want := range map[string]string{"settlement": "success", "capture": "success", "pending": "pending", "expire": "failed", "cancel": "failed", "deny": "failed", "failure": "failed"} {
		fraud := ""
		if provider == "capture" {
			fraud = "accept"
		}
		if got := mapMidtransStatus(provider, fraud); got != want {
			t.Fatalf("%s = %s, want %s", provider, got, want)
		}
	}
}

func createCampaign(t *testing.T, db *gorm.DB, id string, endDate time.Time) domain.Campaign {
	t.Helper()
	organization := domain.Organization{ID: "org-" + id, Name: "Organization " + id, Email: id + "@example.test", PasswordHash: "test"}
	campaign := domain.Campaign{ID: id, OrganizationID: organization.ID, Title: "Campaign " + id, Slug: "campaign-" + id, Category: "infak", TargetAmount: 100000, EndDate: endDate, Status: "active"}
	if err := db.Create(&organization).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&campaign).Error; err != nil {
		t.Fatal(err)
	}
	return campaign
}

func requestJSON(t *testing.T, router http.Handler, method, path, payload string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(payload))
	req.RemoteAddr = "198.51.100.10:1234"
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	return response
}

func adminCookie(t *testing.T, router http.Handler) *http.Cookie {
	t.Helper()
	response := requestJSON(t, router, http.MethodPost, "/api/v1/admin/login", `{"password":"test-admin-password"}`, nil)
	if response.Code != http.StatusOK {
		t.Fatalf("admin login = %d: %s", response.Code, response.Body.String())
	}
	for _, cookie := range response.Result().Cookies() {
		if cookie.Name == adminSessionCookie {
			return cookie
		}
	}
	t.Fatal("admin login did not set session cookie")
	return nil
}

func donationPayloadJSON(campaignID string, amount int) string {
	return fmt.Sprintf(`{"campaign_id":%q,"donor":{"name":"Budi","phone_number":"+6281234567890","email":"budi@example.com"},"amount":%d,"platform_tip":0,"payment_method":"midtrans_snap"}`, campaignID, amount)
}

func responseDonationID(t *testing.T, response *httptest.ResponseRecorder) string {
	t.Helper()
	var payload struct {
		Data struct {
			DonationID string `json:"donation_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Data.DonationID == "" {
		t.Fatalf("missing donation id in %s", response.Body.String())
	}
	return payload.Data.DonationID
}

func TestDonationCheckoutIsIdempotent(t *testing.T) {
	db, router, _, snapCalls := setupAPITest(t, 20)
	campaign := createCampaign(t, db, "campaign-idempotent", time.Now().Add(time.Hour))
	payload := donationPayloadJSON(campaign.ID, 10000)
	first := requestJSON(t, router, http.MethodPost, "/api/v1/donations", payload, map[string]string{"Idempotency-Key": "checkout-1"})
	if first.Code != http.StatusCreated {
		t.Fatalf("first checkout = %d: %s", first.Code, first.Body.String())
	}
	second := requestJSON(t, router, http.MethodPost, "/api/v1/donations", payload, map[string]string{"Idempotency-Key": "checkout-1"})
	if second.Code != http.StatusOK {
		t.Fatalf("retry checkout = %d: %s", second.Code, second.Body.String())
	}
	if responseDonationID(t, first) != responseDonationID(t, second) {
		t.Fatal("retry returned a different donation")
	}
	if got := atomic.LoadInt32(snapCalls); got != 1 {
		t.Fatalf("Snap calls = %d, want 1", got)
	}
	var count int64
	if err := db.Model(&domain.Donation{}).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("donations = %d, want 1", count)
	}
}

func TestProductionDonationRequiresIdempotencyKey(t *testing.T) {
	db, router, _, _ := setupAPITest(t, 20)
	appConfig.Env = "production"
	campaign := createCampaign(t, db, "campaign-production-idempotency", time.Now().Add(time.Hour))
	response := requestJSON(t, router, http.MethodPost, "/api/v1/donations", donationPayloadJSON(campaign.ID, 10000), nil)
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "IDEMPOTENCY_KEY_REQUIRED") {
		t.Fatalf("missing idempotency key = %d: %s", response.Code, response.Body.String())
	}
}

func TestFailedIdempotentCheckoutDoesNotCreateDuplicate(t *testing.T) {
	db, router, _, _ := setupAPITest(t, 20)
	appPayment = payment.NewMidtransClient("sandbox", "")
	campaign := createCampaign(t, db, "campaign-failed-idempotency", time.Now().Add(time.Hour))
	payload := donationPayloadJSON(campaign.ID, 10000)
	headers := map[string]string{"Idempotency-Key": "failed-checkout-1"}
	first := requestJSON(t, router, http.MethodPost, "/api/v1/donations", payload, headers)
	if first.Code != http.StatusBadGateway {
		t.Fatalf("failed checkout = %d: %s", first.Code, first.Body.String())
	}
	second := requestJSON(t, router, http.MethodPost, "/api/v1/donations", payload, headers)
	if second.Code != http.StatusConflict || !strings.Contains(second.Body.String(), "IDEMPOTENCY_PREVIOUSLY_FAILED") {
		t.Fatalf("failed checkout retry = %d: %s", second.Code, second.Body.String())
	}
	var donations, donors int64
	if err := db.Model(&domain.Donation{}).Count(&donations).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&domain.Donor{}).Count(&donors).Error; err != nil {
		t.Fatal(err)
	}
	if donations != 1 || donors != 1 {
		t.Fatalf("failed retry created donations=%d donors=%d, want 1 each", donations, donors)
	}
}

func TestDonationRejectsExpiredCampaignAndInvalidAmount(t *testing.T) {
	db, router, _, snapCalls := setupAPITest(t, 20)
	expired := createCampaign(t, db, "campaign-expired", time.Now().Add(-time.Minute))
	if available, err := appStore.GetActiveCampaignByID(expired.ID); err != nil || available != nil {
		t.Fatalf("expired campaign availability = %#v, %v", available, err)
	}
	response := requestJSON(t, router, http.MethodPost, "/api/v1/donations", donationPayloadJSON(expired.ID, 10000), nil)
	if response.Code != http.StatusNotFound {
		t.Fatalf("expired campaign = %d: %s", response.Code, response.Body.String())
	}
	active := createCampaign(t, db, "campaign-validation", time.Now().Add(time.Hour))
	response = requestJSON(t, router, http.MethodPost, "/api/v1/donations", donationPayloadJSON(active.ID, 1000), nil)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("invalid amount = %d: %s", response.Code, response.Body.String())
	}
	if got := atomic.LoadInt32(snapCalls); got != 0 {
		t.Fatalf("Snap calls = %d, want 0", got)
	}
}

func TestPaymentNotificationContractAndExport(t *testing.T) {
	db, router, _, _ := setupAPITest(t, 20)
	campaign := createCampaign(t, db, "campaign-webhook", time.Now().Add(time.Hour))
	checkout := requestJSON(t, router, http.MethodPost, "/api/v1/donations", donationPayloadJSON(campaign.ID, 10000), nil)
	if checkout.Code != http.StatusCreated {
		t.Fatalf("checkout = %d: %s", checkout.Code, checkout.Body.String())
	}
	donationID := responseDonationID(t, checkout)
	var donation domain.Donation
	if err := db.First(&donation, "id = ?", donationID).Error; err != nil {
		t.Fatal(err)
	}
	invalid := requestJSON(t, router, http.MethodPost, "/api/v1/payments/midtrans/notification", fmt.Sprintf(`{"order_id":%q,"status_code":"200","gross_amount":"10000","transaction_status":"settlement","transaction_id":"trx-1","transaction_time":"2026-07-30 12:00:00","signature_key":"invalid"}`, donation.ProviderOrderID), nil)
	if invalid.Code != http.StatusUnauthorized {
		t.Fatalf("invalid callback = %d: %s", invalid.Code, invalid.Body.String())
	}
	if err := db.First(&donation, "id = ?", donationID).Error; err != nil {
		t.Fatal(err)
	}
	if donation.Status != "pending" {
		t.Fatalf("invalid callback changed donation to %q", donation.Status)
	}
	sum := sha512.Sum512([]byte(donation.ProviderOrderID + "200" + "10000" + "server-key"))
	payload := fmt.Sprintf(`{"order_id":%q,"status_code":"200","gross_amount":"10000","transaction_status":"settlement","transaction_id":"trx-1","transaction_time":"2026-07-30 12:00:00","signature_key":%q}`, donation.ProviderOrderID, hex.EncodeToString(sum[:]))
	for i := 0; i < 2; i++ {
		response := requestJSON(t, router, http.MethodPost, "/api/v1/payments/midtrans/notification", payload, nil)
		if response.Code != http.StatusOK {
			t.Fatalf("settlement %d = %d: %s", i, response.Code, response.Body.String())
		}
	}
	var updated domain.Campaign
	if err := db.First(&updated, "id = ?", campaign.ID).Error; err != nil {
		t.Fatal(err)
	}
	if updated.CollectedAmount != 10000 {
		t.Fatalf("collected = %.0f, want 10000", updated.CollectedAmount)
	}
	cookie := adminCookie(t, router)
	export := requestJSON(t, router, http.MethodGet, "/api/v1/donations/export", "", map[string]string{"Cookie": cookie.String()})
	if export.Code != http.StatusOK || !strings.Contains(export.Body.String(), "provider_order_id") || !strings.Contains(export.Body.String(), donation.ProviderOrderID) || !strings.Contains(export.Body.String(), "settlement") {
		t.Fatalf("export = %d: %s", export.Code, export.Body.String())
	}
}

func TestCampaignListIncludesInactiveOnlyForAdminSession(t *testing.T) {
	db, router, _, _ := setupAPITest(t, 20)
	active := createCampaign(t, db, "campaign-list-active", time.Now().Add(time.Hour))
	paused := createCampaign(t, db, "campaign-list-paused", time.Now().Add(time.Hour))
	if err := db.Model(&domain.Campaign{}).Where("id = ?", paused.ID).Update("status", "paused").Error; err != nil {
		t.Fatal(err)
	}
	public := requestJSON(t, router, http.MethodGet, "/api/v1/campaigns?include_inactive=true", "", nil)
	if public.Code != http.StatusOK || !strings.Contains(public.Body.String(), active.ID) || strings.Contains(public.Body.String(), paused.ID) {
		t.Fatalf("public campaign list = %d: %s", public.Code, public.Body.String())
	}
	cookie := adminCookie(t, router)
	admin := requestJSON(t, router, http.MethodGet, "/api/v1/campaigns?include_inactive=true", "", map[string]string{"Cookie": cookie.String()})
	if admin.Code != http.StatusOK || !strings.Contains(admin.Body.String(), paused.ID) {
		t.Fatalf("admin campaign list = %d: %s", admin.Code, admin.Body.String())
	}
}

func TestHTTPGuards(t *testing.T) {
	options, err := configuredCORS(&config.Config{Env: "test"})
	if err != nil || !strings.Contains(strings.Join(options.AllowedMethods, ","), "PATCH") {
		t.Fatalf("CORS must allow campaign status PATCH: %#v %v", options.AllowedMethods, err)
	}
	if _, err := configuredCORS(&config.Config{Env: "production"}); err == nil {
		t.Fatal("production accepted missing CORS origins")
	}
	if _, err := configuredCORS(&config.Config{Env: "production", CORSAllowedOrigins: "http://localhost:5173"}); err == nil {
		t.Fatal("production accepted localhost CORS origin")
	}
	if _, err := configuredCORS(&config.Config{Env: "production", CORSAllowedOrigins: "https://dashboard.example"}); err != nil {
		t.Fatalf("production rejected HTTPS origin: %v", err)
	}
	db, router, _, _ := setupAPITest(t, 1)
	local := httptest.NewRequest(http.MethodGet, "/health", nil)
	local.Header.Set("Origin", "http://localhost:18481")
	localResponse := httptest.NewRecorder()
	router.ServeHTTP(localResponse, local)
	if got := localResponse.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:18481" {
		t.Fatalf("local CORS origin = %q", got)
	}
	campaign := createCampaign(t, db, "campaign-guards", time.Now().Add(time.Hour))
	first := requestJSON(t, router, http.MethodPost, "/api/v1/donations", donationPayloadJSON(campaign.ID, 10000), nil)
	if first.Code != http.StatusCreated {
		t.Fatalf("first request = %d", first.Code)
	}
	second := requestJSON(t, router, http.MethodPost, "/api/v1/donations", donationPayloadJSON(campaign.ID, 10000), nil)
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("rate limit = %d: %s", second.Code, second.Body.String())
	}
	cookie := adminCookie(t, router)
	for attempt := 1; attempt <= 11; attempt++ {
		login := requestJSON(t, router, http.MethodPost, "/api/v1/admin/login", `{"password":"wrong"}`, nil)
		if attempt == 11 && login.Code != http.StatusTooManyRequests {
			t.Fatalf("admin login rate limit = %d: %s", login.Code, login.Body.String())
		}
	}
	oversized := requestJSON(t, router, http.MethodPost, "/api/v1/campaigns", `{"x":"`+strings.Repeat("a", maxCampaignBody)+`"}`, map[string]string{"Cookie": cookie.String()})
	if oversized.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversized request = %d: %s", oversized.Code, oversized.Body.String())
	}
}

func TestAdminPasswordSessionAndCampaignStatus(t *testing.T) {
	db, router, _, _ := setupAPITest(t, 20)
	campaign := createCampaign(t, db, "campaign-admin", time.Now().Add(time.Hour))
	unauthorized := requestJSON(t, router, http.MethodPatch, "/api/v1/campaigns/"+campaign.ID+"/status", `{"status":"paused"}`, nil)
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status update = %d", unauthorized.Code)
	}
	invalid := requestJSON(t, router, http.MethodPost, "/api/v1/admin/login", `{"password":"wrong"}`, nil)
	if invalid.Code != http.StatusUnauthorized {
		t.Fatalf("invalid login = %d", invalid.Code)
	}
	cookie := adminCookie(t, router)
	if !cookie.HttpOnly || cookie.SameSite != http.SameSiteLaxMode || cookie.Secure {
		t.Fatalf("unexpected development cookie: %+v", cookie)
	}
	invalidStatus := requestJSON(t, router, http.MethodPatch, "/api/v1/campaigns/"+campaign.ID+"/status", `{"status":"draft"}`, map[string]string{"Cookie": cookie.String()})
	if invalidStatus.Code != http.StatusBadRequest {
		t.Fatalf("invalid campaign status = %d", invalidStatus.Code)
	}
	session := requestJSON(t, router, http.MethodGet, "/api/v1/admin/session", "", map[string]string{"Cookie": cookie.String()})
	if session.Code != http.StatusOK || !strings.Contains(session.Body.String(), `"authenticated":true`) {
		t.Fatalf("session = %d: %s", session.Code, session.Body.String())
	}
	updated := requestJSON(t, router, http.MethodPatch, "/api/v1/campaigns/"+campaign.ID+"/status", `{"status":"paused"}`, map[string]string{"Cookie": cookie.String()})
	if updated.Code != http.StatusOK {
		t.Fatalf("status update = %d: %s", updated.Code, updated.Body.String())
	}
	var stored domain.Campaign
	if err := db.First(&stored, "id = ?", campaign.ID).Error; err != nil {
		t.Fatal(err)
	}
	if stored.Status != "paused" {
		t.Fatalf("campaign status = %q", stored.Status)
	}
	logout := requestJSON(t, router, http.MethodPost, "/api/v1/admin/logout", "", map[string]string{"Cookie": cookie.String()})
	if logout.Code != http.StatusNoContent {
		t.Fatalf("logout = %d", logout.Code)
	}
	if cookies := logout.Result().Cookies(); len(cookies) != 1 || cookies[0].MaxAge >= 0 {
		t.Fatalf("logout cookie = %+v", cookies)
	}
	loggedOutSession := requestJSON(t, router, http.MethodGet, "/api/v1/admin/session", "", nil)
	if loggedOutSession.Code != http.StatusOK || !strings.Contains(loggedOutSession.Body.String(), `"authenticated":false`) {
		t.Fatalf("logged-out session = %d: %s", loggedOutSession.Code, loggedOutSession.Body.String())
	}
}

func TestGeneratedAdminPasswordCanCreateSession(t *testing.T) {
	db, _, _, _ := setupAPITest(t, 20)
	hash, err := bcrypt.GenerateFromPassword([]byte("generated-random-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	organization := domain.Organization{ID: "generated-admin", Name: "Admin", Email: "generated@example.test", PasswordHash: string(hash), Status: "active"}
	if err := db.Create(&organization).Error; err != nil {
		t.Fatal(err)
	}
	appConfig.AdminPassword = ""
	appConfig.AdminSessionSecret = ""
	router, err := newRouter(appConfig)
	if err != nil {
		t.Fatal(err)
	}
	login := requestJSON(t, router, http.MethodPost, "/api/v1/admin/login", `{"password":"generated-random-password"}`, nil)
	if login.Code != http.StatusOK || !strings.Contains(login.Body.String(), `"must_change_password":true`) {
		t.Fatalf("generated login = %d: %s", login.Code, login.Body.String())
	}
}

func TestProductionAdminConfigRequiresPasswordAndSessionSecret(t *testing.T) {
	if err := validateAdminConfig(&config.Config{Env: "production"}); err == nil {
		t.Fatal("production accepted missing admin credentials")
	}
	valid := &config.Config{Env: "production", AdminPassword: "long-admin-password", AdminSessionSecret: strings.Repeat("s", 32)}
	if err := validateAdminConfig(valid); err != nil {
		t.Fatalf("valid production admin config rejected: %v", err)
	}
	cookie, err := newAdminSession(valid, time.Now().Add(time.Hour))
	if err != nil || !cookie.Secure || !cookie.HttpOnly || cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("production cookie = %+v, %v", cookie, err)
	}
}

func TestReadinessChecksDatabase(t *testing.T) {
	_, router, _, _ := setupAPITest(t, 20)
	ready := httptest.NewRecorder()
	router.ServeHTTP(ready, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if ready.Code != http.StatusOK {
		t.Fatalf("ready = %d: %s", ready.Code, ready.Body.String())
	}

	database.DB = nil
	unready := httptest.NewRecorder()
	router.ServeHTTP(unready, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if unready.Code != http.StatusServiceUnavailable {
		t.Fatalf("unready = %d: %s", unready.Code, unready.Body.String())
	}
}
