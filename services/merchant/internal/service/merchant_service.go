package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"

	"github.com/openlankapay/openlankapay/pkg/auth"
	"github.com/openlankapay/openlankapay/pkg/notification"

	"github.com/openlankapay/openlankapay/services/merchant/internal/domain"
)

// MerchantRepository defines the data access contract for merchants.
type MerchantRepository interface {
	Create(ctx context.Context, merchant *domain.Merchant) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Merchant, error)
	GetByEmail(ctx context.Context, email string) (*domain.Merchant, error)
	Update(ctx context.Context, merchant *domain.Merchant) error
	List(ctx context.Context, params ListParams) ([]*domain.Merchant, int, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
}

// APIKeyRepository defines the data access contract for API keys.
type APIKeyRepository interface {
	Create(ctx context.Context, key *domain.APIKey) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.APIKey, error)
	GetByKeyID(ctx context.Context, keyID string) (*domain.APIKey, error)
	ListByMerchant(ctx context.Context, merchantID uuid.UUID) ([]*domain.APIKey, error)
	Update(ctx context.Context, key *domain.APIKey) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// UserRepository defines the data access contract for users.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, merchantID uuid.UUID, email string) (*domain.User, error)
	GetByEmailGlobal(ctx context.Context, email string) (*domain.User, error)
	ListByMerchant(ctx context.Context, merchantID uuid.UUID) ([]*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
}

// DirectorRepository defines the data access contract for directors.
type DirectorRepository interface {
	Create(ctx context.Context, d *domain.Director) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Director, error)
	GetByToken(ctx context.Context, token string) (*domain.Director, error)
	ListByMerchant(ctx context.Context, merchantID uuid.UUID) ([]*domain.Director, error)
	CountByMerchant(ctx context.Context, merchantID uuid.UUID) (int, error)
	Update(ctx context.Context, d *domain.Director) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// SubmitDirectorInput holds data for director verification submission.
type SubmitDirectorInput struct {
	FullName          string
	DateOfBirth       *time.Time
	NICPassportNumber string
	Phone             string
	Address           string
	DocumentObjectKey string
	DocumentFilename  string
}

// EventPublisher defines the contract for publishing domain events.
type EventPublisher interface {
	Publish(ctx context.Context, subject string, data any) error
}

// ListParams holds pagination and filtering parameters.
type ListParams struct {
	Page      int
	PerPage   int
	KYCStatus *domain.KYCStatus
	Status    *domain.MerchantStatus
}

// RegisterInput holds the data required to register a new merchant.
type RegisterInput struct {
	BusinessName   string
	BusinessType   string
	RegistrationNo string
	Website        string
	ContactEmail   string
	ContactPhone   string
	ContactName    string
}

// RegisterWithUserInput holds data to register a merchant and create the admin user.
type RegisterWithUserInput struct {
	BusinessName string
	ContactEmail string
	ContactPhone string
	ContactName  string
	AdminEmail   string
	AdminPassword string
	AdminName    string
}

// LoginResult holds the result of a successful login.
type LoginResult struct {
	User         *domain.User
	Merchant     *domain.Merchant
	AccessToken  string
	RefreshToken string
}

// UpdateProfileInput holds data for updating merchant profile/KYC.
type UpdateProfileInput struct {
	BusinessName   *string
	BusinessType   *string
	RegistrationNo *string
	Website        *string
	ContactPhone   *string
	ContactName    *string
	AddressLine1   *string
	AddressLine2   *string
	City           *string
	District       *string
	PostalCode     *string
	BankName       *string
	BankBranch     *string
	BankAccountNo  *string
	BankAccountName *string
	SubmitKYC      bool
}

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountInactive    = errors.New("account is inactive")
)

// MerchantService orchestrates merchant domain operations.
type MerchantService struct {
	merchants  MerchantRepository
	apiKeys    APIKeyRepository
	users      UserRepository
	directors  DirectorRepository
	events     EventPublisher
	notifier   notification.Notifier
	jwtSecret  string
	adminEmail string
	tokenTTL   time.Duration
	refreshTTL time.Duration
}

