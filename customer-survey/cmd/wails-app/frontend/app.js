// Wails runtime bindings
let SubmitSurvey;
let HandleRemindMeLater;
let HandleNoThanks;
let WindowClose;

// Initialize Wails runtime
window.addEventListener('DOMContentLoaded', () => {
  // Get Wails runtime functions
  if (window.runtime) {
    WindowClose = window.runtime.Quit;
  }
  
  // Get Go backend bindings
  if (window.go && window.go.main && window.go.main.App) {
    SubmitSurvey = window.go.main.App.SubmitSurvey;
    HandleRemindMeLater = window.go.main.App.HandleRemindMeLater;
    HandleNoThanks = window.go.main.App.HandleNoThanks;
  }
  
  // Auto-resize textarea
  const textarea = document.getElementById('note');
  if (textarea) {
    textarea.addEventListener('input', function() {
      this.style.height = 'auto';
      this.style.height = Math.max(35, this.scrollHeight) + 'px';
    });
  }
});

// Handle Yes button - show survey form
function handleYes() {
  document.getElementById('promptScreen').classList.add('hidden');
  document.getElementById('surveyFormContainer').classList.remove('hidden');
}

// Handle Remind Me Later button - close window with reminder flag
async function handleRemindLater() {
  const statusEl = document.getElementById('status');
  const buttons = document.querySelectorAll('.prompt-btn');
  
  // Disable all buttons
  buttons.forEach(btn => btn.disabled = true);
  
  statusEl.textContent = 'We\'ll remind you in 7 days!';
  statusEl.className = 'status success show';
  
  try {
    // Call Go backend to save Remind Me Later settings
    if (HandleRemindMeLater) {
      const result = await HandleRemindMeLater();
      if (result && result.success) {
        // Close window after delay
        setTimeout(() => {
          if (WindowClose) {
            WindowClose();
          } else {
            window.close();
          }
        }, 1500);
      } else {
        throw new Error(result ? result.error : 'Failed to save reminder');
      }
    }
  } catch (error) {
    console.error('Error:', error);
    statusEl.textContent = 'Error saving reminder. Please try again.';
    statusEl.className = 'status error show';
    buttons.forEach(btn => btn.disabled = false);
  }
}

// Handle No button - submit declined response
async function handleNo() {
  const statusEl = document.getElementById('status');
  const buttons = document.querySelectorAll('.prompt-btn');
  
  // Disable all buttons
  buttons.forEach(btn => btn.disabled = true);
  
  statusEl.textContent = 'Saving your preference...';
  statusEl.className = 'status show';
  
  try {
    // Call Go backend to save No Thanks settings
    if (HandleNoThanks) {
      const result = await HandleNoThanks();
      
      if (result && result.success) {
        statusEl.textContent = 'Thank you! Your preference has been recorded.';
        statusEl.className = 'status success show';
        
        // Close window after delay
        setTimeout(() => {
          if (WindowClose) {
            WindowClose();
          } else {
            window.close();
          }
        }, 2000);
      } else {
        throw new Error(result ? result.error : 'Failed to save preference');
      }
    } else {
      throw new Error('Wails backend not available');
    }
  } catch (error) {
    console.error('Error:', error);
    statusEl.textContent = 'Error recording response.';
    statusEl.className = 'status error show';
    
    // Re-enable buttons
    buttons.forEach(btn => btn.disabled = false);
  }
}

let submitting = false;

// Submit form with 1-2-3 rating system
async function submitForm() {
  if (submitting) return;
  
  // Get selected ratings (1, 2, or 3)
  const q1Radio = document.querySelector('input[name="q1"]:checked');
  const q2Radio = document.querySelector('input[name="q2"]:checked');
  const q3Radio = document.querySelector('input[name="q3"]:checked');
  
  // Validate all questions are answered
  if (!q1Radio || !q2Radio || !q3Radio) {
    const status = document.getElementById('formStatus');
    status.textContent = 'Please answer all three questions before submitting.';
    status.className = 'status error show';
    
    setTimeout(() => {
      status.classList.remove('show');
    }, 3000);
    return;
  }
  
  submitting = true;
  const submitBtn = document.querySelector('.submit-btn');
  const status = document.getElementById('formStatus');
  const originalText = submitBtn.textContent;
  
  // Disable button and show loading state
  submitBtn.disabled = true;
  submitBtn.textContent = 'Submitting your feedback...';
  submitBtn.style.opacity = '0.7';
  
  const q1 = parseInt(q1Radio ? q1Radio.value : "0", 10);
  const q2 = parseInt(q2Radio ? q2Radio.value : "0", 10);
  const q3 = parseInt(q3Radio ? q3Radio.value : "0", 10);
  const note = document.getElementById('note').value.trim() || "";
  
  try {
    // Submit via Wails backend
    if (SubmitSurvey) {
      const result = await SubmitSurvey('completed', q1, q2, q3, note);
      
      if (result && result.success) {
        // Hide survey form and show thank you screen
        document.getElementById('surveyFormContainer').classList.add('hidden');
        document.getElementById('thankYouScreen').classList.remove('hidden');
        
        // Close window after delay
        setTimeout(() => {
          if (WindowClose) {
            WindowClose();
          } else {
            window.close();
          }
        }, 3500);
      } else {
        throw new Error('Submission failed');
      }
    } else {
      throw new Error('Wails backend not available');
    }
    
  } catch (error) {
    console.error('Submission error:', error);
    status.textContent = 'Something went wrong. Please try again.';
    status.className = 'status error show';
    
    // Reset button
    submitBtn.disabled = false;
    submitBtn.textContent = originalText;
    submitBtn.style.opacity = '1';
    submitting = false;
  }
}
