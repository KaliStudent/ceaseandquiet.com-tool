// Chrome serializes this function. Keep every target check inside it.
export function previewForm(profile) {
  if (window.top !== window || location.origin !== 'http://127.0.0.1:8787' || location.pathname !== '/' || location.search) return {ok: false, reason: 'destination'};
  const forms = document.querySelectorAll('form#cq-demo-form[data-cq-preview="v1"]');
  if (forms.length !== 1) return {ok: false, reason: 'form'};
  const form = forms[0];
  const names = ['givenName', 'familyName', 'email', 'region'];
  const inputs = names.map(name => form.elements.namedItem(name));
  if (inputs.some((input, i) => !(input instanceof HTMLInputElement) || input.type !== (names[i] === 'email' ? 'email' : 'text') || input.disabled || input.readOnly || input.getClientRects().length === 0 || input.value !== '')) return {ok: false, reason: 'fields'};
  if (profile === undefined) return {ok: true};
  if (!profile || Object.keys(profile).length !== names.length || names.some(name => typeof profile[name] !== 'string')) return {ok: false, reason: 'profile'};
  // Check every target before writing anything. Never submit or requestSubmit.
  inputs.forEach((input, i) => { input.value = profile[names[i]]; });
  inputs.forEach(input => input.dispatchEvent(new Event('input', {bubbles: true})));
  return {ok: true};
}

export async function fillPreview(chromeAPI, getProfile) {
  const [tab] = await chromeAPI.tabs.query({active: true, currentWindow: true});
  let url;
  try { url = new URL(tab?.url); } catch { throw new Error('Open the local preview form first.'); }
  if (!tab?.id || url.origin !== 'http://127.0.0.1:8787' || url.pathname !== '/' || url.search || url.username || url.password) throw new Error('Open the local preview form first.');
  const [probe] = await chromeAPI.scripting.executeScript({target: {tabId: tab.id, frameIds: [0]}, func: previewForm});
  if (!probe?.result?.ok || !probe.documentId) throw new Error('Use the empty local preview form. Clear any existing fields first.');
  const profile = getProfile();
  if (!profile) throw new Error('The vault has locked. Unlock it again.');
  // Pin the write to the document that passed preflight, avoiding navigation races.
  const [result] = await chromeAPI.scripting.executeScript({target: {tabId: tab.id, documentIds: [probe.documentId]}, func: previewForm, args: [profile]});
  if (!result?.result?.ok) throw new Error('The page changed. Nothing was submitted. Clear the form and try again.');
}
