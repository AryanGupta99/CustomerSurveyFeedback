// Update rating value display with animation
function updateValue(elementId, value) {
  const element = document.getElementById(elementId);
  element.style.transform = 'scale(1.1)';
  element.textContent = value;
  setTimeout(() => {
    element.style.transform = 'scale(1)';
  }, 150);
}

// Placeholder for future progress tracking functionality
function updateProgress() {
  // Progress tracking removed for cleaner UI
}

let submitting = false;

// Submit form with enhanced UX
async function submitForm() {
  if (submitting) return;
  submitting = true;
  const submitBtn = document.querySelector('.submit-btn');
  const status = document.getElementById('status');
  const originalText = submitBtn.textContent;
  
  // Disable button and show loading state
  submitBtn.disabled = true;
  submitBtn.textContent = '⏳ Submitting your feedback...';
  submitBtn.style.opacity = '0.7';
  
  const q1 = parseInt(document.getElementById('q1').value, 10);
  const q2 = parseInt(document.getElementById('q2').value, 10);
  const q3 = parseInt(document.getElementById('q3').value, 10);
  const note = document.getElementById('note').value.trim();
  
  try {
    const response = await fetch('/submit', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
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
      
      // Optional: Reset form after delay
      setTimeout(() => {
        if (confirm('Would you like to submit another response?')) {
          location.reload();
        }
      }, 3000);
      
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
  // Initialize progress
  updateProgress();
  
  // Add enhanced hover and interaction effects to sliders
  const sliders = document.querySelectorAll('input[type="range"]');
  sliders.forEach(slider => {
    // Add smooth track fill animation
    slider.addEventListener('input', function() {
      const value = (this.value - this.min) / (this.max - this.min);
      const percentage = value * 100;
      this.style.background = `linear-gradient(to right, #0288d1 0%, #00acc1 ${percentage}%, #e0e0e0 ${percentage}%, #e0e0e0 100%)`;
    });
    
    // Initialize track fill
    const value = (slider.value - slider.min) / (slider.max - slider.min);
    const percentage = value * 100;
    slider.style.background = `linear-gradient(to right, #0288d1 0%, #00acc1 ${percentage}%, #e0e0e0 ${percentage}%, #e0e0e0 100%)`;
    
    slider.addEventListener('mouseover', function() {
      this.style.transform = 'scaleY(1.1)';
    });
    
    slider.addEventListener('mouseout', function() {
      this.style.transform = 'scaleY(1)';
    });
  });
  
  // Auto-resize textarea
  const textarea = document.getElementById('note');
  textarea.addEventListener('input', function() {
    this.style.height = 'auto';
    this.style.height = Math.max(80, this.scrollHeight) + 'px';
  });
});
