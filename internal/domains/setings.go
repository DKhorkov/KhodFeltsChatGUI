package domains

type ThemeType int

const (
	ThemeLight ThemeType = iota
	ThemeDark
)

type NotificationConsent int

const (
	ConsentNewMessage NotificationConsent = 1 << iota
)

type Settings struct {
	Theme           ThemeType           `json:"theme"`
	EmailConsents   NotificationConsent `json:"emailConsents"`
	WebPushConsents NotificationConsent `json:"webPushConsents"`
}
