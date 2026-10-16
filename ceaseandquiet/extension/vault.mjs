export const FIELDS = Object.freeze(['givenName', 'familyName', 'email', 'region']);
export const ITERATIONS = 600000;
const CONTEXT = new TextEncoder().encode('CeaseAndQuiet:local-vault:v1');
const bytes = value => new TextEncoder().encode(value);
const base64 = value => btoa(String.fromCharCode(...new Uint8Array(value)));

function fromBase64(value, min, max) {
  if (typeof value !== 'string' || value.length > max * 2 || !/^[A-Za-z0-9+/]+={0,2}$/.test(value)) throw new Error('Invalid vault');
  const result = Uint8Array.from(atob(value), char => char.charCodeAt(0));
  if (result.length < min || result.length > max || base64(result) !== value) throw new Error('Invalid vault');
  return result;
}

export function validateProfile(input) {
  if (!input || typeof input !== 'object' || Array.isArray(input) || Object.keys(input).length !== FIELDS.length) throw new Error('Invalid profile');
  const profile = {};
  for (const name of FIELDS) {
    if (!Object.hasOwn(input, name) || typeof input[name] !== 'string' || input[name].length > (name === 'email' ? 254 : 100) || /[\u0000-\u001f\u007f]/.test(input[name])) throw new Error('Invalid profile');
    profile[name] = input[name].trim();
  }
  if (!profile.givenName || !profile.familyName || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(profile.email)) throw new Error('Enter a first name, last name, and valid email.');
  return profile;
}

async function derive(passphrase, salt) {
  if (typeof passphrase !== 'string' || passphrase.length < 12 || passphrase.length > 1024) throw new Error('Use a passphrase of 12–1024 characters.');
  const material = await crypto.subtle.importKey('raw', bytes(passphrase), 'PBKDF2', false, ['deriveKey']);
  return crypto.subtle.deriveKey({name: 'PBKDF2', salt, iterations: ITERATIONS, hash: 'SHA-256'}, material, {name: 'AES-GCM', length: 256}, false, ['encrypt', 'decrypt']);
}

export async function encryptProfile(input, passphrase) {
  const profile = validateProfile(input);
  const salt = crypto.getRandomValues(new Uint8Array(16));
  const iv = crypto.getRandomValues(new Uint8Array(12));
  const key = await derive(passphrase, salt);
  const plaintext = bytes(JSON.stringify(profile));
  try {
    const ciphertext = await crypto.subtle.encrypt({name: 'AES-GCM', iv, additionalData: CONTEXT}, key, plaintext);
    return {version: 2, kdf: 'PBKDF2-SHA256', iterations: ITERATIONS, cipher: 'AES-256-GCM', salt: base64(salt), iv: base64(iv), ciphertext: base64(ciphertext)};
  } finally { plaintext.fill(0); }
}

export async function decryptProfile(envelope, passphrase) {
  if (!envelope || typeof envelope !== 'object' || Object.keys(envelope).length !== 7 || envelope.version !== 2 || envelope.kdf !== 'PBKDF2-SHA256' || envelope.iterations !== ITERATIONS || envelope.cipher !== 'AES-256-GCM') throw new Error('Invalid or older vault. Erase it and create a new one.');
  const salt = fromBase64(envelope.salt, 16, 16);
  const iv = fromBase64(envelope.iv, 12, 12);
  const ciphertext = fromBase64(envelope.ciphertext, 17, 4096);
  const key = await derive(passphrase, salt);
  const plaintext = new Uint8Array(await crypto.subtle.decrypt({name: 'AES-GCM', iv, additionalData: CONTEXT}, key, ciphertext));
  try { return validateProfile(JSON.parse(new TextDecoder('utf-8', {fatal: true}).decode(plaintext))); }
  finally { plaintext.fill(0); }
}

// Recheck the absolute deadline on every read, even if browser timers pause.
export class VaultSession {
  #profile = null;
  #deadline = 0;
  #generation = 0;
  constructor({now = () => Date.now(), duration = 120000, decrypt = decryptProfile} = {}) {
    this.now = now;
    this.duration = duration;
    this.decrypt = decrypt;
  }
  lock() { this.#generation++; this.#profile = null; this.#deadline = 0; }
  async unlock(envelope, passphrase) {
    this.lock();
    const generation = this.#generation;
    const profile = validateProfile(await this.decrypt(envelope, passphrase));
    if (generation !== this.#generation) throw new Error('Unlock cancelled');
    this.#profile = profile;
    this.#deadline = this.now() + this.duration;
  }
  snapshot() {
    if (!this.#profile || this.now() >= this.#deadline) { this.lock(); return null; }
    return {...this.#profile};
  }
}
