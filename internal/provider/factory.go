package provider

import "fmt"

// NewProvider constructs a Provider implementation based on the given name.
// baseURL is forwarded to the selected provider's BaseURL field.
// apiKey is forwarded to providers that require authentication (e.g. OpenAI);
// it is ignored for local providers such as Ollama.
func NewProvider(name, baseURL, apiKey string) (Provider, error) {
	switch name {
	case "ollama":
		return &OllamaProvider{
			BaseURL: baseURL,
		}, nil
	case "openai":
		return &OpenAIProvider{
			BaseURL: baseURL,
			APIKey:  apiKey,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported provider %q: supported providers are \"ollama\", \"openai\"", name)
	}
}

// RequiresAPIKey reports whether the named provider requires an API key.
// Unknown provider names return false so that unsupported providers reach
// the factory and receive the "unsupported provider" error, not a
// misleading "missing API key" error.
func RequiresAPIKey(name string) bool {
	switch name {
	case "openai":
		return true
	default:
		return false
	}
}
