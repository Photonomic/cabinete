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
	// Define the flags as variables
	var dir string
	var organizeByYear bool
	var organizeByMonth bool

	// Set up the root command
	var rootCmd = &cobra.Command{
		Use:   "cabinete",
		Short: "Organize photos by creation date into a directory structure.",
		Long: `Cabinete organizes photos (or other files) by creation date.
The tool can create a directory structure by year, month, or day, and moves
files into the appropriate subdirectories.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Check if the directory flag is provided
			if dir == "" {
				fmt.Println("Error: directory flag is required")
				os.Exit(1)
			}

			// Run the organizer with the provided options
			runOrganizer(dir, organizeByYear, organizeByMonth)
		},
	}

	// Set up flags
	rootCmd.Flags().StringVarP(&dir, "dir", "d", "", "Directory containing files to organize (required)")
	rootCmd.MarkFlagRequired("dir")
	rootCmd.Flags().BoolVarP(&organizeByYear, "year", "y", false, "Organize files by year")
	rootCmd.Flags().BoolVarP(&organizeByMonth, "month",  "m", false, "Organize files by month within each year")

	// Execute the command
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

// runOrganizer organizes files based on year, month, or day.
func runOrganizer(dir string, organizeByYear bool, organizeByMonth bool) {
	app := tview.NewApplication()
	table := tview.NewTable().SetBorders(true).SetFixed(1, 1)
	table.SetCell(0, 0, tview.NewTableCell("Directory").SetTextColor(tcell.ColorYellow).SetAlign(tview.AlignCenter))
	table.SetCell(0, 1, tview.NewTableCell("Total Files").SetTextColor(tcell.ColorYellow).SetAlign(tview.AlignCenter))

	statusText := tview.NewTextView().SetTextAlign(tview.AlignCenter).SetDynamicColors(true)
	layout := tview.NewFlex().SetDirection(tview.FlexRow).AddItem(table, 0, 1, false).AddItem(statusText, 1, 0, false)

	var wg sync.WaitGroup
	done := make(chan struct{})

	// Counters for directories and file processing
	dirFileCounts := make(map[string]int)
	yearFileCounts := make(map[string]map[string]int)
	var totalFiles, processedFiles int
	var mu sync.Mutex

	// Run file organization in a goroutine
	go func() {
		defer close(done)

		// Count total files first
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				totalFiles++
			}
			return nil
		})

		// Process files and move them to respective directories based on options
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}

			// Determine year and month for the file creation time
			creationTime := info.ModTime()
			yearDir := fmt.Sprintf("%d", creationTime.Year())
			monthDir := fmt.Sprintf("%02d - %s", creationTime.Month(), creationTime.Month().String())

			// Determine target directory based on options
			var targetDir string
			if organizeByYear {
				targetDir = filepath.Join(dir, yearDir)
			}
			if organizeByMonth {
				targetDir = filepath.Join(dir, yearDir, monthDir)
			}
			// Default to organizing by day if no options are set
			if !organizeByYear && !organizeByMonth {
				dayDir := fmt.Sprintf("%02d", creationTime.Day())
				targetDir = filepath.Join(dir, yearDir, monthDir, dayDir)
			}

			// Move file to target directory
			targetPath := filepath.Join(targetDir, info.Name())
			if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
				return err
			}
			if err := os.Rename(path, targetPath); err != nil {
				log.Printf("Failed to move file %s to %s: %v", path, targetDir, err)
			}

			// Update counters
			mu.Lock()
			if organizeByYear {
				if _, exists := yearFileCounts[yearDir]; !exists {
					yearFileCounts[yearDir] = make(map[string]int)
				}
				if organizeByMonth {
					yearFileCounts[yearDir][monthDir]++
				} else {
					dirFileCounts[yearDir]++
				}
			} else {
				dirFileCounts[targetDir]++
			}
			processedFiles++
			mu.Unlock()

			// Queue UI update
			wg.Add(1)
			app.QueueUpdateDraw(func() {
				defer wg.Done()

				// Update table based on options
				table.Clear()
				rowIndex := 1
				if organizeByYear {
					for year, months := range yearFileCounts {
						// Add row for year total
						yearTotal := 0
						for _, count := range months {
							yearTotal += count
						}
						table.SetCell(rowIndex, 0, tview.NewTableCell(fmt.Sprintf("Year: %s", year)).
							SetTextColor(tcell.ColorGreen).
							SetAlign(tview.AlignLeft))
						table.SetCell(rowIndex, 1, tview.NewTableCell(fmt.Sprintf("%d", yearTotal)).
							SetTextColor(tcell.ColorWhite).
							SetAlign(tview.AlignCenter))
						rowIndex++

						// Add row for each month in the year
						if organizeByMonth {
							for month, count := range months {
								table.SetCell(rowIndex, 0, tview.NewTableCell(fmt.Sprintf("  %s", month)).
									SetTextColor(tcell.ColorBlue).
									SetAlign(tview.AlignLeft))
								table.SetCell(rowIndex, 1, tview.NewTableCell(fmt.Sprintf("%d", count)).
									SetTextColor(tcell.ColorWhite).
									SetAlign(tview.AlignCenter))
								rowIndex++
							}
						}
					}
				} else {
					// Default to day-level organization if no options are set
					for dir, count := range dirFileCounts {
						table.SetCell(rowIndex, 0, tview.NewTableCell(dir).
							SetTextColor(tcell.ColorGreen).
							SetAlign(tview.AlignLeft))
						table.SetCell(rowIndex, 1, tview.NewTableCell(fmt.Sprintf("%d", count)).
							SetTextColor(tcell.ColorWhite).
							SetAlign(tview.AlignCenter))
						rowIndex++
					}
				}

				// Update status text
				statusText.SetText(fmt.Sprintf("[green]Processed:[white] %d / [red]Pending:[white] %d", processedFiles, totalFiles-processedFiles))
			})

			return nil
		})
	}()

	// Wait for processing to complete and stop the application
	go func() {
		<-done
		wg.Wait()
		app.Stop()
	}()

	// Run the application
	if err := app.SetRoot(layout, true).EnableMouse(true).Run(); err != nil {
		log.Fatalf("Error starting the application: %v", err)
	}

	fmt.Println("Files have been organized!")
}

