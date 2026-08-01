package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/kebaikankuid/kebaikanku/backend/internal/config"
	"github.com/kebaikankuid/kebaikanku/backend/internal/database"
	"github.com/kebaikankuid/kebaikanku/backend/internal/domain"
	"github.com/kebaikankuid/kebaikanku/backend/internal/payment"
	"github.com/kebaikankuid/kebaikanku/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	maxWaitlistBody       = 64 << 10
	maxCampaignBody       = 64 << 10
	maxDonationBody       = 64 << 10
	maxNotificationBody   = 128 << 10
	maxUploadBody         = 5<<20 + 1024
	maxDonationAmount     = 100_000_000
	publicRateLimitWindow = time.Minute
	adminSessionCookie    = "kebaikanku_admin"
	adminSessionTTL       = 12 * time.Hour
)

var (
	publicRequestWindow  = map[string]rateLimitEntry{}
	publicRequestGuard   sync.Mutex
	publicRequestCleaned time.Time
	waitlistEmailRegex   = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	campaignSlugRegex    = regexp.MustCompile(`^[a-z0-9-]+$`)
	appConfig            *config.Config
	appStore             *repository.Store
	appPayment           *payment.MidtransClient
	appPaymentServerKey  string
	appPaymentGuard      sync.RWMutex
)

type rateLimitEntry struct {
	started time.Time
	count   int
}

func main() {
	// 1. Load Configurations
	cfg := config.Load()
	appConfig = cfg

	// 2. Initialize Database and Auto-migrate
	database.Init(cfg)
	initialPassword, err := database.SeedDefaults(database.DB, cfg.AdminPassword)
	if err != nil {
		panic(fmt.Sprintf("Could not seed initial data: %v", err))
	}
	if initialPassword != "" {
		fmt.Printf("IMPORTANT: generated initial admin password: %s (set ADMIN_PASSWORD to replace it)\n", initialPassword)
	}
	if err := validateAdminConfig(cfg); err != nil {
		panic(fmt.Sprintf("Invalid admin authentication configuration: %v", err))
	}
	appStore = repository.NewStore(database.DB)
	if err := refreshPaymentClient(); err != nil {
		panic("Invalid encrypted payment settings configuration")
	}

	r, err := newRouter(cfg)
	if err != nil {
		panic(fmt.Sprintf("Invalid HTTP configuration: %v", err))
	}

	// 4. Start Server
	serverAddr := fmt.Sprintf(":%s", cfg.Port)
	fmt.Printf("Backend API server is running on http://localhost%s in %s mode\n", serverAddr, cfg.Env)

	err = http.ListenAndServe(serverAddr, r)
	if err != nil {
		panic(fmt.Sprintf("Failed to start server: %v", err))
	}
}

func newRouter(cfg *config.Config) (*chi.Mux, error) {
	if err := validateAdminConfig(cfg); err != nil {
		return nil, err
	}
	corsOptions, err := configuredCORS(cfg)
	if err != nil {
		return nil, err
	}
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Logger, middleware.Recoverer, middleware.Timeout(60*time.Second))
	r.Use(cors.Handler(corsOptions))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy","time":"` + time.Now().Format(time.RFC3339) + `"}`))
	})
	r.Get("/readyz", handleReadiness)
	r.Get("/uploads/{name}", handlePublicUpload)
	r.Group(func(r chi.Router) {
		r.Use(publicRateLimit)
		r.Post("/api/v1/waitlist", handleWaitlistSignup)
		r.Get("/api/v1/campaigns", handleListCampaigns)
		r.Get("/api/v1/campaigns/{slug}", handleGetCampaign)
		r.Post("/api/v1/donations", handleCreateDonation)
		r.Get("/api/v1/donations/{id}/status", handleDonationStatus)
	})
	r.With(adminLoginRateLimit).Post("/api/v1/admin/login", handleAdminLogin)
	r.Post("/api/v1/admin/logout", handleAdminLogout)
	r.Get("/api/v1/admin/session", handleAdminSession)
	r.Group(func(r chi.Router) {
		r.Use(requireAdminSession)
		r.Post("/api/v1/campaigns", handleCreateCampaign)
		r.Put("/api/v1/campaigns/{id}", handleUpdateCampaign)
		r.Patch("/api/v1/campaigns/{id}/status", handleUpdateCampaignStatus)
		r.Get("/api/v1/donations/export", handleExportDonations)
		r.Get("/api/v1/admin/donations", handleAdminDonations)
		r.Post("/api/v1/admin/uploads", handleAdminUpload)
		r.Get("/api/v1/admin/settings/payment", handleGetPaymentSettings)
		r.Put("/api/v1/admin/settings/payment", handlePutPaymentSettings)
	})
	r.Post("/api/v1/payments/midtrans/notification", handleMidtransNotification)
	return r, nil
}

func handleReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if database.DB == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "Database is not ready.")
		return
	}
	sqlDB, err := database.DB.DB()
	if err == nil {
		err = sqlDB.PingContext(r.Context())
	}
	if err != nil {
		writeAPIError(w, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "Database is not ready.")
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
}

func handleListCampaigns(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	limit := queryInt(r, "limit", 20)
	if limit < 1 || limit > 100 {
		limit = 20
	}
	page := queryInt(r, "page", 1)
	if page < 1 {
		page = 1
	}

	category := strings.TrimSpace(r.URL.Query().Get("category"))
	includeInactive := r.URL.Query().Get("include_inactive") == "true" && validAdminSession(r, appConfig)
	var campaigns []domain.Campaign
	var err error
	if includeInactive {
		campaigns, err = appStore.ListCampaigns(category, limit, (page-1)*limit)
	} else {
		campaigns, err = appStore.ListActiveCampaigns(category, limit, (page-1)*limit)
	}
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "DB_ERROR", "Could not load campaigns.")
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(apiResponse{
		Success: true,
		Data: map[string]any{
			"campaigns": campaigns,
			"page":      page,
			"limit":     limit,
		},
	})
}

func handleGetCampaign(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	campaign, err := appStore.GetActiveCampaignBySlug(strings.TrimSpace(chi.URLParam(r, "slug")))
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "DB_ERROR", "Could not load campaign.")
		return
	}
	if campaign == nil {
		writeAPIError(w, http.StatusNotFound, "CAMPAIGN_NOT_FOUND", "Campaign was not found.")
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(apiResponse{
		Success: true,
		Data: map[string]any{
			"campaign": campaign,
		},
	})
}

