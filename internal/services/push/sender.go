package push

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"house-bartender-go/internal/db"

	webpush "github.com/SherClockHolmes/webpush-go"
)

type webPushSender struct {
	options webpush.Options
}

func newWebPushSender(cfg Config) Sender {
	return &webPushSender{
		options: webpush.Options{
			Subscriber:      cfg.Subject,
			VAPIDPublicKey:  cfg.PublicKey,
			VAPIDPrivateKey: cfg.PrivateKey,
			TTL:             60,
		},
	}
}

func (s *webPushSender) Send(subscription db.PushSubscription, payload []byte) (DeliveryResult, error) {
	resp, err := webpush.SendNotification(payload, &webpush.Subscription{
		Endpoint: subscription.Endpoint,
		Keys: webpush.Keys{
			P256dh: subscription.P256DH,
			Auth:   subscription.Auth,
		},
	}, &s.options)
	if err != nil {
		return DeliveryResult{}, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	result := DeliveryResult{
		StatusCode: resp.StatusCode,
		Details:    strings.TrimSpace(string(body)),
	}
	if resp.StatusCode >= http.StatusMultipleChoices {
		return result, fmt.Errorf("push gateway returned %s", resp.Status)
	}
	return result, nil
}
