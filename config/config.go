package config

import (
    "fmt"
    "os"
    "strconv"

    "github.com/joho/godotenv"
)

type config struct {
    Port   int    `env:"APP_PORT" envDefault: 5000`
    Prefix string `env:"APP_PREFIX" `
    Token  string `env:"APP_TOKEN"`
}

var Config = new(config)

func init() {
    env := os.Getenv("APP_ENV")
    if "" == env {
        env = "development"
    }

    // https://github.com/bkeepers/dotenv#what-other-env-files-can-i-use
    godotenv.Load(".env." + env + ".local")
    if "test" != env {
        godotenv.Load(".env.local")
    }
    godotenv.Load(".env." + env)
    godotenv.Load() // The Original .env

    port, err := strconv.Atoi(os.Getenv("APP_PORT"))
    if err != nil {
        panic(fmt.Errorf("err port %s", err.Error()))
    }
    Config.Port = port

    Config.Prefix = os.Getenv("APP_PREFIX")

    Config.Token = os.Getenv("APP_TOKEN")
}
