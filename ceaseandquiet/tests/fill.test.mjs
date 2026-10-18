import test from 'node:test';
import assert from 'node:assert/strict';
import {runInNewContext} from 'node:vm';
import {fillPreview, previewForm} from '../extension/fill.mjs';

const profile = {givenName: 'Taylor', familyName: 'Example', email: 'taylor@example.test', address: '100 Example Street', city: 'Example City', region: 'Test State', postalCode: '00000', dateOfBirth: '2000-01-02', phone: '+1 202-555-0100'};
function mockChrome(url = 'http://127.0.0.1:8787/#vault') {
  const calls = [];
  return {calls, tabs: {query: async () => [{id: 7, url}]}, scripting: {executeScript: async input => {
    calls.push(input); return [{documentId: 'approved-document', result: {ok: true}}];
  }}};
}
test('preflight carries no profile; write targets the approved document only', async () => {
  for (const url of ['http://127.0.0.1:8787/#vault', 'https://abcw.xyz:8443/#vault']) {
    const api = mockChrome(url); await fillPreview(api, () => profile);
    assert.equal(api.calls.length, 2);
    assert.equal(api.calls[0].args, undefined);
    assert.deepEqual(api.calls[0].target.frameIds, [0]);
    assert.deepEqual(api.calls[1].target.documentIds, ['approved-document']);
    assert.deepEqual(api.calls[1].args, [profile]);
  }
});

test('serialized preflight rejects insecure, lookalike, wrong-port and framed staging destinations', () => {
  class Input {
    value = ''; disabled = false; readOnly = false;
    constructor(type) { this.type = type; }
    getClientRects() { return [{}]; }
  }
  const inputs = Object.fromEntries(Object.keys(profile).map(name => [name, new Input(({email: 'email', dateOfBirth: 'date', phone: 'tel'}[name] || 'text'))]));
  const sandbox = {
    window: {}, location: {origin: 'https://abcw.xyz:8443', pathname: '/', search: ''},
    HTMLInputElement: Input,
    document: {querySelectorAll: () => [{elements: {namedItem: name => inputs[name]}}]}
  };
  sandbox.window.top = sandbox.window;
  const probe = () => runInNewContext('(' + previewForm.toString() + ')()', sandbox);
  assert.equal(probe().ok, true);
  for (const origin of ['http://abcw.xyz:8443', 'https://abcw.xyz', 'https://abcw.xyz:8444', 'https://abcw.xyz.evil.test:8443']) {
    sandbox.location.origin = origin; assert.equal(probe().ok, false);
  }
  sandbox.location.origin = 'https://abcw.xyz:8443';
  sandbox.window.top = {}; assert.equal(probe().ok, false);
  sandbox.window.top = sandbox.window;
  inputs.email.value = 'occupied@example.test'; assert.equal(probe().ok, false);
  assert.equal(inputs.email.value, 'occupied@example.test');
});
test('unsupported origins, credential URLs, queries and paths never receive an injection', async () => {
  for (const url of ['https://ceaseandquiet.com/', 'https://unrelated.example/', 'http://127.0.0.1:8787.evil.test/', 'http://127.0.0.1:8787/other', 'http://127.0.0.1:8787/?token=sample', 'http://user@127.0.0.1:8787/', 'http://abcw.xyz:8443/', 'https://abcw.xyz/', 'https://abcw.xyz:8444/', 'https://abcw.xyz:8443.evil.test/', 'https://abcw.xyz:8443/other', 'https://abcw.xyz:8443/?token=sample', 'https://user@abcw.xyz:8443/', 'chrome://extensions/']) {
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
  const inputs = names.map(name => new Input(({email: 'email', dateOfBirth: 'date', phone: 'tel'}[name] || 'text')));
  globalThis.HTMLInputElement = Input;
  globalThis.window = {}; window.top = window;
  globalThis.location = {origin: 'http://127.0.0.1:8787', pathname: '/'};
  globalThis.document = {querySelectorAll: () => [{elements: {namedItem: name => inputs[names.indexOf(name)]}}]};
  try {
    inputs[3].disabled = true;
    assert.equal(previewForm(profile).ok, false); assert.ok(inputs.every(input => input.value === ''));
    inputs[3].disabled = false;
    location.origin = 'https://unrelated.example'; assert.equal(previewForm(profile).ok, false);
    location.origin = 'https://abcw.xyz:8443'; assert.equal(previewForm().ok, true);
    location.pathname = '/other'; assert.equal(previewForm(profile).ok, false);
    location.pathname = '/'; location.search = '?token=sample'; assert.equal(previewForm(profile).ok, false);
    location.search = '';
    location.origin = 'http://127.0.0.1:8787'; assert.equal(previewForm().ok, true);
    assert.equal(previewForm(profile).ok, true); assert.deepEqual(inputs.map(input => input.value), Object.values(profile));
    assert.equal(previewForm(profile).ok, false); // Never overwrite a filled form.
  } finally { for (const key of ['window','location','document','HTMLInputElement']) delete globalThis[key]; }
});
