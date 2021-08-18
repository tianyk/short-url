package utils

import (
    "crypto/rand"
    "fmt"
    "math"
    "strings"

    "github.com/pkg/errors"
)

// RandomString 随机生成字符串
func RandomString(length int) string {
    byteLength := int(math.Ceil(float64(length) / 2))
    buffer := make([]byte, byteLength)

    _, err := rand.Read(buffer)
    if err != nil {
        panic(errors.WithMessage(err, "RandomString"))
    }

    var sb strings.Builder
    for _, b := range buffer {
        sb.WriteString(fmt.Sprintf("%02x", b))
    }

    return sb.String()[0:length]
}