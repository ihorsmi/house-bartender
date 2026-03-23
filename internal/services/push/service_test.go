package push

import (
	"errors"
	"io"
	"log/slog"
	"sort"
	"strings"
	"testing"
	"time"

	"house-bartender-go/internal/db"
)

type fakeSender struct {
	sent    []string
	payload []byte
	results map[string]DeliveryResult
	errs    map[string]error
}

func (f *fakeSender) Send(subscription db.PushSubscription, payload []byte) (DeliveryResult, error) {
	f.sent = append(f.sent, subscription.Endpoint)
	f.payload = append([]byte(nil), payload...)
	if result, ok := f.results[subscription.Endpoint]; ok {
		return result, f.errs[subscription.Endpoint]
	}
	return DeliveryResult{StatusCode: 201}, nil
}

func TestSaveSubscriptionUpsertsDuplicateEndpoint(t *testing.T) {
	repo := newFakeRepo()
	bartenderID := repo.addUser("BARTENDER", true, true)
	service, err := newService(repo, testLogger(), Config{
		PublicKey:  "public",
		PrivateKey: "private",
		Subject:    "mailto:test@example.com",
	}, &fakeSender{})
	if err != nil {
		t.Fatalf("newService() error = %v", err)
	}

	if err := service.SaveSubscription(bartenderID, "agent-one", SubscriptionInput{
		Endpoint: "https://push.example/sub-1",
		P256DH:   "p256dh-one",
		Auth:     "auth-one",
	}); err != nil {
		t.Fatalf("SaveSubscription() first error = %v", err)
	}

	if err := service.SaveSubscription(bartenderID, "agent-two", SubscriptionInput{
		Endpoint: "https://push.example/sub-1",
		P256DH:   "p256dh-two",
		Auth:     "auth-two",
	}); err != nil {
		t.Fatalf("SaveSubscription() second error = %v", err)
	}

	subs := repo.listPushSubscriptionsForUser(bartenderID)
	if len(subs) != 1 {
		t.Fatalf("expected 1 subscription, got %d", len(subs))
	}
	if subs[0].UserAgent != "agent-two" {
		t.Fatalf("expected user agent to be updated, got %q", subs[0].UserAgent)
	}
	if subs[0].P256DH != "p256dh-two" || subs[0].Auth != "auth-two" {
		t.Fatalf("expected keys to be updated, got p256dh=%q auth=%q", subs[0].P256DH, subs[0].Auth)
	}
}

func TestNotifyNewOrderTargetsOnlyOnDutyBartenderSubscriptions(t *testing.T) {
	repo := newFakeRepo()
	userID := repo.addUser("USER", true, false)
	bartenderA := repo.addUser("BARTENDER", true, true)
	bartenderB := repo.addUser("BARTENDER", true, true)
	offDutyBartender := repo.addUser("BARTENDER", true, false)
	_ = repo.addUser("ADMIN", true, true)

	repo.mustUpsertSub(t, bartenderA, "https://push.example/a-phone", "a-phone", "a-phone-auth")
	repo.mustUpsertSub(t, bartenderA, "https://push.example/a-pwa", "a-pwa", "a-pwa-auth")
	repo.mustUpsertSub(t, bartenderB, "https://push.example/b-phone", "b-phone", "b-phone-auth")
	repo.mustUpsertSub(t, offDutyBartender, "https://push.example/off-phone", "off-phone", "off-phone-auth")
	repo.mustUpsertSub(t, userID, "https://push.example/user-phone", "user-phone", "user-phone-auth")

	orderID := repo.addOrder(db.Order{
		ID:           1,
		UserID:       userID,
		Quantity:     2,
		Location:     "Table 4",
		Notes:        "No lime",
		CocktailName: "Daiquiri",
		CreatedAt:    time.Unix(1710800000, 0),
	})

	sender := &fakeSender{}
	service, err := newService(repo, testLogger(), Config{
		PublicKey:  "public",
		PrivateKey: "private",
		Subject:    "mailto:test@example.com",
	}, sender)
	if err != nil {
		t.Fatalf("newService() error = %v", err)
	}

	if err := service.NotifyNewOrder(orderID); err != nil {
		t.Fatalf("NotifyNewOrder() error = %v", err)
	}

	sort.Strings(sender.sent)
	want := []string{
		"https://push.example/a-phone",
		"https://push.example/a-pwa",
		"https://push.example/b-phone",
	}
	sort.Strings(want)
	if len(sender.sent) != len(want) {
		t.Fatalf("expected %d sends, got %d (%v)", len(want), len(sender.sent), sender.sent)
	}
	for i := range want {
		if sender.sent[i] != want[i] {
			t.Fatalf("expected send to %v, got %v", want, sender.sent)
		}
	}
	if got := string(sender.payload); !containsAll(got, `"title":"New order"`, `"order_id":`, `"url":"/bartender/orders?highlight=`) {
		t.Fatalf("unexpected payload %q", got)
	}
}

