package startup

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMarkAndCheckDone(t *testing.T) {
	// Clean up before test
	defer ResetAll()

	// Initially should not be done
	if IsSurveyDone() {
		t.Error("Survey should not be marked as done initially")
	}

	// Mark as done
	if err := MarkSurveyDone(); err != nil {
		t.Fatalf("Failed to mark survey as done: %v", err)
	}

	// Should now be done
	if !IsSurveyDone() {
		t.Error("Survey should be marked as done")
	}

	// Verify file exists
	flagPath := filepath.Join(GetAppDataDir(), "done.flag")
	if _, err := os.Stat(flagPath); os.IsNotExist(err) {
		t.Error("done.flag file should exist")
	}
}

func TestMarkAndCheckNoThanks(t *testing.T) {
	defer ResetAll()

	if IsNoThanks() {
		t.Error("NoThanks should not be set initially")
	}

	if err := MarkNoThanks(); err != nil {
		t.Fatalf("Failed to mark NoThanks: %v", err)
	}

	if !IsNoThanks() {
		t.Error("NoThanks should be set")
	}
}

func TestRemindLater(t *testing.T) {
	defer ResetAll()

	// Initially should not skip
	shouldSkip, err := ShouldRemindLater()
	if err != nil {
		t.Fatalf("Error checking remind later: %v", err)
	}
	if shouldSkip {
		t.Error("Should not skip initially")
	}

	// Mark remind later
	if err := MarkRemindLater(); err != nil {
		t.Fatalf("Failed to mark remind later: %v", err)
	}

	// Should now skip (within 7 days)
	shouldSkip, err = ShouldRemindLater()
	if err != nil {
		t.Fatalf("Error checking remind later: %v", err)
	}
	if !shouldSkip {
		t.Error("Should skip within remind window")
	}

	// Verify the date is approximately 7 days in future
	remindPath := filepath.Join(GetAppDataDir(), "remind.txt")
	data, err := os.ReadFile(remindPath)
	if err != nil {
		t.Fatalf("Failed to read remind.txt: %v", err)
	}

	remindDate, err := time.Parse(time.RFC3339, string(data))
	if err != nil {
		t.Fatalf("Invalid date in remind.txt: %v", err)
	}

	expectedDate := time.Now().Add(7 * 24 * time.Hour)
	diff := remindDate.Sub(expectedDate)
	if diff < -time.Minute || diff > time.Minute {
		t.Errorf("Remind date should be ~7 days from now, got difference: %v", diff)
	}
}

func TestShouldShowSurvey(t *testing.T) {
	defer ResetAll()

	// Initially should show
	shouldShow, err := ShouldShowSurvey()
	if err != nil {
		t.Fatalf("Error checking ShouldShowSurvey: %v", err)
	}
	if !shouldShow {
		t.Error("Should show survey initially")
	}

	// Mark as done
	MarkSurveyDone()
	shouldShow, _ = ShouldShowSurvey()
	if shouldShow {
		t.Error("Should not show survey after marking done")
	}

	// Reset and mark no thanks
	ResetAll()
	MarkNoThanks()
	shouldShow, _ = ShouldShowSurvey()
	if shouldShow {
		t.Error("Should not show survey after NoThanks")
	}

	// Reset and mark remind later
	ResetAll()
	MarkRemindLater()
	shouldShow, _ = ShouldShowSurvey()
	if shouldShow {
		t.Error("Should not show survey within remind window")
	}
}

func TestGetStatus(t *testing.T) {
	defer ResetAll()

	status := GetStatus()
	if status != "Survey should be shown" {
		t.Errorf("Expected 'Survey should be shown', got: %s", status)
	}

	MarkSurveyDone()
	status = GetStatus()
	if status != "Survey completed" {
		t.Errorf("Expected 'Survey completed', got: %s", status)
	}

	ResetAll()
	MarkNoThanks()
	status = GetStatus()
	if status != "User opted out (No Thanks)" {
		t.Errorf("Expected 'User opted out (No Thanks)', got: %s", status)
	}
}
