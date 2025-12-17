package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/browser"
	_ "modernc.org/sqlite"

	"github.com/joho/godotenv"
	"github.com/rusq/fsadapter"
	"github.com/rusq/slack"
	"github.com/rusq/slackdump/v3"
	"github.com/rusq/slackdump/v3/auth"
	"github.com/rusq/slackdump/v3/internal/chunk/backend/dbase"
	"github.com/rusq/slackdump/v3/internal/chunk/backend/dbase/repository"
	"github.com/rusq/slackdump/v3/internal/chunk/control"
	"github.com/rusq/slackdump/v3/internal/convert/transform/fileproc"
	"github.com/rusq/slackdump/v3/internal/structures"
	"github.com/rusq/slackdump/v3/internal/viewer"
	"github.com/rusq/slackdump/v3/source"
	"github.com/rusq/slackdump/v3/stream"
)

type channelItem struct {
	channel  slack.Channel
	selected bool
}

type conversationItem struct {
	conversation slack.Channel
	displayName  string
	selected     bool
}

// openFileExplorer opens the native file explorer at the given path
func openFileExplorer(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", path)
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("explorer", path)
	default:
		return fmt.Errorf("unsupported platform")
	}
	return cmd.Start()
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Kobe User Research Slack Export Tool")
	myWindow.Resize(fyne.NewSize(900, 700))

	var sess *slackdump.Session
	var channelItems []*channelItem
	var channelsMux sync.Mutex
	var conversationItems []*conversationItem
	var conversationsMux sync.Mutex

	wizardContent := container.NewStack()
	step1 := createAuthStep(myWindow, &sess, &channelItems, &channelsMux, &conversationItems, &conversationsMux, wizardContent)
	wizardContent.Objects = []fyne.CanvasObject{step1}

	myWindow.SetContent(wizardContent)
	myWindow.ShowAndRun()
}

