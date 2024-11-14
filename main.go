package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "cabinete",
		Short: "Organize files into folders by creation day",
		Run:   runOrganizer,
	}

	rootCmd.Flags().StringP("dir", "d", "", "Directory to organize")
	rootCmd.MarkFlagRequired("dir")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runOrganizer(cmd *cobra.Command, args []string) {
	dir, _ := cmd.Flags().GetString("dir")

	// Set up the tview application
	app := tview.NewApplication()
	table := tview.NewTable().
		SetBorders(true).
		SetFixed(1, 1)

	table.SetCell(0, 0, tview.NewTableCell("Directory (Day)").
		SetTextColor(tcell.ColorYellow).
		SetAlign(tview.AlignCenter))
	table.SetCell(0, 1, tview.NewTableCell("File Count").
		SetTextColor(tcell.ColorYellow).
		SetAlign(tview.AlignCenter))

	statusText := tview.NewTextView().SetTextAlign(tview.AlignCenter)

	// Track processed and pending files
	var processedFiles int
	totalFiles := 0
	dirFileCounts := make(map[string]int)
	var mutex sync.Mutex

	// Pre-count total files for pending counter
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			totalFiles++
		}
		return nil
	})

	// Update statusText function
	updateStatus := func() {
		statusText.SetText(fmt.Sprintf("Processed: %d / %d", processedFiles, totalFiles))
	}

	// Run file organization in a goroutine
	done := make(chan struct{})
	var wg sync.WaitGroup

	go func() {
		defer close(done)
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			// Get file creation day (using modification time as a proxy)
			creationTime := info.ModTime()
			dayDir := fmt.Sprintf("%02d", creationTime.Day())
			targetDir := filepath.Join(dir, dayDir)

			// Create directory if it doesn't exist
			if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
				return err
			}

			// Move file
			targetPath := filepath.Join(targetDir, info.Name())
			if err := os.Rename(path, targetPath); err != nil {
				return err
			}

			// Update counters
			mutex.Lock()
			dirFileCounts[dayDir]++
			processedFiles++
			mutex.Unlock()

			// Queue a UI update
			wg.Add(1)
			app.QueueUpdateDraw(func() {
				defer wg.Done()

				// Update directory row with file count
				row := findOrCreateRow(table, dayDir)
				table.SetCell(row, 1, tview.NewTableCell(fmt.Sprintf("%d", dirFileCounts[dayDir])).
					SetTextColor(tcell.ColorGreen).
					SetAlign(tview.AlignCenter))

				// Update processed/pending status
				updateStatus()
			})

			return nil
		})

		if err != nil {
			log.Fatalf("Error organizing files: %v", err)
		}
	}()

	// Wait for processing to complete and stop the application
	go func() {
		<-done
		wg.Wait() // Wait for all UI updates to complete
		app.Stop()
	}()

	// Run the application
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(table, 0, 1, false).
		AddItem(statusText, 1, 0, false)

	if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
		log.Fatalf("Error starting the application: %v", err)
	}

	fmt.Println("Files have been organized by creation day!")
}

// Helper to find or create a row in the table for a given directory
func findOrCreateRow(table *tview.Table, dayDir string) int {
	for row := 1; row < table.GetRowCount(); row++ {
		cell := table.GetCell(row, 0)
		if cell.Text == dayDir {
			return row
		}
	}

	// If not found, create a new row
	row := table.GetRowCount()
	table.SetCell(row, 0, tview.NewTableCell(dayDir).
		SetTextColor(tcell.ColorWhite).
		SetAlign(tview.AlignLeft))
	table.SetCell(row, 1, tview.NewTableCell("0").
		SetTextColor(tcell.ColorGreen).
		SetAlign(tview.AlignCenter))
	return row
}

