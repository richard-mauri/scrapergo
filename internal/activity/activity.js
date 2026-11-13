// ScraperGo browser presence, no personal tokens or article URLs are transmitted.
// A random per-tab identifier is ephemeral and origin-scoped.
(() => {
  let id='';
  const key='scrapergo-activity-tab-v1';
  try {
    id=sessionStorage.getItem(key)||'';
    if (!/^[a-f0-9]{32}$/.test(id)) {
      id=Array.from(crypto.getRandomValues(new Uint8Array(16)),b=>b.toString(16).padStart(2,'0')).join('');
      sessionStorage.setItem(key,id);
    }
  } catch (_) {
    id=Array.from(crypto.getRandomValues(new Uint8Array(16)),b=>b.toString(16).padStart(2,'0')).join('');
  }
  // Accept a Homepage-issued child handoff without changing public reader access.
  // Fragments are removed before any article/network request.
  const handoffKey='scrapergo-suite-handoff-v1';
  let bearer='';
  try {
    const hash=new URLSearchParams(location.hash.slice(1));
    if(hash.has('suite_mode') && hash.get('suite_mode')==='local') {
      sessionStorage.removeItem(handoffKey);
    } else if(hash.has('suite_token')) {
      const token=hash.get('suite_token')||'';
      if(/^[a-f0-9]{64}$/.test(token))sessionStorage.setItem(handoffKey,token);
    }
    if(hash.has('suite_mode')||hash.has('suite_token'))
      history.replaceState(null,'',location.pathname+location.search);
    bearer=sessionStorage.getItem(handoffKey)||'';
    if(!/^[a-f0-9]{64}$/.test(bearer))bearer='';
  }catch(_){bearer=''}
  const statusElement=document.getElementById('suite-auth-status');
  let statusPending=false;
  // Only a tab that was handed a Homepage credential attempts background
  // recovery. Direct public visits remain guests.
  let hadHandoff=!!bearer;
  const homepageRecoverURL='https://home.themauris.org/api/handoff/scrapergo-recover';
  async function recoverHandoff() {
    const response=await fetch(homepageRecoverURL,{
      method:'POST',mode:'cors',credentials:'include',cache:'no-store'
    });
    if(response.status===401 || response.status===403)return 'expired';
    if(!response.ok)return 'unavailable';
    const data=await response.json();
    if(!/^[a-f0-9]{64}$/.test(data.token)||!Number.isFinite(Date.parse(data.expiresAt)))
      return 'unavailable';
    bearer=data.token;
    try{sessionStorage.setItem(handoffKey,bearer)}catch(_){}
    return 'recovered';
  }
  async function identityResult() {
    const response=await fetch('/api/activity/identity',{
      method:'GET',cache:'no-store',credentials:'same-origin',
      headers:{Authorization:'Bearer '+bearer}
    });
    if(!response.ok)return {state:'unavailable'};
    return response.json();
  }
  async function updateIdentityStatus() {
    if(!statusElement || statusPending)return;
    if(!hadHandoff) {statusElement.textContent='Guest — not signed in';return;}
    statusPending=true;
    try {
      statusElement.textContent='Checking sign-in…';
      let identity=bearer?await identityResult():{state:'guest'};
      if(identity.state==='guest'){
        // Never navigate: a same-site credentialed request to Homepage
        // obtains a new, parent-bound child bearer if the login survives.
        const recovery=await recoverHandoff();
        if(recovery==='expired'){
          bearer='';
          try{sessionStorage.removeItem(handoffKey)}catch(_){}
          statusElement.textContent='Session expired — Sign in again';
          return;
        }
        if(recovery!=='recovered'){
          statusElement.textContent='Unable to verify sign-in';
          return;
        }
        identity=await identityResult();
      }
      if(identity.state==='signed-in' && typeof identity.user==='string' && identity.user){
        statusElement.textContent='Signed in as '+identity.user;
      }else{
        statusElement.textContent='Unable to verify sign-in';
      }
    }catch(_){statusElement.textContent='Unable to verify sign-in'}
    finally{statusPending=false}
  }
  void updateIdentityStatus();
  setInterval(()=>{if(!document.hidden)void updateIdentityStatus()},60000);
  document.addEventListener('visibilitychange',()=>{if(!document.hidden)void updateIdentityStatus()});
  let interacted=false;
  let sentAt=0;
  function send() {
    if (document.hidden)return;
    const now=Date.now();
    if (now-sentAt<25000)return;
    sentAt=now;
    const body=JSON.stringify({id,visible:true,interacted});
    interacted=false;
    try {
      const headers={'Content-Type':'application/json'};
      if(bearer)headers.Authorization='Bearer '+bearer;
      fetch('/api/activity/ping',{method:'POST',cache:'no-store',credentials:'same-origin',
        headers,body,keepalive:true}).catch(()=>{});
    }catch(_){}
  }
  function mark() {interacted=true;send()}
  document.addEventListener('pointerdown',mark,{passive:true});
  document.addEventListener('keydown',mark,{passive:true});
  document.addEventListener('visibilitychange',()=>{if(!document.hidden)send()});
  window.addEventListener('pageshow',send);
  window.addEventListener('focus',send);
  send();
  setInterval(send,60000);
})();
