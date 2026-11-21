// Optional aggregate activity reporting only. Homepage authentication is
// mandatory and handled independently by the embedded suite_auth.js bootstrap.
(() => {
 'use strict';
 const auth=window.scrapergoSuiteAuth;
 if(!auth)return;
 const key='scrapergo-activity-tab-v1';
 let id='';
 try{
  id=sessionStorage.getItem(key)||'';
  if(!/^[a-f0-9]{32}$/.test(id)){
   id=Array.from(crypto.getRandomValues(new Uint8Array(16)),b=>b.toString(16).padStart(2,'0')).join('');
   sessionStorage.setItem(key,id);
  }
 }catch(_){id=Array.from(crypto.getRandomValues(new Uint8Array(16)),b=>b.toString(16).padStart(2,'0')).join('')}
 let initialIdentityChecked=false,interacted=false,sentAt=0;
 function send(){
  if(!initialIdentityChecked||document.hidden||!auth.isVerified())return;
  const now=Date.now();
  if(now-sentAt<25000)return;
  sentAt=now;
  const body=JSON.stringify({id,visible:true,interacted});
  interacted=false;
  try{
   void auth.fetch('/api/activity/ping',{method:'POST',cache:'no-store',
    headers:{'Content-Type':'application/json'},body,keepalive:true}).catch(()=>{});
  }catch(_){}
 }
 function mark(){interacted=true;send()}
 document.addEventListener('pointerdown',mark,{passive:true});
 document.addEventListener('keydown',mark,{passive:true});
 document.addEventListener('visibilitychange',()=>{if(!document.hidden)send()});
 window.addEventListener('pageshow',send);
 window.addEventListener('focus',send);
 void auth.whenReady().then(()=>{initialIdentityChecked=true;send()});
 setInterval(send,60000);
})();