func handleCreateCampaign(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")

	var req campaignPayload
	if !decodeJSONBody(w, r, &req, maxCampaignBody) {
		return
	}

	endDate, err := time.Parse(time.RFC3339, strings.TrimSpace(req.EndDate))
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_FAILED", "end_date must use RFC3339 format.")
		return
	}

	campaign := domain.Campaign{
		ID:              randomID(),
		OrganizationID:  strings.TrimSpace(req.OrganizationID),
		Title:           strings.TrimSpace(req.Title),
		Slug:            strings.TrimSpace(req.Slug),
		Description:     strings.TrimSpace(req.Description),
		Category:        strings.TrimSpace(req.Category),
		Subcategory:     strings.TrimSpace(req.Subcategory),
		CampaignType:    defaultString(strings.TrimSpace(req.CampaignType), "target_deadline"),
		BannerURL:       strings.TrimSpace(req.BannerURL),
		Location:        strings.TrimSpace(req.Location),
		BeneficiaryNote: strings.TrimSpace(req.BeneficiaryNote),
		TargetAmount:    req.TargetAmount,
		EndDate:         endDate,
		Status:          "active",
	}
	if campaign.OrganizationID == "" || campaign.Title == "" || campaign.Slug == "" || campaign.Category == "" || campaign.TargetAmount <= 0 {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_FAILED", "organization_id, title, slug, category, and positive target_amount are required.")
		return
	}
	if !campaignSlugRegex.MatchString(campaign.Slug) {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_FAILED", "slug must only contain lowercase alphanumeric characters and hyphens.")
		return
	}

	if err := appStore.CreateCampaign(&campaign); err != nil {
		if isDuplicateDBError(err) {
			writeAPIError(w, http.StatusConflict, "DUPLICATE_CAMPAIGN", "Campaign slug is already used.")
			return
		}
		writeAPIError(w, http.StatusInternalServerError, "DB_ERROR", "Could not create campaign.")
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(apiResponse{
		Success: true,
		Data: map[string]any{
			"campaign": campaign,
		},
	})
}

func handleUpdateCampaign(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")

	var req campaignPayload
	if !decodeJSONBody(w, r, &req, maxCampaignBody) {
		return
	}

	endDate, err := time.Parse(time.RFC3339, strings.TrimSpace(req.EndDate))
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_FAILED", "end_date must use RFC3339 format.")
		return
	}

	campaign := domain.Campaign{
		ID:              strings.TrimSpace(chi.URLParam(r, "id")),
		Title:           strings.TrimSpace(req.Title),
		Slug:            strings.TrimSpace(req.Slug),
		Description:     strings.TrimSpace(req.Description),
		Category:        strings.TrimSpace(req.Category),
		Subcategory:     strings.TrimSpace(req.Subcategory),
		CampaignType:    defaultString(strings.TrimSpace(req.CampaignType), "target_deadline"),
		BannerURL:       strings.TrimSpace(req.BannerURL),
		Location:        strings.TrimSpace(req.Location),
		BeneficiaryNote: strings.TrimSpace(req.BeneficiaryNote),
		TargetAmount:    req.TargetAmount,
		EndDate:         endDate,
	}
	if campaign.ID == "" || campaign.Title == "" || campaign.Slug == "" || campaign.Category == "" || campaign.TargetAmount <= 0 {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_FAILED", "id, title, slug, category, and positive target_amount are required.")
		return
	}
	if !campaignSlugRegex.MatchString(campaign.Slug) {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_FAILED", "slug must only contain lowercase alphanumeric characters and hyphens.")
		return
	}
	if err := appStore.UpdateCampaign(&campaign); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeAPIError(w, http.StatusNotFound, "CAMPAIGN_NOT_FOUND", "Campaign was not found.")
			return
		}
		if isDuplicateDBError(err) {
			writeAPIError(w, http.StatusConflict, "DUPLICATE_CAMPAIGN", "Campaign slug is already used.")
			return
		}
		writeAPIError(w, http.StatusInternalServerError, "DB_ERROR", "Could not update campaign.")
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(apiResponse{Success: true})
}

func handleUpdateCampaignStatus(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")

	var req campaignStatusPayload
	if !decodeJSONBody(w, r, &req, maxCampaignBody) {
		return
	}
	status := strings.TrimSpace(req.Status)
	if status != "active" && status != "paused" && status != "completed" {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_FAILED", "status must be active, paused, or completed.")
		return
	}
	if err := appStore.UpdateCampaignStatus(strings.TrimSpace(chi.URLParam(r, "id")), status); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeAPIError(w, http.StatusNotFound, "CAMPAIGN_NOT_FOUND", "Campaign was not found.")
			return
		}
		writeAPIError(w, http.StatusInternalServerError, "DB_ERROR", "Could not update campaign status.")
		return
	}
	_ = json.NewEncoder(w).Encode(apiResponse{Success: true, Data: map[string]any{"status": status}})
}

