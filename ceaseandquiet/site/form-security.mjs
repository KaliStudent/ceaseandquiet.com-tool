export async function attachFormSecurity({form, container, status, emailStatus, code, send, confirm, clearSuggestions, onReady = () => {}}) {
  const email = form.elements.namedItem('email');
  let human = false, emailEnabled = false, generation = 0, busy = false;
  const post = async (path, body) => {
    const response = await fetch('/api/v1/form-security/' + path, {method:'POST',mode:'same-origin',credentials:'same-origin',referrerPolicy:'same-origin',cache:'no-store',headers:{'Content-Type':'application/json'},body:JSON.stringify({...body,website:form.elements.namedItem('website').value}),signal:AbortSignal.timeout(15000)});
    if (!response.ok) throw new Error(response.status === 429 ? 'Please wait before trying again.' : response.status === 503 ? 'Email delivery is not connected yet.' : 'Verification failed. Refresh the bot check or request a new code.');
    return response.json();
  };
  const render = () => { send.disabled = !human || !emailEnabled || busy; confirm.disabled = !human || !emailEnabled || busy; };
  email.addEventListener('input', () => { generation++; code.value=''; emailStatus.textContent='Email not verified. Verify the address you intend to use.'; });
  form.addEventListener('reset', () => { generation++; code.value=''; emailStatus.textContent='Enter an email address to check or verify it.'; });
  form.addEventListener('profileclear', () => { generation++; code.value=''; emailStatus.textContent='Enter an email address to check or verify it.'; });
  form.addEventListener('profilechange', async () => {
    const current=++generation;code.value='';emailStatus.textContent='Checking email verification…';
    if(!human || !emailEnabled){emailStatus.textContent=emailEnabled?'Complete the bot check before verifying email.':'Email verification is awaiting the site mail connection.';return;}
    try{const result=await post('email/status',{email:email.value.trim()});if(current===generation)emailStatus.textContent=result.verified?'Email confirmed for this browser session.':'Email not verified. Verify the address you intend to use.';}
    catch{if(current===generation)emailStatus.textContent='Email verification could not be checked. Verify the address again.';}
  });
  send.addEventListener('click', async () => {
    if (busy || !human || !emailEnabled) return;
    if (!email.validity.valid || !email.value.trim()) { emailStatus.textContent='Enter a valid email address first.'; return; }
    busy=true; render(); const current=generation;
    try { await post('email/start',{email:email.value.trim()}); if(current===generation) emailStatus.textContent='Check that inbox for a six-digit code. It expires in 10 minutes. You can request another after one minute.'; }
    catch(error){if(current===generation) emailStatus.textContent=error.message;}
    finally {busy=false;render();}
  });
  confirm.addEventListener('click', async () => {
    if(busy || !human || !emailEnabled) return;
    busy=true;render(); const current=generation;
    try {await post('email/confirm',{email:email.value.trim(),code:code.value.trim()});if(current===generation){code.value='';emailStatus.textContent='Email confirmed for this browser session. Changing the email requires verification again.';}}
    catch(error){if(current===generation)emailStatus.textContent=error.message;}
    finally{busy=false;render();}
  });
  try {
    const response=await fetch('/api/v1/form-security/config',{cache:'no-store',credentials:'same-origin'});
    if(!response.ok)throw new Error('Bot check unavailable. Refresh this page.');
    const config=await response.json();emailEnabled=config.emailEnabled;
    if(!emailEnabled)emailStatus.textContent='Email verification is awaiting the site mail connection. No requests can be sent yet.';
    if(!config.botEnabled){status.textContent='Bot check is awaiting site configuration. Request limits are active.';render();return;}
    const ready=()=>{human=true;status.textContent='Bot check complete. You can use address suggestions.';render();onReady();};
    if(config.humanVerified){ready();return;}
    const script=document.createElement('script');script.src='https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit';script.async=true;
    script.onerror=()=>{status.textContent='Could not load the bot check. Please check your connection and refresh.';};
    script.onload=()=>window.turnstile.render(container,{sitekey:config.siteKey,action:'profile_form',theme:'dark',callback:async token=>{
      try{await post('session',{token});ready();}
      catch{human=false;status.textContent='Bot check could not be confirmed. Refresh this page and try again.';render();}
    },'expired-callback':()=>{human=false;clearSuggestions();status.textContent='Bot check expired. Complete the check again.';render();},'error-callback':()=>{human=false;status.textContent='Bot check unavailable. Please refresh and try again.';render();}});
    document.head.append(script);status.textContent='Complete the bot check to use address suggestions and email verification.';
  }catch(error){status.textContent=error.message;}
  render();
}
