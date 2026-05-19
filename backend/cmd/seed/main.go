package main

import (
	"fmt"
	"os"

	pkgauth "github.com/your-org/itsm_ops/backend/pkg/auth"
)

func main() {
	password := "admin@123"
	if len(os.Args) > 1 {
		password = os.Args[1]
	}
	hashed, err := pkgauth.HashPassword(password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("INSERT OR IGNORE INTO users (username, password, display_name, status) VALUES ('admin', '%s', '系统管理员', 1);\n", hashed)
}
