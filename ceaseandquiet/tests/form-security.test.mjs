import test from 'node:test';
import assert from 'node:assert/strict';
import {attachFormSecurity} from '../site/form-security.mjs';

function element() { return {value:'',validity:{valid:true},listeners:{},addEventListener(name,fn){this.listeners[name]=fn;}}; }
function ui(){
  const email=element(),website=element(), form=element();form.elements={namedItem:name=>name==='email'?email:website};
  const controls=Object.fromEntries(['container','status','emailStatus','code','send','confirm'].map(name=>[name,element()]));
  return {form,email,website,...controls,clearSuggestions(){}};
}
test('website bot check creates a server session and email changes clear verification status',async()=>{
  const oldFetch=globalThis.fetch,oldDoc=globalThis.document,oldWin=globalThis.window;
  let script,options;const calls=[];
  globalThis.document={createElement(){return {};},head:{append(item){script=item;}}};
  globalThis.window={turnstile:{render(container,opts){options=opts;}}};
  globalThis.fetch=async(url,opts)=>{
    calls.push({url,opts});
    if(url.endsWith('/config'))return {ok:true,json:async()=>({botEnabled:true,emailEnabled:true,siteKey:'synthetic-public-site-key'})};
    return {ok:true,json:async()=>({ok:true,verified:true})};
  };
  try{
    const page=ui();await attachFormSecurity(page);
    assert.equal(page.send.disabled,true);script.onload();assert.equal(options.theme,'dark');assert.equal(options.action,'profile_form');
    await options.callback('synthetic-token');assert.equal(page.send.disabled,false);
    page.email.value='taylor@example.test';await page.send.listeners.click();
    const request=calls.find(item=>item.url.endsWith('/email/start'));
    assert.deepEqual(JSON.parse(request.opts.body),{email:'taylor@example.test',website:''});
    assert.equal(request.opts.referrerPolicy,'same-origin');
    page.code.value='123456';await page.confirm.listeners.click();assert.match(page.emailStatus.textContent,/Email confirmed/);
    page.email.value='different@example.test';page.email.listeners.input();assert.match(page.emailStatus.textContent,/not verified/);assert.equal(page.code.value,'');
    options['expired-callback']();assert.equal(page.send.disabled,true);
  }finally{globalThis.fetch=oldFetch;globalThis.document=oldDoc;globalThis.window=oldWin;}
});

test('website does not offer email sending when SMTP is not configured',async()=>{
  const original=globalThis.fetch;
  globalThis.fetch=async()=>({ok:true,json:async()=>({botEnabled:true,humanVerified:true,emailEnabled:false,siteKey:'synthetic-public-site-key'})});
  try{const page=ui();await attachFormSecurity(page);assert.equal(page.send.disabled,true);assert.equal(page.confirm.disabled,true);assert.match(page.emailStatus.textContent,/mail connection/);}
  finally{globalThis.fetch=original;}
});