// NewMerchantService creates a new MerchantService.
func NewMerchantService(
	merchants MerchantRepository,
	apiKeys APIKeyRepository,
	users UserRepository,
	events EventPublisher,
	jwtSecret string,
	notifier notification.Notifier,
	adminEmail string,
	directors DirectorRepository,
) *MerchantService {
	return &MerchantService{
		merchants:  merchants,
		apiKeys:    apiKeys,
		users:      users,
		directors:  directors,
		events:     events,
		notifier:   notifier,
		jwtSecret:  jwtSecret,
		adminEmail: adminEmail,
		tokenTTL:   24 * time.Hour,
		refreshTTL: 7 * 24 * time.Hour,
	}
}

// RegisterWithUser creates a new merchant and an admin user, returning JWT tokens.
func (s *MerchantService) RegisterWithUser(ctx context.Context, input RegisterWithUserInput) (*LoginResult, error) {
	// Validate all inputs BEFORE creating anything in the database
	merchant, err := domain.NewMerchant(input.BusinessName, input.ContactEmail)
	if err != nil {
		return nil, err
	}

	merchant.ContactPhone = input.ContactPhone
	merchant.ContactName = input.ContactName

	// Validate user (including password) before persisting merchant
	user, err := domain.NewUser(merchant.ID, input.AdminEmail, input.AdminPassword, input.AdminName, domain.RoleAdmin, nil)
	if err != nil {
		return nil, err
	}

	// Now persist — merchant first, then user
	if err := s.merchants.Create(ctx, merchant); err != nil {
		return nil, err
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}

	accessToken, err := auth.GenerateToken(user.ID, merchant.ID, user.Role, nil, s.jwtSecret, s.tokenTTL)
	if err != nil {
		return nil, fmt.Errorf("generating access token: %w", err)
	}

	refreshToken, err := auth.GenerateRefreshToken(user.ID, s.jwtSecret, s.refreshTTL)
	if err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}

	_ = s.events.Publish(ctx, "merchant.registered", merchant)

	// Send welcome/onboarding email
	if s.notifier != nil {
		s.notifier.SendEmail(ctx, notification.SendEmailInput{
			MerchantID: merchant.ID,
			Recipient:  merchant.ContactEmail,
			Subject:    "Welcome to Open Pay!",
			Body: fmt.Sprintf(
				`<p>Hi <strong>%s</strong>,</p>
				<p>Welcome to Open Pay! Your merchant account for <strong>%s</strong> has been created successfully.</p>
				<p>Here's what to do next:</p>
				<ol>
					<li><strong>Complete KYC verification</strong> — Submit your business documents to unlock full payment processing.</li>
					<li><strong>Set up your payment methods</strong> — Configure how you want to accept crypto payments.</li>
					<li><strong>Start accepting payments</strong> — Create payment links or integrate our API.</li>
				</ol>
				<p>You can get started by logging into your dashboard.</p>
				<p>If you have any questions, our support team is here to help.</p>`,
				input.ContactName, merchant.BusinessName,
			),
			EventType: "merchant.welcome",
		})
	}

	return &LoginResult{
		User:         user,
		Merchant:     merchant,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// Login authenticates a user by email and password.
func (s *MerchantService) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	user, err := s.users.GetByEmailGlobal(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, ErrAccountInactive
	}

	if !user.VerifyPassword(password) {
		return nil, ErrInvalidCredentials
	}

	merchant, err := s.merchants.GetByID(ctx, user.MerchantID)
	if err != nil {
		return nil, fmt.Errorf("fetching merchant: %w", err)
	}

	now := time.Now().UTC()
	user.LastLoginAt = &now
	_ = s.users.Update(ctx, user)

	accessToken, err := auth.GenerateToken(user.ID, merchant.ID, user.Role, user.BranchID, s.jwtSecret, s.tokenTTL)
	if err != nil {
		return nil, fmt.Errorf("generating access token: %w", err)
	}

	refreshToken, err := auth.GenerateRefreshToken(user.ID, s.jwtSecret, s.refreshTTL)
	if err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}

	return &LoginResult{
		User:         user,
		Merchant:     merchant,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshToken generates a new access token from a valid refresh token.
func (s *MerchantService) RefreshToken(ctx context.Context, refreshTokenStr string) (*LoginResult, error) {
	userID, err := auth.ValidateRefreshToken(refreshTokenStr, s.jwtSecret)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, ErrAccountInactive
	}

	merchant, err := s.merchants.GetByID(ctx, user.MerchantID)
	if err != nil {
		return nil, fmt.Errorf("fetching merchant: %w", err)
	}

	accessToken, err := auth.GenerateToken(user.ID, merchant.ID, user.Role, user.BranchID, s.jwtSecret, s.tokenTTL)
	if err != nil {
		return nil, fmt.Errorf("generating access token: %w", err)
	}

	newRefreshToken, err := auth.GenerateRefreshToken(user.ID, s.jwtSecret, s.refreshTTL)
	if err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}

	return &LoginResult{
		User:         user,
		Merchant:     merchant,
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// GetUserByID returns a user by ID.
func (s *MerchantService) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.users.GetByID(ctx, id)
}

