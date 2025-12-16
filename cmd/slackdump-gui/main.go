package main

import (
<<<<<<< HEAD
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
||||||| parent of f50e4866 (add shell.nix)
"context"
"fmt"
"log/slog"
"os"
"os/exec"
"path/filepath"
"runtime"
"strconv"
"strings"
"sync"
"time"
=======
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
>>>>>>> f50e4866 (add shell.nix)

<<<<<<< HEAD
"fyne.io/fyne/v2"
"fyne.io/fyne/v2/app"
"fyne.io/fyne/v2/container"
"fyne.io/fyne/v2/dialog"
"fyne.io/fyne/v2/widget"
"github.com/jmoiron/sqlx"
"github.com/pkg/browser"
||||||| parent of f50e4866 (add shell.nix)
"fyne.io/fyne/v2"
"fyne.io/fyne/v2/app"
"fyne.io/fyne/v2/container"
"fyne.io/fyne/v2/dialog"
"fyne.io/fyne/v2/widget"
"github.com/jmoiron/sqlx"
=======
	_ "modernc.org/sqlite"
>>>>>>> f50e4866 (add shell.nix)

<<<<<<< HEAD
"github.com/joho/godotenv"
"github.com/rusq/fsadapter"
"github.com/rusq/slack"
"github.com/rusq/slackdump/v3"
"github.com/rusq/slackdump/v3/auth"
"github.com/rusq/slackdump/v3/internal/chunk/backend/dbase"
"github.com/rusq/slackdump/v3/internal/chunk/backend/dbase/repository"
"github.com/rusq/slackdump/v3/internal/chunk/control"
"github.com/rusq/slackdump/v3/internal/convert/transform/fileproc"
"github.com/rusq/slackdump/v3/internal/network"
"github.com/rusq/slackdump/v3/internal/structures"
"github.com/rusq/slackdump/v3/internal/viewer"
"github.com/rusq/slackdump/v3/source"
"github.com/rusq/slackdump/v3/stream"
)
||||||| parent of f50e4866 (add shell.nix)
"github.com/joho/godotenv"
"github.com/rusq/fsadapter"
"github.com/rusq/slack"
"github.com/rusq/slackdump/v3"
"github.com/rusq/slackdump/v3/auth"
"github.com/rusq/slackdump/v3/internal/chunk/backend/dbase"
"github.com/rusq/slackdump/v3/internal/chunk/backend/dbase/repository"
"github.com/rusq/slackdump/v3/internal/chunk/control"
"github.com/rusq/slackdump/v3/internal/convert/transform/fileproc"
"github.com/rusq/slackdump/v3/internal/network"
"github.com/rusq/slackdump/v3/internal/structures"
"github.com/rusq/slackdump/v3/source"
"github.com/rusq/slackdump/v3/stream"
)
=======
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/jmoiron/sqlx"
>>>>>>> f50e4866 (add shell.nix)

	"github.com/rusq/fsadapter"
	"github.com/rusq/slack"
	"github.com/rusq/slackdump/v3"
	"github.com/rusq/slackdump/v3/auth"
	"github.com/rusq/slackdump/v3/internal/chunk/backend/dbase"
	"github.com/rusq/slackdump/v3/internal/chunk/backend/dbase/repository"
	"github.com/rusq/slackdump/v3/internal/chunk/control"
	"github.com/rusq/slackdump/v3/internal/convert/transform/fileproc"
	"github.com/rusq/slackdump/v3/internal/network"
	"github.com/rusq/slackdump/v3/internal/structures"
	"github.com/rusq/slackdump/v3/source"
	"github.com/rusq/slackdump/v3/stream"
	)

	return container.NewBorder(
		container.NewVBox(title, description, widget.NewSeparator()),
		nextBtn, nil, nil,
		container.NewScroll(form),
	)
}

func filterChannels(items []*channelItem, query string) []*channelItem {
	if query == "" {
		return append([]*channelItem(nil), items...)
	}
	var res []*channelItem
	q := strings.ToLower(query)
	for _, item := range items {
		if strings.Contains(strings.ToLower(item.channel.Name), q) {
			res = append(res, item)
		}
	}
	return res
}

func whatDo() (choice, error) {
	return choiceWizard, nil
}

