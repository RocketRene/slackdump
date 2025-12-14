package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/joho/godotenv"
	"github.com/rusq/slack"
	"github.com/rusq/slackdump/v3"
	"github.com/rusq/slackdump/v3/auth"
	"github.com/rusq/slackdump/v3/internal/network"
	"github.com/rusq/slackdump/v3/types"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Kobe User Research Slack Export Tool")
	myWindow.Resize(fyne.NewSize(700, 600))

	// Title and description
	title := widget.NewLabel("Kobe User Research Slack Export Tool")
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter
	
	description := widget.NewLabel("For Research Purposes Only")
	description.TextStyle = fyne.TextStyle{Italic: true}
	description.Alignment = fyne.TextAlignCenter

	// Authentication fields
	subdomainEntry := widget.NewEntry()
	subdomainEntry.SetPlaceHolder("Enter workspace subdomain (e.g., myworkspace)")
	
	cookieEntry := widget.NewEntry()
	cookieEntry.SetPlaceHolder("Enter 'd' cookie value")
	// Make cookie visible by default (not password masked)

	// Load credentials from .env if available
	loadCredentials := func() {
		// Try to load from current directory .env
		if err := godotenv.Load(".env"); err == nil {
			if subdomain := os.Getenv("SLACK_SUBDOMAIN"); subdomain != "" {
				subdomainEntry.SetText(subdomain)
			}
			if cookie := os.Getenv("SLACK_COOKIE"); cookie != "" {
				cookieEntry.SetText(cookie)
			}
		}
	}
	loadCredentials()

	// Function to save credentials to .env
	saveCredentials := func(subdomain, cookie string) error {
		// Get current directory
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		
		envPath := filepath.Join(cwd, ".env")
		content := fmt.Sprintf("SLACK_SUBDOMAIN=%s\nSLACK_COOKIE=%s\n", subdomain, cookie)
		
		return os.WriteFile(envPath, []byte(content), 0600)
	}

	// Output folder - default to ~/Documents/koberesearch/
	outputFolderEntry := widget.NewEntry()
	outputFolderEntry.SetPlaceHolder("~/Documents/koberesearch/")
	outputFolderEntry.SetText("~/Documents/koberesearch/")

	// Options
	ignoreMediaCheck := widget.NewCheck("Ignore Media (files/images)", nil)
	ignoreMediaCheck.Checked = true // Default to true

	// Speed settings
	speedLabel := widget.NewLabel("Fetch Speed:")
	speedSelect := widget.NewSelect([]string{"Default", "Fast", "Maximum"}, nil)
	speedSelect.SetSelected("Fast") // Default to Fast
	
	// Date selection - default to 2025
	yearLabel := widget.NewLabel("Export Year:")
	yearSelect := widget.NewSelect([]string{"2025", "2024", "2023", "2022", "2021"}, nil)
	yearSelect.SetSelected("2025")

	// Status label
	statusLabel := widget.NewLabel("")
	statusLabel.Wrapping = fyne.TextWrapWord

	var channels types.Channels
	var channelsMux sync.Mutex
	var sess *slackdump.Session

	// Channels list
	channelList := widget.NewList(
		func() int {
			channelsMux.Lock()
			defer channelsMux.Unlock()
			return len(channels)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			channelsMux.Lock()
			defer channelsMux.Unlock()
			if id < len(channels) {
				label := obj.(*widget.Label)
				ch := channels[id]
				label.SetText(fmt.Sprintf("%s (%s)", ch.Name, ch.ID))
			}
		},
	)

	// Get Channels button (declare variable first for closure)
	var getChannelsBtn *widget.Button
	getChannelsBtn = widget.NewButton("Get Full List of Channels", func() {
		subdomain := strings.TrimSpace(subdomainEntry.Text)
		cookie := strings.TrimSpace(cookieEntry.Text)

		if subdomain == "" || cookie == "" {
			dialog.ShowError(fmt.Errorf("subdomain and cookie are required"), myWindow)
			return
		}

		// Save credentials to .env file
		if err := saveCredentials(subdomain, cookie); err != nil {
			// Don't fail if save fails, just log
			statusLabel.SetText(fmt.Sprintf("Warning: Failed to save credentials: %v", err))
		}

		// Disable button during fetch
		getChannelsBtn.Disable()
		
		// Clear previous channels
		channelsMux.Lock()
		channels = nil
		channelsMux.Unlock()
		channelList.Refresh()

		statusLabel.SetText("Authenticating...")

		// Run in goroutine to keep UI responsive
		go func() {
			defer getChannelsBtn.Enable()
			
			// Authenticate
			ctx := context.Background()
			authProvider, err := auth.NewCookieOnlyAuth(ctx, subdomain, cookie)
			if err != nil {
				statusLabel.SetText(fmt.Sprintf("Authentication failed: %v", err))
				return
			}

			// Test authentication
			_, err = authProvider.Test(ctx)
			if err != nil {
				statusLabel.SetText(fmt.Sprintf("Authentication test failed: %v", err))
				return
			}

			statusLabel.SetText("Authenticated! Getting channels...")

			// Determine speed limits based on user selection
			var limits network.Limits
			switch speedSelect.Selected {
			case "Maximum":
				limits = network.NoLimits
			case "Fast":
				// Custom fast limits - higher than default but not unlimited
				// Create a copy to avoid modifying global defaults
				limits = network.DefLimits
				limits.Tier2.Boost = 60  // Increase from 20
				limits.Tier2.Burst = 10  // Increase from 3
			default: // "Default"
				limits = network.DefLimits
			}

			// Create session with configured limits
			opts := []slackdump.Option{
				slackdump.WithLimits(limits),
			}
			sess, err = slackdump.New(ctx, authProvider, opts...)
			if err != nil {
				statusLabel.SetText(fmt.Sprintf("Failed to create session: %v", err))
				return
			}

			// Stream channels as they arrive
			statusLabel.SetText("Fetching channels...")
			err = sess.StreamChannels(ctx, slackdump.AllChanTypes, func(ch slack.Channel) error {
				// Add channel to list (thread-safe)
				channelsMux.Lock()
				channels = append(channels, ch)
				count := len(channels)
				channelsMux.Unlock()
				
				// Update UI - Fyne handles thread-safety internally
				statusLabel.SetText(fmt.Sprintf("Fetching channels... (%d found)", count))
				channelList.Refresh()
				
				return nil
			})

			if err != nil {
				statusLabel.SetText(fmt.Sprintf("Failed to get channels: %v", err))
				return
			}

			channelsMux.Lock()
			finalCount := len(channels)
			channelsMux.Unlock()
			
			statusLabel.SetText(fmt.Sprintf("✓ Retrieved %d channels", finalCount))
			channelList.Refresh()
		}()
	})

	// Export button (for future enhancement)
	exportBtn := widget.NewButton("Export Selected Channels", func() {
		if sess == nil {
			dialog.ShowError(fmt.Errorf("please authenticate and get channels first"), myWindow)
			return
		}

		selectedYear := yearSelect.Selected
		if selectedYear == "" {
			selectedYear = "2025"
		}

		// Parse year for date filtering
		year, err := strconv.Atoi(selectedYear)
		if err != nil {
			dialog.ShowError(fmt.Errorf("invalid year: %w", err), myWindow)
			return
		}
		oldest := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)

		channelsMux.Lock()
		channelCount := len(channels)
		channelsMux.Unlock()

		outputFolder := outputFolderEntry.Text
		if outputFolder == "" {
			outputFolder = "~/Documents/koberesearch/"
		}
		ignoreMedia := ignoreMediaCheck.Checked

		statusLabel.SetText(fmt.Sprintf("Export would save to: %s (ignore media: %v)", outputFolder, ignoreMedia))
		dialog.ShowInformation("Export", 
			fmt.Sprintf("This is a minimal MVP. Export would save %d channels from %s onwards to:\n%s\n\nIgnore Media: %v", 
			channelCount, oldest.Format("2006-01-02"), outputFolder, ignoreMedia), myWindow)
	})

	// Layout
	authForm := container.NewVBox(
		widget.NewLabel("Workspace Subdomain:"),
		subdomainEntry,
		widget.NewLabel("Cookie (d value):"),
		cookieEntry,
		widget.NewLabel("Output Folder:"),
		outputFolderEntry,
		widget.NewSeparator(),
		ignoreMediaCheck,
		container.NewHBox(speedLabel, speedSelect),
		container.NewHBox(yearLabel, yearSelect),
		widget.NewSeparator(),
		getChannelsBtn,
	)

	content := container.NewBorder(
		container.NewVBox(
			title,
			description,
			widget.NewSeparator(),
			authForm,
			widget.NewSeparator(),
			statusLabel,
		),
		exportBtn,
		nil,
		nil,
		container.NewScroll(channelList),
	)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}