// ChangePassword changes a user's password after verifying the current one.
func (s *MerchantService) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if !user.VerifyPassword(currentPassword) {
		return domain.ErrInvalidCredentials
	}
	if err := domain.ValidatePassword(newPassword); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}
	user.PasswordHash = string(hash)
	user.UpdatedAt = time.Now().UTC()
	return s.users.Update(ctx, user)
}

// SetupTOTP generates a new TOTP secret for a user and returns the provisioning URI.
func (s *MerchantService) SetupTOTP(ctx context.Context, userID uuid.UUID) (secret, uri string, err error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return "", "", err
	}
	if user.TOTPEnabled {
		return "", "", fmt.Errorf("2FA is already enabled")
	}
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "OpenLankaPay",
		AccountName: user.Email,
	})
	if err != nil {
		return "", "", fmt.Errorf("generating TOTP: %w", err)
	}
	user.TOTPSecret = key.Secret()
	user.UpdatedAt = time.Now().UTC()
	if err := s.users.Update(ctx, user); err != nil {
		return "", "", err
	}
	return key.Secret(), key.URL(), nil
}

// VerifyAndEnableTOTP verifies a TOTP code and enables 2FA for the user.
func (s *MerchantService) VerifyAndEnableTOTP(ctx context.Context, userID uuid.UUID, code string) ([]string, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.TOTPSecret == "" {
		return nil, fmt.Errorf("2FA not set up — call setup first")
	}
	if user.TOTPEnabled {
		return nil, fmt.Errorf("2FA is already enabled")
	}
	if !totp.Validate(code, user.TOTPSecret) {
		return nil, domain.ErrInvalidTOTP
	}
	codes := make([]string, 8)
	for i := range codes {
		codes[i] = uuid.New().String()[:8]
	}
	user.TOTPEnabled = true
	user.TOTPBackupCodes = codes
	user.UpdatedAt = time.Now().UTC()
	if err := s.users.Update(ctx, user); err != nil {
		return nil, err
	}
	return codes, nil
}

// DisableTOTP disables 2FA after verifying a TOTP code.
func (s *MerchantService) DisableTOTP(ctx context.Context, userID uuid.UUID, code string) error {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if !user.TOTPEnabled {
		return fmt.Errorf("2FA is not enabled")
	}
	if !totp.Validate(code, user.TOTPSecret) {
		return domain.ErrInvalidTOTP
	}
	user.TOTPEnabled = false
	user.TOTPSecret = ""
	user.TOTPBackupCodes = nil
	user.UpdatedAt = time.Now().UTC()
	return s.users.Update(ctx, user)
}

// ValidateTOTP checks if a TOTP code or backup code is valid for login.
func (s *MerchantService) ValidateTOTP(ctx context.Context, userID uuid.UUID, code string) error {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if !user.TOTPEnabled {
		return nil
	}
	if totp.Validate(code, user.TOTPSecret) {
		return nil
	}
	for i, bc := range user.TOTPBackupCodes {
		if bc == code {
			user.TOTPBackupCodes = append(user.TOTPBackupCodes[:i], user.TOTPBackupCodes[i+1:]...)
			user.UpdatedAt = time.Now().UTC()
			_ = s.users.Update(ctx, user)
			return nil
		}
	}
	return domain.ErrInvalidTOTP
}