func createAuthStep(myWindow fyne.Window, sess **slackdump.Session, channelItems *[]*channelItem, channelsMux *sync.Mutex, conversationItems *[]*conversationItem, conversationsMux *sync.Mutex, wizardContent *fyne.Container) fyne.CanvasObject {
	title := widget.NewLabel("Kobe User Research Slack Export Tool")
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	description := widget.NewLabel("Export Slack workspace channels to SQLite database for research analysis")
	description.Wrapping = fyne.TextWrapWord
	description.Alignment = fyne.TextAlignCenter

	statusLabel := widget.NewLabel("Enter credentials to begin...")
	statusLabel.Wrapping = fyne.TextWrapWord

	progressBar := widget.NewProgressBar()
	progressBar.Hide()

	godotenv.Load()

	subdomainEntry := widget.NewEntry()
	subdomainEntry.SetPlaceHolder("your-workspace")
	subdomainEntry.Text = os.Getenv("SLACK_SUBDOMAIN")

	cookieEntry := widget.NewEntry()
	cookieEntry.SetPlaceHolder("xoxd-...")
	cookieEntry.Text = os.Getenv("SLACK_COOKIE")

	outputFolderEntry := widget.NewEntry()
	outputFolderEntry.SetPlaceHolder("~/Documents/koberesearch/")
	outputFolderEntry.Text = os.Getenv("OUTPUT_FOLDER")
	if outputFolderEntry.Text == "" {
		outputFolderEntry.Text = "~/Documents/koberesearch/"
	}

	currentYear := time.Now().Year()
	years := []string{fmt.Sprintf("%d", currentYear)}
	for i := 1; i <= 10; i++ {
		years = append(years, fmt.Sprintf("%d", currentYear-i))
	}
	yearSelect := widget.NewSelect(years, nil)
	yearSelect.Selected = fmt.Sprintf("%d", currentYear)

	nextBtn := widget.NewButton("Authenticate & Fetch Channels →", nil)
	nextBtn.OnTapped = func() {
		if subdomainEntry.Text == "" || cookieEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("subdomain and cookie required"), myWindow)
			return
		}

		nextBtn.Disable()
		statusLabel.SetText("Authenticating...")
		progressBar.Show()
		progressBar.SetValue(0.3)

		go func() {
			ctx := context.Background()

			prov, err := auth.NewCookieOnlyAuth(ctx, subdomainEntry.Text, cookieEntry.Text)
			if err != nil {
				statusLabel.SetText(fmt.Sprintf("Auth failed: %v", err))
				nextBtn.Enable()
				progressBar.Hide()
				return
			}

			*sess, err = slackdump.New(ctx, prov)
			if err != nil {
				statusLabel.SetText(fmt.Sprintf("Session failed: %v", err))
				nextBtn.Enable()
				progressBar.Hide()
				return
			}

			statusLabel.SetText("Fetching channels...")
			progressBar.SetValue(0.6)

			done := make(chan bool, 1)
			ticker := time.NewTicker(500 * time.Millisecond)

			go func() {
				for {
					select {
					case <-done:
						ticker.Stop()
						return
					case <-ticker.C:
						channelsMux.Lock()
						chCount := len(*channelItems)
						channelsMux.Unlock()
						conversationsMux.Lock()
						convCount := len(*conversationItems)
						conversationsMux.Unlock()
						if chCount > 0 || convCount > 0 {
							statusLabel.SetText(fmt.Sprintf("Fetching... (%d channels, %d DMs)", chCount, convCount))
						}
					}
				}
			}()

			err = (*sess).StreamChannels(ctx, slackdump.AllChanTypes, func(ch slack.Channel) error {
				if ch.IsIM || ch.IsMpIM {
					// This is a DM or Group DM
					conversationsMux.Lock()
					displayName := ch.Name
					if ch.IsIM {
						displayName = fmt.Sprintf("DM: %s", ch.User)
					} else if ch.IsMpIM {
						displayName = fmt.Sprintf("Group DM: %s", ch.Purpose.Value)
					}
					*conversationItems = append(*conversationItems, &conversationItem{
						conversation: ch,
						displayName:  displayName,
						selected:     false,
					})
					conversationsMux.Unlock()
				} else {
					// This is a regular channel
					channelsMux.Lock()
					*channelItems = append(*channelItems, &channelItem{channel: ch, selected: false})
					channelsMux.Unlock()
				}
				return nil
			})

			done <- true

			if err != nil {
				statusLabel.SetText(fmt.Sprintf("Failed: %v", err))
				nextBtn.Enable()
				progressBar.Hide()
				return
			}

			channelsMux.Lock()
			channelCount := len(*channelItems)
			channelsMux.Unlock()

			conversationsMux.Lock()
			convCount := len(*conversationItems)
			conversationsMux.Unlock()

			statusLabel.SetText(fmt.Sprintf("✓ %d channels, %d DMs/Group DMs fetched", channelCount, convCount))
			progressBar.Hide()

			step2 := createChannelSelectionStep(myWindow, sess, channelItems, channelsMux, conversationItems, conversationsMux,
				outputFolderEntry.Text, yearSelect.Selected, wizardContent)
			wizardContent.Objects = []fyne.CanvasObject{step2}
			wizardContent.Refresh()
		}()
	}

	form := container.NewVBox(
		widget.NewLabel("Subdomain:"), subdomainEntry,
		widget.NewLabel("Cookie:"), cookieEntry,
		widget.NewLabel("Output:"), outputFolderEntry,
		widget.NewSeparator(),
		container.NewHBox(widget.NewLabel("Year:"), yearSelect),
		widget.NewSeparator(),
		statusLabel, progressBar,
	)

	return container.NewBorder(
		container.NewVBox(title, description, widget.NewSeparator()),
		nextBtn, nil, nil,
		container.NewScroll(form),
	)
}

