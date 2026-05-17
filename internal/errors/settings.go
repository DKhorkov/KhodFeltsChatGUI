package errors

import "errors"

var (
	ErrSettingsNotFound            = errors.New("settings not found")
	ErrWebPushSubscriptionNotFound = errors.New("web-push subscription not found")
	ErrWebPushSubscriptionExpired  = errors.New("web-push subscription expired")
)
