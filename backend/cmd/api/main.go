package main

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
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
	"gorm.io/gorm"
)

const (
	waitlistCooldown = 20 * time.Second
	maxWaitlistBody  = 1 << 20
)

var (
	waitlistRequestWindow = map[string]time.Time{}
	waitlistRequestGuard  sync.Mutex
	waitlistEmailRegex    = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	campaignSlugRegex     = regexp.MustCompile(`^[a-z0-9-]+$`)
	appConfig             *config.Config
	appStore              *repository.Store
	appPayment            *payment.MidtransClient
)

func main() {
	// 1. Load Configurations
	cfg := config.Load()
	appConfig = cfg

	// 2. Initialize Database and Auto-migrate
	database.Init(cfg)
	appStore = repository.NewStore(database.DB)
	appPayment = payment.NewMidtransClient(cfg.MidtransEnv, cfg.MidtransServerKey)

	// 3. Setup Router
	r := chi.NewRouter()

	// Standard middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS config
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:*", "http://127.0.0.1:*", "https://*.pages.dev", "https://kebaikanku.id", "https://*.kebaikanku.id"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by browsers
	}))

	// 4. Routes
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy","time":"` + time.Now().Format(time.RFC3339) + `"}`))
	})
	r.Post("/api/v1/waitlist", handleWaitlistSignup)
	r.Get("/api/v1/campaigns", handleListCampaigns)
	r.Get("/api/v1/campaigns/{slug}", handleGetCampaign)
	r.Post("/api/v1/campaigns", handleCreateCampaign)
	r.Put("/api/v1/campaigns/{id}", handleUpdateCampaign)
	r.Post("/api/v1/donations", handleCreateDonation)
	r.Get("/api/v1/donations/export", handleExportDonations)
	r.Post("/api/v1/payments/midtrans/notification", handleMidtransNotification)

	// 5. Start Server
	serverAddr := fmt.Sprintf(":%s", cfg.Port)
	fmt.Printf("Backend API server is running on http://localhost%s in %s mode\n", serverAddr, cfg.Env)

	err := http.ListenAndServe(serverAddr, r)
	if err != nil {
		panic(fmt.Sprintf("Failed to start server: %v", err))
	}
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

	campaigns, err := appStore.ListActiveCampaigns(strings.TrimSpace(r.URL.Query().Get("category")), limit, (page-1)*limit)
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

	if !hasAdminToken(r) {
		writeAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Valid campaign admin token is required.")
		return
	}

	var req campaignPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_JSON", "Payload must be valid JSON.")
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

	if !hasAdminToken(r) {
		writeAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Valid campaign admin token is required.")
		return
	}

	var req campaignPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_JSON", "Payload must be valid JSON.")
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
		writeAPIError(w, http.StatusInternalServerError, "DB_ERROR", "Could not update campaign.")
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(apiResponse{Success: true})
}

func handleCreateDonation(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")

	var req donationPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_JSON", "Payload must be valid JSON.")
		return
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
	if req.Anonymous && donor.Name == "" {
		donor.Name = "Hamba Allah"
	}
	if donor.PhoneNumber == "" {
		// ponytail: phone is optional for alpha; synthetic value keeps the existing unique donor constraint.
		donor.PhoneNumber = "guest-" + randomID()
	}
	if donor.Name == "" || req.Amount <= 0 || req.PlatformTip < 0 {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_FAILED", "donor name or anonymous, positive amount, and non-negative platform_tip are required.")
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
	donation.ProviderOrderID = donation.ID
	if err := appStore.CreateDonation(&donation); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "DB_ERROR", "Could not create donation.")
		return
	}

	snap, err := appPayment.CreateSnapTransaction(r.Context(), payment.SnapRequest{
		OrderID:     donation.ProviderOrderID,
		GrossAmount: grossAmount,
		DonorName:   donorRecord.Name,
		DonorEmail:  donorRecord.Email,
		DonorPhone:  donorRecord.PhoneNumber,
		ItemName:    campaign.Title,
	})
	if err != nil {
		errPayload, _ := json.Marshal(map[string]any{"snap_init_error": err.Error()})
		_, _ = appStore.ApplyPaymentStatus(donation.ProviderOrderID, "snap_init_failed", "", string(errPayload), "failed", nil)
		writeAPIError(w, http.StatusBadGateway, "PAYMENT_PROVIDER_ERROR", "Could not start Midtrans payment.")
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(apiResponse{
		Success: true,
		Data: map[string]any{
			"donation_id": donation.ID,
			"status":      donation.Status,
			"payment": map[string]any{
				"provider":     "midtrans",
				"snap_token":   snap.Token,
				"redirect_url": snap.RedirectURL,
			},
		},
	})
}

func handleMidtransNotification(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")

	if appConfig.MidtransNotifyKey != "" && subtle.ConstantTimeCompare([]byte(r.Header.Get("X-Notification-Token")), []byte(appConfig.MidtransNotifyKey)) != 1 {
		writeAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Valid notification token is required.")
		return
	}

	var req midtransNotificationPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_JSON", "Payload must be valid JSON.")
		return
	}

	if !payment.VerifyMidtransSignature(req.OrderID, req.StatusCode, req.GrossAmount, appConfig.MidtransServerKey, req.SignatureKey) {
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
	if !hasAdminToken(r) {
		writeAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Valid campaign admin token is required.")
		return
	}

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
	_ = writer.Write([]string{"id", "campaign", "donor_name", "donor_phone", "amount", "platform_tip", "status", "provider_status", "created_at", "paid_at"})
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

func hasAdminToken(r *http.Request) bool {
	if appConfig == nil || appConfig.CampaignAdminToken == "" {
		return false
	}
	token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	return subtle.ConstantTimeCompare([]byte(token), []byte(appConfig.CampaignAdminToken)) == 1
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
	r.Body = http.MaxBytesReader(w, r.Body, maxWaitlistBody)

	var req waitlistPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeWaitlistError(w, http.StatusBadRequest, "INVALID_JSON", "Payload must be valid JSON.")
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
	if blockedByRateLimit(clientIP) {
		writeWaitlistError(w, http.StatusTooManyRequests, "RATE_LIMITED", "Too many waitlist requests. Please try again later.")
		return
	}

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

func blockedByRateLimit(ip string) bool {
	now := time.Now().UTC()
	waitlistRequestGuard.Lock()
	defer waitlistRequestGuard.Unlock()

	cutoff := now.Add(-waitlistCooldown)
	for key, timestamp := range waitlistRequestWindow {
		if timestamp.Before(cutoff) {
			delete(waitlistRequestWindow, key)
		}
	}

	lastRequest, ok := waitlistRequestWindow[ip]
	if ok && now.Sub(lastRequest) < waitlistCooldown {
		return true
	}
	waitlistRequestWindow[ip] = now
	return false
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
