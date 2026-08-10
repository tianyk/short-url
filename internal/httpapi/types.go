package httpapi

// CreateShortURLRequest is the request payload for creating a short URL.
type CreateShortURLRequest struct {
	LongUrl string `form:"longUrl" json:"longUrl" binding:"required"`
	// 有效期，格式 time.ParseDuration 格式 https://github.com/xhit/go-str2duration
	MaxAge string `form:"maxAge" json:"maxAge,omitempty"`
	// Visibility 控制短地址是否公开，取值 "public"/"private"。
	// 字段名与 JSON key 不一致是有意为之：JSON key "scope" 是外部 API 契约，保持兼容。
	Visibility string `form:"scope" json:"scope,omitempty"`
}

// CreateShortURLResponse is the response payload after creating a short URL.
type CreateShortURLResponse struct {
	LongUrl  string `json:"longUrl"`
	ShortUrl string `json:"shortUrl"`
	Password string `json:"password,omitempty"`
}
