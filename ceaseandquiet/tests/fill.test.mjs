import test from 'node:test';
import assert from 'node:assert/strict';
import {fillPreview, previewForm} from '../extension/fill.mjs';

const profile = {givenName: 'Taylor', familyName: 'Example', email: 'taylor@example.test', region: 'Test region'};
function mockChrome(url = 'http://127.0.0.1:8787/#vault') {
  const calls = [];
  return {calls, tabs: {query: async () => [{id: 7, url}]}, scripting: {executeScript: async input => {
    calls.push(input); return [{documentId: 'approved-document', result: {ok: true}}];
  }}};
}
test('preflight carries no profile; write targets the approved document only', async () => {
  const api = mockChrome(); await fillPreview(api, () => profile);
  assert.equal(api.calls.length, 2);
  assert.equal(api.calls[0].args, undefined);
  assert.deepEqual(api.calls[0].target.frameIds, [0]);
  assert.deepEqual(api.calls[1].target.documentIds, ['approved-document']);
  assert.deepEqual(api.calls[1].args, [profile]);
});
test('unsupported origins, credential URLs, queries and paths never receive an injection', async () => {
  for (const url of ['https://ceaseandquiet.com/', 'https://unrelated.example/', 'http://127.0.0.1:8787.evil.test/', 'http://127.0.0.1:8787/other', 'http://127.0.0.1:8787/?token=sample', 'http://user@127.0.0.1:8787/', 'chrome://extensions/']) {
    const api = mockChrome(url); await assert.rejects(fillPreview(api, () => profile)); assert.equal(api.calls.length, 0);
  }
});
test('failed preflight and expired session never transmit profile', async () => {
  const api = mockChrome();
  api.scripting.executeScript = async input => { api.calls.push(input); return [{result: {ok: false}}]; };
  await assert.rejects(fillPreview(api, () => profile)); assert.equal(api.calls.length, 1);
  const expired = mockChrome(); await assert.rejects(fillPreview(expired, () => null)); assert.equal(expired.calls.length, 1);
});
test('injected function rechecks origin and validates all fields before writing', () => {
  class Input {
    value = ''; disabled = false; readOnly = false;
    constructor(type = 'text') { this.type = type; }
    getClientRects() { return [{}]; }
    dispatchEvent() {}
  }
  const names = Object.keys(profile);
  const inputs = names.map(name => new Input(name === 'email' ? 'email' : 'text'));
  globalThis.HTMLInputElement = Input;
  globalThis.window = {}; window.top = window;
  globalThis.location = {origin: 'http://127.0.0.1:8787', pathname: '/'};
  globalThis.document = {querySelectorAll: () => [{elements: {namedItem: name => inputs[names.indexOf(name)]}}]};
  try {
    inputs[3].disabled = true;
    assert.equal(previewForm(profile).ok, false); assert.ok(inputs.every(input => input.value === ''));
    inputs[3].disabled = false;
    location.origin = 'https://unrelated.example'; assert.equal(previewForm(profile).ok, false);
    location.origin = 'http://127.0.0.1:8787'; assert.equal(previewForm().ok, true);
    assert.equal(previewForm(profile).ok, true); assert.deepEqual(inputs.map(input => input.value), Object.values(profile));
    assert.equal(previewForm(profile).ok, false); // Never overwrite a filled form.
  } finally { for (const key of ['window','location','document','HTMLInputElement']) delete globalThis[key]; }
});
