const $ = id => document.getElementById(id);
const fields = ['givenName','familyName','email','region'];
let unlocked = null;
const encode = text => new TextEncoder().encode(text);
const decode = bytes => new TextDecoder().decode(bytes);
const b64 = bytes => btoa(String.fromCharCode(...new Uint8Array(bytes)));
const unb64 = text => Uint8Array.from(atob(text), c => c.charCodeAt(0));
const message = text => { $('state').textContent = text; };
const profile = () => Object.fromEntries(fields.map(name => [name, $(name).value.trim()]));
const show = data => fields.forEach(name => { $(name).value = data[name] || ''; });
async function key(passphrase, salt) {
  const material = await crypto.subtle.importKey('raw', encode(passphrase), 'PBKDF2', false, ['deriveKey']);
  return crypto.subtle.deriveKey({name:'PBKDF2',salt,iterations:600000,hash:'SHA-256'},material,{name:'AES-GCM',length:256},false,['encrypt','decrypt']);
}
async function save() {
  if ($('passphrase').value.length < 12) { message('Use a passphrase of at least 12 characters.'); return; }
  const data = profile();
  if (!data.givenName || !data.familyName || !data.email) { message('First name, last name and email are required.'); return; }
  const salt = crypto.getRandomValues(new Uint8Array(16));
  const iv = crypto.getRandomValues(new Uint8Array(12));
  const secret = await key($('passphrase').value,salt);
  const ciphertext = await crypto.subtle.encrypt({name:'AES-GCM',iv},secret,encode(JSON.stringify(data)));
  await chrome.storage.local.set({vault:{version:1,kdf:'PBKDF2-SHA256',iterations:600000,cipher:'AES-256-GCM',salt:b64(salt),iv:b64(iv),ciphertext:b64(ciphertext)}});
  unlocked = data; $('passphrase').value = ''; message('Encrypted vault saved locally.');
}
async function unlock() {
  const {vault} = await chrome.storage.local.get('vault');
  if (!vault || vault.version !== 1) { message('No compatible vault saved.'); return; }
  try {
    const secret = await key($('passphrase').value,unb64(vault.salt));
    unlocked = JSON.parse(decode(await crypto.subtle.decrypt({name:'AES-GCM',iv:unb64(vault.iv)},secret,unb64(vault.ciphertext))));
    show(unlocked); $('passphrase').value = ''; message('Unlocked for this popup session.');
  } catch { unlocked = null; message('Could not unlock. Check your passphrase.'); }
}
async function fill() {
  if (!unlocked) { message('Unlock the vault first.'); return; }
  const [tab] = await chrome.tabs.query({active:true,currentWindow:true});
  if (!tab?.url) { message('Open the Cease & Quiet preview form first.'); return; }
  const url = new URL(tab.url);
  if (!['https://ceaseandquiet.com','http://127.0.0.1:8787'].includes(url.origin)) { message('This page is not an approved preview origin.'); return; }
  const injection = await chrome.scripting.executeScript({target:{tabId:tab.id},func:data=>{
    const form = document.querySelector('#cq-demo-form');
    if (!form) return false;
    for (const [name,value] of Object.entries(data)) {
      const input = form.elements.namedItem(name);
      if (!(input instanceof HTMLInputElement)) return false;
      input.value = value;
      input.dispatchEvent(new Event('input',{bubbles:true}));
    }
    return true;
  },args:[unlocked]});
  message(injection[0]?.result ? 'Filled the local preview form. Nothing was submitted.' : 'Preview form not found.');
}
async function erase() {
  await chrome.storage.local.remove('vault');
  unlocked = null; show({}); $('passphrase').value = '';
  message('Saved vault deleted from extension storage.');
}
for (const [id,fn] of [['save',save],['unlock',unlock],['fill',fill],['erase',erase]]) $(id).addEventListener('click',()=>fn().catch(()=>message('Action failed. No form was submitted.')));
chrome.storage.local.get('vault').then(({vault})=>message(vault?'Vault saved. Enter passphrase to unlock.':'No vault saved. Enter details and encrypt.'));
