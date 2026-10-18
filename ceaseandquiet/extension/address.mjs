// Shared opt-in UI. Callers choose the transport; no direct provider requests.
export function attachAddressSuggestions({address, city, region, postalCode, enabled, results, status, lookup}) {
  let timer, generation = 0, controller;
  const clear = () => {
    generation++; clearTimeout(timer); controller?.abort(); results.replaceChildren(); results.hidden = true;
  };
  enabled.addEventListener('change', () => { clear(); status.textContent = enabled.checked ? 'Geoapify receives the partial address you type. Enter at least four characters.' : 'Address suggestions off. Enter the address manually.'; });
  address.addEventListener('input', () => {
    clear();
    if (!enabled.checked || address.readOnly || address.value.trim().length < 4) return;
    const current = generation, text = address.value.trim();
    timer = setTimeout(async () => {
      controller = new AbortController();
      status.textContent = 'Looking up address suggestions…';
      try {
        const items = await lookup(text, controller.signal);
        if (current !== generation || !enabled.checked || address.readOnly) return;
        for (const item of items.slice(0, 5)) {
          const row = document.createElement('li'), button = document.createElement('button');
          button.type = 'button'; button.className = 'glow-button secondary'; button.textContent = item.label;
          button.addEventListener('click', () => {
            clear();
            for (const [field, value] of [[address,item.address],[city,item.city],[region,item.state],[postalCode,item.postalCode]]) {
              field.value = value; field.dispatchEvent(new Event('input', {bubbles: true}));
            }
            clear(); status.textContent = 'Address selected. Check every field and add an apartment or unit if needed.';
          });
          row.append(button); results.append(row);
        }
        results.hidden = items.length === 0;
        status.textContent = items.length ? 'Choose an address, then check the filled fields.' : 'No matches. Continue entering the address manually.';
      } catch (error) {
        if (current !== generation || error.name === 'AbortError') return;
        const messages = {
          401: 'Sign in to staging again, then retry the address.',
          403: 'The address request was blocked. Refresh this page and retry.',
          429: 'Too many address searches. Wait one minute and try again.',
          502: 'The address provider did not respond. Try again shortly.',
          503: 'The address service is not configured.',
          connection: 'Could not connect to address suggestions. Check your connection and retry.'
        };
        status.textContent = (messages[error.code] || 'Suggestions unavailable.') + ' You can enter the address manually.';
      }
    }, 400);
  });
  return {clear};
}

// Executes within the approved page using its temporary activeTab grant. The
// popup keeps connect-src none and gains no permanent host permission.
export async function lookupAddressOnPage(text) {
  if (window.top !== window || !['http://127.0.0.1:8787', 'https://abcw.xyz:8443'].includes(location.origin) || location.pathname !== '/' || location.search || typeof text !== 'string' || text.length < 4 || text.length > 200) throw new Error('Use the approved preview page.');
  const response = await fetch('/api/v1/address-suggestions', {method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify({text}), cache: 'no-store', signal: AbortSignal.timeout(6000)});
  if (!response.ok) throw new Error('Suggestions unavailable');
  const data = await response.json();
  return data.items;
}
