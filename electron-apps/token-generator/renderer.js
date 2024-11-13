const crypto = require('crypto');

function generateToken() {
  const tokenSecret = document.getElementById('tokenSecret').value.trim();
  const bucketName = document.getElementById('bucketName').value.trim();
  const validMinutes = parseInt(document.getElementById('validMinutes').value);

  // Input validation
  if (!tokenSecret || !bucketName || isNaN(validMinutes) || validMinutes <= 0) {
    alert('Please enter valid inputs for all fields.');
    return;
  }

  // Calculate expiration time (Unix timestamp)
  const expirationTime = Math.floor(Date.now() / 1000) + validMinutes * 60;

  // Prepare token data
  const tokenData = `${bucketName}.${expirationTime}`;

  // Generate HMAC hash
  const hash = crypto
    .createHmac('sha256', tokenSecret)
    .update(tokenData)
    .digest('hex');

  // Construct the token
  const token = `${bucketName}.${expirationTime}.${hash}`;

  // Display the token
  document.getElementById('generatedToken').value = token;

  // Start the countdown timer
  startCountdown(expirationTime);
}

function startCountdown(expirationTime) {
  const timerElement = document.getElementById('countdownTimer');
  const interval = setInterval(() => {
    const currentTime = Math.floor(Date.now() / 1000);
    const timeLeft = expirationTime - currentTime;

    if (timeLeft <= 0) {
      clearInterval(interval);
      timerElement.textContent = 'Token has expired';
      timerElement.style.color = 'red';
      return;
    }

    const hours = Math.floor(timeLeft / 3600);
    const minutes = Math.floor((timeLeft % 3600) / 60);
    const seconds = timeLeft % 60;

    timerElement.textContent = `${hours}h ${minutes}m ${seconds}s`;
  }, 1000);
}
