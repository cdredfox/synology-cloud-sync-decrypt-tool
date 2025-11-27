package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/synology-cloud-sync-decrypt-tool/syndecrypt-go/pkg/core"
	"github.com/synology-cloud-sync-decrypt-tool/syndecrypt-go/pkg/files"
)

const version = "1.0.0"
const appTitle = "Synology Cloud Sync Decrypt Tool"

type guiApp struct {
	app       fyne.App
	window    fyne.Window
	password  *widget.Entry
	inputPath *widget.Entry
	outputDir *widget.Entry
	recursive *widget.Check
	logView   *widget.List
	decryptBtn *widget.Button
	progress  *widget.ProgressBar
	status    *widget.Label

	logEntries []LogEntry
	isRunning  bool
}

type LogEntry struct {
	Time    string
	Message string
	Type    string // success, error, info, progress
}

func main() {
	gui := &guiApp{}
	gui.app = app.New()
	gui.window = gui.app.NewWindow(appTitle)
	gui.window.Resize(fyne.NewSize(800, 600))
	gui.window.SetMaster()

	gui.logEntries = make([]LogEntry, 0)

	gui.createUI()
	gui.window.ShowAndRun()
}

func (g *guiApp) createUI() {
	// Password input
	g.password = widget.NewPasswordEntry()
	g.password.SetPlaceHolder("Enter decryption password")

	// Input path (file or directory)
	g.inputPath = widget.NewEntry()
	g.inputPath.SetPlaceHolder("Select encrypted files or directory")

	inputBrowseBtn := widget.NewButtonWithIcon("Browse", theme.FolderOpenIcon(), func() {
		g.browseInput()
	})

	inputEntryContainer := container.NewBorder(nil, nil, nil, inputBrowseBtn, g.inputPath)

	// Output directory
	g.outputDir = widget.NewEntry()
	g.outputDir.SetText("output")

	outputBrowseBtn := widget.NewButtonWithIcon("Browse", theme.FolderOpenIcon(), func() {
		g.browseOutput()
	})

	outputEntryContainer := container.NewBorder(nil, nil, nil, outputBrowseBtn, g.outputDir)

	// Options
	g.recursive = widget.NewCheck("Process subdirectories recursively", nil)

	// Progress bar
	g.progress = widget.NewProgressBar()
	g.progress.Hide()

	// Status label
	g.status = widget.NewLabel("")
	g.status.Hide()

	// Log view
	g.logView = widget.NewList(
		func() int {
			return len(g.logEntries)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			label := o.(*widget.Label)
			entry := g.logEntries[i]
			label.SetText(fmt.Sprintf("[%s] %s", entry.Time, entry.Message))
			label.Wrapping = fyne.TextWrapWord

			// Set color based on type
			switch entry.Type {
			case "success":
				label.TextStyle = fyne.TextStyle{Bold: true}
				label.Importance = widget.SuccessImportance
			case "error":
				label.TextStyle = fyne.TextStyle{Bold: true}
				label.Importance = widget.DangerImportance
			case "progress":
				label.Importance = widget.WarningImportance
			default:
				label.Importance = widget.MediumImportance
			}
		},
	)

	scrollLog := container.NewScroll(g.logView)
	scrollLog.SetMinSize(fyne.NewSize(0, 200))

	// Decrypt button
	g.decryptBtn = widget.NewButtonWithIcon("Start Decryption", theme.ConfirmIcon(), g.startDecryption)
	g.decryptBtn.Importance = widget.HighImportance

	// Clear log button
	clearBtn := widget.NewButtonWithIcon("Clear Log", theme.ContentClearIcon(), g.clearLog)

	// Main form
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Password", Widget: g.password},
			{Text: "Input (File/Dir)", Widget: inputEntryContainer},
			{Text: "Output Directory", Widget: outputEntryContainer},
		},
	}

	// Layout
	content := container.NewBorder(
		nil, // top
		container.NewVBox(
			g.progress,
			g.status,
			container.NewGridWithColumns(2,
				g.decryptBtn,
				clearBtn,
			),
		), // bottom
		nil, // left
		nil, // right
		container.NewVBox(
			widget.NewLabelWithStyle(appTitle, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewSeparator(),
			form,
			g.recursive,
			scrollLog,
		), // center
	)

	g.window.SetContent(content)
}

func (g *guiApp) browseInput() {
	choice := widget.NewSelect([]string{"Single File", "Directory"}, func(s string) {})
	choice.SetSelected("Single File")

	dialog.ShowCustom("Select Input Type", "OK",
		container.NewVBox(
			widget.NewLabel("What would you like to decrypt?"),
			choice,
		),
		g.window,
	)

	if choice.Selected == "Directory" {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			g.inputPath.SetText(uri.Path())
			g.recursive.SetChecked(true)
		}, g.window)
	} else {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			g.inputPath.SetText(reader.URI().Path())
			g.recursive.SetChecked(false)
		}, g.window)
	}
}