func createChannelSelectionStep(myWindow fyne.Window, sess **slackdump.Session, channelItems *[]*channelItem,
	channelsMux *sync.Mutex, conversationItems *[]*conversationItem, conversationsMux *sync.Mutex, outputFolder, selectedYear string, wizardContent *fyne.Container) fyne.CanvasObject {

	title := widget.NewLabel("Step 2: Select Channels")
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	statusLabel := widget.NewLabel("Select channels to export")
	statusLabel.Wrapping = fyne.TextWrapWord

	channelList := widget.NewList(
		func() int {
			channelsMux.Lock()
			defer channelsMux.Unlock()
			return len(*channelItems)
		},
		func() fyne.CanvasObject {
			return widget.NewCheck("", nil)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			channelsMux.Lock()
			defer channelsMux.Unlock()
			if id < len(*channelItems) {
				check := obj.(*widget.Check)
				item := (*channelItems)[id]
				check.Text = fmt.Sprintf("%s (%s)", item.channel.Name, item.channel.ID)
				check.Checked = item.selected
				check.OnChanged = func(checked bool) {
					channelsMux.Lock()
					item.selected = checked
					channelsMux.Unlock()
				}
				check.Refresh()
			}
		},
	)

	selectAllBtn := widget.NewButton("Select All", func() {
		channelsMux.Lock()
		for _, item := range *channelItems {
			item.selected = true
		}
		channelsMux.Unlock()
		channelList.Refresh()
	})

	deselectAllBtn := widget.NewButton("Deselect All", func() {
		channelsMux.Lock()
		for _, item := range *channelItems {
			item.selected = false
		}
		channelsMux.Unlock()
		channelList.Refresh()
	})

	backBtn := widget.NewButton("← Back", func() {
		step1 := createAuthStep(myWindow, sess, channelItems, channelsMux, conversationItems, conversationsMux, wizardContent)
		wizardContent.Objects = []fyne.CanvasObject{step1}
		wizardContent.Refresh()
	})

	nextBtn := widget.NewButton("Next: Select DMs →", func() {
		step3 := createConversationSelectionStep(myWindow, sess, channelItems, channelsMux, conversationItems, conversationsMux,
			outputFolder, selectedYear, wizardContent)
		wizardContent.Objects = []fyne.CanvasObject{step3}
		wizardContent.Refresh()
	})

	var exportBtn *widget.Button
	exportBtn = widget.NewButton("Skip to Export", func() {
		channelsMux.Lock()
		var selectedChannels []slack.Channel
		for _, item := range *channelItems {
			if item.selected {
				selectedChannels = append(selectedChannels, item.channel)
			}
		}
		channelsMux.Unlock()

		year, _ := strconv.Atoi(selectedYear)
		if year == 0 {
			year = 2025
		}
		oldest := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)

		outputPath := outputFolder
		if outputPath == "" {
			outputPath = "~/Documents/koberesearch/"
		}
		if strings.HasPrefix(outputPath, "~/") {
			home, _ := os.UserHomeDir()
			outputPath = filepath.Join(home, outputPath[2:])
		}

		if err := os.MkdirAll(outputPath, 0755); err != nil {
			dialog.ShowError(err, myWindow)
			return
		}

		step4 := createExportStep(myWindow, sess, selectedChannels, []slack.Channel{}, outputPath, oldest, wizardContent, channelItems, channelsMux, conversationItems, conversationsMux)
		wizardContent.Objects = []fyne.CanvasObject{step4}
		wizardContent.Refresh()
	})

	controls := container.NewHBox(selectAllBtn, deselectAllBtn)

	return container.NewBorder(
		container.NewVBox(title, widget.NewSeparator(), statusLabel, widget.NewSeparator(), controls),
		container.NewHBox(backBtn, exportBtn, nextBtn),
		nil, nil,
		container.NewScroll(channelList),
	)
}

