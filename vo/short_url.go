package vo

type ShortUrlVo struct {
    LongUrl string `form:"longUrl" json:"longUrl" binding:"required"`
    // 有效期，格式 time.ParseDuration 格式
    Expire string `form:"expire" json:"expire" binding:"regexp:^[0-9]+()"`
}
