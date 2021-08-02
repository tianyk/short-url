package vo

type ShortUrlVo struct {
    LongUrl string `form:"longUrl" json:"longUrl" binding:"required"`
    // 有效期，格式 time.ParseDuration 格式 https://github.com/xhit/go-str2duration
    MaxAge string `form:"maxAge" json:"maxAge"`
}