func createExportStep(myWindow fyne.Window, sess **slackdump.Session, selectedChannels []slack.Channel,
	outputPath string, oldest time.Time, wizardContent *fyne.Container, channelItems *[]*channelItem, channelsMux *sync.Mutex) fyne.CanvasObject {

	title := widget.NewLabel("Step 3: Exporting Channels")
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

		// Create channel IDs list for EntityList
		var channelIDs []string
		for _, ch := range selectedChannels {
			channelIDs = append(channelIDs, ch.ID)
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

		statusLabel.SetText(fmt.Sprintf("✓ Export completed successfully!\nMessages: %d\nUsers: %d\nChannels: %s\nDatabase: %s", msgCount, userCount, channelNames, dbFile))
	}()

	backBtn := widget.NewButton("← Back to Selection", func() {
		step2 := createChannelSelectionStep(myWindow, sess, channelItems, channelsMux,
			outputPath, wizardContent)
		wizardContent.Objects = []fyne.CanvasObject{step2}
		wizardContent.Refresh()
	})

	openFolderBtn := widget.NewButton("📁 Open Export Folder", func() {
		if err := openFileExplorer(exportFolder); err != nil {
			dialog.ShowError(fmt.Errorf("failed to open folder: %v", err), myWindow)
		}
	})

<<<<<<< HEAD
go func() {
for {
select {
case <-done:
ticker.Stop()
return
case <-ticker.C:
channelsMux.Lock()
count := len(*channelItems)
channelsMux.Unlock()
if count > 0 {
statusLabel.SetText(fmt.Sprintf("Fetching channels... (%d found)", count))
}
}
}
}()

err = (*sess).StreamChannels(ctx, slackdump.AllChanTypes, func(ch slack.Channel) error {
channelsMux.Lock()
*channelItems = append(*channelItems, &channelItem{channel: ch, selected: false})
channelsMux.Unlock()
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
finalCount := len(*channelItems)
channelsMux.Unlock()

statusLabel.SetText(fmt.Sprintf("✓ %d channels - Click Next", finalCount))
progressBar.Hide()

step2 := createChannelSelectionStep(myWindow, sess, channelItems, channelsMux, 
outputFolderEntry.Text, yearSelect.Selected, ignoreMediaCheck.Checked, wizardContent)
wizardContent.Objects = []fyne.CanvasObject{step2}
wizardContent.Refresh()
}()
})

form := container.NewVBox(
widget.NewLabel("Subdomain:"), subdomainEntry,
widget.NewLabel("Cookie:"), cookieEntry,
widget.NewLabel("Output:"), outputFolderEntry,
widget.NewSeparator(),
ignoreMediaCheck,
container.NewHBox(widget.NewLabel("Speed:"), speedSelect),
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
channelsMux *sync.Mutex, outputFolder, selectedYear string, ignoreMedia bool, wizardContent *fyne.Container) fyne.CanvasObject {

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
step1 := createAuthStep(myWindow, sess, channelItems, channelsMux, wizardContent)
wizardContent.Objects = []fyne.CanvasObject{step1}
wizardContent.Refresh()
})

var exportBtn *widget.Button
exportBtn = widget.NewButton("Export Selected", func() {
channelsMux.Lock()
var selectedChannels []slack.Channel
for _, item := range *channelItems {
if item.selected {
selectedChannels = append(selectedChannels, item.channel)
}
}
channelsMux.Unlock()

if len(selectedChannels) == 0 {
dialog.ShowError(fmt.Errorf("select at least one channel"), myWindow)
return
}

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

step3 := createExportStep(myWindow, sess, selectedChannels, outputPath, oldest, wizardContent, channelItems, channelsMux)
wizardContent.Objects = []fyne.CanvasObject{step3}
wizardContent.Refresh()
})

controls := container.NewHBox(selectAllBtn, deselectAllBtn)

return container.NewBorder(
container.NewVBox(title, widget.NewSeparator(), statusLabel, widget.NewSeparator(), controls),
container.NewHBox(backBtn, exportBtn),
nil, nil,
container.NewScroll(channelList),
)
}

