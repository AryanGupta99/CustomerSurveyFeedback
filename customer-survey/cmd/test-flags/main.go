package main

import (
	"customer-survey/pkg/startup"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	fmt.Println("=== Testing Startup Package Functions (No UI) ===")

	// Show AppData location
	appDataDir := startup.GetAppDataDir()
	fmt.Printf("AppData Directory: %s\n\n", appDataDir)

	// Test 1: Check initial state
	fmt.Println("[Test 1] Initial State Check")
	fmt.Printf("  IsSurveyDone: %v\n", startup.IsSurveyDone())
	fmt.Printf("  IsNoThanks: %v\n", startup.IsNoThanks())
	shouldSkip, _ := startup.ShouldRemindLater()
	fmt.Printf("  ShouldRemindLater: %v\n", shouldSkip)
	shouldShow, _ := startup.ShouldShowSurvey()
	fmt.Printf("  ShouldShowSurvey: %v\n", shouldShow)
	fmt.Printf("  Status: %s\n\n", startup.GetStatus())

	// Test 2: Mark survey as done
	fmt.Println("[Test 2] Marking Survey as Done")
	if err := startup.MarkSurveyDone(); err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else {
		fmt.Println("  ✓ done.flag created")
		flagPath := filepath.Join(appDataDir, "done.flag")
		if content, err := os.ReadFile(flagPath); err == nil {
			fmt.Printf("  Content: %s\n", string(content))
		}
	}

	// Check state again
	fmt.Printf("  IsSurveyDone: %v\n", startup.IsSurveyDone())
	shouldShow, _ = startup.ShouldShowSurvey()
	fmt.Printf("  ShouldShowSurvey: %v\n", shouldShow)
	fmt.Printf("  Status: %s\n\n", startup.GetStatus())

	// Test 3: Reset and mark No Thanks
	fmt.Println("[Test 3] Reset and Mark No Thanks")
	startup.ResetAll()
	fmt.Println("  ✓ Reset complete")

	if err := startup.MarkNoThanks(); err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else {
		fmt.Println("  ✓ nothanks.flag created")
	}

	fmt.Printf("  IsNoThanks: %v\n", startup.IsNoThanks())
	shouldShow, _ = startup.ShouldShowSurvey()
	fmt.Printf("  ShouldShowSurvey: %v\n", shouldShow)
	fmt.Printf("  Status: %s\n\n", startup.GetStatus())

	// Test 4: Reset and mark Remind Later
	fmt.Println("[Test 4] Reset and Mark Remind Later")
	startup.ResetAll()
	fmt.Println("  ✓ Reset complete")

	if err := startup.MarkRemindLater(); err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else {
		fmt.Println("  ✓ remind.txt created")
		remindPath := filepath.Join(appDataDir, "remind.txt")
		if content, err := os.ReadFile(remindPath); err == nil {
			fmt.Printf("  Remind until: %s\n", string(content))
		}
	}

	shouldSkip, _ = startup.ShouldRemindLater()
	fmt.Printf("  ShouldRemindLater: %v\n", shouldSkip)
	shouldShow, _ = startup.ShouldShowSurvey()
	fmt.Printf("  ShouldShowSurvey: %v\n", shouldShow)
	fmt.Printf("  Status: %s\n\n", startup.GetStatus())

	// Final cleanup
	fmt.Println("[Cleanup] Removing test files")
	startup.ResetAll()
	fmt.Println("  ✓ All test files removed")

	fmt.Println("\n=== All Tests Complete (No UI Launched) ===")
}
