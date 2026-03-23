package push

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"house-bartender-go/internal/db"
)

var ErrNotConfigured = errors.New("push notifications are not configured")
var ErrInvalidSubscription = errors.New("invalid push subscription")

type Config struct {
	PublicKey  string
	PrivateKey string
	Subject    string
}

type Service struct {
	repo      Repository
	log       *slog.Logger
	sender    Sender
	enabled   bool
	publicKey string
}

type Repository interface {
	UpsertPushSubscription(p db.UpsertPushSubscriptionParams) error
	DisablePushSubscriptionForUser(userID int64, endpoint string) error
	GetOrderByID(id int64) (*db.Order, error)
	ListPushSubscriptionsForOnDutyBartenders() ([]db.PushSubscription, error)
	MarkPushSubscriptionSuccess(endpoint string) error
	MarkPushSubscriptionFailure(endpoint string) error
	DisablePushSubscriptionByEndpoint(endpoint string) error
}

type Sender interface {
	Send(subscription db.PushSubscription, payload []byte) (DeliveryResult, error)
}

type DeliveryResult struct {
	StatusCode int
	Details    string
}

type SubscriptionInput struct {
	Endpoint    string
	P256DH      string
	Auth        string
	DeviceLabel string
}

type NotificationPayload struct {
	Title     string `json:"title"`
	Body      string `json:"body"`
	OrderID   int64  `json:"order_id,omitempty"`
	URL       string `json:"url"`
	Tag       string `json:"tag"`
	Timestamp int64  `json:"timestamp"`
}

func New(repo Repository, logger *slog.Logger, cfg Config) (*Service, error) {
	return newService(repo, logger, cfg, nil)
}

func newService(repo Repository, logger *slog.Logger, cfg Config, sender Sender) (*Service, error) {
	if logger == nil {
		logger = slog.Default()
	}
	if repo == nil {
		return nil, errors.New("push repository is required")
	}

	publicKey := strings.TrimSpace(cfg.PublicKey)
	privateKey := strings.TrimSpace(cfg.PrivateKey)
	subject := strings.TrimSpace(cfg.Subject)

	allEmpty := publicKey == "" && privateKey == "" && subject == ""
	if allEmpty {
		logger.Info("web push disabled: VAPID keys are not configured")
		return &Service{repo: repo, log: logger}, nil
	}
	if publicKey == "" || privateKey == "" || subject == "" {
		return nil, errors.New("VAPID_PUBLIC_KEY, VAPID_PRIVATE_KEY, and VAPID_SUBJECT must all be set together")
	}
	if !strings.HasPrefix(subject, "mailto:") && !strings.HasPrefix(subject, "https://") {
		return nil, errors.New("VAPID_SUBJECT must start with mailto: or https://")
	}
	if sender == nil {
		sender = newWebPushSender(Config{
			PublicKey:  publicKey,
			PrivateKey: privateKey,
			Subject:    subject,
		})
	}

	return &Service{
		repo:      repo,
		log:       logger,
		sender:    sender,
		enabled:   true,
		publicKey: publicKey,
	}, nil
}

func (s *Service) Enabled() bool {
	return s != nil && s.enabled
}

func (s *Service) PublicKey() string {
	if s == nil {
		return ""
	}
	return s.publicKey
}

func (s *Service) SaveSubscription(userID int64, userAgent string, input SubscriptionInput) error {
	if !s.Enabled() {
		return ErrNotConfigured
	}
	if userID <= 0 {
		return fmt.Errorf("%w: bartender user is required", ErrInvalidSubscription)
	}
	if err := validateSubscription(input); err != nil {
		return err
	}

	return s.repo.UpsertPushSubscription(db.UpsertPushSubscriptionParams{
		BartenderUserID: userID,
		Endpoint:        strings.TrimSpace(input.Endpoint),
		P256DH:          strings.TrimSpace(input.P256DH),
		Auth:            strings.TrimSpace(input.Auth),
		UserAgent:       strings.TrimSpace(userAgent),
		DeviceLabel:     strings.TrimSpace(input.DeviceLabel),
	})
}

func (s *Service) DisableSubscription(userID int64, endpoint string) error {
	if userID <= 0 {
		return fmt.Errorf("%w: bartender user is required", ErrInvalidSubscription)
	}
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return fmt.Errorf("%w: endpoint is required", ErrInvalidSubscription)
	}
	return s.repo.DisablePushSubscriptionForUser(userID, endpoint)
}

