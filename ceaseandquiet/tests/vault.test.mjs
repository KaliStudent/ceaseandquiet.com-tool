import test from 'node:test';
import assert from 'node:assert/strict';
import {encryptProfile, decryptProfile, validateProfile, VaultSession} from '../extension/vault.mjs';

const profile = {givenName: 'Taylor', familyName: 'Example', email: 'taylor@example.test', region: 'Test region'};
const password = 'unique test phrase 247!';
const envelope = await encryptProfile(profile, password);

test('authenticated encryption round trip stores no profile or passphrase', async () => {
  assert.deepEqual(await decryptProfile(envelope, password), profile);
  const serialized = JSON.stringify(envelope);
  for (const value of [...Object.values(profile), password]) assert.ok(!serialized.includes(value));
  const second = await encryptProfile(profile, password);
  assert.notEqual(second.salt, envelope.salt);
  assert.notEqual(second.iv, envelope.iv);
  assert.notEqual(second.ciphertext, envelope.ciphertext);
});
test('wrong passwords and altered authenticated data are rejected', async () => {
  await assert.rejects(decryptProfile(envelope, 'incorrect test password'));
  for (const field of ['ciphertext', 'iv', 'salt']) {
    const altered = {...envelope};
    const bytes = Buffer.from(altered[field], 'base64'); bytes[0] ^= 1; altered[field] = bytes.toString('base64');
    await assert.rejects(decryptProfile(altered, password));
  }
});
test('malformed, oversized, legacy, and weakened envelope settings fail closed', async () => {
  for (const change of [{version: 1}, {iterations: 1}, {cipher: 'AES-CBC'}, {kdf: 'other'}, {salt: '!!'}, {iv: 'AA=='}, {ciphertext: 'A'.repeat(20000)}, {unexpected: 'field'}]) {
    await assert.rejects(decryptProfile({...envelope, ...change}, password));
  }
  await assert.rejects(encryptProfile(profile, 'short'));
});
test('profile schema rejects unsupported fields and unsafe or oversized values', () => {
  for (const value of [{...profile, ssn: 'synthetic'}, {...profile, givenName: ''}, {...profile, email: 'invalid'}, {...profile, region: 'x'.repeat(101)}, {...profile, givenName: 'bad\nname'}]) assert.throws(() => validateProfile(value));
});
test('session expires at its absolute deadline and returns defensive copies', async () => {
  let now = 0;
  const session = new VaultSession({now: () => now, duration: 50, decrypt: async () => profile});
  await session.unlock({}, password);
  const copy = session.snapshot(); copy.givenName = 'changed';
  assert.equal(session.snapshot().givenName, 'Taylor');
  now = 50; assert.equal(session.snapshot(), null);
});
test('locking while unlock is pending prevents late plaintext reveal', async () => {
  let finish;
  const session = new VaultSession({decrypt: () => new Promise(resolve => { finish = resolve; })});
  const unlock = session.unlock({}, password);
  session.lock(); finish(profile);
  await assert.rejects(unlock, /cancelled/);
  assert.equal(session.snapshot(), null);
});
test('a failed unlock clears an earlier unlocked session', async () => {
  let fail = false;
  const session = new VaultSession({decrypt: async () => { if (fail) throw new Error('failed'); return profile; }});
  await session.unlock({}, password); fail = true;
  await assert.rejects(session.unlock({}, password));
  assert.equal(session.snapshot(), null);
});
