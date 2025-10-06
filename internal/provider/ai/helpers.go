package ai

// NewTextContent 创建文本内容
func NewTextContent(text string) Content {
	return Content{
		Type: ContentTypeText,
		Text: text,
	}
}

// NewImageContentFromURL 创建图像内容（URL）
func NewImageContentFromURL(url string, detail ...string) Content {
	c := Content{
		Type: ContentTypeImage,
		URL:  url,
	}
	if len(detail) > 0 {
		c.Detail = detail[0]
	}
	return c
}

// NewImageContentFromBase64 创建图像内容（Base64）
func NewImageContentFromBase64(base64 string, detail ...string) Content {
	c := Content{
		Type:   ContentTypeImage,
		Base64: base64,
	}
	if len(detail) > 0 {
		c.Detail = detail[0]
	}
	return c
}

// NewAudioContentFromURL 创建音频内容（URL）
func NewAudioContentFromURL(url string) Content {
	return Content{
		Type: ContentTypeAudio,
		URL:  url,
	}
}

// NewAudioContentFromBase64 创建音频内容（Base64）
func NewAudioContentFromBase64(base64 string) Content {
	return Content{
		Type:   ContentTypeAudio,
		Base64: base64,
	}
}

// NewVideoContentFromURL 创建视频内容（URL）
func NewVideoContentFromURL(url string) Content {
	return Content{
		Type: ContentTypeVideo,
		URL:  url,
	}
}

// NewVideoContentFromBase64 创建视频内容（Base64）
func NewVideoContentFromBase64(base64 string) Content {
	return Content{
		Type:   ContentTypeVideo,
		Base64: base64,
	}
}

// NewMessage 创建消息
func NewMessage(role Role, contents ...Content) Message {
	return Message{
		Role:    role,
		Content: contents,
	}
}

// NewSystemMessage 创建系统消息
func NewSystemMessage(text string) Message {
	return Message{
		Role: RoleSystem,
		Content: []Content{
			NewTextContent(text),
		},
	}
}

// NewUserMessage 创建用户消息
func NewUserMessage(contents ...Content) Message {
	return Message{
		Role:    RoleUser,
		Content: contents,
	}
}

// NewUserTextMessage 创建用户文本消息
func NewUserTextMessage(text string) Message {
	return Message{
		Role: RoleUser,
		Content: []Content{
			NewTextContent(text),
		},
	}
}

// NewAssistantMessage 创建助手消息
func NewAssistantMessage(text string) Message {
	return Message{
		Role: RoleAssistant,
		Content: []Content{
			NewTextContent(text),
		},
	}
}

// GetTextFromResponse 从响应中提取文本
func GetTextFromResponse(resp *CompletionResponse) string {
	if len(resp.Choices) == 0 {
		return ""
	}

	var text string
	for _, content := range resp.Choices[0].Message.Content {
		if content.Type == ContentTypeText {
			text += content.Text
		}
	}
	return text
}

// ValidateConfig 验证配置
func ValidateConfig(config *Config) error {
	if config == nil {
		return ErrInvalidConfig
	}
	if config.BaseURL == "" {
		return ErrInvalidConfig
	}
	if config.APIKey == "" {
		return ErrInvalidConfig
	}
	return nil
}
