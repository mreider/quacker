async function validateToken() {
  const token = document.getElementById('tokenInput').value;
  try {
    const response = await fetch('https://your-domain.com/validate_token', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ token })
    });
    const data = await response.json();
    if (response.ok) {
      document.getElementById('urls').value = data.urls.join('\n');
    } else {
      document.getElementById('urls').value = data.error || "Tell the admin to check the token.";
    }
  } catch (error) {
    document.getElementById('urls').value = "Error contacting server.";
  }
}
