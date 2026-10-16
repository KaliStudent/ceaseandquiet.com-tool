import test from 'node:test';
import assert from 'node:assert/strict';

const profile = {givenName: 'Taylor', familyName: 'Example', email: 'taylor@example.test', region: 'Test region'};
const password = 'unique test phrase 247!';
let counter = 0;

async function popup({stored, failDelete = false, failStorage = false} = {}) {
  const ids = ['givenName','familyName','email','region','passphrase','profile-panel','unlock-panel','mode','save','unlock','fill','lock','erase','erase-after','state'];
  const elements = Object.fromEntries(ids.map(id => [id, {value: '', hidden: false, disabled: false, checked: true, listeners: {}, addEventListener(name, callback) { this.listeners[name] = callback; }}]));
  const events = {}, docEvents = {};
  let storage = stored;
  globalThis.document = {hidden: false, getElementById: id => elements[id], addEventListener: (name, callback) => { docEvents[name] = callback; }};
  globalThis.window = {addEventListener: (name, callback) => { events[name] = callback; }};
  globalThis.chrome = {
    storage: {local: {
      setAccessLevel: async input => { assert.equal(input.accessLevel, 'TRUSTED_CONTEXTS'); if (failStorage) throw new Error('Unavailable'); },
      get: async () => ({vault: storage}),
      set: async input => { storage = structuredClone(input.vault); },
      remove: async () => { if (failDelete) throw new Error('Deletion failed'); storage = undefined; }
    }},
    tabs: {query: async () => [{id: 7, url: 'http://127.0.0.1:8787/'}]},
    scripting: {executeScript: async () => [{documentId: 'same-document', result: {ok: true}}]}
  };
  await import('../extension/popup.js?test=' + counter++);
  return {elements, events, docEvents, storage: () => storage, click: async id => { await elements[id].listeners.click(); }, close: () => { events.pagehide(); delete globalThis.document; delete globalThis.window; delete globalThis.chrome; }};
}

test('popup save, reopen, failed unlock, successful fill, erase and lock lifecycle', async () => {
  let ui = await popup();
  try {
    for (const [name,value] of Object.entries(profile)) ui.elements[name].value = value;
    ui.elements.passphrase.value = password; await ui.click('save');
    assert.equal(ui.elements['profile-panel'].hidden, true);
    assert.ok(Object.keys(profile).every(name => ui.elements[name].value === ''));
    const stored = ui.storage(); assert.ok(!JSON.stringify(stored).includes(profile.email));
    ui.close(); ui = await popup({stored});
    ui.elements.passphrase.value = 'wrong test passphrase'; await ui.click('unlock');
    assert.equal(ui.elements['profile-panel'].hidden, true);
    assert.equal(ui.elements.passphrase.value, '');
    ui.elements.passphrase.value = password; await ui.click('unlock');
    assert.equal(ui.elements.givenName.value, profile.givenName);
    assert.equal(ui.elements['profile-panel'].hidden, false);
    await ui.click('fill');
    assert.equal(ui.storage(), undefined);
    assert.equal(ui.elements.givenName.value, '');
    assert.equal(ui.elements.fill.hidden, true);
  } finally { ui.close(); }
});
test('fill deletion failure keeps encrypted record and still clears revealed values', async () => {
  const ui = await popup({failDelete: true});
  try {
    for (const [name,value] of Object.entries(profile)) ui.elements[name].value = value;
    ui.elements.passphrase.value = password; await ui.click('save');
    ui.elements.passphrase.value = password; await ui.click('unlock');
    await ui.click('fill');
    assert.ok(ui.storage()); assert.equal(ui.elements.givenName.value, '');
    assert.match(ui.elements.state.textContent, /deletion failed/);
    assert.equal(ui.elements['profile-panel'].hidden, true);
  } finally { ui.close(); }
});
test('hiding the popup clears visible fields; unavailable storage disables operations', async () => {
  let ui = await popup();
  try {
    ui.elements.email.value = profile.email;
    document.hidden = true; ui.docEvents.visibilitychange();
    assert.equal(ui.elements.email.value, '');
    ui.close(); ui = await popup({failStorage: true});
    assert.ok(['save','unlock','fill','erase'].every(id => ui.elements[id].disabled));
  } finally { ui.close(); }
});
