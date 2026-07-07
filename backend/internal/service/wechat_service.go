package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	apperrors "pickup-helper/internal/errors"
)

const (
	wxTokenURL = "https://api.weixin.qq.com/cgi-bin/token"
	wxLoginURL = "https://api.weixin.qq.com/sns/jscode2session"
	wxPhoneURL = "https://api.weixin.qq.com/wxa/business/getuserphonenumber"
	wxTokenTTL = 7100
)

// WechatService handles WeChat mini-program login integration.
type WechatService struct {
	appID     string
	appSecret string
	cli       *http.Client
	log       *slog.Logger

	mu             sync.Mutex
	accessToken    string
	tokenExpiresAt time.Time
}

func NewWechatService(appID, appSecret string) *WechatService {
	return &WechatService{
		appID:     appID,
		appSecret: appSecret,
		cli:       &http.Client{Timeout: 10 * time.Second},
		log:       slog.Default(),
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
		"&js_code=" + url.QueryEscape(code) + "&grant_type=authorization_code"

	resp, err := s.cli.Get(url)
	if err != nil {
		s.log.ErrorContext(ctx, "wechat code2session http fail",
			slog.String("err", err.Error()))
		return "", "", apperrors.New(apperrors.ErrInternal, "微信登录请求失败")
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	s.log.InfoContext(ctx, "wechat code2session response",
		slog.Int("http_status", resp.StatusCode),
		slog.String("body", string(body)))

	var result struct {
		OpenID     string `json:"openid"`
		SessionKey string `json:"session_key"`
		UnionID    string `json:"unionid"`
		ErrCode    int    `json:"errcode"`
		ErrMsg     string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", "", apperrors.New(apperrors.ErrInternal, "微信响应异常")
	}
	if result.ErrCode != 0 {
		s.log.ErrorContext(ctx, "wechat code2session error",
			slog.Int("errcode", result.ErrCode),
			slog.String("errmsg", result.ErrMsg))

		msg := result.ErrMsg
		if msg == "" {
			msg = wechatErrText(result.ErrCode)
		}
		return "", "", apperrors.New(apperrors.ErrUnauthenticated,
			fmt.Sprintf("微信登录失败 [%d] %s", result.ErrCode, msg))
	}
	return result.OpenID, result.SessionKey, nil
}

func wechatErrText(code int) string {
	switch code {
	case -1: return "系统繁忙"
	case 40029: return "code无效(已过期或已使用,请重新wx.login)"
	case 45011: return "频率限制"
	case 40013: return "appid无效"
	case 40125: return "appsecret无效"
	default: return ""
	}
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
