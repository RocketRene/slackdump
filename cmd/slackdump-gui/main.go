package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/rusq/slackdump/v3"
	"github.com/rusq/slackdump/v3/auth"
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
	cookieEntry.Password = true

	// Date selection - default to 2025
	yearLabel := widget.NewLabel("Export Year:")
	yearSelect := widget.NewSelect([]string{"2025", "2024", "2023", "2022", "2021"}, nil)
	yearSelect.SetSelected("2025")

	// Status label
	statusLabel := widget.NewLabel("")
	statusLabel.Wrapping = fyne.TextWrapWord

	// Channels list
	channelList := widget.NewList(
		func() int { return 0 },
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
		},
	)

	var channels types.Channels
	var sess *slackdump.Session

	// Get Channels button
	getChannelsBtn := widget.NewButton("Get Full List of Channels", func() {
		subdomain := strings.TrimSpace(subdomainEntry.Text)
		cookie := strings.TrimSpace(cookieEntry.Text)

		if subdomain == "" || cookie == "" {
			dialog.ShowError(fmt.Errorf("subdomain and cookie are required"), myWindow)
			return
		}

		statusLabel.SetText("Authenticating...")
		myWindow.Canvas().Refresh(statusLabel)

		// Authenticate
		ctx := context.Background()
		authProvider, err := auth.NewCookieOnlyAuth(ctx, subdomain, cookie)
		if err != nil {
			statusLabel.SetText(fmt.Sprintf("Authentication failed: %v", err))
			dialog.ShowError(err, myWindow)
			return
		}

		// Test authentication
		_, err = authProvider.Test(ctx)
		if err != nil {
			statusLabel.SetText(fmt.Sprintf("Authentication test failed: %v", err))
			dialog.ShowError(err, myWindow)
			return
		}

		statusLabel.SetText("Authenticated! Getting channels...")
		myWindow.Canvas().Refresh(statusLabel)

		// Create session
		sess, err = slackdump.New(ctx, authProvider)
		if err != nil {
			statusLabel.SetText(fmt.Sprintf("Failed to create session: %v", err))
			dialog.ShowError(err, myWindow)
			return
		}

		// Get channels
		channels, err = sess.GetChannels(ctx, slackdump.AllChanTypes...)
		if err != nil {
			statusLabel.SetText(fmt.Sprintf("Failed to get channels: %v", err))
			dialog.ShowError(err, myWindow)
			return
		}

		statusLabel.SetText(fmt.Sprintf("Retrieved %d channels", len(channels)))

		// Update the list
		channelList.Length = func() int { return len(channels) }
		channelList.CreateItem = func() fyne.CanvasObject {
			return widget.NewLabel("")
		}
		channelList.UpdateItem = func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(channels) {
				label := obj.(*widget.Label)
				ch := channels[id]
				label.SetText(fmt.Sprintf("%s (%s)", ch.Name, ch.ID))
			}
		}
		channelList.Refresh()
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

		statusLabel.SetText(fmt.Sprintf("Export functionality would export messages from %s onwards", oldest.Format("2006-01-02")))
		dialog.ShowInformation("Export", 
			fmt.Sprintf("This is a minimal MVP. Export functionality would export %d channels from %s onwards.", 
			len(channels), oldest.Format("2006-01-02")), myWindow)
	})

	// Layout
	authForm := container.NewVBox(
		widget.NewLabel("Workspace Subdomain:"),
		subdomainEntry,
		widget.NewLabel("Cookie (d value):"),
		cookieEntry,
		container.NewHBox(yearLabel, yearSelect),
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
