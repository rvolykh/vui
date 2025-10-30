package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/rvolykh/vui/internal/config"
)

func (am *AuthManager) authenticateWithOIDC(client *api.Client, profile *config.Profile) error {
	role := profile.AuthConfig.OIDCRole
	if role == "" {
		return fmt.Errorf("oidc_role is required for OIDC authentication")
	}

	clientNonce, err := oidcClientNonce()
	if err != nil {
		return fmt.Errorf("failed to generate client nonce: %w", err)
	}

	var (
		secretCh = make(chan *api.Secret)
		errCh    = make(chan error)
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/oidc/callback", oidcCallbackHandler(client, clientNonce, secretCh))
	tmpSrv := &http.Server{
		Addr:    "localhost:8250",
		Handler: mux,
	}
	defer func() {
		tmpSrv.Shutdown(context.TODO())

		close(secretCh)
		close(errCh)
	}()

	go func() {
		err := tmpSrv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("failed to listen and serve: %w", err)
		}
	}()

	if err := oidcOpenBrowser(client, role, clientNonce); err != nil {
		return fmt.Errorf("failed to open browser for OIDC auth: %w", err)
	}

	select {
	case secret := <-secretCh:
		if secret == nil || secret.Auth == nil {
			return fmt.Errorf("no authentication data returned from OIDC")
		}
		client.SetToken(secret.Auth.ClientToken)
		return nil
	case err := <-errCh:
		return fmt.Errorf("failed to authenticate with OIDC: %w", err)
	case <-time.After(1 * time.Minute):
		return fmt.Errorf("timed out waiting for OIDC authentication")
	}
}

func oidcCallbackHandler(client *api.Client, clientNonce string, secretCh chan<- *api.Secret) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		data := map[string][]string{
			"code":         {req.FormValue("code")},
			"client_nonce": {clientNonce},
			"id_token":     {req.FormValue("id_token")},
			"state":        {req.FormValue("state")},
		}

		secret, err := client.Logical().ReadWithData("auth/oidc/oidc/callback", data)
		if err != nil {
			oidcRenderHTML(w, fmt.Errorf("failed to read with data: %v", err))
			return
		}

		secretCh <- secret
		oidcRenderHTML(w, nil)
	}
}

func oidcRenderHTML(w http.ResponseWriter, failure error) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	data := map[string]string{
		"Status":  "Success",
		"Message": "OIDC authentication completed successfully.",
	}
	if failure != nil {
		data["Status"] = "Failed"
		data["Message"] = failure.Error()
	}
	tmpl, err := template.New("oidcHTML").Parse(`
		<!DOCTYPE html>
		<html>
		<head>
			<title>VUI OIDC Authentication</title>
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
		</head>
		<body style="font-family:sans-serif;max-width:480px;margin:2em auto;text-align:center;">
			<h2>Authentication {{ .Status }}</h2>
			<p>{{ .Message }}</p>
			<p>You may close this page.</p>
		</body>
		</html>
	`)
	if err != nil {
		w.Write([]byte(fmt.Sprintf("failed to parse template: %v", err)))
		return
	}
	tmpl.Execute(w, data)
}

func oidcClientNonce() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes for nonce: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func oidcOpenBrowser(client *api.Client, role string, clientNonce string) error {
	data := map[string]interface{}{
		"role":         role,
		"redirect_uri": "http://localhost:8250/oidc/callback",
		"client_nonce": clientNonce,
	}

	secret, err := client.Logical().Write("auth/oidc/oidc/auth_url", data)
	if err != nil {
		return fmt.Errorf("failed to write auth URL: %w", err)
	}
	if secret == nil {
		return fmt.Errorf("no secret returned for auth URL")
	}

	url := secret.Data["auth_url"].(string)
	if url == "" {
		return fmt.Errorf("no auth URL returned for auth URL")
	}

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		return fmt.Errorf("failed to open browser for OIDC auth: %w", err)
	}

	return nil
}