func createConversationSelectionStep(myWindow fyne.Window, sess **slackdump.Session, channelItems *[]*channelItem,
	channelsMux *sync.Mutex, conversationItems *[]*conversationItem, conversationsMux *sync.Mutex, outputFolder, selectedYear string, wizardContent *fyne.Container) fyne.CanvasObject {

	title := widget.NewLabel("Step 3: Select DMs and Group DMs")
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	statusLabel := widget.NewLabel("Select DMs and Group DMs to export")
	statusLabel.Wrapping = fyne.TextWrapWord

	conversationList := widget.NewList(
		func() int {
			conversationsMux.Lock()
			defer conversationsMux.Unlock()
			return len(*conversationItems)
		},
		func() fyne.CanvasObject {
			return widget.NewCheck("", nil)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			conversationsMux.Lock()
			defer conversationsMux.Unlock()
			if id < len(*conversationItems) {
				check := obj.(*widget.Check)
				item := (*conversationItems)[id]
				check.Text = fmt.Sprintf("%s (%s)", item.displayName, item.conversation.ID)
				check.Checked = item.selected
				check.OnChanged = func(checked bool) {
					conversationsMux.Lock()
					item.selected = checked
					conversationsMux.Unlock()
				}
				check.Refresh()
			}
		},
	)

	selectAllBtn := widget.NewButton("Select All", func() {
		conversationsMux.Lock()
		for _, item := range *conversationItems {
			item.selected = true
		}
		conversationsMux.Unlock()
		conversationList.Refresh()
	})

	deselectAllBtn := widget.NewButton("Deselect All", func() {
		conversationsMux.Lock()
		for _, item := range *conversationItems {
			item.selected = false
		}
		conversationsMux.Unlock()
		conversationList.Refresh()
	})

	backBtn := widget.NewButton("← Back", func() {
		step2 := createChannelSelectionStep(myWindow, sess, channelItems, channelsMux, conversationItems, conversationsMux,
			outputFolder, selectedYear, wizardContent)
		wizardContent.Objects = []fyne.CanvasObject{step2}
		wizardContent.Refresh()
	})

	exportBtn := widget.NewButton("Export Selected →", func() {
		channelsMux.Lock()
		var selectedChannels []slack.Channel
		for _, item := range *channelItems {
			if item.selected {
				selectedChannels = append(selectedChannels, item.channel)
			}
		}
		channelsMux.Unlock()

		conversationsMux.Lock()
		var selectedConversations []slack.Channel
		for _, item := range *conversationItems {
			if item.selected {
				selectedConversations = append(selectedConversations, item.conversation)
			}
		}
		conversationsMux.Unlock()

		year, _ := strconv.Atoi(selectedYear)
		if year == 0 {
			year = 2025
		}
		oldest := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)

		outputPath := outputFolder
		if outputPath == "" {
			outputPath = "~/Documents/koberesearch/"
		}
		if strings.HasPrefix(outputPath, "~/") {
			home, _ := os.UserHomeDir()
			outputPath = filepath.Join(home, outputPath[2:])
		}

		if err := os.MkdirAll(outputPath, 0755); err != nil {
			dialog.ShowError(err, myWindow)
			return
		}

		step4 := createExportStep(myWindow, sess, selectedChannels, selectedConversations, outputPath, oldest, wizardContent, channelItems, channelsMux, conversationItems, conversationsMux)
		wizardContent.Objects = []fyne.CanvasObject{step4}
		wizardContent.Refresh()
	})

	controls := container.NewHBox(selectAllBtn, deselectAllBtn)

	return container.NewBorder(
		container.NewVBox(title, widget.NewSeparator(), statusLabel, widget.NewSeparator(), controls),
		container.NewHBox(backBtn, exportBtn),
		nil, nil,
		container.NewScroll(conversationList),
	)
}

