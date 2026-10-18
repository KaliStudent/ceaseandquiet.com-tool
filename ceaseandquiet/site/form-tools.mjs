export async function lookupWebsiteAddress(text, signal) {
  let response;
  try {
    response = await fetch('/api/v1/address-suggestions', {
      method: 'POST', mode: 'same-origin', credentials: 'same-origin', referrerPolicy: 'same-origin',
      headers: {'Content-Type': 'application/json'}, body: JSON.stringify({text}), cache: 'no-store', signal
    });
  } catch (error) {
    if (error.name === 'AbortError') throw error;
    const failure = new Error('Address connection failed'); failure.code = 'connection'; throw failure;
  }
  if (!response.ok) {
    const error = new Error('Address lookup failed'); error.code = response.status; throw error;
  }
  const data = await response.json();
  if (!Array.isArray(data.items)) throw new Error('Invalid address response');
  return data.items;
}

export function attachBirthDatePicker(input) {
  input.max = new Date().toISOString().slice(0, 10);
  input.addEventListener('click', () => {
    if (input.disabled || input.readOnly || typeof input.showPicker !== 'function') return;
    try { input.showPicker(); } catch { /* Native keyboard/date entry remains available. */ }
  });
}