func handleCreateDonation(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")

	var req donationPayload
	if !decodeJSONBody(w, r, &req, maxDonationBody) {
		return
	}
	idempotencyKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if idempotencyKey == "" && isProduction(appConfig) {
		writeAPIError(w, http.StatusBadRequest, "IDEMPOTENCY_KEY_REQUIRED", "Idempotency-Key is required for donations in production.")
		return
	}
	if len(idempotencyKey) > 255 {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_FAILED", "Idempotency-Key must not exceed 255 characters.")
		return
	}
	if idempotencyKey != "" {
		existing, err := appStore.GetDonationByIdempotencyKey(idempotencyKey)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "DB_ERROR", "Could not load prior donation request.")
			return
		}
		if existing != nil {
			writeIdempotentDonation(w, existing)
			return
		}
	}

	campaign, err := appStore.GetActiveCampaignByID(strings.TrimSpace(req.CampaignID))
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "DB_ERROR", "Could not load campaign.")
		return
	}
	if campaign == nil {
		writeAPIError(w, http.StatusNotFound, "CAMPAIGN_NOT_FOUND", "Campaign was not found.")
		return
	}

	donor := domain.Donor{
		ID:          randomID(),
		Name:        strings.TrimSpace(req.Donor.Name),
		PhoneNumber: strings.TrimSpace(req.Donor.PhoneNumber),
		Email:       strings.TrimSpace(req.Donor.Email),
	}
	paymentPhone := donor.PhoneNumber
	if req.Anonymous && donor.Name == "" {
		donor.Name = "Hamba Allah"
	}
	if donor.PhoneNumber == "" {
		// ponytail: phone is optional for alpha; synthetic value keeps the existing unique donor constraint.
		donor.PhoneNumber = "guest-" + randomID()
	}
	if donor.Name == "" || req.Amount < 2000 || req.Amount > maxDonationAmount || req.PlatformTip < 0 {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_FAILED", "donor name or anonymous, donation amount between 2000 and 100000000, and non-negative platform_tip are required.")
		return
	}
	if donor.Email != "" && !waitlistEmailRegex.MatchString(donor.Email) {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_FAILED", "donor email must be a valid email address.")
		return
	}
	if req.Amount != math.Trunc(req.Amount) || req.PlatformTip != math.Trunc(req.PlatformTip) {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_FAILED", "amount and platform_tip must be whole numbers.")
		return
	}
	grossAmount := int64(req.Amount + req.PlatformTip)
	if grossAmount < 1 {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_FAILED", "amount + platform_tip must be at least 1.")
		return
	}

	donorRecord, err := appStore.FindOrCreateDonor(&donor)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "DB_ERROR", "Could not save donor.")
		return
	}

	donation := domain.Donation{
		ID:             randomID(),
		CampaignID:     campaign.ID,
		DonorID:        donorRecord.ID,
		Amount:         req.Amount,
		PlatformTip:    req.PlatformTip,
		Status:         "pending",
		PaymentMethod:  defaultString(strings.TrimSpace(req.PaymentMethod), "midtrans_snap"),
		Provider:       "midtrans",
		ProviderStatus: "pending",
	}
	if idempotencyKey != "" {
		donation.IdempotencyKey = &idempotencyKey
	}
	donation.ProviderOrderID = donation.ID
	if err := appStore.CreateDonation(&donation); err != nil {
		if idempotencyKey != "" && isDuplicateDBError(err) {
			existing, findErr := appStore.GetDonationByIdempotencyKey(idempotencyKey)
			if findErr == nil && existing != nil {
				writeIdempotentDonation(w, existing)
				return
			}
		}
		writeAPIError(w, http.StatusInternalServerError, "DB_ERROR", "Could not create donation.")
		return
	}

	paymentClient, err := effectivePaymentClient()
	if err != nil {
		writeAPIError(w, http.StatusServiceUnavailable, "PAYMENT_CONFIGURATION_ERROR", "Payment configuration is unavailable.")
		return
	}
	snap, err := paymentClient.CreateSnapTransaction(r.Context(), payment.SnapRequest{
		OrderID:         donation.ProviderOrderID,
		GrossAmount:     grossAmount,
		DonorName:       donorRecord.Name,
		DonorEmail:      donorRecord.Email,
		DonorPhone:      paymentPhone,
		ItemName:        campaign.Title,
		FinishURL:       strings.TrimRight(appConfig.PublicLandingURL, "/") + "/payments/" + url.PathEscape(donation.ID),
		NotificationURL: appConfig.MidtransNotificationURL,
	})
	if err != nil {
		errPayload, _ := json.Marshal(map[string]any{"snap_init_error": err.Error()})
		_, _ = appStore.ApplyPaymentStatus(donation.ProviderOrderID, "snap_init_failed", "", string(errPayload), "failed", nil)
		writeAPIError(w, http.StatusBadGateway, "PAYMENT_PROVIDER_ERROR", "Could not start Midtrans payment.")
		return
	}
	if err := appStore.SaveCheckout(donation.ID, snap.Token, snap.RedirectURL); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "DB_ERROR", "Could not save payment checkout.")
		return
	}
	donation.CheckoutToken = snap.Token
	donation.CheckoutRedirectURL = snap.RedirectURL

	w.WriteHeader(http.StatusCreated)
	writeDonationCheckout(w, &donation)
}

func handleDonationStatus(w http.ResponseWriter, r *http.Request) {
	donation, err := appStore.GetDonation(strings.TrimSpace(chi.URLParam(r, "id")))
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "DB_ERROR", "Could not load donation.")
		return
	}
	if donation == nil {
		writeAPIError(w, http.StatusNotFound, "DONATION_NOT_FOUND", "Donation was not found.")
		return
	}
	client, err := effectivePaymentClient()
	if err != nil {
		writeAPIError(w, http.StatusServiceUnavailable, "PAYMENT_CONFIGURATION_ERROR", "Payment configuration is unavailable.")
		return
	}
	provider, err := client.GetTransactionStatus(r.Context(), donation.ProviderOrderID)
	if err != nil {
		if errors.Is(err, payment.ErrTransactionNotFound) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(apiResponse{Success: true, Data: map[string]any{
				"donation_id": donation.ID, "status": donation.Status, "provider_status": donation.ProviderStatus,
				"paid_at": donation.PaidAt, "counted": false,
			}})
			return
		}
		writeAPIError(w, http.StatusBadGateway, "PAYMENT_PROVIDER_ERROR", "Could not verify payment status with Midtrans.")
		return
	}
	raw, _ := json.Marshal(provider)
	status := mapMidtransStatus(provider.TransactionStatus, provider.FraudStatus)
	counted, err := appStore.ApplyPaymentStatus(donation.ProviderOrderID, provider.TransactionStatus, provider.TransactionID, string(raw), status, parseMidtransTime(provider.TransactionTime))
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "DB_ERROR", "Could not update payment status.")
		return
	}
	updated, err := appStore.GetDonation(donation.ID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "DB_ERROR", "Could not load payment status.")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(apiResponse{Success: true, Data: map[string]any{
		"donation_id": updated.ID, "status": updated.Status, "provider_status": updated.ProviderStatus,
		"paid_at": updated.PaidAt, "counted": counted,
	}})
}

