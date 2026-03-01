package main

import (
  "fmt"
  "time"

  auth "sercherai/backend/internal/platform/auth"
)

func main() {
  token, err := auth.SignToken("sercherai_dev_secret_change_me", "admin_001", "ADMIN", "ACCESS", 2*time.Hour)
  if err != nil {
    panic(err)
  }
  fmt.Print(token)
}
