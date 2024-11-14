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

	table.SetCell(0, 0, tview.NewTableCell("File Name").
		SetTextColor(tcell.ColorYellow).
		SetAlign(tview.AlignCenter))
	table.SetCell(0, 1, tview.NewTableCell("Day Directory").
		SetTextColor(tcell.ColorYellow).
		SetAlign(tview.AlignCenter))
	table.SetCell(0, 2, tview.NewTableCell("Status").
		SetTextColor(tcell.ColorYellow).
		SetAlign(tview.AlignCenter))

	var wg sync.WaitGroup
	rowIndex := 1
	done := make(chan struct{})

	// Start a goroutine for processing files
	go func() {
		defer close(done)
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			// Getting file creation time (using modification time as proxy if unsupported)
			creationTime := info.ModTime()
			dayDir := fmt.Sprintf("%02d", creationTime.Day())

			// Target directory based on day
			targetDir := filepath.Join(dir, dayDir)
			if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
				return err
			}

			// Move file to the day directory
			targetPath := filepath.Join(targetDir, info.Name())
			moveStatus := "Moved"
			if err := os.Rename(path, targetPath); err != nil {
				moveStatus = "Failed"
			}

			// Schedule a table update
			wg.Add(1)
			app.QueueUpdateDraw(func() {
				defer wg.Done()
				table.SetCell(rowIndex, 0, tview.NewTableCell(info.Name()).
					SetTextColor(tcell.ColorWhite).
					SetAlign(tview.AlignLeft))
				table.SetCell(rowIndex, 1, tview.NewTableCell(dayDir).
					SetTextColor(tcell.ColorGreen).
					SetAlign(tview.AlignCenter))
				table.SetCell(rowIndex, 2, tview.NewTableCell(moveStatus).
					SetTextColor(tcell.ColorLightGray).
					SetAlign(tview.AlignCenter))
				rowIndex++
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
	if err := app.SetRoot(table, true).EnableMouse(true).Run(); err != nil {
		log.Fatalf("Error starting the application: %v", err)
	}

	fmt.Println("Files have been organized by creation day!")
}

