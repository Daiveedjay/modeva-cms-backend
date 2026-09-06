// Command reset_admin_pw generates a bcrypt hash for a new admin password
// and prints ready-to-run SQL to reset a locked-out admin account.
//
// It never touches the database itself - you copy the SQL into whichever
// psql/DB shell you trust. Uses the same bcrypt cost as the backend.
//
// Usage:
//
//	go run ./cmd/reset_admin_pw -email you@modeva.biz
//	  (prompts for the password without echoing it)
//
//	go run ./cmd/reset_admin_pw -email you@modeva.biz -password 'NewPass123'
//	  (non-interactive; avoid on shared shells - it lands in shell history)
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

func main() {
	email := flag.String("email", "", "email of the admin account to reset (required)")
	password := flag.String("password", "", "new password (optional; prompted securely if omitted)")
	promote := flag.Bool("promote", false, "also set role=super_admin and status=active")
	flag.Parse()

	if strings.TrimSpace(*email) == "" {
		fmt.Fprintln(os.Stderr, "error: -email is required")
		flag.Usage()
		os.Exit(1)
	}

	pw := *password
	if pw == "" {
		pw = promptPassword()
	}

	if len(pw) < 8 {
		fmt.Fprintln(os.Stderr, "error: password must be at least 8 characters (matches backend ValidatePassword)")
		os.Exit(1)
	}

	// bcrypt.DefaultCost matches services/admin_auth_service.go HashPassword.
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error hashing password: %v\n", err)
		os.Exit(1)
	}

	// Single-quotes in an email would break the SQL literal; guard against it.
	safeEmail := strings.ReplaceAll(*email, "'", "''")

	fmt.Println("\n-- Run this against the CMS database (local, Neon, or Railway):")
	if *promote {
		fmt.Printf(
			"UPDATE admins SET password_hash = '%s', role = 'super_admin', status = 'active', updated_at = now() WHERE email = '%s';\n",
			hash, safeEmail,
		)
	} else {
		fmt.Printf(
			"UPDATE admins SET password_hash = '%s', updated_at = now() WHERE email = '%s';\n",
			hash, safeEmail,
		)
	}
	fmt.Println("\n-- Expect: UPDATE 1. If it says UPDATE 0, the email doesn't match any row -")
	fmt.Println("-- check with: SELECT email, role, status FROM admins;")
}

func promptPassword() string {
	fmt.Fprint(os.Stderr, "New password: ")
	// Read without echo when stdin is a real terminal.
	if term.IsTerminal(int(os.Stdin.Fd())) {
		b, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading password: %v\n", err)
			os.Exit(1)
		}
		return string(b)
	}
	// Fallback for piped input.
	s, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.TrimRight(s, "\r\n")
}