func createExportStep(myWindow fyne.Window, sess **slackdump.Session, selectedChannels []slack.Channel, 
outputPath string, oldest time.Time, wizardContent *fyne.Container, channelItems *[]*channelItem, channelsMux *sync.Mutex) fyne.CanvasObject {

title := widget.NewLabel("Step 3: Exporting Channels")
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

// Create channel IDs list for EntityList
var channelIDs []string
for _, ch := range selectedChannels {
channelIDs = append(channelIDs, ch.ID)
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
statusLabel.SetText(fmt.Sprintf("✓ Export completed successfully!\nDatabase: %s", dbFile))

// Start the viewer automatically
statusLabel.SetText("Loading viewer...")
if err := startViewer(ctx, dbFile, statusLabel); err != nil {
statusLabel.SetText(fmt.Sprintf("Export complete, but viewer failed to start: %v\nDatabase: %s", err, dbFile))
slog.ErrorContext(ctx, "failed to start viewer", "error", err)
} else {
statusLabel.SetText(fmt.Sprintf("✓ Export complete! Viewer is running.\nDatabase: %s", dbFile))
}
}()

backBtn := widget.NewButton("← Back to Selection", func() {
step2 := createChannelSelectionStep(myWindow, sess, channelItems, channelsMux,
outputPath, "2025", true, wizardContent)
wizardContent.Objects = []fyne.CanvasObject{step2}
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
||||||| parent of f50e4866 (add shell.nix)
go func() {
for {
select {
case <-done:
ticker.Stop()
return
case <-ticker.C:
channelsMux.Lock()
count := len(*channelItems)
channelsMux.Unlock()
if count > 0 {
statusLabel.SetText(fmt.Sprintf("Fetching channels... (%d found)", count))
}
}
}
}()

err = (*sess).StreamChannels(ctx, slackdump.AllChanTypes, func(ch slack.Channel) error {
channelsMux.Lock()
*channelItems = append(*channelItems, &channelItem{channel: ch, selected: false})
channelsMux.Unlock()
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
finalCount := len(*channelItems)
channelsMux.Unlock()

statusLabel.SetText(fmt.Sprintf("✓ %d channels - Click Next", finalCount))
progressBar.Hide()

step2 := createChannelSelectionStep(myWindow, sess, channelItems, channelsMux, 
outputFolderEntry.Text, yearSelect.Selected, ignoreMediaCheck.Checked, wizardContent)
wizardContent.Objects = []fyne.CanvasObject{step2}
wizardContent.Refresh()
}()
})

form := container.NewVBox(
widget.NewLabel("Subdomain:"), subdomainEntry,
widget.NewLabel("Cookie:"), cookieEntry,
widget.NewLabel("Output:"), outputFolderEntry,
widget.NewSeparator(),
ignoreMediaCheck,
container.NewHBox(widget.NewLabel("Speed:"), speedSelect),
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
channelsMux *sync.Mutex, outputFolder, selectedYear string, ignoreMedia bool, wizardContent *fyne.Container) fyne.CanvasObject {

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
step1 := createAuthStep(myWindow, sess, channelItems, channelsMux, wizardContent)
wizardContent.Objects = []fyne.CanvasObject{step1}
wizardContent.Refresh()
})

var exportBtn *widget.Button
exportBtn = widget.NewButton("Export Selected", func() {
channelsMux.Lock()
var selectedChannels []slack.Channel
for _, item := range *channelItems {
if item.selected {
selectedChannels = append(selectedChannels, item.channel)
}
}
channelsMux.Unlock()

if len(selectedChannels) == 0 {
dialog.ShowError(fmt.Errorf("select at least one channel"), myWindow)
return
}

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

step3 := createExportStep(myWindow, sess, selectedChannels, outputPath, oldest, wizardContent, channelItems, channelsMux)
wizardContent.Objects = []fyne.CanvasObject{step3}
wizardContent.Refresh()
})

controls := container.NewHBox(selectAllBtn, deselectAllBtn)

return container.NewBorder(
container.NewVBox(title, widget.NewSeparator(), statusLabel, widget.NewSeparator(), controls),
container.NewHBox(backBtn, exportBtn),
nil, nil,
container.NewScroll(channelList),
)
}

func createExportStep(myWindow fyne.Window, sess **slackdump.Session, selectedChannels []slack.Channel, 
outputPath string, oldest time.Time, wizardContent *fyne.Container, channelItems *[]*channelItem, channelsMux *sync.Mutex) fyne.CanvasObject {

title := widget.NewLabel("Step 3: Exporting Channels")
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

// Create channel IDs list for EntityList
var channelIDs []string
for _, ch := range selectedChannels {
channelIDs = append(channelIDs, ch.ID)
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
statusLabel.SetText(fmt.Sprintf("✓ Export completed successfully!\nDatabase: %s", dbFile))
}()

backBtn := widget.NewButton("← Back to Selection", func() {
step2 := createChannelSelectionStep(myWindow, sess, channelItems, channelsMux,
outputPath, "2025", true, wizardContent)
wizardContent.Objects = []fyne.CanvasObject{step2}
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
=======
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
>>>>>>> f50e4866 (add shell.nix)
}

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