func handleMidtransNotification(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")

	if appConfig.MidtransNotifyKey != "" && subtle.ConstantTimeCompare([]byte(r.Header.Get("X-Notification-Token")), []byte(appConfig.MidtransNotifyKey)) != 1 {
		writeAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Valid notification token is required.")
		return
	}

	var req midtransNotificationPayload
	if !decodeJSONBody(w, r, &req, maxNotificationBody) {
		return
	}

	serverKey, err := effectiveMidtransServerKey()
	if err != nil {
		writeAPIError(w, http.StatusServiceUnavailable, "PAYMENT_CONFIGURATION_ERROR", "Payment configuration is unavailable.")
		return
	}
	if !payment.VerifyMidtransSignature(req.OrderID, req.StatusCode, req.GrossAmount, serverKey, req.SignatureKey) {
		writeAPIError(w, http.StatusUnauthorized, "INVALID_SIGNATURE", "Midtrans signature is invalid.")
		return
	}

	rawPayload, _ := json.Marshal(req)
	status := mapMidtransStatus(req.TransactionStatus, req.FraudStatus)
	paidAt := parseMidtransTime(req.TransactionTime)
	counted, err := appStore.ApplyPaymentStatus(req.OrderID, req.TransactionStatus, req.TransactionID, string(rawPayload), status, paidAt)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			writeAPIError(w, http.StatusNotFound, "DONATION_NOT_FOUND", "Donation was not found.")
			return
		}
		writeAPIError(w, http.StatusInternalServerError, "DB_ERROR", "Could not process payment notification.")
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(apiResponse{
		Success: true,
		Data: map[string]any{
			"message": "notification processed successfully",
			"status":  status,
			"counted": counted,
		},
	})
}

