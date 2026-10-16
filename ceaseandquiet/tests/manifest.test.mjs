import test from 'node:test';
import assert from 'node:assert/strict';
import {readFile} from 'node:fs/promises';

test('extension ships without permanent site access, background execution, or network access', async () => {
  const manifest = JSON.parse(await readFile(new URL('../extension/manifest.json', import.meta.url), 'utf8'));
  assert.equal(manifest.manifest_version, 3);
  assert.deepEqual([...manifest.permissions].sort(), ['activeTab', 'scripting', 'storage']);
  for (const property of ['host_permissions', 'optional_host_permissions', 'content_scripts', 'background', 'externally_connectable', 'web_accessible_resources']) assert.equal(manifest[property], undefined);
  const policy = manifest.content_security_policy.extension_pages;
  assert.ok(policy.includes("connect-src 'none'"));
  assert.ok(policy.includes("form-action 'none'"));
  assert.ok(!policy.includes('unsafe-eval') && !policy.includes('unsafe-inline'));
});
