// Website-only controller. Only authenticated ciphertext is written to storage.
// Crypto is injected so the website and extension use the same envelope format.
export function attachWebsiteVault({form, controls, fields, encrypt, decrypt, storage, clearSuggestions, document: doc = document, window: win = window, setTimer = setTimeout, clearTimer = clearTimeout, now = Date.now}) {
  const key = 'ceaseandquiet.website.vault.v3';
  let saved = false, available = false, busy = false, revealed = false, generation = 0, timer, deadline = 0;
  const input = name => form.elements.namedItem(name);
  const say = text => { controls.status.textContent = text; };
  const clear = () => {
    fields.forEach(name => { input(name).value = ''; });
    controls.passphrase.value = ''; clearSuggestions();
    form.dispatchEvent?.(new Event('profileclear'));
  };
  const render = () => {
    controls.mode.textContent = saved ? revealed ? 'PROFILE UNLOCKED' : 'ENCRYPTED PROFILE SAVED · LOCKED' : 'CREATE YOUR WEBSITE PROFILE';
    for (const name of fields) input(name).readOnly = busy || (saved && !revealed);
    controls.save.disabled = !available || busy || (saved && !revealed);
    controls.save.textContent = saved ? 'Encrypt and save changes' : 'Encrypt and save profile';
    controls.unlock.hidden = !saved || revealed;
    controls.unlock.disabled = !available || busy;
    controls.erase.hidden = !saved;
    controls.erase.disabled = !available || busy;
    controls.lock.disabled = false;
    controls.passphrase.disabled = busy || !available;
  };
  const lock = (message = 'Profile locked. Visible details and passphrase cleared.') => {
    generation++; clearTimer(timer); revealed = false; deadline = 0; clear(); render(); say(message);
  };
  const run = async action => {
    if (busy || !available) return;
    if (revealed && now() >= deadline) { lock('Profile locked after two minutes. Unlock it again.'); return; }
    busy = true; render();
    try { await action(); }
    catch (error) { say(error.message || 'Action failed. Your saved profile has not been erased.'); }
    finally { busy = false; controls.passphrase.value = ''; render(); }
  };
  controls.save.addEventListener('click', () => run(async () => {
    const current = generation;
    const profile = Object.fromEntries(fields.map(name => [name, input(name).value]));
    const envelope = await encrypt(profile, controls.passphrase.value);
    if (current !== generation) return;
    // setItem may fail (private browsing, full disk, site storage policy). Keep
    // the previous envelope and visible fields until the write succeeds.
    storage.setItem(key, JSON.stringify(envelope)); saved = true;
    lock('Encrypted profile saved in this browser for this website. Enter your passphrase to unlock.');
  }));
  controls.unlock.addEventListener('click', () => run(async () => {
    const passphrase = controls.passphrase.value;
    lock('Unlocking your saved profile…');
    const current = generation;
    try {
      const raw = storage.getItem(key);
      if (!raw || raw.length > 16384) throw new Error('Invalid saved profile');
      const profile = await decrypt(JSON.parse(raw), passphrase);
      if (current !== generation) return;
      fields.forEach(name => { input(name).value = profile[name]; });
      form.dispatchEvent?.(new Event('profilechange'));
      revealed = true; deadline = now() + 120000; render();
      timer = setTimer(() => lock('Profile locked after two minutes.'), 120000);
      say('Profile unlocked for two minutes. Edit details here, then enter your passphrase to save changes. Nothing has been sent.');
    } catch {
      if (current === generation) lock('Could not unlock. Check your passphrase. The encrypted profile is still saved.');
    }
  }));
  controls.lock.addEventListener('click', () => lock());
  controls.erase.addEventListener('click', () => run(async () => {
    lock(); storage.removeItem(key); saved = false;
    say('Saved website profile erased. You can create a new profile.');
  }));
  form.addEventListener('reset', () => lock('Form cleared. Any encrypted saved profile remains locked.'));
  win.addEventListener('pagehide', () => lock());
  doc.addEventListener('visibilitychange', () => { if (doc.hidden) lock(); });
  win.addEventListener('focus', () => { if (revealed && now() >= deadline) lock('Profile locked after two minutes.'); });
  win.addEventListener('storage', event => {
    if (event.key !== key && event.key !== null) return;
    lock('Saved profile changed in another tab. Unlock the current copy again.');
    try { saved = storage.getItem(key) !== null; } catch { available = false; }
    render();
  });
  try {
    saved = storage.getItem(key) !== null;
    available = true;
    say(saved ? 'Your encrypted website profile is saved in this browser. Enter your passphrase to unlock.' : 'Enter your details and a unique passphrase to save an encrypted profile in this browser.');
  } catch { say('Browser storage is unavailable. Manual entry works, but saving and unlocking are disabled.'); }
  render();
  return {lock};
}
