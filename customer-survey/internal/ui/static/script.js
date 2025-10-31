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
  
  statusEl.textContent = '⏰ We\'ll remind you next time!';
  statusEl.className = 'status success show';
  
  // Submit reminder response
  try {
    await fetch('/submit', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        survey_response: 'remind_later',
        server_performance: 0,
        technical_support: 0,
        overall_support: 0,
        note: 'User requested reminder later'
      })
    });
  } catch (error) {
    console.error('Error:', error);
  }
  
  // Close window after delay
  setTimeout(() => {
    window.close();
  }, 1500);
}

// Handle No button - submit declined response
async function handleNo() {
  const statusEl = document.getElementById('status');
  const buttons = document.querySelectorAll('.prompt-btn');
  
  // Disable all buttons
  buttons.forEach(btn => btn.disabled = true);
  
  statusEl.textContent = '⏳ Saving your preference...';
  statusEl.className = 'status show';
  
  try {
    // Submit declined response with just user details
    const response = await fetch('/submit', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        survey_response: 'declined',
        server_performance: 0,
        technical_support: 0,
        overall_support: 0,
        note: 'User declined to participate in survey'
      })
    });
    
    if (response.ok) {
      statusEl.textContent = '✅ Thank you! Your preference has been recorded.';
      statusEl.className = 'status success show';
      
      // Close window after delay
      setTimeout(() => {
        window.close();
      }, 2000);
    } else {
      throw new Error('Submission failed');
    }
  } catch (error) {
    console.error('Error:', error);
    statusEl.textContent = '❌ Error recording response.';
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
    status.textContent = '⚠️ Please answer all three questions before submitting.';
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
  submitBtn.textContent = '⏳ Submitting your feedback...';
  submitBtn.style.opacity = '0.7';
  
  const q1 = parseInt(q1Radio ? q1Radio.value : "0", 10);
  const q2 = parseInt(q2Radio ? q2Radio.value : "0", 10);
  const q3 = parseInt(q3Radio ? q3Radio.value : "0", 10);
  const note = document.getElementById('note').value.trim() || "";
  
  try {
    const response = await fetch('/submit', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        survey_response: 'completed',
        server_performance: q1,
        technical_support: q2,
        overall_support: q3,
        note: note
      })
    });
    
    if (response.ok) {
      // Success animation
      status.textContent = '✅ Thank you! Your feedback helps us improve our services.';
      status.className = 'status success show';
      submitBtn.textContent = '✅ Feedback Submitted!';
      
      // Close window after delay
      setTimeout(() => {
        window.close();
      }, 2500);
      
    } else {
      const errorText = await response.text();
      throw new Error(errorText);
    }
    
  } catch (error) {
    console.error('Submission error:', error);
    status.textContent = '❌ Something went wrong. Please try again.';
    status.className = 'status error show';
    
    // Reset button
    submitBtn.disabled = false;
    submitBtn.textContent = originalText;
    submitBtn.style.opacity = '1';
    submitting = false;
  }
}

// Add smooth interactions on page load
document.addEventListener('DOMContentLoaded', function() {
  // Auto-resize textarea
  const textarea = document.getElementById('note');
  if (textarea) {
    textarea.addEventListener('input', function() {
      this.style.height = 'auto';
      this.style.height = Math.max(60, this.scrollHeight) + 'px';
    });
  }
});
