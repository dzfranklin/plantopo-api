package authn

import (
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/json"
	"github.com/workos/workos-go/v4/pkg/usermanagement"
	"net/http"
)

type WorkOS struct {
	jwks   jose.JSONWebKeySet
	client *usermanagement.Client
}

func NewWorkOS(clientID, apiKey string) (*WorkOS, error) {
	jwks, err := fetchJWKS(clientID)
	if err != nil {
		return nil, err
	}
	client := usermanagement.NewClient(apiKey)
	return &WorkOS{jwks: jwks, client: client}, nil
}

func (w *WorkOS) Verify(token string) (string, error) {
	sig, err := jose.ParseSigned(token, []jose.SignatureAlgorithm{jose.RS256})
	if err != nil {
		return "", err
	}
	payload, err := sig.Verify(w.jwks)
	if err != nil {
		return "", err
	}
	userId, err := unmarshalPayload(payload)
	if err != nil {
		return "", err

	}
	return userId, nil
}

func unmarshalPayload(payload []byte) (string, error) {
	var data struct {
		Sub string `json:"sub"`
	}
	err := json.Unmarshal(payload, &data)
	if err != nil {
		return "", err
	}
	return data.Sub, nil
}

func fetchJWKS(clientID string) (jose.JSONWebKeySet, error) {
	url, err := usermanagement.GetJWKSURL(clientID)
	if err != nil {
		return jose.JSONWebKeySet{}, err
	}

	resp, err := http.Get(url.String())
	if err != nil {
		return jose.JSONWebKeySet{}, err
	}
	defer resp.Body.Close()

	var jwks jose.JSONWebKeySet
	err = json.NewDecoder(resp.Body).Decode(&jwks)
	if err != nil {
		return jose.JSONWebKeySet{}, err
	}

	return jwks, nil
}