func createExportStep(myWindow fyne.Window, sess **slackdump.Session, selectedChannels []slack.Channel, selectedConversations []slack.Channel,
	outputPath string, oldest time.Time, wizardContent *fyne.Container, channelItems *[]*channelItem, channelsMux *sync.Mutex, conversationItems *[]*conversationItem, conversationsMux *sync.Mutex) fyne.CanvasObject {

	title := widget.NewLabel("Step 4: Exporting Data")
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	currentChannelLabel := widget.NewLabel("")
	currentChannelLabel.TextStyle = fyne.TextStyle{Bold: true}
	currentChannelLabel.Wrapping = fyne.TextWrapWord

	statusLabel := widget.NewLabel("Starting export...")
	statusLabel.Wrapping = fyne.TextWrapWord

	exportProgressBar := widget.NewProgressBar()
	exportProgressBar.SetValue(0)

	exportFolder := filepath.Join(outputPath, fmt.Sprintf("export_%s", time.Now().Format("20060102_150405")))
	os.MkdirAll(exportFolder, 0755)

	dbFile := filepath.Join(exportFolder, source.DefaultDBFile)
	dbPathLabel := widget.NewLabel(fmt.Sprintf("Database: %s", dbFile))
	dbPathLabel.Wrapping = fyne.TextWrapWord

	// Start export in goroutine
	go func() {
		ctx := context.Background()

		// Create channel IDs list for EntityList (both channels and conversations)
		var channelIDs []string
		for _, ch := range selectedChannels {
			channelIDs = append(channelIDs, ch.ID)
		}
		for _, conv := range selectedConversations {
			channelIDs = append(channelIDs, conv.ID)
		}

		// Create entity list from channel IDs
		entityList, err := structures.NewEntityList(channelIDs)
		if err != nil {
			statusLabel.SetText(fmt.Sprintf("Error creating entity list: %v", err))
			dialog.ShowError(err, myWindow)
			return
		}

		statusLabel.SetText("Initializing database...")

		// Open SQLite database connection
		conn, err := sqlx.Open(repository.Driver, dbFile)
		if err != nil {
			statusLabel.SetText(fmt.Sprintf("Error opening database: %v", err))
			dialog.ShowError(err, myWindow)
			return
		}
		defer conn.Close()

		statusLabel.SetText("Creating database controller...")

		// Create database controller using the same pattern as archive command
		ctrl, err := createDBController(ctx, conn, *sess, exportFolder, oldest, time.Now())
		if err != nil {
			statusLabel.SetText(fmt.Sprintf("Error creating controller: %v", err))
			dialog.ShowError(err, myWindow)
			return
		}
		defer func() {
			if err := ctrl.Close(); err != nil {
				slog.ErrorContext(ctx, "unable to close database controller", "error", err)
			}
		}()

		statusLabel.SetText("Exporting channels to database...")
		currentChannelLabel.SetText("Streaming data to SQLite database...")

		// Run the controller with the entity list
		if err := ctrl.RunNoTransform(ctx, entityList); err != nil {
			statusLabel.SetText(fmt.Sprintf("Export failed: %v", err))
			dialog.ShowError(err, myWindow)
			return
		}

		exportProgressBar.SetValue(1.0)
		currentChannelLabel.SetText("✓ Export Complete!")

		// Query stats
		var msgCount, userCount int
		var channelNames string
		if err := conn.Get(&msgCount, "SELECT COUNT(*) FROM MESSAGE"); err != nil {
			slog.WarnContext(ctx, "failed to count messages", "error", err)
		}
		if err := conn.Get(&userCount, "SELECT COUNT(*) FROM S_USER"); err != nil {
			slog.WarnContext(ctx, "failed to count users", "error", err)
		}
		if err := conn.Get(&channelNames, "SELECT GROUP_CONCAT(NAME, ', ') FROM (SELECT DISTINCT NAME FROM CHANNEL ORDER BY NAME LIMIT 20)"); err != nil {
			slog.WarnContext(ctx, "failed to list channels", "error", err)
			channelNames = "N/A"
		}

		statusLabel.SetText(fmt.Sprintf("✓ Export completed successfully!\nChannels: %d\nDMs/Group DMs: %d\nMessages: %d\nUsers: %d\nSample Channels: %s\nDatabase: %s",
			len(selectedChannels), len(selectedConversations), msgCount, userCount, channelNames, dbFile))
	}()

	backBtn := widget.NewButton("← Back to Selection", func() {
		step3 := createConversationSelectionStep(myWindow, sess, channelItems, channelsMux, conversationItems, conversationsMux,
			outputPath, fmt.Sprintf("%d", oldest.Year()), wizardContent)
		wizardContent.Objects = []fyne.CanvasObject{step3}
		wizardContent.Refresh()
	})

	openFolderBtn := widget.NewButton("📁 Open Export Folder", func() {
		if err := openFileExplorer(exportFolder); err != nil {
			dialog.ShowError(fmt.Errorf("failed to open folder: %v", err), myWindow)
		}
	})

	return container.NewBorder(
		container.NewVBox(
			title,
			widget.NewSeparator(),
			currentChannelLabel,
			exportProgressBar,
			statusLabel,
			widget.NewSeparator(),
			dbPathLabel,
		),
		container.NewHBox(backBtn, openFolderBtn),
		nil, nil,
		widget.NewLabel(""),
	)
}

