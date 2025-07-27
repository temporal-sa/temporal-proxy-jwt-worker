package vaultauth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"io"
	"net/http"
	"sync"
	"time"
)

type VaultToken struct {
	Token  string
	Expiry time.Time
}

type VaultConfig struct {
	VaultAddr string
	Username  string
	Password  string
	OidcRole  string
}

func (v VaultConfig) GenerateToken(ctx context.Context) (*VaultToken, error) {
	// log in to Vault
	loginURL := fmt.Sprintf("%s/v1/auth/userpass/login/%s", v.VaultAddr, v.Username)

	payload := map[string]string{"password": v.Password}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(loginURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("login failed: %s", data)
	}

	var loginResp struct {
		Auth struct {
			ClientToken string `json:"client_token"`
		} `json:"auth"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&loginResp)

	vaultToken := loginResp.Auth.ClientToken

	// fetch token
	oidcURL := fmt.Sprintf("%s/v1/identity/oidc/token/%s", v.VaultAddr, v.OidcRole)

	req, err := http.NewRequest("GET", oidcURL, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("X-Vault-Token", vaultToken)

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OIDC token fetch failed: %s", data)
	}

	var oidcResp struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&oidcResp)

	token, _, err := jwt.NewParser().ParseUnverified(oidcResp.Data.Token, jwt.MapClaims{})
	if err != nil {
		return nil, err
	}

	claims := token.Claims.(jwt.MapClaims)

	expUnix, ok := claims["exp"].(float64)
	if !ok {
		return nil, fmt.Errorf("no 'exp' claim or wrong type")
	}

	expiry := time.Unix(int64(expUnix), 0)

	return &VaultToken{oidcResp.Data.Token, expiry}, nil
}

type VaultHeadersProvider struct {
	Config     VaultConfig
	Token      *VaultToken
	TokenLock  sync.Mutex
	WorkloadId string
}

func (v *VaultHeadersProvider) GetHeaders(ctx context.Context) (map[string]string, error) {
	var err error

	v.TokenLock.Lock()
	token := v.Token
	defer v.TokenLock.Unlock()

	if token == nil || time.Now().After(token.Expiry) {
		token, err = v.Config.GenerateToken(ctx)
		if err != nil {
			return nil, err
		}
		v.Token = token
	}

	return map[string]string{
		"Authorization": "Bearer " + token.Token,
		"Workload-Id":   v.WorkloadId,
	}, nil
}
