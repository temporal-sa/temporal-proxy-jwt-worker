package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	vaultAddr = "http://127.0.0.1:8200"
	username  = "brendan"
	password  = "secret"
	oidcRole  = "temporal-worker"
)

func main() {
	// 1. Login with userpass
	loginURL := fmt.Sprintf("%s/v1/auth/userpass/login/%s", vaultAddr, username)

	payload := map[string]string{"password": password}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(loginURL, "application/json", bytes.NewReader(body))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		panic(fmt.Sprintf("Login failed: %s", data))
	}

	var loginResp struct {
		Auth struct {
			ClientToken string `json:"client_token"`
		} `json:"auth"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&loginResp)

	vaultToken := loginResp.Auth.ClientToken
	fmt.Println("✅ Vault token acquired")

	// 2. Fetch the OIDC token
	oidcURL := fmt.Sprintf("%s/v1/identity/oidc/token/%s", vaultAddr, oidcRole)

	req, err := http.NewRequest("GET", oidcURL, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("X-Vault-Token", vaultToken)

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		panic(fmt.Sprintf("OIDC token fetch failed: %s", data))
	}

	var oidcResp struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&oidcResp)

	fmt.Println("Resp body:", resp.Body)

	fmt.Println("🎫 JWT:", oidcResp.Data.Token)
}
