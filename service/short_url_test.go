package service

import (
    "testing"
    "time"

    "github.com/pkg/errors"
    "github.com/stretchr/testify/assert"
    leveldbErrors "github.com/syndtr/goleveldb/leveldb/errors"

    "short-url/proto"
)

func TestCreateShortUrl(t *testing.T) {
    message := &proto.ShortUrlMessage{
        LongUrl: "Test",
        Expire:  time.Now().Add(time.Minute).Unix(),
    }

    urlId, err := CreateShortUrl(message)
    if err != nil {
        t.Error(err)
    }

    assert.NotEmpty(t, urlId)
}

func TestFindLongUrl(t *testing.T) {
    message := &proto.ShortUrlMessage{
        LongUrl: "Test",
        Expire:  time.Now().Add(time.Second).Unix(),
    }

    urlId, err := CreateShortUrl(message)
    if err != nil {
        t.Error(err)
    }

    longUrl, err := FindLongUrl(urlId)
    if err != nil {
        t.Error(err)
    }
    assert.Equal(t, longUrl, "Test")

    time.Sleep(2 * time.Second)
    _, err = FindLongUrl(urlId)
    assert.Equal(t, leveldbErrors.ErrNotFound, errors.Cause(err))
}