// UpdateMerchantProfile updates merchant profile and optionally submits KYC.
func (s *MerchantService) UpdateMerchantProfile(ctx context.Context, id uuid.UUID, input UpdateProfileInput) (*domain.Merchant, error) {
	merchant, err := s.merchants.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.BusinessName != nil {
		merchant.BusinessName = *input.BusinessName
	}
	if input.BusinessType != nil {
		merchant.BusinessType = *input.BusinessType
	}
	if input.RegistrationNo != nil {
		merchant.RegistrationNo = *input.RegistrationNo
	}
	if input.Website != nil {
		merchant.Website = *input.Website
	}
	if input.ContactPhone != nil {
		merchant.ContactPhone = *input.ContactPhone
	}
	if input.ContactName != nil {
		merchant.ContactName = *input.ContactName
	}
	if input.AddressLine1 != nil {
		merchant.AddressLine1 = *input.AddressLine1
	}
	if input.AddressLine2 != nil {
		merchant.AddressLine2 = *input.AddressLine2
	}
	if input.City != nil {
		merchant.City = *input.City
	}
	if input.District != nil {
		merchant.District = *input.District
	}
	if input.PostalCode != nil {
		merchant.PostalCode = *input.PostalCode
	}
	if input.BankName != nil {
		merchant.BankName = *input.BankName
	}
	if input.BankBranch != nil {
		merchant.BankBranch = *input.BankBranch
	}
	if input.BankAccountNo != nil {
		merchant.BankAccountNo = *input.BankAccountNo
	}
	if input.BankAccountName != nil {
		merchant.BankAccountName = *input.BankAccountName
	}

	if input.SubmitKYC && merchant.KYCStatus == domain.KYCPending {
		now := time.Now().UTC()
		merchant.KYCSubmittedAt = &now
		if err := merchant.TransitionKYC(domain.KYCInstantAccess); err != nil {
			return nil, err
		}
	}

	if err := s.merchants.Update(ctx, merchant); err != nil {
		return nil, err
	}

	// Send KYC submitted notification
	if input.SubmitKYC && s.notifier != nil {
		s.notifier.SendEmail(ctx, notification.SendEmailInput{
			MerchantID: merchant.ID,
			Recipient:  merchant.ContactEmail,
			Subject:    "KYC Application Submitted — Open Pay",
			Body:       fmt.Sprintf(`<p>Thank you for submitting your KYC application for <strong>%s</strong>.</p><p>Our team will review your application and get back to you within 1-3 business days. You have been granted instant access with a limited transaction volume in the meantime.</p><p>If you have any questions, please contact our support team.</p>`, merchant.BusinessName),
			EventType:  "kyc.submitted",
		})

		// Notify admin about new KYC submission
		if s.adminEmail != "" {
			s.notifier.SendEmail(ctx, notification.SendEmailInput{
				MerchantID: merchant.ID,
				Recipient:  s.adminEmail,
				Subject:    "New KYC Submission — Open Pay",
				Body: fmt.Sprintf(
					`<p>A new KYC application has been submitted and requires review.</p>
					<p><strong>Business:</strong> %s<br/>
					<strong>Contact:</strong> %s<br/>
					<strong>Email:</strong> %s</p>
					<p>Please review the application in the admin dashboard.</p>`,
					merchant.BusinessName, merchant.ContactName, merchant.ContactEmail,
				),
				EventType: "kyc.admin_review_needed",
			})
		}
	}

	return merchant, nil
}

// Register creates a new merchant (legacy, without user).
func (s *MerchantService) Register(ctx context.Context, input RegisterInput) (*domain.Merchant, error) {
	merchant, err := domain.NewMerchant(input.BusinessName, input.ContactEmail)
	if err != nil {
		return nil, err
	}

	merchant.BusinessType = input.BusinessType
	merchant.RegistrationNo = input.RegistrationNo
	merchant.Website = input.Website
	merchant.ContactPhone = input.ContactPhone
	merchant.ContactName = input.ContactName

	if err := s.merchants.Create(ctx, merchant); err != nil {
		return nil, err
	}

	_ = s.events.Publish(ctx, "merchant.registered", merchant)

	return merchant, nil
}

// GetByID returns a merchant by ID.
func (s *MerchantService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Merchant, error) {
	return s.merchants.GetByID(ctx, id)
}

// List returns paginated merchants.
func (s *MerchantService) List(ctx context.Context, params ListParams) ([]*domain.Merchant, int, error) {
	return s.merchants.List(ctx, params)
}

