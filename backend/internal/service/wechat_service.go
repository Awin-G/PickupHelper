package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	apperrors "pickup-helper/internal/errors"
)

const (
	wxTokenURL = "https://api.weixin.qq.com/cgi-bin/token"
	wxLoginURL = "https://api.weixin.qq.com/sns/jscode2session"
	wxPhoneURL = "https://api.weixin.qq.com/wxa/business/getuserphonenumber"
	wxTokenTTL = 7100 // seconds, actual TTL 7200, cache margin 100s
)

// WechatService handles WeChat mini-program login integration.
type WechatService struct {
	appID     string
	appSecret string
	cli       *http.Client

	mu             sync.Mutex
	accessToken    string
	tokenExpiresAt time.Time
}

func NewWechatService(appID, appSecret string) *WechatService {
	return &WechatService{
		appID:     appID,
		appSecret: appSecret,
		cli:       &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *WechatService) getAccessToken(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.accessToken != "" && time.Now().Before(s.tokenExpiresAt) {
		return s.accessToken, nil
	}

	url := wxTokenURL + "?grant_type=client_credential&appid=" + s.appID + "&secret=" + s.appSecret
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := s.cli.Do(req)
	if err != nil {
		return "", fmt.Errorf("wx token: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("wx token parse: %w", err)
	}
	if result.ErrCode != 0 {
		return "", fmt.Errorf("wx token error: [%d] %s", result.ErrCode, result.ErrMsg)
	}
	s.accessToken = result.AccessToken
	s.tokenExpiresAt = time.Now().Add(wxTokenTTL * time.Second)
	return s.accessToken, nil
}

// Code2Session exchanges wx.login() code for openid and session_key.
func (s *WechatService) Code2Session(ctx context.Context, code string) (string, string, error) {
	url := wxLoginURL + "?appid=" + s.appID + "&secret=" + s.appSecret +
		"&js_code=" + code + "&grant_type=authorization_code"

	resp, err := s.cli.Get(url)
	if err != nil {
		return "", "", apperrors.New(apperrors.ErrInternal, "微信登录失败")
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		OpenID     string `json:"openid"`
		SessionKey string `json:"session_key"`
		ErrCode    int    `json:"errcode"`
		ErrMsg     string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", "", apperrors.New(apperrors.ErrInternal, "微信响应异常")
	}
	if result.ErrCode != 0 {
		return "", "", apperrors.New(apperrors.ErrUnauthenticated,
			fmt.Sprintf("微信登录失败 [%d]", result.ErrCode))
	}
	return result.OpenID, result.SessionKey, nil
}

// GetPhoneNumber exchanges a phone_code for the user's phone number.
func (s *WechatService) GetPhoneNumber(ctx context.Context, phoneCode string) (string, error) {
	token, err := s.getAccessToken(ctx)
	if err != nil {
		return "", apperrors.New(apperrors.ErrInternal, "微信token获取失败")
	}

	reqBody := fmt.Sprintf(`{"code":"%s"}`, phoneCode)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		wxPhoneURL+"?access_token="+token,
		strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.cli.Do(req)
	if err != nil {
		return "", apperrors.New(apperrors.ErrInternal, "获取手机号失败")
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		ErrCode   int    `json:"errcode"`
		ErrMsg    string `json:"errmsg"`
		PhoneInfo struct {
			PhoneNumber     string `json:"phoneNumber"`
			PurePhoneNumber string `json:"purePhoneNumber"`
			CountryCode     string `json:"countryCode"`
		} `json:"phone_info"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", apperrors.New(apperrors.ErrInternal, "解析手机号失败")
	}
	if result.ErrCode != 0 {
		return "", apperrors.New(apperrors.ErrInternal,
			fmt.Sprintf("获取手机号失败 [%d]", result.ErrCode))
	}
	phone := result.PhoneInfo.PurePhoneNumber
	if phone == "" {
		phone = result.PhoneInfo.PhoneNumber
	}
	return phone, nil
}