// createDBController creates a database controller similar to the archive command's DBController
// createDBController creates a database controller similar to the archive command's DBController
func createDBController(ctx context.Context, conn *sqlx.DB, sess *slackdump.Session, dirname string, oldest, latest time.Time) (*control.Controller, error) {
	lg := slog.Default()

	// Create session info
	sessionInfo := dbase.SessionInfo{
		FromTS:         &oldest,
		ToTS:           &latest,
		FilesEnabled:   false, // disable file download for GUI
		AvatarsEnabled: false, // disable avatar download for GUI
		Mode:           "gui-export",
		Args:           "Kobe User Research Export",
	}

	// Create database processor
	dbp, err := dbase.New(ctx, conn, sessionInfo)
	if err != nil {
		return nil, err
	}

	// Create stream options
	sopts := []stream.Option{
		stream.OptLatest(latest),
		stream.OptOldest(oldest),
		stream.OptFastSearch(),
		stream.OptResultFn(func(sr stream.Result) error {
			lg.Info("stream", "result", sr.String())
			return nil
		}),
	}

	// Use Session's Stream method to create a streamer with the correct client
	streamer := sess.Stream(sopts...)

	// Create stub downloaders (not used but required by controller)
	dl := fileproc.NewDownloader(
		ctx,
		false, // WithFiles disabled
		sess.Client(),
		fsadapter.NewDirectory(dirname),
		lg,
	)
	avdl := fileproc.NewDownloader(
		ctx,
		false, // WithAvatars disabled
		sess.Client(),
		fsadapter.NewDirectory(dirname),
		lg,
	)

	// Create controller
	ctrl, err := control.New(
		ctx,
		streamer,
		dbp,
		control.WithFiler(fileproc.New(dl)),
		control.WithAvatarProcessor(fileproc.NewAvatarProc(avdl)),
		control.WithFlags(control.Flags{MemberOnly: false, RecordFiles: false, ChannelUsers: false}),
	)
	if err != nil {
		return nil, err
	}

	return ctrl, nil
}

// startViewer starts the slackdump viewer for the given database file
func startViewer(ctx context.Context, dbFile string, statusLabel *widget.Label) error {
// Load the database source
src, err := source.OpenDatabase(ctx, dbFile)
if err != nil {
return fmt.Errorf("failed to open database: %w", err)
}

// Create viewer on localhost with a dynamic port
listenAddr := "127.0.0.1:8080"
statusLabel.SetText(fmt.Sprintf("Starting viewer on %s...", listenAddr))

v, err := viewer.New(ctx, listenAddr, src)
if err != nil {
return fmt.Errorf("failed to create viewer: %w", err)
}

// Create a channel to signal when the server is ready
ready := make(chan struct{})

// Start the viewer in a goroutine
// Note: The server will run for the lifetime of the application.
// It will be automatically cleaned up when the application exits.
go func() {
// Signal that we're about to start the server
close(ready)
if err := v.ListenAndServe(); err != nil && err != http.ErrServerClosed {
slog.ErrorContext(ctx, "viewer server error", "error", err)
}
}()

// Wait for the server to be ready
<-ready

// Give the server a brief moment to bind to the port
time.Sleep(100 * time.Millisecond)

// Open the browser (blocking call is fine)
viewerURL := fmt.Sprintf("http://%s", listenAddr)
if err := browser.OpenURL(viewerURL); err != nil {
slog.ErrorContext(ctx, "failed to open browser", "error", err)
} else {
slog.InfoContext(ctx, "opened browser", "url", viewerURL)
}

return nil
}