func handleExportDonations(w http.ResponseWriter, r *http.Request) {
	limit := queryInt(r, "limit", 1000)
	if limit < 1 || limit > 5000 {
		limit = 1000
	}
	donations, err := appStore.ListDonations(limit)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "DB_ERROR", "Could not export donations.")
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="donations.csv"`)
	writer := csv.NewWriter(w)
	_ = writer.Write([]string{"id", "provider_order_id", "campaign", "donor_name", "donor_phone", "amount", "platform_tip", "status", "provider_status", "created_at", "paid_at"})
	for _, donation := range donations {
		campaignTitle := ""
		if donation.Campaign != nil {
			campaignTitle = donation.Campaign.Title
		}
		donorName := ""
		donorPhone := ""
		if donation.Donor != nil {
			donorName = donation.Donor.Name
			donorPhone = donation.Donor.PhoneNumber
		}
		paidAt := ""
		if donation.PaidAt != nil {
			paidAt = donation.PaidAt.Format(time.RFC3339)
		}
		_ = writer.Write([]string{
			donation.ID,
			donation.ProviderOrderID,
			campaignTitle,
			donorName,
			donorPhone,
			fmt.Sprintf("%.0f", donation.Amount),
			fmt.Sprintf("%.0f", donation.PlatformTip),
			donation.Status,
			donation.ProviderStatus,
			donation.CreatedAt.Format(time.RFC3339),
			paidAt,
		})
	}
	writer.Flush()
}

func handleAdminDonations(w http.ResponseWriter, r *http.Request) {
	page, limit := queryInt(r, "page", 1), queryInt(r, "limit", 25)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 25
	}
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	if status != "" && status != "pending" && status != "success" && status != "failed" {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_FAILED", "status must be pending, success, or failed.")
		return
	}
	campaignID := strings.TrimSpace(r.URL.Query().Get("campaign_id"))
	sort := strings.TrimSpace(r.URL.Query().Get("sort"))
	if sort == "" {
		sort = "latest"
	}
	if sort != "latest" && sort != "oldest" && sort != "amount_desc" && sort != "amount_asc" {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_FAILED", "sort must be latest, oldest, amount_desc, or amount_asc.")
		return
	}
	donations, total, err := appStore.ListDonationsPage(status, campaignID, sort, limit, (page-1)*limit)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "DB_ERROR", "Could not load donations.")
		return
	}
	pages := int((total + int64(limit) - 1) / int64(limit))
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(apiResponse{Success: true, Data: map[string]any{"donations": donations, "pagination": map[string]any{"page": page, "limit": limit, "total": total, "pages": pages}}})
}

func handleAdminUpload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBody)
	file, _, err := r.FormFile("file")
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_UPLOAD", "A JPEG, PNG, or WebP file is required.")
		return
	}
	defer file.Close()
	head := make([]byte, 512)
	n, err := io.ReadFull(file, head)
	if err != nil && err != io.ErrUnexpectedEOF {
		writeAPIError(w, http.StatusBadRequest, "INVALID_UPLOAD", "Could not read uploaded file.")
		return
	}
	mime := http.DetectContentType(head[:n])
	extensions := map[string]string{"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp"}
	ext, ok := extensions[mime]
	if !ok {
		writeAPIError(w, http.StatusUnsupportedMediaType, "INVALID_UPLOAD_TYPE", "Only JPEG, PNG, and WebP images are allowed.")
		return
	}
	name := randomID() + ext
	if err := os.MkdirAll(appConfig.UploadDir, 0o750); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "UPLOAD_ERROR", "Could not prepare upload storage.")
		return
	}
	destination, err := os.OpenFile(filepath.Join(appConfig.UploadDir, name), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "UPLOAD_ERROR", "Could not save uploaded file.")
		return
	}
	defer destination.Close()
	if _, err := destination.Write(head[:n]); err == nil {
		_, err = io.Copy(destination, file)
	}
	if err != nil {
		_ = os.Remove(destination.Name())
		writeAPIError(w, http.StatusInternalServerError, "UPLOAD_ERROR", "Could not save uploaded file.")
		return
	}
	base := strings.TrimRight(appConfig.PublicUploadBaseURL, "/")
	if base == "" {
		base = "/uploads"
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(apiResponse{Success: true, Data: map[string]any{"url": base + "/" + name}})
}

func handlePublicUpload(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" || filepath.Base(name) != name {
		http.NotFound(w, r)
		return
	}
	ext := strings.ToLower(filepath.Ext(name))
	if ext != ".jpg" && ext != ".png" && ext != ".webp" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeFile(w, r, filepath.Join(appConfig.UploadDir, name))
}

func handleGetPaymentSettings(w http.ResponseWriter, r *http.Request) {
	view, err := paymentSettingsView()
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "SETTINGS_ERROR", "Could not load payment settings.")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(apiResponse{Success: true, Data: map[string]any{"payment": view}})
}

func handlePutPaymentSettings(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req paymentSettingsPayload
	if !decodeJSONBody(w, r, &req, maxCampaignBody) {
		return
	}
	req.Mode = strings.ToLower(strings.TrimSpace(req.Mode))
	req.ServerKey = strings.TrimSpace(req.ServerKey)
	req.ClientKey = strings.TrimSpace(req.ClientKey)
	if req.Mode != "sandbox" && req.Mode != "production" {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_FAILED", "mode must be sandbox or production.")
		return
	}
	if (req.ServerKey == "") != (req.ClientKey == "") {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_FAILED", "server_key and client_key must be provided together.")
		return
	}
	if req.ServerKey == "" && (appConfig.MidtransServerKey != "" || appConfig.MidtransClientKey != "") && !validMidtransKeyPair(req.Mode, appConfig.MidtransServerKey, appConfig.MidtransClientKey) {
		writeAPIError(w, http.StatusBadRequest, "INVALID_MIDTRANS_KEYS", "Environment Midtrans keys do not match the selected mode.")
		return
	}
	if req.ServerKey == "" {
		if err := appStore.DeletePaymentSetting(); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "DB_ERROR", "Could not reset payment settings.")
			return
		}
		if err := refreshPaymentClient(); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "SETTINGS_ERROR", "Could not activate payment settings.")
			return
		}
		view, _ := paymentSettingsView()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(apiResponse{Success: true, Data: map[string]any{"payment": view}})
		return
	}
	setting := domain.PaymentSetting{ID: "midtrans", Mode: req.Mode}
	if req.ServerKey != "" {
		if !validMidtransKeyPair(req.Mode, req.ServerKey, req.ClientKey) {
			writeAPIError(w, http.StatusBadRequest, "INVALID_MIDTRANS_KEYS", "Midtrans keys do not match the selected mode.")
			return
		}
		key, err := payment.DecodeSettingsKey(appConfig.AdminSettingsEncryptionKey)
		if err != nil {
			writeAPIError(w, http.StatusServiceUnavailable, "ENCRYPTION_UNAVAILABLE", "Settings encryption is not configured.")
			return
		}
		setting.ServerKeyCipher, err = payment.EncryptSetting(key, req.ServerKey)
		if err == nil {
			setting.ClientKeyCipher, err = payment.EncryptSetting(key, req.ClientKey)
		}
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "ENCRYPTION_ERROR", "Could not encrypt payment settings.")
			return
		}
	}
	if err := appStore.SavePaymentSetting(&setting); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "DB_ERROR", "Could not save payment settings.")
		return
	}
	if err := refreshPaymentClient(); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "SETTINGS_ERROR", "Could not activate payment settings.")
		return
	}
	view, _ := paymentSettingsView()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(apiResponse{Success: true, Data: map[string]any{"payment": view}})
}

func validMidtransKeyPair(mode, serverKey, clientKey string) bool {
	if mode == "sandbox" {
		return strings.HasPrefix(serverKey, "SB-Mid-server-") && strings.HasPrefix(clientKey, "SB-Mid-client-")
	}
	return strings.HasPrefix(serverKey, "Mid-server-") && strings.HasPrefix(clientKey, "Mid-client-")
}

func paymentSettingsView() (map[string]any, error) {
	setting, err := appStore.GetPaymentSetting()
	if err != nil {
		return nil, err
	}
	mode := appConfig.MidtransEnv
	serverKey, clientKey := appConfig.MidtransServerKey, appConfig.MidtransClientKey
	if setting != nil {
		mode = setting.Mode
		if setting.ServerKeyCipher != "" {
			key, decodeErr := payment.DecodeSettingsKey(appConfig.AdminSettingsEncryptionKey)
			if decodeErr != nil {
				return nil, decodeErr
			}
			serverKey, decodeErr = payment.DecryptSetting(key, setting.ServerKeyCipher)
			if decodeErr == nil {
				clientKey, decodeErr = payment.DecryptSetting(key, setting.ClientKeyCipher)
			}
			if decodeErr != nil {
				return nil, decodeErr
			}
		}
	}
	return map[string]any{"mode": mode, "source": map[bool]string{true: "override", false: "environment"}[setting != nil], "server_key_configured": serverKey != "", "client_key_configured": clientKey != "", "server_key_masked": maskKey(serverKey), "client_key_masked": maskKey(clientKey)}, nil
}

func maskKey(value string) string {
	if len(value) <= 8 {
		return strings.Repeat("•", len(value))
	}
	return value[:min(14, len(value)-4)] + "••••••••" + value[len(value)-4:]
}

func effectivePaymentConfig() (string, string, error) {
	setting, err := appStore.GetPaymentSetting()
	if err != nil {
		return "", "", err
	}
	if setting == nil {
		return appConfig.MidtransEnv, appConfig.MidtransServerKey, nil
	}
	if setting.ServerKeyCipher == "" {
		return setting.Mode, appConfig.MidtransServerKey, nil
	}
	key, err := payment.DecodeSettingsKey(appConfig.AdminSettingsEncryptionKey)
	if err != nil {
		return "", "", err
	}
	serverKey, err := payment.DecryptSetting(key, setting.ServerKeyCipher)
	if err == nil {
		_, err = payment.DecryptSetting(key, setting.ClientKeyCipher)
	}
	return setting.Mode, serverKey, err
}

func effectivePaymentClient() (*payment.MidtransClient, error) {
	appPaymentGuard.RLock()
	defer appPaymentGuard.RUnlock()
	return appPayment, nil
}

func effectiveMidtransServerKey() (string, error) {
	appPaymentGuard.RLock()
	defer appPaymentGuard.RUnlock()
	return appPaymentServerKey, nil
}

func refreshPaymentClient() error {
	mode, serverKey, err := effectivePaymentConfig()
	if err != nil {
		return err
	}
	appPaymentGuard.Lock()
	appPayment = payment.NewMidtransClient(mode, serverKey)
	appPaymentServerKey = serverKey
	appPaymentGuard.Unlock()
	return nil
}

func queryInt(r *http.Request, key string, fallback int) int {
	value := strings.TrimSpace(r.URL.Query().Get(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func writeDonationCheckout(w http.ResponseWriter, donation *domain.Donation) {
	_ = json.NewEncoder(w).Encode(apiResponse{
		Success: true,
		Data: map[string]any{
			"donation_id": donation.ID,
			"status":      donation.Status,
			"payment": map[string]any{
				"provider":     donation.Provider,
				"snap_token":   donation.CheckoutToken,
				"redirect_url": donation.CheckoutRedirectURL,
			},
		},
	})
}

func writeIdempotentDonation(w http.ResponseWriter, donation *domain.Donation) {
	if donation.Status == "failed" {
		writeAPIError(w, http.StatusConflict, "IDEMPOTENCY_PREVIOUSLY_FAILED", "The original payment checkout failed; use a new Idempotency-Key to retry.")
		return
	}
	if donation.CheckoutToken == "" || donation.CheckoutRedirectURL == "" {
		writeAPIError(w, http.StatusConflict, "IDEMPOTENCY_IN_PROGRESS", "The original payment checkout is not ready yet.")
		return
	}
	w.WriteHeader(http.StatusOK)
	writeDonationCheckout(w, donation)
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst any, limit int64) bool {
	if r.ContentLength > limit {
		writeAPIError(w, http.StatusRequestEntityTooLarge, "REQUEST_TOO_LARGE", "Payload exceeds the allowed size.")
		return false
	}
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(dst); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			writeAPIError(w, http.StatusRequestEntityTooLarge, "REQUEST_TOO_LARGE", "Payload exceeds the allowed size.")
			return false
		}
		writeAPIError(w, http.StatusBadRequest, "INVALID_JSON", "Payload must be valid JSON.")
		return false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeAPIError(w, http.StatusBadRequest, "INVALID_JSON", "Payload must contain one JSON value.")
		return false
	}
	return true
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func mapMidtransStatus(transactionStatus, fraudStatus string) string {
	switch strings.ToLower(transactionStatus) {
	case "capture":
		if strings.EqualFold(fraudStatus, "accept") {
			return "success"
		}
		return "pending"
	case "settlement":
		return "success"
	case "deny", "cancel", "expire", "failure":
		return "failed"
	default:
		return "pending"
	}
}

func parseMidtransTime(value string) *time.Time {
	loc := time.FixedZone("WIB", 7*3600)
	parsed, err := time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(value), loc)
	if err != nil {
		return nil
	}
	return &parsed
}

func handleAdminLogin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	var req adminLoginPayload
	if !decodeJSONBody(w, r, &req, maxWaitlistBody) {
		return
	}
	if appConfig == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "ADMIN_AUTH_UNAVAILABLE", "Admin authentication is not configured.")
		return
	}
	valid := appConfig.AdminPassword != "" && subtle.ConstantTimeCompare([]byte(req.Password), []byte(appConfig.AdminPassword)) == 1
	if !valid && appConfig.AdminPassword == "" {
		hash, err := appStore.GetDefaultAdminPasswordHash()
		valid = err == nil && bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) == nil
	}
	if !valid {
		writeAPIError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Password is invalid.")
		return
	}
	cookie, err := newAdminSession(appConfig, time.Now().Add(adminSessionTTL))
	if err != nil {
		writeAPIError(w, http.StatusServiceUnavailable, "ADMIN_AUTH_UNAVAILABLE", "Admin authentication is not configured.")
		return
	}
	http.SetCookie(w, cookie)
	_ = json.NewEncoder(w).Encode(apiResponse{Success: true, Data: map[string]any{"authenticated": true, "must_change_password": appConfig.AdminPassword == ""}})
}

func handleAdminLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, expiredAdminSession(appConfig))
	w.WriteHeader(http.StatusNoContent)
}

func handleAdminSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	authenticated := validAdminSession(r, appConfig)
	_ = json.NewEncoder(w).Encode(apiResponse{Success: true, Data: map[string]any{"authenticated": authenticated, "must_change_password": authenticated && appConfig != nil && appConfig.AdminPassword == ""}})
}

func requireAdminSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !validAdminSession(r, appConfig) {
			writeAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Admin session is required.")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func validateAdminConfig(cfg *config.Config) error {
	if !isProduction(cfg) {
		return nil
	}
	if cfg.AdminPassword != "" && len(strings.TrimSpace(cfg.AdminPassword)) < 12 {
		return errors.New("ADMIN_PASSWORD must be at least 12 characters in production")
	}
	if len(cfg.AdminSessionSecret) < 32 {
		return errors.New("ADMIN_SESSION_SECRET must be at least 32 characters in production")
	}
	return nil
}

func newAdminSession(cfg *config.Config, expiresAt time.Time) (*http.Cookie, error) {
	secret := adminSessionSecret(cfg)
	if secret == "" {
		return nil, errors.New("admin session secret is not configured")
	}
	payload := strconv.FormatInt(expiresAt.Unix(), 10)
	return &http.Cookie{
		Name:     adminSessionCookie,
		Value:    payload + "." + signAdminSession(secret, payload),
		Path:     "/",
		Expires:  expiresAt,
		MaxAge:   int(time.Until(expiresAt).Seconds()),
		HttpOnly: true,
		Secure:   isProduction(cfg),
		SameSite: http.SameSiteLaxMode,
	}, nil
}

func expiredAdminSession(cfg *config.Config) *http.Cookie {
	return &http.Cookie{Name: adminSessionCookie, Value: "", Path: "/", MaxAge: -1, Expires: time.Unix(1, 0), HttpOnly: true, Secure: isProduction(cfg), SameSite: http.SameSiteLaxMode}
}

func validAdminSession(r *http.Request, cfg *config.Config) bool {
	secret := adminSessionSecret(cfg)
	if secret == "" {
		return false
	}
	cookie, err := r.Cookie(adminSessionCookie)
	if err != nil {
		return false
	}
	payload, signature, ok := strings.Cut(cookie.Value, ".")
	if !ok || subtle.ConstantTimeCompare([]byte(signature), []byte(signAdminSession(secret, payload))) != 1 {
		return false
	}
	expiresAt, err := strconv.ParseInt(payload, 10, 64)
	return err == nil && time.Now().Unix() < expiresAt
}

func adminSessionSecret(cfg *config.Config) string {
	if cfg == nil {
		return ""
	}
	if cfg.AdminSessionSecret != "" {
		return cfg.AdminSessionSecret
	}
	if cfg.AdminPassword != "" {
		// ponytail: development may reuse ADMIN_PASSWORD; production requires a separate session secret.
		return cfg.AdminPassword
	}
	if !isProduction(cfg) && appStore != nil {
		hash, _ := appStore.GetDefaultAdminPasswordHash()
		return hash
	}
	return ""
}

func signAdminSession(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func isProduction(cfg *config.Config) bool {
	return cfg != nil && (strings.EqualFold(cfg.Env, "production") || strings.EqualFold(cfg.Env, "prod"))
}

type waitlistPayload struct {
	Email   string `json:"email"`
	Website string `json:"website"`
	Source  string `json:"source"`
}

type campaignPayload struct {
	OrganizationID  string  `json:"organization_id"`
	Title           string  `json:"title"`
	Slug            string  `json:"slug"`
	Description     string  `json:"description"`
	Category        string  `json:"category"`
	Subcategory     string  `json:"subcategory"`
	CampaignType    string  `json:"campaign_type"`
	BannerURL       string  `json:"banner_url"`
	Location        string  `json:"location"`
	BeneficiaryNote string  `json:"beneficiary_note"`
	TargetAmount    float64 `json:"target_amount"`
	EndDate         string  `json:"end_date"`
}

type campaignStatusPayload struct {
	Status string `json:"status"`
}

type adminLoginPayload struct {
	Password string `json:"password"`
}

type paymentSettingsPayload struct {
	Mode      string `json:"mode"`
	ServerKey string `json:"server_key"`
	ClientKey string `json:"client_key"`
}

type donorPayload struct {
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email"`
}

