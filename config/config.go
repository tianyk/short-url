package config

import (
    "fmt"
    "os"
    "strconv"

    _ "github.com/joho/godotenv/autoload"
)

type config struct {
    Port   int    `env:"APP_PORT" envDefault: 5000`
    Prefix string `env:"APP_PREFIX" `
}

var Config = new(config)

func init() {
    port, err := strconv.Atoi(os.Getenv("APP_PORT"))
    if err != nil {
        panic(fmt.Errorf("err port %s", err.Error()))
    }

    Config.Port = port
    Config.Prefix = os.Getenv("APP_PREFIX")
}
