package main

import (
"context"
"encoding/json"
"fmt"
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

"github.com/joho/godotenv"
"github.com/rusq/slack"
"github.com/rusq/slackdump/v3"
"github.com/rusq/slackdump/v3/auth"
"github.com/rusq/slackdump/v3/internal/network"
)

type channelItem struct {
channel  slack.Channel
selected bool
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

wizardContent := container.NewStack()
step1 := createAuthStep(myWindow, &sess, &channelItems, &channelsMux, wizardContent)
wizardContent.Objects = []fyne.CanvasObject{step1}

myWindow.SetContent(wizardContent)
myWindow.ShowAndRun()
}

func createAuthStep(myWindow fyne.Window, sess **slackdump.Session, channelItems *[]*channelItem, channelsMux *sync.Mutex, wizardContent *fyne.Container) fyne.CanvasObject {
title := widget.NewLabel("Kobe User Research Slack Export Tool")
title.TextStyle = fyne.TextStyle{Bold: true}
title.Alignment = fyne.TextAlignCenter

description := widget.NewLabel("For Research Purposes Only - Step 1: Configuration")
description.TextStyle = fyne.TextStyle{Italic: true}
description.Alignment = fyne.TextAlignCenter

subdomainEntry := widget.NewEntry()
subdomainEntry.SetPlaceHolder("workspace subdomain")

cookieEntry := widget.NewEntry()
cookieEntry.SetPlaceHolder("d cookie value")

if err := godotenv.Load(".env"); err == nil {
if subdomain := os.Getenv("SLACK_SUBDOMAIN"); subdomain != "" {
subdomainEntry.SetText(subdomain)
}
if cookie := os.Getenv("SLACK_COOKIE"); cookie != "" {
cookieEntry.SetText(cookie)
}
}

outputFolderEntry := widget.NewEntry()
outputFolderEntry.SetText("~/Documents/koberesearch/")

ignoreMediaCheck := widget.NewCheck("Ignore Media", nil)
ignoreMediaCheck.Checked = true

speedSelect := widget.NewSelect([]string{"Default", "Fast", "Maximum"}, nil)
speedSelect.SetSelected("Fast")

yearSelect := widget.NewSelect([]string{"2025", "2024", "2023", "2022", "2021"}, nil)
yearSelect.SetSelected("2025")

statusLabel := widget.NewLabel("")
statusLabel.Wrapping = fyne.TextWrapWord

progressBar := widget.NewProgressBarInfinite()
progressBar.Hide()

var nextBtn *widget.Button
nextBtn = widget.NewButton("Next: Select Channels", func() {
subdomain := strings.TrimSpace(subdomainEntry.Text)
cookie := strings.TrimSpace(cookieEntry.Text)

if subdomain == "" || cookie == "" {
dialog.ShowError(fmt.Errorf("subdomain and cookie required"), myWindow)
return
}

cwd, _ := os.Getwd()
envPath := filepath.Join(cwd, ".env")
content := fmt.Sprintf("SLACK_SUBDOMAIN=%s\nSLACK_COOKIE=%s\n", subdomain, cookie)
os.WriteFile(envPath, []byte(content), 0600)

nextBtn.Disable()
progressBar.Show()
statusLabel.SetText("Authenticating...")

go func() {
ctx := context.Background()
authProvider, err := auth.NewCookieOnlyAuth(ctx, subdomain, cookie)
if err != nil {
statusLabel.SetText(fmt.Sprintf("Auth failed: %v", err))
nextBtn.Enable()
progressBar.Hide()
return
}

_, err = authProvider.Test(ctx)
if err != nil {
statusLabel.SetText(fmt.Sprintf("Auth test failed: %v", err))
nextBtn.Enable()
progressBar.Hide()
return
}

var limits network.Limits
switch speedSelect.Selected {
case "Maximum":
limits = network.NoLimits
case "Fast":
limits = network.DefLimits
limits.Tier2.Boost = 60
limits.Tier2.Burst = 10
default:
limits = network.DefLimits
}

*sess, err = slackdump.New(ctx, authProvider, slackdump.WithLimits(limits))
if err != nil {
statusLabel.SetText(fmt.Sprintf("Session failed: %v", err))
nextBtn.Enable()
progressBar.Hide()
return
}

statusLabel.SetText("Fetching channels...")

// Update count periodically while fetching
ticker := time.NewTicker(500 * time.Millisecond)
done := make(chan bool)

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

dbPathLabel := widget.NewLabel(fmt.Sprintf("Output: %s", exportFolder))
dbPathLabel.Wrapping = fyne.TextWrapWord

// Start export in goroutine
go func() {
ctx := context.Background()
total := len(selectedChannels)
exported := 0
skipped := 0
failed := 0

for i, ch := range selectedChannels {
progress := float64(i) / float64(total)
currentChannel := fmt.Sprintf("Scraping: %s (%s)", ch.Name, ch.ID)
status := fmt.Sprintf("Channel %d/%d", i+1, total)

exportProgressBar.SetValue(progress)
currentChannelLabel.SetText(currentChannel)
statusLabel.SetText(status)

// Dump channel
conv, err := (*sess).Dump(ctx, ch.ID, oldest, time.Now())
if err != nil {
statusLabel.SetText(fmt.Sprintf("Error dumping %s: %v", ch.Name, err))
failed++
time.Sleep(2 * time.Second)
continue
}

// Skip empty channels
if len(conv.Messages) == 0 {
statusLabel.SetText(fmt.Sprintf("Skipping %s: no messages found", ch.Name))
skipped++
time.Sleep(1 * time.Second)
continue
}

// Save to JSON file
filename := filepath.Join(exportFolder, fmt.Sprintf("%s_%s.json", ch.Name, ch.ID))
data, err := json.MarshalIndent(conv, "", "  ")
if err != nil {
statusLabel.SetText(fmt.Sprintf("Error marshaling %s: %v", ch.Name, err))
failed++
time.Sleep(2 * time.Second)
continue
}

if err := os.WriteFile(filename, data, 0644); err != nil {
statusLabel.SetText(fmt.Sprintf("Error writing %s: %v", ch.Name, err))
failed++
time.Sleep(2 * time.Second)
continue
}

// Success!
statusLabel.SetText(fmt.Sprintf("✓ Saved %s: %d messages", ch.Name, len(conv.Messages)))
exported++
time.Sleep(500 * time.Millisecond)
}

exportProgressBar.SetValue(1.0)
currentChannelLabel.SetText("✓ Export Complete!")
statusLabel.SetText(fmt.Sprintf("✓ Complete: %d exported, %d skipped (empty), %d failed\nOutput: %s", exported, skipped, failed, exportFolder))
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
}
