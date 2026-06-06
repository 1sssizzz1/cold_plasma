package utils

import (
    "math/rand"
    "time"
)

var (
    rng = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func GenerateCode() string {
    const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    code := make([]byte, 6)
    for i := range code {
        code[i] = chars[rng.Intn(len(chars))]
    }
    return string(code)
}

