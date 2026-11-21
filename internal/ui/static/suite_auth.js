// The Homepage child bearer is verified by ScraperGo on every application API
// request. This bootstrap is always embedded, independently of telemetry.
(() => {
 'use strict';
 const homepage='https://home.themauris.org';
 const tokenKey='scrapergo-suite-handoff-v1';
 const tokenPattern=/^[a-f0-9]{64}$/;
 let bearer='',verified=false,verifying=false,verifiedUser='';
 let pendingReady=[];
 const status=document.getElementById('suite-auth-status');
 const form=document.getElementById('form');
 const input=document.getElementById('url');
 const read=document.getElementById('read');
 const dialog=document.createElement('dialog');
 dialog.className='suite-signin-dialog';
 dialog.setAttribute('aria-labelledby','suite-signin-title');
 dialog.innerHTML='<h2 id="suite-signin-title">Homepage sign-in required</h2>';
 const info=document.createElement('p');
 info.textContent='Sign in to Homepage with a personal account or choose an authenticated Guest session. A guest can read articles without personal settings.';
 const actions=document.createElement('div');actions.className='suite-signin-actions';
 const link=document.createElement('a');
 link.href=homepage+'/';link.textContent='Go to Homepage';
 actions.append(link);dialog.append(info,actions);
 dialog.addEventListener('cancel',e=>e.preventDefault());
 document.body.append(dialog);
 function setAccess(ok,showDialog=false){
  verified=ok;
  document.documentElement.classList.toggle('suite-auth-ready',ok);
  if(input)input.disabled=!ok;
  if(read)read.disabled=!ok;
  if(form)form.setAttribute('aria-disabled',String(!ok));
  if(ok){
   if(dialog.open)dialog.close();
   for(const resolve of pendingReady.splice(0))resolve();
  }else if(showDialog&&!dialog.open&&typeof dialog.showModal==='function')dialog.showModal();
 }
 function clearBearer(){
  bearer='';verifiedUser='';
  try{sessionStorage.removeItem(tokenKey)}catch(_){}
  setAccess(false,true);
 }
 function acceptHandoff(){
  try{
   const hash=new URLSearchParams(location.hash.slice(1));
   const delivered=hash.get('suite_token')||'';
   if(hash.has('suite_mode')||hash.has('suite_token'))history.replaceState(null,'',location.pathname+location.search);
   if(hash.get('suite_mode')==='local'){clearBearer();return false}
   if(tokenPattern.test(delivered)){
    bearer=delivered;sessionStorage.setItem(tokenKey,bearer);
   }else{bearer=sessionStorage.getItem(tokenKey)||''}
  }catch(_){bearer=''}
  if(!tokenPattern.test(bearer)){bearer='';return false}
  return true;
 }
 async function recover(){
  // Existing recovery contract; the browser never receives Homepage cookies
  // through ScraperGo. Same-site credentialed requests stay with Homepage.
  const response=await fetch(homepage+'/api/handoff/scrapergo-recover',{
   method:'POST',mode:'cors',credentials:'include',cache:'no-store'
  });
  if(!response.ok)return false;
  const data=await response.json();
  if(!tokenPattern.test(data.token))return false;
  bearer=data.token;
  try{sessionStorage.setItem(tokenKey,bearer)}catch(_){}
  return true;
 }
 async function checkIdentity(){
  if(verifying)return;
  verifying=true;
  try{
   if(!tokenPattern.test(bearer)&&!(await recover())){clearBearer();return}
   let response=await fetch('/api/suite/identity',{
    cache:'no-store',headers:{Authorization:'Bearer '+bearer}
   });
   if(response.status===401||response.status===403){
    // A revoked child may be replaced only through a valid Homepage parent.
    bearer='';
    if(!(await recover())){clearBearer();return}
    response=await fetch('/api/suite/identity',{
     cache:'no-store',headers:{Authorization:'Bearer '+bearer}
    });
   }
   if(!response.ok){clearBearer();return}
   const profile=await response.json();
   if(profile.state!=='signed-in'||typeof profile.user!=='string'||!profile.user||!['user','admin','guest'].includes(profile.role)){
    clearBearer();return
   }
   verifiedUser=profile.user;
   if(status&&typeof SuiteBannerStatus!=='undefined')SuiteBannerStatus.signedIn(profile.user);
   else if(status)status.textContent='Signed in: '+profile.user;
   setAccess(true);
  }catch(_){
   clearBearer();
   if(status)status.textContent='Homepage verification unavailable';
  }finally{verifying=false}
 }
 window.scrapergoSuiteAuth={
  isVerified:()=>verified&&tokenPattern.test(bearer),
  currentBearer:()=>verified?bearer:'',
  whenReady:()=>verified?Promise.resolve():new Promise(resolve=>pendingReady.push(resolve)),
  fetch:async (url,options={})=>{
   if(!verified||!tokenPattern.test(bearer))throw Error('Homepage sign-in or Guest required');
   const response=await fetch(url,{...options,
    headers:{...(options.headers||{}),Authorization:'Bearer '+bearer}
   });
   if([401,403,503].includes(response.status)){
    clearBearer();
    if(status)status.textContent=response.status===503?'Homepage verification unavailable':'Session expired or revoked';
    throw Error('Homepage session unavailable: sign in again');
   }
   return response;
  },
  refresh:checkIdentity
 };
 setAccess(false);
 // An explicit legacy local-mode handoff is not a Guest identity.
 const parsed=new URLSearchParams(location.hash.slice(1));
 const rejectedLocal=parsed.get('suite_mode')==='local';
 if(acceptHandoff()||!rejectedLocal)void checkIdentity();
 document.addEventListener('visibilitychange',()=>{if(!document.hidden)void checkIdentity()});
 window.addEventListener('focus',()=>{void checkIdentity()});
})();
