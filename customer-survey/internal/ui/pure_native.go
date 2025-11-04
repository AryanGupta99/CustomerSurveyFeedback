package ui

import (
	"context"
	"log"
	"os"
	"syscall"
	"time"
	"unsafe"

	"customer-survey/internal/survey"
	"customer-survey/pkg/model"
)

var (
	user32                  = syscall.NewLazyDLL("user32.dll")
	procMessageBoxW         = user32.NewProc("MessageBoxW")
	procCreateWindowExW     = user32.NewProc("CreateWindowExW")
	procDefWindowProcW      = user32.NewProc("DefWindowProcW")
	procPostQuitMessage     = user32.NewProc("PostQuitMessage")
	procGetMessageW         = user32.NewProc("GetMessageW")
	procTranslateMessage    = user32.NewProc("TranslateMessage")
	procDispatchMessageW    = user32.NewProc("DispatchMessageW")
	procRegisterClassExW    = user32.NewProc("RegisterClassExW")
	procLoadCursorW         = user32.NewProc("LoadCursorW")
	procShowWindow          = user32.NewProc("ShowWindow")
	procUpdateWindow        = user32.NewProc("UpdateWindow")
)

const (
	MB_OK                = 0x00000000
	MB_OKCANCEL          = 0x00000001
	MB_YESNOCANCEL       = 0x00000003
	MB_YESNO             = 0x00000004
	MB_ICONINFORMATION   = 0x00000040
	IDOK                 = 1
	IDCANCEL             = 2
	IDYES                = 6
	IDNO                 = 7
)

type SurveyData struct {
	Performance int
	Support     int
	Overall     int
	Feedback    string
}

// RunPureNativeGUI creates a pure Windows native GUI using only Windows API
func RunPureNativeGUI() error {
	// Step 1: Welcome prompt
	title, _ := syscall.UTF16PtrFromString("ACE Customer Survey 🏢")
	msg, _ := syscall.UTF16PtrFromString("Your Opinion Matters!\n\n" +
		"Would you like to take a quick survey\n" +
		"to help us improve our services?\n\n" +
		"Click YES to start the survey.")
	
	ret, _, _ := procMessageBoxW.Call(
		0,
		uintptr(unsafe.Pointer(msg)),
		uintptr(unsafe.Pointer(title)),
		MB_YESNO|MB_ICONINFORMATION,
	)
	
	if ret != IDYES {
		return nil // User clicked No
	}
	
	// Step 2: Collect survey responses
	data := &SurveyData{}
	
	// Question 1: Server Performance
	if !askRatingQuestion("Server Performance", "How would you rate our SERVER PERFORMANCE?", &data.Performance) {
		return nil
	}
	
	// Question 2: Technical Support
	if !askRatingQuestion("Technical Support", "How would you rate our TECHNICAL SUPPORT?", &data.Support) {
		return nil
	}
	
	// Question 3: Overall Rating
	if !askRatingQuestion("Overall Experience", "What is your OVERALL RATING?", &data.Overall) {
		return nil
	}
	
	// Question 4: Feedback (optional)
	feedbackTitle, _ := syscall.UTF16PtrFromString("Additional Feedback")
	feedbackMsg, _ := syscall.UTF16PtrFromString("Would you like to add any comments?\n\n" +
		"(Note: Text input will be collected in next version.\n" +
		"Click OK if you have feedback, Cancel to skip)")
	
	ret, _, _ = procMessageBoxW.Call(
		0,
		uintptr(unsafe.Pointer(feedbackMsg)),
		uintptr(unsafe.Pointer(feedbackTitle)),
		MB_OKCANCEL|MB_ICONINFORMATION,
	)
	
	if ret == IDOK {
		data.Feedback = "User indicated they have feedback"
	}
	
	// Step 3: Submit
	submitSurveyData(data)
	
	return nil
}

func askRatingQuestion(title, question string, rating *int) bool {
	titlePtr, _ := syscall.UTF16PtrFromString("ACE Survey - " + title)
	msgText := question + "\n\n" +
		"😊 YES = Good (Excellent)\n" +
		"😐 CANCEL = Okay (Average)\n" +
		"☹️ NO = Bad (Needs Improvement)\n\n" +
		"Choose your rating:"
	
	msgPtr, _ := syscall.UTF16PtrFromString(msgText)
	
	ret, _, _ := procMessageBoxW.Call(
		0,
		uintptr(unsafe.Pointer(msgPtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		MB_YESNOCANCEL|MB_ICONINFORMATION,
	)
	
	switch ret {
	case IDYES:
		*rating = 3 // Good
	case IDCANCEL:
		*rating = 2 // Okay
	case IDNO:
		*rating = 1 // Bad
	default:
		return false // User cancelled
	}
	
	return true
}

func submitSurveyData(data *SurveyData) {
	// Show submitting message
	title, _ := syscall.UTF16PtrFromString("Submitting... ⏳")
	msg, _ := syscall.UTF16PtrFromString("Submitting your feedback...\n\nPlease wait...")
	procMessageBoxW.Call(
		0,
		uintptr(unsafe.Pointer(msg)),
		uintptr(unsafe.Pointer(title)),
		MB_OK|MB_ICONINFORMATION,
	)
	
	// Get system info
	username := os.Getenv("USERNAME")
	if username == "" {
		username = "user@company.com"
	}
	
	servername, _ := os.Hostname()
	if servername == "" {
		servername = "unknown-server"
	}
	
	// Create response
	resp := model.SurveyResponse{
		ServerName:        servername,
		UserName:          username,
		SurveyResponse:    "Completed",
		ServerPerformance: data.Performance,
		TechnicalSupport:  data.Support,
		OverallSupport:    data.Overall,
		Note:              data.Feedback,
	}
	
	// Submit to webhook
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if err := survey.SubmitSurvey(ctx, resp); err != nil {
		log.Printf("Submission error: %v (saved locally)", err)
		title, _ = syscall.UTF16PtrFromString("✓ Saved Offline")
		msg, _ = syscall.UTF16PtrFromString("✓ Feedback saved locally!\n\n" +
			"Your response has been saved.\n" +
			"We'll sync it when you're back online.")
		procMessageBoxW.Call(
			0,
			uintptr(unsafe.Pointer(msg)),
			uintptr(unsafe.Pointer(title)),
			MB_OK|MB_ICONINFORMATION,
		)
	} else {
		title, _ = syscall.UTF16PtrFromString("✓ Thank You!")
		msg, _ = syscall.UTF16PtrFromString("✓ Thank you for your feedback!\n\n" +
			"Your response has been submitted successfully.\n\n" +
			"We appreciate your time and valuable input!")
		procMessageBoxW.Call(
			0,
			uintptr(unsafe.Pointer(msg)),
			uintptr(unsafe.Pointer(title)),
			MB_OK|MB_ICONINFORMATION,
		)
	}
}
