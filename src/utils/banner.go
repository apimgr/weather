package utils

import (
	"fmt"
	"os"
)

// DisplayFirstRunBanner displays the startup banner with setup information
// AI.md: Host-specific values detected at runtime
func DisplayFirstRunBanner(port int, setupToken string, isDockerized bool, torOnion string) {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "localhost"
	}

	localURL := fmt.Sprintf("http://%s:%d", hostname, port)
	dockerURL := ""
	if isDockerized {
		// AI.md: Detect Docker gateway at runtime, not hardcoded
		if gwIP := GetDockerGatewayIP(); gwIP != "" {
			dockerURL = fmt.Sprintf("http://%s:%d", gwIP, port)
		}
	}

	// Banner width
	width := 65

	fmt.Println()
	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Printf("║%s║\n", centerText("🌤️  Weather Service - First Run Setup", width))
	fmt.Println("║                                                               ║")
	fmt.Printf("║  🌐 Local:  %-49s ║\n", localURL)

	if isDockerized && dockerURL != "" {
		fmt.Printf("║  🐳 Docker: %-49s ║\n", dockerURL)
	}

	if torOnion != "" {
		fmt.Printf("║  🧅 Tor:    %-49s ║\n", torOnion)
	}

	fmt.Println("║                                                               ║")
	fmt.Println("║  ⚡ Server started successfully (before setup)                ║")
	fmt.Println("║                                                               ║")
	fmt.Printf("║  🔐 Setup Token: %-44s ║\n", setupToken)
	fmt.Println("║     Use this ONE TIME to complete server setup               ║")
	fmt.Println("║     Navigate to /admin/server/setup in your browser          ║")
	fmt.Println("║                                                               ║")
	fmt.Println("║  📝 Auto-generated server.yml created                         ║")
	fmt.Println("║  📧 SMTP auto-detected and configured                         ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
	fmt.Println()
}

// DisplayNormalBanner displays the normal startup banner (not first run)
// AI.md: Host-specific values detected at runtime
func DisplayNormalBanner(version, buildDate string, port int, isDockerized bool, torOnion string) {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "localhost"
	}

	localURL := fmt.Sprintf("http://%s:%d", hostname, port)
	dockerURL := ""
	if isDockerized {
		// AI.md: Detect Docker gateway at runtime, not hardcoded
		if gwIP := GetDockerGatewayIP(); gwIP != "" {
			dockerURL = fmt.Sprintf("http://%s:%d", gwIP, port)
		}
	}

	width := 65

	fmt.Println()
	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Printf("║%s║\n", centerText(fmt.Sprintf("🌤️  Weather Service v%s", version), width))
	fmt.Println("║                                                               ║")
	fmt.Printf("║  🌐 Local:  %-49s ║\n", localURL)

	if isDockerized && dockerURL != "" {
		fmt.Printf("║  🐳 Docker: %-49s ║\n", dockerURL)
	}

	if torOnion != "" {
		fmt.Printf("║  🧅 Tor:    %-49s ║\n", torOnion)
	}

	fmt.Println("║                                                               ║")
	fmt.Printf("║  Built: %-54s ║\n", buildDate)
	fmt.Println("║                                                               ║")
	fmt.Println("║  ✅ Server ready                                              ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
	fmt.Println()
}

// centerText centers text within a given width
func centerText(text string, width int) string {
	if len(text) >= width {
		return text[:width]
	}
	padding := width - len(text)
	leftPad := padding / 2
	rightPad := padding - leftPad

	result := ""
	for i := 0; i < leftPad; i++ {
		result += " "
	}
	result += text
	for i := 0; i < rightPad; i++ {
		result += " "
	}
	return result
}