type donationPayload struct {
	CampaignID    string       `json:"campaign_id"`
	Donor         donorPayload `json:"donor"`
	Anonymous     bool         `json:"anonymous"`
	Amount        float64      `json:"amount"`
	PlatformTip   float64      `json:"platform_tip"`
	PaymentMethod string       `json:"payment_method"`
}

type midtransNotificationPayload struct {
	TransactionTime   string `json:"transaction_time"`
	TransactionStatus string `json:"transaction_status"`
	TransactionID     string `json:"transaction_id"`
	FraudStatus       string `json:"fraud_status"`
	StatusCode        string `json:"status_code"`
	SignatureKey      string `json:"signature_key"`
	PaymentType       string `json:"payment_type"`
	OrderID           string `json:"order_id"`
	GrossAmount       string `json:"gross_amount"`
	Currency          string `json:"currency"`
}

type apiErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type apiResponse struct {
	Success bool              `json:"success"`
	Data    map[string]any    `json:"data,omitempty"`
	Error   *apiErrorResponse `json:"error,omitempty"`
}

func handleWaitlistSignup(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")

	var req waitlistPayload
	if !decodeJSONBody(w, r, &req, maxWaitlistBody) {
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	if !waitlistEmailRegex.MatchString(email) {
		writeWaitlistError(w, http.StatusBadRequest, "INVALID_EMAIL", "Email is required and must be valid.")
		return
	}

	if strings.TrimSpace(req.Website) != "" {
		writeWaitlistError(w, http.StatusBadRequest, "BOT_DETECTED", "Request rejected as suspicious.")
		return
	}

	clientIP := getClientIP(r)

	var exists domain.Waitlist
	err := database.DB.Where("email = ?", email).First(&exists).Error
	if err == nil {
		writeWaitlistError(w, http.StatusConflict, "DUPLICATE_EMAIL", "Email is already registered in the waitlist.")
		return
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		writeWaitlistError(w, http.StatusInternalServerError, "DB_ERROR", "Could not process waitlist registration.")
		return
	}

	entry := domain.Waitlist{
		ID:        randomID(),
		Email:     email,
		Source:    strings.TrimSpace(req.Source),
		IPAddress: clientIP,
		UserAgent: r.Header.Get("User-Agent"),
	}
	if err := database.DB.Create(&entry).Error; err != nil {
		if isDuplicateDBError(err) {
			writeWaitlistError(w, http.StatusConflict, "DUPLICATE_EMAIL", "Email is already registered in the waitlist.")
			return
		}
		writeWaitlistError(w, http.StatusInternalServerError, "DB_ERROR", "Could not process waitlist registration.")
		return
	}

	emailQueued := isSMTPConfigured(appConfig)
	adminEmailQueued := isAdminEmailConfigured(appConfig)
	if emailQueued || adminEmailQueued {
		go deliverWaitlistEmails(appConfig, entry)
	}

	response := apiResponse{
		Success: true,
		Data: map[string]any{
			"id":                 entry.ID,
			"email_queued":       emailQueued,
			"admin_email_queued": adminEmailQueued,
		},
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(response)
}

func writeWaitlistError(w http.ResponseWriter, status int, code, message string) {
	writeAPIError(w, status, code, message)
}

func writeAPIError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	response := apiResponse{
		Success: false,
		Error: &apiErrorResponse{
			Code:    code,
			Message: message,
		},
	}
	_ = json.NewEncoder(w).Encode(response)
}

func publicRateLimit(next http.Handler) http.Handler {
	return rateLimit(func() int {
		if appConfig != nil && appConfig.PublicRateLimit > 0 {
			return appConfig.PublicRateLimit
		}
		return 120
	})(next)
}

func adminLoginRateLimit(next http.Handler) http.Handler {
	return rateLimit(func() int { return 10 })(next)
}

func rateLimit(limit func() int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			now := time.Now().UTC()
			key := r.URL.Path + "|" + getClientIP(r)
			publicRequestGuard.Lock()
			if publicRequestCleaned.IsZero() || now.Sub(publicRequestCleaned) >= publicRateLimitWindow {
				for oldKey, entry := range publicRequestWindow {
					if now.Sub(entry.started) >= publicRateLimitWindow {
						delete(publicRequestWindow, oldKey)
					}
				}
				publicRequestCleaned = now
			}
			entry := publicRequestWindow[key]
			if entry.started.IsZero() || now.Sub(entry.started) >= publicRateLimitWindow {
				entry = rateLimitEntry{started: now}
			}
			entry.count++
			publicRequestWindow[key] = entry
			publicRequestGuard.Unlock()
			if entry.count > limit() {
				w.Header().Set("Retry-After", strconv.Itoa(int(publicRateLimitWindow.Seconds())))
				writeAPIError(w, http.StatusTooManyRequests, "RATE_LIMITED", "Too many requests. Please try again later.")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func configuredCORS(cfg *config.Config) (cors.Options, error) {
	production := isProduction(cfg)
	origins := []string{}
	if cfg != nil {
		for _, rawOrigin := range strings.Split(cfg.CORSAllowedOrigins, ",") {
			origin := strings.TrimSpace(rawOrigin)
			if origin == "" {
				continue
			}
			if strings.Contains(origin, "*") {
				return cors.Options{}, fmt.Errorf("CORS_ALLOWED_ORIGINS must not contain wildcard origins")
			}
			parsed, err := url.ParseRequestURI(origin)
			if err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
				return cors.Options{}, fmt.Errorf("invalid CORS origin %q", origin)
			}
			if production && (parsed.Scheme != "https" || parsed.Hostname() == "localhost" || net.ParseIP(parsed.Hostname()) != nil) {
				return cors.Options{}, fmt.Errorf("production CORS origin %q must be a public HTTPS origin", origin)
			}
			origins = append(origins, origin)
		}
	}
	if production && len(origins) == 0 {
		return cors.Options{}, errors.New("CORS_ALLOWED_ORIGINS is required in production")
	}
	options := cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "Idempotency-Key", "X-CSRID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}
	if len(origins) == 0 {
		options.AllowOriginFunc = func(_ *http.Request, origin string) bool {
			parsed, err := url.ParseRequestURI(origin)
			return err == nil && parsed.Scheme == "http" && (parsed.Hostname() == "localhost" || net.ParseIP(parsed.Hostname()).IsLoopback())
		}
	}
	return options, nil
}

func getClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return strings.TrimSpace(r.RemoteAddr)
	}
	return host
}