func (g *guiApp) browseOutput() {
	dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil || uri == nil {
			return
		}
		g.outputDir.SetText(uri.Path())
	}, g.window)
}

func (g *guiApp) startDecryption() {
	if g.isRunning {
		return
	}

	// Validate inputs
	password := g.password.Text
	if password == "" {
		dialog.ShowError(fmt.Errorf("please enter password"), g.window)
		return
	}

	inputPath := g.inputPath.Text
	if inputPath == "" {
		dialog.ShowError(fmt.Errorf("please select input file or directory"), g.window)
		return
	}

	outputDir := g.outputDir.Text
	if outputDir == "" {
		dialog.ShowError(fmt.Errorf("please select output directory"), g.window)
		return
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		dialog.ShowError(fmt.Errorf("failed to create output directory: %v", err), g.window)
		return
	}

	// Disable UI during decryption
	g.setUIEnabled(false)
	g.isRunning = true
	g.progress.Show()
	g.status.Show()
	g.clearLog()

	// Add initial log
	g.addLogEntry(fmt.Sprintf("Starting decryption: %s", inputPath), "info")
	g.addLogEntry(fmt.Sprintf("Output directory: %s", outputDir), "info")

	// Run decryption in background
	go func() {
		defer func() {
			g.isRunning = false
			g.setUIEnabled(true)
		}()

		config := core.DecryptConfig{
			Password: []byte(password),
		}

		// Process input
		info, err := os.Stat(inputPath)
		if err != nil {
			g.addLogEntry(fmt.Sprintf("❌ Error accessing path: %v", err), "error")
			return
		}

		startTime := time.Now()

		if info.IsDir() {
			// Process directory
			g.addLogEntry("Processing directory recursively...", "info")
			results, err := files.DecryptDirectory(inputPath, outputDir, config)

			if err != nil {
				g.addLogEntry(fmt.Sprintf("❌ Directory decryption failed: %v", err), "error")
				return
			}

			elapsed := time.Now().Sub(startTime)
			g.addLogEntry(fmt.Sprintf("✅ Decryption complete!"), "success")
			g.addLogEntry(fmt.Sprintf("Total files: %d, Successful: %d, Failed: %d, Time: %v",
				results.TotalFiles, results.SuccessCount, results.FailedCount, elapsed), "info")
		} else {
			// Process single file
			outputFile := g.generateOutputFileName(inputPath, outputDir)
			g.addLogEntry(fmt.Sprintf("Decrypting: %s", filepath.Base(inputPath)), "info")

			err := files.DecryptFile(inputPath, outputFile, config)
			if err != nil {
				g.addLogEntry(fmt.Sprintf("❌ Failed: %v", err), "error")
				return
			}

			elapsed := time.Now().Sub(startTime)
			g.addLogEntry(fmt.Sprintf("✅ Successfully decrypted: %s", filepath.Base(inputPath)), "success")
			g.addLogEntry(fmt.Sprintf("Output saved to: %s (Time: %v)", outputFile, elapsed), "info")
		}
	}()
}

func (g *guiApp) generateOutputFileName(inputFile, outputDir string) string {
	baseName := filepath.Base(inputFile)
	ext := filepath.Ext(baseName)

	// Remove encrypted extension if present
	encryptedExts := []string{".cse", ".enc", ".cloudsync", ".csenc"}
	extLower := strings.ToLower(ext)
	for _, encryptedExt := range encryptedExts {
		if extLower == encryptedExt {
			baseName = baseName[:len(baseName)-len(ext)]
			break
		}
	}

	return filepath.Join(outputDir, baseName)
}

func (g *guiApp) clearLog() {
	g.logEntries = make([]LogEntry, 0)
	g.logView.Refresh()
}

func (g *guiApp) addLogEntry(message, entryType string) {
	entry := LogEntry{
		Time:    time.Now().Format("15:04:05"),
		Message: message,
		Type:    entryType,
	}

	g.logEntries = append(g.logEntries, entry)
	g.logView.Refresh()

	// Scroll to bottom
	g.logView.ScrollToBottom()
}

func (g *guiApp) setUIEnabled(enabled bool) {
	g.password.Disable()
	g.inputPath.Disable()
	g.outputDir.Disable()
	g.recursive.Disable()

	if enabled {
		g.decryptBtn.Enable()
		g.decryptBtn.SetText("Start Decryption")
		g.progress.Hide()
		g.status.Hide()

	}
	g.password.Enable()
	g.inputPath.Enable()
	g.outputDir.Enable()
	g.recursive.Enable()
}
