package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/injectionator/n8r/internal/auth"
	"github.com/injectionator/n8r/internal/config"
)

const banner = `
 ███╗   ██╗ █████╗ ██████╗
 ████╗  ██║██╔══██╗██╔══██╗
 ██╔██╗ ██║╚█████╔╝██████╔╝
 ██║╚██╗██║██╔══██╗██╔══██╗
 ██║ ╚████║╚█████╔╝██║  ██║
 ╚═╝  ╚═══╝ ╚════╝ ╚═╝  ╚═╝
`

func main() {
	if len(os.Args) < 2 {
		printWelcome()
		os.Exit(0)
	}

	cmd := os.Args[1]

	switch cmd {
	case "login":
		cmdLogin()
	case "logout":
		cmdLogout()
	case "status":
		cmdStatus()
	case "profile":
		cmdProfile()
	case "version", "--version", "-v":
		fmt.Printf("n8r v%s\n", config.Version)
	case "help", "--help", "-h":
		printWelcome()
	default:
		token, err := auth.LoadToken()
		if err != nil || token == nil || token.IsExpired() {
			fmt.Fprintln(os.Stderr, "You must be logged in to use this command.")
			fmt.Fprintln(os.Stderr, "Run `n8r login` to authenticate.")
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printWelcome() {
	fmt.Print(banner)
	fmt.Printf(" Injectionator CLI v%s\n", config.Version)
	fmt.Println(" https://injectionator.com")
	fmt.Println()

	token, _ := auth.LoadToken()
	if token == nil || token.IsExpired() {
		fmt.Println(" You are not logged in.")
		fmt.Println(" This is an alpha product with limited access.")
		fmt.Println(" Run `n8r login` to authenticate with your Injectionator account.")
		fmt.Println()
	} else {
		fmt.Println(" You are logged in.")
		fmt.Println()
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Commands:")
	fmt.Println("  n8r login          Authenticate with Injectionator")
	fmt.Println("  n8r logout         Remove stored credentials")
	fmt.Println("  n8r status         Show authentication status")
	fmt.Println("  n8r profile        View your Injectionator profile")
	fmt.Println("  n8r version        Print version")
	fmt.Println()
}

func requireToken() *auth.StoredToken {
	token, err := auth.LoadToken()
	if err != nil || token == nil || token.IsExpired() {
		fmt.Fprintln(os.Stderr, "You must be logged in to use this command.")
		fmt.Fprintln(os.Stderr, "Run `n8r login` to authenticate.")
		os.Exit(1)
	}
	return token
}

func cmdProfile() {
	token := requireToken()

	req, err := http.NewRequest("GET", config.BaseURL+"/api/auth/profile", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Origin", config.BaseURL)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		fmt.Fprintln(os.Stderr, "Session expired. Run `n8r login` to re-authenticate.")
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Error (HTTP %d): %s\n", resp.StatusCode, string(body))
		os.Exit(1)
	}

	var profile struct {
		User struct {
			Email        string   `json:"email"`
			Name         string   `json:"name"`
			Role         string   `json:"role"`
			Organization string   `json:"organization"`
			Interests    []string `json:"interests"`
		} `json:"user"`
		Cohorts     []string `json:"cohorts"`
		Points      int      `json:"points"`
		AlphaAccess bool     `json:"alpha_access"`
		Badges      []struct {
			Name    string `json:"name"`
			Emoji   string `json:"emoji"`
			Mission string `json:"mission"`
		} `json:"badges"`
		Missions struct {
			Completed int `json:"completed"`
			Total     int `json:"total"`
		} `json:"missions"`
	}

	if err := json.Unmarshal(body, &profile); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing profile: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(banner)
	fmt.Println(" Your Injectionator Profile")
	fmt.Println(" ──────────────────────────")
	fmt.Println()

	name := profile.User.Name
	if name == "" {
		name = "(not set)"
	}
	fmt.Printf("  Name:         %s\n", name)
	fmt.Printf("  Email:        %s\n", profile.User.Email)

	if profile.User.Role != "" {
		fmt.Printf("  Role:         %s\n", profile.User.Role)
	}
	if profile.User.Organization != "" {
		fmt.Printf("  Organization: %s\n", profile.User.Organization)
	}
	if len(profile.User.Interests) > 0 {
		fmt.Printf("  Interests:    %s\n", strings.Join(profile.User.Interests, ", "))
	}

	fmt.Println()

	if len(profile.Cohorts) > 0 {
		fmt.Printf("  Cohorts:      %s\n", strings.Join(profile.Cohorts, ", "))
	}
	if profile.AlphaAccess {
		fmt.Println("  Alpha Access: Yes")
	}

	fmt.Println()
	fmt.Printf("  Points:       %d\n", profile.Points)
	fmt.Printf("  Missions:     %d/%d completed\n", profile.Missions.Completed, profile.Missions.Total)

	if len(profile.Badges) > 0 {
		fmt.Println()
		fmt.Println("  Badges:")
		for _, b := range profile.Badges {
			emoji := b.Emoji
			if emoji == "" {
				emoji = " "
			}
			fmt.Printf("    %s %s (%s)\n", emoji, b.Name, b.Mission)
		}
	}

	fmt.Println()
}

func cmdLogin() {
	fmt.Print(banner)
	fmt.Println(" Authenticate with Injectionator")
	fmt.Println()

	existing, _ := auth.LoadToken()
	if existing != nil && !existing.IsExpired() {
		fmt.Println("You are already authenticated.")
		fmt.Println("Run `n8r logout` first to re-authenticate.")
		return
	}

	dcr, err := auth.RequestDeviceCode()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Visit %s and enter code:\n\n", dcr.VerificationURI)
	fmt.Printf("  >>> %s <<<\n\n", dcr.UserCode)
	fmt.Println("Waiting for authorization...")

	token, err := auth.PollForToken(dcr.DeviceCode, dcr.Interval, dcr.ExpiresIn)
	if err != nil {
		if err == auth.ErrExpiredToken {
			fmt.Fprintln(os.Stderr, "Error: Device code expired. Please run `n8r login` again.")
		} else if err == auth.ErrAccessDenied {
			fmt.Fprintln(os.Stderr, "Error: Authorization was denied.")
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}

	if err := auth.SaveToken(*token); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving credentials: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully authenticated!")
}

func cmdLogout() {
	if err := auth.DeleteToken(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Logged out. Credentials removed.")
}

func cmdStatus() {
	token, err := auth.LoadToken()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading credentials: %v\n", err)
		os.Exit(1)
	}

	if token == nil {
		fmt.Println("Not authenticated. Run `n8r login` to get started.")
		return
	}

	if token.IsExpired() {
		fmt.Println("Status: Token expired")
		fmt.Printf("Expired at: %s\n", token.ExpiresAt.Format(time.RFC3339))
		fmt.Println("Run `n8r login` to re-authenticate.")
		return
	}

	fmt.Println("Status: Authenticated")
	fmt.Printf("Token type: %s\n", token.TokenType)
	fmt.Printf("Expires at: %s\n", token.ExpiresAt.Format(time.RFC3339))
	fmt.Printf("Saved at:   %s\n", token.SavedAt.Format(time.RFC3339))
}