func isDuplicateDBError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "unique constraint") ||
		strings.Contains(message, "duplicate key") ||
		strings.Contains(message, "duplicated key") ||
		strings.Contains(message, "1062")
}

func deliverWaitlistEmails(cfg *config.Config, entry domain.Waitlist) {
	if err := sendWaitlistConfirmation(cfg, entry.Email); err != nil {
		fmt.Printf("Failed to send waitlist confirmation: %v\n", err)
	}
	if err := sendWaitlistAdminNotification(cfg, entry); err != nil {
		fmt.Printf("Failed to send waitlist admin notification: %v\n", err)
	}
}

func sendWaitlistConfirmation(cfg *config.Config, recipient string) error {
	if !isSMTPConfigured(cfg) {
		return nil
	}

	from := mail.Address{Name: cfg.SMTPFromName, Address: cfg.SMTPFrom}
	to := mail.Address{Address: recipient}
	subject := "Anda sudah masuk waitlist kebaikanku.id"
	body := fmt.Sprintf(`Halo,

Terima kasih sudah bergabung ke waitlist kebaikanku.id.

Kami akan mengirimkan update saat dashboard pengelola kampanye dan integrasi payment gateway sudah siap dirilis.

Pantau halaman ini untuk update berikutnya:
%s

Salam,
Tim kebaikanku.id
`, cfg.WaitlistEmailURL)

	headers := map[string]string{
		"From":         from.String(),
		"To":           to.String(),
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/plain; charset=UTF-8",
	}

	var message strings.Builder
	for key, value := range headers {
		message.WriteString(key)
		message.WriteString(": ")
		message.WriteString(value)
		message.WriteString("\r\n")
	}
	message.WriteString("\r\n")
	message.WriteString(body)

	return sendSMTPMessage(cfg, recipient, message.String())
}

