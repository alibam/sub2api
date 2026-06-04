package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAccount_IsAnthropicAPIKeyPassthroughEnabled(t *testing.T) {
	t.Run("Anthropic API Key 开启", func(t *testing.T) {
		account := &Account{
			Platform: PlatformAnthropic,
			Type:     AccountTypeAPIKey,
			Extra: map[string]any{
				"anthropic_passthrough": true,
			},
		}
		require.True(t, account.IsAnthropicAPIKeyPassthroughEnabled())
	})

	t.Run("Anthropic API Key 关闭", func(t *testing.T) {
		account := &Account{
			Platform: PlatformAnthropic,
			Type:     AccountTypeAPIKey,
			Extra: map[string]any{
				"anthropic_passthrough": false,
			},
		}
		require.False(t, account.IsAnthropicAPIKeyPassthroughEnabled())
	})

	t.Run("字段类型非法默认关闭", func(t *testing.T) {
		account := &Account{
			Platform: PlatformAnthropic,
			Type:     AccountTypeAPIKey,
			Extra: map[string]any{
				"anthropic_passthrough": "true",
			},
		}
		require.False(t, account.IsAnthropicAPIKeyPassthroughEnabled())
	})

	t.Run("非 Anthropic API Key 账号始终关闭", func(t *testing.T) {
		oauth := &Account{
			Platform: PlatformAnthropic,
			Type:     AccountTypeOAuth,
			Extra: map[string]any{
				"anthropic_passthrough": true,
			},
		}
		require.False(t, oauth.IsAnthropicAPIKeyPassthroughEnabled())

		openai := &Account{
			Platform: PlatformOpenAI,
			Type:     AccountTypeAPIKey,
			Extra: map[string]any{
				"anthropic_passthrough": true,
			},
		}
		require.False(t, openai.IsAnthropicAPIKeyPassthroughEnabled())
	})
}

func TestAccount_GetAnthropicAPIKeyAuthHeader(t *testing.T) {
	t.Run("默认使用 x-api-key", func(t *testing.T) {
		account := &Account{
			Platform:    PlatformAnthropic,
			Type:        AccountTypeAPIKey,
			Credentials: map[string]any{},
		}
		require.Equal(t, AnthropicAPIKeyAuthHeaderXAPIKey, account.GetAnthropicAPIKeyAuthHeader())
		require.False(t, account.UsesAnthropicAPIKeyBearerAuth())
	})

	t.Run("显式 bearer 使用 Authorization Bearer", func(t *testing.T) {
		account := &Account{
			Platform: PlatformAnthropic,
			Type:     AccountTypeAPIKey,
			Credentials: map[string]any{
				"auth_header": "bearer",
			},
		}
		require.Equal(t, AnthropicAPIKeyAuthHeaderBearer, account.GetAnthropicAPIKeyAuthHeader())
		require.True(t, account.UsesAnthropicAPIKeyBearerAuth())
	})

	t.Run("非 Anthropic API Key 忽略配置", func(t *testing.T) {
		account := &Account{
			Platform: PlatformOpenAI,
			Type:     AccountTypeAPIKey,
			Credentials: map[string]any{
				"auth_header": "bearer",
			},
		}
		require.Equal(t, AnthropicAPIKeyAuthHeaderXAPIKey, account.GetAnthropicAPIKeyAuthHeader())
	})
}