// Approve transitions a merchant's KYC status to APPROVED.
func (s *MerchantService) Approve(ctx context.Context, id uuid.UUID, force bool, forceReason string) error {
	merchant, err := s.merchants.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if !force {
		allVerified, err := s.AllDirectorsVerified(ctx, id)
		if err != nil {
			return err
		}
		if !allVerified {
			return domain.ErrDirectorsNotVerified
		}
	}

	if err := merchant.TransitionKYC(domain.KYCApproved); err != nil {
		return err
	}

	if force && forceReason != "" {
		merchant.KYCReviewNotes = forceReason
	}

	if err := s.merchants.Update(ctx, merchant); err != nil {
		return err
	}

	_ = s.events.Publish(ctx, "merchant.approved", merchant)

	if s.notifier != nil {
		s.notifier.SendEmail(ctx, notification.SendEmailInput{
			MerchantID: merchant.ID,
			Recipient:  merchant.ContactEmail,
			Subject:    "KYC Approved — Open Pay",
			Body:       fmt.Sprintf(`<p>Congratulations! Your KYC application for <strong>%s</strong> has been approved.</p><p>You now have full access to all Open Pay features with no transaction limits. Start accepting payments today!</p>`, merchant.BusinessName),
			EventType:  "kyc.approved",
		})
	}

	return nil
}

// Reject transitions a merchant's KYC status to REJECTED with a reason.
func (s *MerchantService) Reject(ctx context.Context, id uuid.UUID, reason string) error {
	merchant, err := s.merchants.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := merchant.TransitionKYC(domain.KYCRejected); err != nil {
		return err
	}

	merchant.KYCRejectionReason = reason

	if err := s.merchants.Update(ctx, merchant); err != nil {
		return err
	}

	_ = s.events.Publish(ctx, "merchant.rejected", merchant)

	// Send rejection notification email
	if s.notifier != nil {
		s.notifier.SendEmail(ctx, notification.SendEmailInput{
			MerchantID: merchant.ID,
			Recipient:  merchant.ContactEmail,
			Subject:    "KYC Application Update — Open Pay",
			Body:       fmt.Sprintf(`<p>We regret to inform you that your KYC application for <strong>%s</strong> was not approved.</p><p><strong>Reason:</strong> %s</p><p>Please review the feedback and resubmit your application with the required changes. If you have questions, contact our support team.</p>`, merchant.BusinessName, reason),
			EventType:  "kyc.rejected",
		})
	}

	return nil
}

// Deactivate sets a merchant's status to INACTIVE (soft deactivation).
func (s *MerchantService) Deactivate(ctx context.Context, id uuid.UUID) error {
	merchant, err := s.merchants.GetByID(ctx, id)
	if err != nil {
		return err
	}

	merchant.Status = domain.MerchantInactive
	merchant.UpdatedAt = time.Now().UTC()

	if err := s.merchants.Update(ctx, merchant); err != nil {
		return err
	}

	_ = s.events.Publish(ctx, "merchant.deactivated", merchant)

	return nil
}

// CreateAPIKey generates a new API key for a merchant.
func (s *MerchantService) CreateAPIKey(ctx context.Context, merchantID uuid.UUID, env, name string) (*domain.APIKey, string, error) {
	_, err := s.merchants.GetByID(ctx, merchantID)
	if err != nil {
		return nil, "", err
	}

	key, plainSecret, err := domain.NewAPIKey(merchantID, env, name)
	if err != nil {
		return nil, "", fmt.Errorf("generating API key: %w", err)
	}

	if err := s.apiKeys.Create(ctx, key); err != nil {
		return nil, "", fmt.Errorf("storing API key: %w", err)
	}

	_ = s.events.Publish(ctx, "merchant.apikey.created", key)

	return key, plainSecret, nil
}

// ValidateAPIKey checks a key ID and secret, returning the associated merchant.
func (s *MerchantService) ValidateAPIKey(ctx context.Context, keyID, secret string) (*domain.Merchant, error) {
	key, err := s.apiKeys.GetByKeyID(ctx, keyID)
	if err != nil {
		return nil, fmt.Errorf("invalid API key: %w", err)
	}

	if !key.IsActive {
		return nil, errors.New("API key is revoked")
	}

	if !key.VerifySecret(secret) {
		return nil, errors.New("invalid API secret")
	}

	merchant, err := s.merchants.GetByID(ctx, key.MerchantID)
	if err != nil {
		return nil, fmt.Errorf("merchant not found for key: %w", err)
	}

	return merchant, nil
}