func (s *Service) NotifyNewOrder(orderID int64) error {
	if !s.Enabled() {
		return nil
	}

	order, err := s.repo.GetOrderByID(orderID)
	if err != nil {
		s.log.Error("push notify: load order failed", "order_id", orderID, "err", err)
		return err
	}
	if order == nil {
		return nil
	}

	subscriptions, err := s.repo.ListPushSubscriptionsForOnDutyBartenders()
	if err != nil {
		s.log.Error("push notify: load subscriptions failed", "order_id", orderID, "err", err)
		return err
	}
	if len(subscriptions) == 0 {
		return nil
	}

	payload, err := json.Marshal(buildNotificationPayload(*order))
	if err != nil {
		s.log.Error("push notify: encode payload failed", "order_id", orderID, "err", err)
		return err
	}

	s.log.Info("push notify: dispatching", "order_id", orderID, "subscription_count", len(subscriptions))

	successes := 0
	failures := 0
	for _, sub := range subscriptions {
		result, sendErr := s.sender.Send(sub, payload)
		if sendErr != nil {
			failures++
			if result.StatusCode == 404 || result.StatusCode == 410 {
				_ = s.repo.DisablePushSubscriptionByEndpoint(sub.Endpoint)
				s.log.Warn("push notify: subscription disabled after invalid response",
					"order_id", orderID,
					"bartender_user_id", sub.BartenderUserID,
					"endpoint", sub.Endpoint,
					"status_code", result.StatusCode,
					"details", result.Details,
					"err", sendErr,
				)
				continue
			}

			_ = s.repo.MarkPushSubscriptionFailure(sub.Endpoint)
			s.log.Warn("push notify: delivery failed",
				"order_id", orderID,
				"bartender_user_id", sub.BartenderUserID,
				"endpoint", sub.Endpoint,
				"status_code", result.StatusCode,
				"details", result.Details,
				"err", sendErr,
			)
			continue
		}

		successes++
		_ = s.repo.MarkPushSubscriptionSuccess(sub.Endpoint)
	}

	s.log.Info("push notify: completed", "order_id", orderID, "successes", successes, "failures", failures)
	return nil
}

func validateSubscription(input SubscriptionInput) error {
	input.Endpoint = strings.TrimSpace(input.Endpoint)
	input.P256DH = strings.TrimSpace(input.P256DH)
	input.Auth = strings.TrimSpace(input.Auth)
	input.DeviceLabel = strings.TrimSpace(input.DeviceLabel)

	switch {
	case input.Endpoint == "":
		return fmt.Errorf("%w: endpoint is required", ErrInvalidSubscription)
	case len(input.Endpoint) > 2048:
		return fmt.Errorf("%w: endpoint is too long", ErrInvalidSubscription)
	case input.P256DH == "":
		return fmt.Errorf("%w: p256dh is required", ErrInvalidSubscription)
	case len(input.P256DH) > 512:
		return fmt.Errorf("%w: p256dh is too long", ErrInvalidSubscription)
	case input.Auth == "":
		return fmt.Errorf("%w: auth is required", ErrInvalidSubscription)
	case len(input.Auth) > 256:
		return fmt.Errorf("%w: auth is too long", ErrInvalidSubscription)
	case len(input.DeviceLabel) > 255:
		return fmt.Errorf("%w: device label is too long", ErrInvalidSubscription)
	default:
		return nil
	}
}

func buildNotificationPayload(order db.Order) NotificationPayload {
	return NotificationPayload{
		Title:     "New order",
		Body:      formatOrderBody(order),
		OrderID:   order.ID,
		URL:       fmt.Sprintf("/bartender/orders?highlight=%d#order-%d", order.ID, order.ID),
		Tag:       fmt.Sprintf("order-%d", order.ID),
		Timestamp: order.CreatedAt.UnixMilli(),
	}
}

func formatOrderBody(order db.Order) string {
	var parts []string
	if location := strings.TrimSpace(order.Location); location != "" {
		parts = append(parts, location)
	}

	switch {
	case strings.TrimSpace(order.CocktailName) != "" && order.Quantity > 1:
		parts = append(parts, fmt.Sprintf("%dx %s", order.Quantity, order.CocktailName))
	case strings.TrimSpace(order.CocktailName) != "":
		parts = append(parts, order.CocktailName)
	case order.Quantity > 1:
		parts = append(parts, fmt.Sprintf("%d items", order.Quantity))
	default:
		parts = append(parts, "1 item")
	}

	return strings.Join(parts, " - ")
}
