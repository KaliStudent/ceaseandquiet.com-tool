import {FIELDS, encryptProfile, VaultSession} from './vault.mjs';
import {fillPreview} from './fill.mjs';

const $ = id => document.getElementById(id);
const session = new VaultSession();
let saved = false, busy = false, ready = false, lifecycle = 0, timer;
const message = text => { $('state').textContent = text; };
const clearFields = () => { FIELDS.forEach(name => { $(name).value = ''; }); $('passphrase').value = ''; };

function render() {
  const open = session.snapshot() !== null;
  $('profile-panel').hidden = saved && !open;
  $('unlock-panel').hidden = open;
  $('mode').textContent = open ? 'UNLOCKED · THIS POPUP ONLY' : saved ? 'ENCRYPTED · LOCKED' : 'CREATE YOUR LOCAL VAULT';
  $('save').hidden = saved;
  $('unlock').hidden = !saved || open;
  $('fill').hidden = !open;
  $('lock').hidden = !open;
  $('erase').hidden = !saved;
  for (const id of ['save', 'unlock', 'fill', 'erase']) $(id).disabled = busy || !ready;
  for (const name of FIELDS) $(name).readOnly = open;
}

function lock(text = 'Vault locked. Revealed fields cleared.') {
  lifecycle++; clearTimeout(timer); session.lock(); clearFields(); render(); message(text);
}

async function save() {
  const profile = Object.fromEntries(FIELDS.map(name => [name, $(name).value]));
  const envelope = await encryptProfile(profile, $('passphrase').value);
  await chrome.storage.local.set({vault: envelope});
  saved = true;
  lock('Vault encrypted and saved on this device. Unlock to fill.');
}

async function unlock() {
  const passphrase = $('passphrase').value;
  clearFields(); session.lock();
  const generation = lifecycle;
  try {
    const {vault} = await chrome.storage.local.get('vault');
    if (generation !== lifecycle) return;
    if (!vault) { saved = false; throw new Error('No vault is saved.'); }
    await session.unlock(vault, passphrase);
    if (generation !== lifecycle) { session.lock(); return; }
    const profile = session.snapshot();
    if (!profile) return;
    FIELDS.forEach(name => { $(name).value = profile[name]; });
    timer = setTimeout(() => lock('Vault locked after two minutes.'), 120000);
    message('Unlocked for up to two minutes. Check the details before filling.');
  } catch { lock('Could not unlock. Check the passphrase. Older prototype vaults must be erased and recreated.'); }
}

async function fill() {
  await fillPreview(chrome, () => session.snapshot());
  const eraseAfter = $('erase-after').checked;
  // Always lock after a successful fill, including when local deletion fails.
  lock('Form filled locally. Nothing was submitted.');
  if (eraseAfter) {
    try { await chrome.storage.local.remove('vault'); saved = false; message('Form filled. Saved vault erased. Clear the page when finished.'); }
    catch { message('Form filled, but saved vault deletion failed. Use Erase vault to retry.'); }
  }
}

async function erase() {
  lock(); await chrome.storage.local.remove('vault'); saved = false;
  message('Saved vault erased. Copies already filled into a page must be cleared there.');
}

for (const [id, action] of [['save', save], ['unlock', unlock], ['fill', fill], ['erase', erase]]) $(id).addEventListener('click', async () => {
  if (busy || !ready) return;
  busy = true; render();
  try { await action(); }
  catch (error) { message(error.message || 'Action failed. Please try again.'); }
  finally { $('passphrase').value = ''; busy = false; render(); }
});
$('lock').addEventListener('click', () => lock());
window.addEventListener('pagehide', () => lock());
document.addEventListener('visibilitychange', () => { if (document.hidden) lock(); });

try {
  await chrome.storage.local.setAccessLevel({accessLevel: 'TRUSTED_CONTEXTS'});
  const {vault} = await chrome.storage.local.get('vault'); saved = Boolean(vault); ready = true;
  message(saved ? 'Vault saved on this device. Enter your passphrase to unlock.' : 'Use sample details during beta testing. Choose a unique passphrase.');
} catch { message('Local storage is unavailable. The vault is disabled.'); }
render();