// Freeze freezes a merchant's funds due to unauthorized activity.
func (s *MerchantService) Freeze(ctx context.Context, id uuid.UUID, reason string) error {
	merchant, err := s.merchants.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := merchant.TransitionStatus(domain.MerchantFrozen, reason); err != nil {
		return err
	}

	if err := s.merchants.Update(ctx, merchant); err != nil {
		return err
	}

	_ = s.events.Publish(ctx, "merchant.frozen", merchant)

	if s.notifier != nil {
		s.notifier.SendEmail(ctx, notification.SendEmailInput{
			MerchantID: merchant.ID,
			Recipient:  merchant.ContactEmail,
			Subject:    "Account Frozen — Open Pay",
			Body: fmt.Sprintf(
				`<p>Your merchant account for <strong>%s</strong> has been frozen.</p>
				<p><strong>Reason:</strong> %s</p>
				<p>All payment processing and withdrawals have been temporarily suspended. Please contact our support team for further assistance.</p>`,
				merchant.BusinessName, reason,
			),
			EventType: "merchant.frozen",
		})
	}

	return nil
}

// Unfreeze releases a frozen merchant account.
func (s *MerchantService) Unfreeze(ctx context.Context, id uuid.UUID) error {
	merchant, err := s.merchants.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := merchant.TransitionStatus(domain.MerchantActive, ""); err != nil {
		return err
	}

	if err := s.merchants.Update(ctx, merchant); err != nil {
		return err
	}

	_ = s.events.Publish(ctx, "merchant.unfrozen", merchant)

	if s.notifier != nil {
		s.notifier.SendEmail(ctx, notification.SendEmailInput{
			MerchantID: merchant.ID,
			Recipient:  merchant.ContactEmail,
			Subject:    "Account Restored — Open Pay",
			Body: fmt.Sprintf(
				`<p>Your merchant account for <strong>%s</strong> has been restored.</p>
				<p>Payment processing and withdrawals have been re-enabled. Thank you for your cooperation.</p>`,
				merchant.BusinessName,
			),
			EventType: "merchant.unfrozen",
		})
	}

	return nil
}

// Terminate permanently terminates a merchant account for terms violation.
func (s *MerchantService) Terminate(ctx context.Context, id uuid.UUID, reason string) error {
	merchant, err := s.merchants.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := merchant.TransitionStatus(domain.MerchantTerminated, reason); err != nil {
		return err
	}

	if err := s.merchants.Update(ctx, merchant); err != nil {
		return err
	}

	_ = s.events.Publish(ctx, "merchant.terminated", merchant)

	if s.notifier != nil {
		s.notifier.SendEmail(ctx, notification.SendEmailInput{
			MerchantID: merchant.ID,
			Recipient:  merchant.ContactEmail,
			Subject:    "Account Terminated — Open Pay",
			Body: fmt.Sprintf(
				`<p>Your merchant account for <strong>%s</strong> has been permanently terminated.</p>
				<p><strong>Reason:</strong> %s</p>
				<p>If you believe this is an error, please contact our support team.</p>`,
				merchant.BusinessName, reason,
			),
			EventType: "merchant.terminated",
		})
	}

	return nil
}

// RevokeAPIKey deactivates an API key.
func (s *MerchantService) RevokeAPIKey(ctx context.Context, id uuid.UUID, reason string) error {
	key, err := s.apiKeys.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := key.Revoke(reason); err != nil {
		return err
	}

	return s.apiKeys.Update(ctx, key)
}

