import test from 'node:test';
import assert from 'node:assert/strict';
import {readFile} from 'node:fs/promises';
import {FIELDS} from '../extension/vault.mjs';

test('both forms expose the exact encrypted profile fields without duplicate IDs', async () => {
  const popup = await readFile(new URL('../extension/popup.html', import.meta.url), 'utf8');
  const website = await readFile(new URL('../site/index.html', import.meta.url), 'utf8');
  const ids = [...popup.matchAll(/\bid="([^"]+)"/g)].map(match => match[1]);
  assert.equal(new Set(ids).size, ids.length);
  for (const name of FIELDS) {
    assert.match(popup, new RegExp('<input[^>]*id="' + name + '"'));
    assert.match(website, new RegExp('<input[^>]*name="' + name + '"'));
  }
  for (const html of [popup, website]) {
    assert.match(html, /type="date"/); assert.match(html, /type="tel"/);
    assert.match(html, /id="address-enabled" type="checkbox">/);
    assert.doesNotMatch(html, /id="address-enabled"[^>]*checked/);
  }
});
