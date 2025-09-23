package utils

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	serviceresponse "sen1or/letslive/auth/response"
)

func CheckCAPTCHA(token string, userIPAddress string) *serviceresponse.Response[any] {
	formData := url.Values{}
	formData.Set("secret", os.Getenv("CLOUDFLARE_TURNSTILE_SECRET_KEY"))
	formData.Set("response", token)
	if len(userIPAddress) == 0 {
		formData.Set("remoteip", userIPAddress)
	}

	// Send verification request to Cloudflare
	resp, err := http.PostForm("https://challenges.cloudflare.com/turnstile/v0/siteverify", formData)
	if err != nil {
		return serviceresponse.NewResponseFromTemplate[any](serviceresponse.RES_ERR_CAPTCHA_FAILED, nil, nil, nil)
	}
	defer resp.Body.Close()

	type TurnstileResponse struct {
		Success bool `json:"success"`
	}

	// Parse response
	var outcome TurnstileResponse
	if err := json.NewDecoder(resp.Body).Decode(&outcome); err != nil {
		return serviceresponse.NewResponseFromTemplate[any](serviceresponse.RES_ERR_CAPTCHA_FAILED, nil, nil, nil)
	}

	if outcome.Success {
		return nil
	}

	return serviceresponse.NewResponseFromTemplate[any](serviceresponse.RES_ERR_CAPTCHA_FAILED, nil, nil, nil)
}