func TestNotifyNewOrderDisablesGoneSubscriptions(t *testing.T) {
	repo := newFakeRepo()
	userID := repo.addUser("USER", true, false)
	bartenderID := repo.addUser("BARTENDER", true, true)

	repo.mustUpsertSub(t, bartenderID, "https://push.example/good", "good", "good-auth")
	repo.mustUpsertSub(t, bartenderID, "https://push.example/gone", "gone", "gone-auth")

	orderID := repo.addOrder(db.Order{
		ID:           1,
		UserID:       userID,
		Quantity:     1,
		Location:     "Bar top",
		CocktailName: "Negroni",
		CreatedAt:    time.Unix(1710801000, 0),
	})

	sender := &fakeSender{
		results: map[string]DeliveryResult{
			"https://push.example/gone": {StatusCode: 410, Details: "subscription is gone"},
		},
		errs: map[string]error{
			"https://push.example/gone": errors.New("gone"),
		},
	}
	service, err := newService(repo, testLogger(), Config{
		PublicKey:  "public",
		PrivateKey: "private",
		Subject:    "mailto:test@example.com",
	}, sender)
	if err != nil {
		t.Fatalf("newService() error = %v", err)
	}

	if err := service.NotifyNewOrder(orderID); err != nil {
		t.Fatalf("NotifyNewOrder() error = %v", err)
	}

	subs := repo.listPushSubscriptionsForUser(bartenderID)
	byEndpoint := map[string]db.PushSubscription{}
	for _, sub := range subs {
		byEndpoint[sub.Endpoint] = sub
	}

	if !byEndpoint["https://push.example/good"].Enabled {
		t.Fatalf("expected good subscription to stay enabled")
	}
	if byEndpoint["https://push.example/good"].LastSuccessAt.IsZero() {
		t.Fatalf("expected good subscription last success timestamp to be set")
	}

	if byEndpoint["https://push.example/gone"].Enabled {
		t.Fatalf("expected gone subscription to be disabled")
	}
	if byEndpoint["https://push.example/gone"].FailureCount != 1 {
		t.Fatalf("expected gone subscription failure count 1, got %d", byEndpoint["https://push.example/gone"].FailureCount)
	}
	if byEndpoint["https://push.example/gone"].LastFailureAt.IsZero() {
		t.Fatalf("expected gone subscription last failure timestamp to be set")
	}
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func containsAll(haystack string, needles ...string) bool {
	for _, needle := range needles {
		if !strings.Contains(haystack, needle) {
			return false
		}
	}
	return true
}

type fakeRepo struct {
	nextUserID int64
	nextSubID  int64
	orders     map[int64]db.Order
	users      map[int64]fakeUser
	subs       map[string]db.PushSubscription
}

type fakeUser struct {
	role   string
	active bool
	onDuty bool
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		orders: map[int64]db.Order{},
		users:  map[int64]fakeUser{},
		subs:   map[string]db.PushSubscription{},
	}
}

func (r *fakeRepo) addUser(role string, active, onDuty bool) int64 {
	r.nextUserID++
	r.users[r.nextUserID] = fakeUser{role: role, active: active, onDuty: onDuty}
	return r.nextUserID
}

