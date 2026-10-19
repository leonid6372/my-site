package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/leonid6372/my-site/backend/internal/i18n"
)

// Endpoint names a localised content response.
type Endpoint string

// Content endpoints.
const (
	EndpointMain   Endpoint = "main"
	EndpointCV     Endpoint = "cv"
	EndpointTravel Endpoint = "travel"
)

// Render reads ep from the store and returns it as JSON in lang. Nothing is
// cached: every call queries the database, so responses always reflect it.
func (s *Service) Render(ctx context.Context, ep Endpoint, lang i18n.Lang) ([]byte, error) {
	var (
		dto any
		err error
	)
	switch ep {
	case EndpointMain:
		dto, err = s.Main(ctx, lang)
	case EndpointCV:
		dto, err = s.CV(ctx, lang)
	case EndpointTravel:
		dto, err = s.Travel(ctx, lang)
	default:
		return nil, fmt.Errorf("unknown endpoint %q", ep)
	}
	if err != nil {
		return nil, err
	}
	body, err := json.Marshal(dto)
	if err != nil {
		return nil, fmt.Errorf("encode %s: %w", ep, err)
	}
	return body, nil
}
