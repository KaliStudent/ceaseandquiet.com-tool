import test from 'node:test';
import assert from 'node:assert/strict';
import {lookupWebsiteAddress, attachBirthDatePicker} from '../site/form-tools.mjs';

test('website lookup sends only address text and keeps authentication and referrer inside this origin', async () => {
  const original = globalThis.fetch;
  globalThis.fetch = async (url, options) => {
    assert.equal(url, '/api/v1/address-suggestions');
    assert.equal(options.mode, 'same-origin');
    assert.equal(options.credentials, 'same-origin');
    assert.equal(options.referrerPolicy, 'same-origin');
    assert.deepEqual(JSON.parse(options.body), {text: '100 Example'});
    return {ok: true, json: async () => ({items: [{address: '100 Example Street'}]})};
  };
  try { assert.equal((await lookupWebsiteAddress('100 Example'))[0].address, '100 Example Street'); }
  finally { globalThis.fetch = original; }
});

test('website lookup distinguishes rejected requests and unavailable providers without showing response bodies', async () => {
  const original = globalThis.fetch;
  try {
    for (const status of [401, 403, 429, 502, 503]) {
      globalThis.fetch = async () => ({ok: false, status});
      await assert.rejects(lookupWebsiteAddress('100 Example'), error => error.code === status);
    }
    globalThis.fetch = async () => { throw new TypeError('private request details'); };
    await assert.rejects(lookupWebsiteAddress('100 Example'), error => error.code === 'connection' && !error.message.includes('private'));
  } finally { globalThis.fetch = original; }
});

test('clicking date of birth opens the native calendar and prevents future dates', () => {
  let click, opened = 0;
  const input = {addEventListener: (name, action) => { assert.equal(name, 'click'); click = action; }, showPicker: () => { opened++; }};
  attachBirthDatePicker(input);
  assert.match(input.max, /^\d{4}-\d{2}-\d{2}$/);
  click(); assert.equal(opened, 1);
  input.readOnly = true; click(); assert.equal(opened, 1);
  input.readOnly = false; input.showPicker = undefined; assert.doesNotThrow(click);
});