func sendWaitlistAdminNotification(cfg *config.Config, entry domain.Waitlist) error {
	if !isAdminEmailConfigured(cfg) {
		return nil
	}

	from := mail.Address{Name: cfg.SMTPFromName, Address: cfg.SMTPFrom}
	to := mail.Address{Address: cfg.WaitlistAdminEmail}
	subject := "Waitlist baru kebaikanku.id"
	body := fmt.Sprintf(`Ada pendaftar baru di waitlist kebaikanku.id.

Email: %s
Source: %s
IP Address: %s
User Agent: %s
Created At: %s
ID: %s
`, entry.Email, entry.Source, entry.IPAddress, entry.UserAgent, entry.CreatedAt.Format(time.RFC3339), entry.ID)

	headers := map[string]string{
		"From":         from.String(),
		"To":           to.String(),
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/plain; charset=UTF-8",
	}

	var message strings.Builder
	for key, value := range headers {
		message.WriteString(key)
		message.WriteString(": ")
		message.WriteString(value)
		message.WriteString("\r\n")
	}
	message.WriteString("\r\n")
	message.WriteString(body)

	return sendSMTPMessage(cfg, cfg.WaitlistAdminEmail, message.String())
}

func isSMTPConfigured(cfg *config.Config) bool {
	return cfg != nil && cfg.SMTPHost != "" && cfg.SMTPPort != "" && cfg.SMTPFrom != ""
}

func isAdminEmailConfigured(cfg *config.Config) bool {
	return isSMTPConfigured(cfg) && cfg.WaitlistAdminEmail != ""
}

func sendSMTPMessage(cfg *config.Config, recipient string, message string) error {
	addr := net.JoinHostPort(cfg.SMTPHost, cfg.SMTPPort)
	var auth smtp.Auth
	if cfg.SMTPUser != "" && cfg.SMTPPass != "" && strings.ToLower(cfg.SMTPEncryption) != "null" {
		auth = smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPHost)
	}

	return smtp.SendMail(addr, auth, cfg.SMTPFrom, []string{recipient}, []byte(message))
}

func randomID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%s-%s-%s-%s-%s", hex.EncodeToString(b[0:4]), hex.EncodeToString(b[4:6]), hex.EncodeToString(b[6:8]), hex.EncodeToString(b[8:10]), hex.EncodeToString(b[10:16]))
}
