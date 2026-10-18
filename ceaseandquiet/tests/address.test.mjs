import test from 'node:test';
import assert from 'node:assert/strict';
import {attachAddressSuggestions} from '../extension/address.mjs';

function element() { return {value: '', checked: false, readOnly: false, hidden: true, children: [], listeners: {}, addEventListener(name, action) { this.listeners[name] = action; }, replaceChildren() { this.children = []; }, append(value) { this.children.push(value); }, dispatchEvent(event) { this.listeners[event.type]?.(); }}; }

test('address suggestions require opt-in and selection fills only address components', async () => {
  const savedDocument = globalThis.document, savedTimer = globalThis.setTimeout, savedClear = globalThis.clearTimeout;
  let pending, calls = 0;
  const fields = Object.fromEntries(['address','city','region','postalCode','enabled','results','status'].map(name => [name, element()]));
  globalThis.document = {createElement: element};
  globalThis.setTimeout = action => { pending = action; return 1; };
  globalThis.clearTimeout = () => { pending = null; };
  try {
    attachAddressSuggestions({...fields, lookup: async text => { calls++; assert.equal(text, '100 Exam'); return [{label: '100 Example Street, Example City, TS 00000', address: '100 Example Street', city: 'Example City', state: 'TS', postalCode: '00000'}]; }});
    fields.address.value = '100 Exam'; fields.address.dispatchEvent(new Event('input'));
    assert.equal(pending, null); assert.equal(calls, 0);
    fields.enabled.checked = true; fields.enabled.listeners.change(); fields.address.dispatchEvent(new Event('input'));
    assert.equal(calls, 0); await pending(); assert.equal(calls, 1);
    const button = fields.results.children[0].children[0]; button.listeners.click();
    assert.equal(fields.address.value, '100 Example Street'); assert.equal(fields.city.value, 'Example City'); assert.equal(fields.region.value, 'TS'); assert.equal(fields.postalCode.value, '00000');
    assert.equal(fields.results.hidden, true); assert.equal(pending, null);
  } finally { globalThis.document = savedDocument; globalThis.setTimeout = savedTimer; globalThis.clearTimeout = savedClear; }
});

test('late lookup results cannot repopulate suggestions after reset or lock', async () => {
  const savedDocument = globalThis.document, savedTimer = globalThis.setTimeout, savedClear = globalThis.clearTimeout;
  let pending, finish;
  const fields = Object.fromEntries(['address','city','region','postalCode','enabled','results','status'].map(name => [name, element()]));
  globalThis.document = {createElement: element}; globalThis.setTimeout = action => { pending = action; return 1; }; globalThis.clearTimeout = () => { pending = null; };
  try {
    const api = attachAddressSuggestions({...fields, lookup: () => new Promise(resolve => { finish = resolve; })});
    fields.enabled.checked = true; fields.address.value = '100 Exam'; fields.address.dispatchEvent(new Event('input'));
    const work = pending(); api.clear(); finish([{label: 'stale result'}]); await work;
    assert.equal(fields.results.children.length, 0); assert.equal(fields.results.hidden, true);
  } finally { globalThis.document = savedDocument; globalThis.setTimeout = savedTimer; globalThis.clearTimeout = savedClear; }
});