func (r *fakeRepo) addOrder(order db.Order) int64 {
	r.orders[order.ID] = order
	return order.ID
}

func (r *fakeRepo) listPushSubscriptionsForUser(userID int64) []db.PushSubscription {
	var out []db.PushSubscription
	for _, sub := range r.subs {
		if sub.BartenderUserID == userID {
			out = append(out, sub)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Endpoint < out[j].Endpoint
	})
	return out
}

func (r *fakeRepo) mustUpsertSub(t *testing.T, userID int64, endpoint, p256dh, auth string) {
	t.Helper()
	if err := r.UpsertPushSubscription(db.UpsertPushSubscriptionParams{
		BartenderUserID: userID,
		Endpoint:        endpoint,
		P256DH:          p256dh,
		Auth:            auth,
		UserAgent:       "android-chrome",
	}); err != nil {
		t.Fatalf("UpsertPushSubscription(%s) error = %v", endpoint, err)
	}
}

func (r *fakeRepo) UpsertPushSubscription(p db.UpsertPushSubscriptionParams) error {
	now := time.Now()
	sub, ok := r.subs[p.Endpoint]
	if !ok {
		r.nextSubID++
		sub.ID = r.nextSubID
		sub.CreatedAt = now
	}
	sub.BartenderUserID = p.BartenderUserID
	sub.Endpoint = p.Endpoint
	sub.P256DH = p.P256DH
	sub.Auth = p.Auth
	if strings.TrimSpace(p.UserAgent) != "" {
		sub.UserAgent = p.UserAgent
	}
	if strings.TrimSpace(p.DeviceLabel) != "" {
		sub.DeviceLabel = p.DeviceLabel
	}
	sub.Enabled = true
	sub.UpdatedAt = now
	sub.LastSeenAt = now
	sub.FailureCount = 0
	sub.LastFailureAt = time.Time{}
	r.subs[p.Endpoint] = sub
	return nil
}

func (r *fakeRepo) DisablePushSubscriptionForUser(userID int64, endpoint string) error {
	sub, ok := r.subs[endpoint]
	if !ok || sub.BartenderUserID != userID {
		return nil
	}
	sub.Enabled = false
	sub.UpdatedAt = time.Now()
	r.subs[endpoint] = sub
	return nil
}

func (r *fakeRepo) GetOrderByID(id int64) (*db.Order, error) {
	order, ok := r.orders[id]
	if !ok {
		return nil, nil
	}
	copy := order
	return &copy, nil
}

func (r *fakeRepo) ListPushSubscriptionsForOnDutyBartenders() ([]db.PushSubscription, error) {
	var out []db.PushSubscription
	for _, sub := range r.subs {
		user := r.users[sub.BartenderUserID]
		if !sub.Enabled || user.role != "BARTENDER" || !user.active || !user.onDuty {
			continue
		}
		out = append(out, sub)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Endpoint < out[j].Endpoint
	})
	return out, nil
}

func (r *fakeRepo) MarkPushSubscriptionSuccess(endpoint string) error {
	sub, ok := r.subs[endpoint]
	if !ok {
		return nil
	}
	now := time.Now()
	sub.Enabled = true
	sub.UpdatedAt = now
	sub.LastSuccessAt = now
	sub.FailureCount = 0
	r.subs[endpoint] = sub
	return nil
}

func (r *fakeRepo) MarkPushSubscriptionFailure(endpoint string) error {
	sub, ok := r.subs[endpoint]
	if !ok {
		return nil
	}
	now := time.Now()
	sub.UpdatedAt = now
	sub.LastFailureAt = now
	sub.FailureCount++
	r.subs[endpoint] = sub
	return nil
}

func (r *fakeRepo) DisablePushSubscriptionByEndpoint(endpoint string) error {
	sub, ok := r.subs[endpoint]
	if !ok {
		return nil
	}
	now := time.Now()
	sub.Enabled = false
	sub.UpdatedAt = now
	sub.LastFailureAt = now
	sub.FailureCount++
	r.subs[endpoint] = sub
	return nil
}
