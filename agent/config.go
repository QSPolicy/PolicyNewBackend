package agent

// LLMConfig 单个 LLM 提供者的配置
type LLMConfig struct {
	Name    string `koanf:"name"`
	APIKey  string `koanf:"api_key"`
	BaseURL string `koanf:"base_url"`
	Model   string `koanf:"model"`
}

// Config Agent 模块配置
type Config struct {
	LLMConfigs        []LLMConfig `koanf:"llm_configs"`
	BaiduSearchAPIKey string      `koanf:"baidu_search_api_key"`
}

// DefaultConfig 返回 Agent 模块的默认配置
func DefaultConfig() Config {
	return Config{
		LLMConfigs:        []LLMConfig{},
		BaiduSearchAPIKey: "",
	}
}
