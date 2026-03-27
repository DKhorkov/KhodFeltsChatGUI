package interfaces

import "net/http"

//go:generate mockgen -source=http.go -destination=../../mocks/http/client.go -package=mockhttp -exclude_interfaces=
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
