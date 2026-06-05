//go:build unit

package service

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestAccountTestService_AnthropicAPIKeyBearerUsesMinimalNonStreamingRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := newTestContext()

	resp := newJSONResponse(http.StatusOK, `{"content":[{"type":"text","text":"ok"}]}`)
	repo := &mockAccountRepoForGemini{}
	upstream := &queuedHTTPUpstream{responses: []*http.Response{resp}}
	svc := &AccountTestService{accountRepo: repo, httpUpstream: upstream, cfg: &config.Config{}}
	account := &Account{
		ID:          301,
		Platform:    PlatformAnthropic,
		Type:        AccountTypeAPIKey,
		Concurrency: 1,
		Credentials: map[string]any{
			"api_key":       "jdcloud-key",
			"auth_header":   "bearer",
			"base_url":      "https://modelservice.jdcloud.com/anthropic/v1/messages",
			"model_mapping": map[string]any{"claude-opus-4-7": "T-C-2"},
		},
	}

	err := svc.testClaudeAccountConnection(ctx, account, "claude-opus-4-7")
	require.NoError(t, err)
	require.Len(t, upstream.requests, 1)

	req := upstream.requests[0]
	require.Equal(t, "https://modelservice.jdcloud.com/anthropic/v1/messages", req.URL.String())
	require.Equal(t, "Bearer jdcloud-key", getHeaderRaw(req.Header, "authorization"))
	require.Empty(t, getHeaderRaw(req.Header, "x-api-key"))

	body, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	require.Equal(t, "T-C-2", gjson.GetBytes(body, "model").String())
	require.False(t, gjson.GetBytes(body, "system").Exists())
	require.False(t, gjson.GetBytes(body, "metadata").Exists())
	require.False(t, gjson.GetBytes(body, "messages.0.content.0.cache_control").Exists())
	require.False(t, gjson.GetBytes(body, "stream").Bool())
}

func TestGatewayService_AnthropicAPIKeyBearerPassthroughBuildsJDCloudRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptestRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
	c.Request.Header.Set("Authorization", "Bearer inbound-token")
	c.Request.Header.Set("Anthropic-Beta", "oauth-2025-04-20")

	account := &Account{
		ID:       302,
		Platform: PlatformAnthropic,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			"api_key":     "jdcloud-key",
			"auth_header": "bearer",
			"base_url":    "https://modelservice.jdcloud.com/anthropic/v1/messages",
		},
		Extra: map[string]any{"anthropic_passthrough": true},
	}

	svc := &GatewayService{cfg: &config.Config{}}
	req, _, err := svc.buildUpstreamRequestAnthropicAPIKeyPassthrough(
		c.Request.Context(), c, account, []byte(`{"model":"T-C-2","messages":[],"stream":false}`), "jdcloud-key",
	)
	require.NoError(t, err)
	require.Equal(t, "https://modelservice.jdcloud.com/anthropic/v1/messages", req.URL.String())
	require.Equal(t, "Bearer jdcloud-key", getHeaderRaw(req.Header, "authorization"))
	require.Empty(t, getHeaderRaw(req.Header, "x-api-key"))
}

func httptestRecorder() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}

func TestGatewayService_AnthropicAPIKeyBearerCountTokensUsesBearer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptestRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages/count_tokens", strings.NewReader(`{}`))

	account := &Account{
		ID:       303,
		Platform: PlatformAnthropic,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			"api_key":     "jdcloud-key",
			"auth_header": "bearer",
			"base_url":    "https://modelservice.jdcloud.com/anthropic",
		},
		Extra: map[string]any{"anthropic_passthrough": true},
	}

	svc := &GatewayService{cfg: &config.Config{}}
	req, err := svc.buildCountTokensRequestAnthropicAPIKeyPassthrough(
		c.Request.Context(), c, account, []byte(`{"model":"T-C-2","messages":[]}`), "jdcloud-key",
	)
	require.NoError(t, err)
	require.Equal(t, "https://modelservice.jdcloud.com/anthropic/v1/messages/count_tokens", req.URL.String())
	require.Equal(t, "Bearer jdcloud-key", getHeaderRaw(req.Header, "authorization"))
	require.Empty(t, getHeaderRaw(req.Header, "x-api-key"))
}
