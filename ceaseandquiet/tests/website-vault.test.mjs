import test from 'node:test';
import assert from 'node:assert/strict';
import {attachWebsiteVault} from '../site/profile-vault.mjs';
import {FIELDS, encryptProfile, decryptProfile} from '../extension/vault.mjs';

const profile = {givenName: 'Taylor', familyName: 'Example', email: 'taylor@example.test', address: '100 Example Street', city: 'Example City', region: 'TS', postalCode: '00000', dateOfBirth: '2000-01-02', phone: '+1 202-555-0100'};
const password = 'unique website test passphrase';
function element() { return {value: '', listeners: {}, addEventListener(name, action) { this.listeners[name] = action; }}; }
function website({initial, failWrite = false, failRemove = false, decrypt = decryptProfile, unavailable = false} = {}) {
  const inputs = Object.fromEntries(FIELDS.map(name => [name, element()]));
  const controls = Object.fromEntries(['save','unlock','lock','erase','mode','status','passphrase'].map(name => [name, element()]));
  const form = {...element(), elements: {namedItem: name => inputs[name]}};
  const doc = element(), win = element(); let raw = initial, expiry;
  const storage = {
    getItem() { if (unavailable) throw new Error('Storage blocked'); return raw ?? null; },
    setItem(key, value) { if (failWrite) throw new Error('Storage write failed'); raw = value; },
    removeItem() { if (failRemove) throw new Error('Storage deletion failed'); raw = undefined; }
  };
  attachWebsiteVault({form, controls, fields: FIELDS, encrypt: encryptProfile, decrypt, storage, clearSuggestions() {}, document: doc, window: win, setTimer(action) { expiry = action; return 1; }, clearTimer() { expiry = undefined; }});
  return {inputs, controls, form, doc, win, stored: () => raw, click: id => controls[id].listeners.click(), expire: () => expiry(), enter() { for (const name of FIELDS) inputs[name].value = profile[name]; controls.passphrase.value = password; }};
}

test('website profile saves ciphertext, survives reload, unlocks for editing and erases without the extension', async () => {
  let ui = website(); ui.enter(); await ui.click('save');
  assert.ok(ui.stored()); assert.ok(!ui.stored().includes(profile.email));
  assert.equal(ui.controls.passphrase.value, '');
  assert.ok(FIELDS.every(name => ui.inputs[name].value === '' && ui.inputs[name].readOnly));
  ui = website({initial: ui.stored()});
  ui.controls.passphrase.value = 'wrong website password'; await ui.click('unlock');
  assert.ok(FIELDS.every(name => ui.inputs[name].value === '')); assert.ok(ui.stored());
  ui.controls.passphrase.value = password; await ui.click('unlock');
  assert.equal(ui.inputs.email.value, profile.email); assert.equal(ui.inputs.email.readOnly, false);
  ui.inputs.city.value = 'Updated Example City'; ui.controls.passphrase.value = password;
  await ui.click('save');
  assert.equal((await decryptProfile(JSON.parse(ui.stored()), password)).city, 'Updated Example City');
  await ui.click('erase');
  assert.equal(ui.stored(), undefined); assert.equal(ui.inputs.email.value, '');
  assert.equal(ui.inputs.email.readOnly, false);
});

test('website lock, hidden page, expiry and cross-tab replacement clear visible details', async () => {
  const initial = JSON.stringify(await encryptProfile(profile, password));
  for (const action of ['lock', 'hide', 'expire', 'storage', 'reset']) {
    const ui = website({initial}); ui.controls.passphrase.value = password; await ui.click('unlock');
    if (action === 'lock') await ui.click('lock');
    if (action === 'hide') { ui.doc.hidden = true; ui.doc.listeners.visibilitychange(); }
    if (action === 'expire') ui.expire();
    if (action === 'storage') ui.win.listeners.storage({key: 'ceaseandquiet.website.vault.v3'});
    if (action === 'reset') ui.form.listeners.reset();
    assert.ok(FIELDS.every(name => ui.inputs[name].value === ''), action);
    assert.equal(ui.controls.passphrase.value, '', action);
    assert.ok(ui.stored(), action);
  }
});

test('hiding during website unlock prevents late plaintext reveal', async () => {
  let finish;
  const ui = website({initial: '{}', decrypt: () => new Promise(resolve => { finish = resolve; })});
  ui.controls.passphrase.value = password;
  const pending = ui.click('unlock'); ui.win.listeners.pagehide(); finish(profile); await pending;
  assert.ok(FIELDS.every(name => ui.inputs[name].value === ''));
});

test('website storage failures are visible and do not pretend that a profile was saved or erased', async () => {
  let ui = website({failWrite: true}); ui.enter(); await ui.click('save');
  assert.equal(ui.stored(), undefined); assert.equal(ui.inputs.email.value, profile.email);
  assert.match(ui.controls.status.textContent, /write failed/);
  ui = website({initial: '{}', failRemove: true}); await ui.click('erase');
  assert.equal(ui.stored(), '{}'); assert.match(ui.controls.status.textContent, /deletion failed/);
  ui = website({unavailable: true}); assert.equal(ui.controls.save.disabled, true); assert.equal(ui.inputs.email.readOnly, false);
});
