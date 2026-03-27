package domains

type ThemeType int

const (
	ThemeLight ThemeType = iota
	ThemeDark
)

type Settings struct {
	Theme ThemeType `json:"theme"`
}
