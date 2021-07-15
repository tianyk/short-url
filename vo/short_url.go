package vo

type ShortUrlVo struct {
    LongUrl string `form:"longUrl" json:"longUrl" binding:"required"`
    ShortUrl string `form:"shortUrl" json:"shortUrl"`
}