// CreateDirector creates a new director record and sends a verification email.
func (s *MerchantService) CreateDirector(ctx context.Context, merchantID uuid.UUID, email string) (*domain.Director, error) {
	merchant, err := s.merchants.GetByID(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	count, err := s.directors.CountByMerchant(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	if count >= domain.MaxDirectorsPerMerchant {
		return nil, domain.ErrMaxDirectors
	}
	director, err := domain.NewDirector(merchantID, email)
	if err != nil {
		return nil, err
	}
	if err := s.directors.Create(ctx, director); err != nil {
		return nil, err
	}
	s.sendDirectorVerificationEmail(ctx, merchant.BusinessName, director)
	return director, nil
}

// ListDirectors returns all directors for a merchant.
func (s *MerchantService) ListDirectors(ctx context.Context, merchantID uuid.UUID) ([]*domain.Director, error) {
	return s.directors.ListByMerchant(ctx, merchantID)
}

// ResendDirectorVerification regenerates the token and resends the verification email.
func (s *MerchantService) ResendDirectorVerification(ctx context.Context, merchantID, directorID uuid.UUID) error {
	merchant, err := s.merchants.GetByID(ctx, merchantID)
	if err != nil {
		return err
	}
	director, err := s.directors.GetByID(ctx, directorID)
	if err != nil {
		return err
	}
	if director.MerchantID != merchantID {
		return domain.ErrDirectorNotFound
	}
	if director.Status == domain.DirectorStatusVerified {
		return fmt.Errorf("%w: director already verified", domain.ErrInvalidDirector)
	}
	if err := director.RegenerateToken(); err != nil {
		return err
	}
	if err := s.directors.Update(ctx, director); err != nil {
		return err
	}
	s.sendDirectorVerificationEmail(ctx, merchant.BusinessName, director)
	return nil
}

// RemoveDirector deletes an unverified director record.
func (s *MerchantService) RemoveDirector(ctx context.Context, merchantID, directorID uuid.UUID) error {
	director, err := s.directors.GetByID(ctx, directorID)
	if err != nil {
		return err
	}
	if director.MerchantID != merchantID {
		return domain.ErrDirectorNotFound
	}
	if director.Status == domain.DirectorStatusVerified {
		return fmt.Errorf("%w: cannot remove a verified director", domain.ErrInvalidDirector)
	}
	return s.directors.Delete(ctx, directorID)
}

// GetDirectorByToken looks up a director and their merchant by verification token.
func (s *MerchantService) GetDirectorByToken(ctx context.Context, token string) (*domain.Director, *domain.Merchant, error) {
	director, err := s.directors.GetByToken(ctx, token)
	if err != nil {
		return nil, nil, err
	}
	merchant, err := s.merchants.GetByID(ctx, director.MerchantID)
	if err != nil {
		return nil, nil, err
	}
	return director, merchant, nil
}

// SubmitDirectorVerification processes a director's identity submission.
func (s *MerchantService) SubmitDirectorVerification(ctx context.Context, token string, input SubmitDirectorInput) (*domain.Director, error) {
	director, err := s.directors.GetByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if err := director.Verify(
		input.FullName, input.NICPassportNumber, input.Phone, input.Address,
		input.DateOfBirth, input.DocumentObjectKey, input.DocumentFilename,
	); err != nil {
		return nil, err
	}
	if err := s.directors.Update(ctx, director); err != nil {
		return nil, err
	}
	merchant, _ := s.merchants.GetByID(ctx, director.MerchantID)
	if s.notifier != nil && merchant != nil {
		s.notifier.SendEmail(ctx, notification.SendEmailInput{
			MerchantID: director.MerchantID,
			Recipient:  merchant.ContactEmail,
			Subject:    "Director Verification Complete — Open Pay",
			Body: fmt.Sprintf(
				`<p>Director <strong>%s</strong> (%s) has verified their identity for <strong>%s</strong>.</p><p>You can check the KYC status in your dashboard.</p>`,
				director.FullName, director.Email, merchant.BusinessName,
			),
			EventType: "director.verified",
		})
	}
	return director, nil
}

// AllDirectorsVerified returns true if all directors for the merchant are verified.
func (s *MerchantService) AllDirectorsVerified(ctx context.Context, merchantID uuid.UUID) (bool, error) {
	directors, err := s.directors.ListByMerchant(ctx, merchantID)
	if err != nil {
		return false, err
	}
	for _, d := range directors {
		if d.Status != domain.DirectorStatusVerified {
			return false, nil
		}
	}
	return true, nil
}

func (s *MerchantService) sendDirectorVerificationEmail(ctx context.Context, businessName string, director *domain.Director) {
	if s.notifier == nil {
		return
	}
	portalURL := os.Getenv("MERCHANT_PORTAL_URL")
	if portalURL == "" {
		portalURL = "http://localhost:4600"
	}
	verifyLink := fmt.Sprintf("%s/verify/director/%s", portalURL, director.VerificationToken)
	s.notifier.SendEmail(ctx, notification.SendEmailInput{
		MerchantID: director.MerchantID,
		Recipient:  director.Email,
		Subject:    "Director Identity Verification — Open Pay",
		Body: fmt.Sprintf(
			`<p>You have been listed as a director for <strong>%s</strong> on the Open Pay platform.</p><p>As part of the KYC verification process, we need to confirm your identity.</p><p>Please click the link below to verify your identity. This link expires in 7 days.</p><p><a href="%s" style="display:inline-block;padding:12px 24px;background:#f97316;color:#fff;text-decoration:none;border-radius:6px;font-weight:600;">Verify My Identity</a></p><p>If you are not associated with this business, please ignore this email.</p>`,
			businessName, verifyLink,
		),
		EventType: "director.verification",
	})
}
