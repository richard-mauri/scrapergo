// CI-only behavioral regression for the embedded Homepage handoff.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const vm = require('node:vm');
const source = fs.readFileSync('internal/ui/static/suite_auth.js','utf8');
const valid = 'a'.repeat(64);
const settle = async () => { for(let i=0;i<6;i++) await new Promise(resolve=>setImmediate(resolve)); };
async function scenario({hash='',stored='',identity=200,recover=false}){
 const calls={opened:0,closed:0,checked:0,recovered:0};
 const data = new Map(stored?[['scrapergo-suite-handoff-v1',stored]]:[]);
 const elements = {
  'suite-auth-status':{textContent:''},
  form:{setAttribute(){}},
  url:{disabled:false},
  read:{disabled:false}
 };
 const dialog={open:false,className:'',setAttribute(){},innerHTML:'',
  append(){},addEventListener(){},
  showModal(){this.open=true;calls.opened++},
  close(){this.open=false;calls.closed++}
 };
 const document={
  body:{append(){}},
  documentElement:{classList:{toggle(){}}},
  createElement(tag){return tag==='dialog'?dialog:{textContent:'',className:'',href:'',append(){},setAttribute(){},addEventListener(){}}},
  getElementById(id){return elements[id]},
  addEventListener(){}
 };
 let finishIdentity;
 const identityResponse=new Promise(resolve=>finishIdentity=resolve);
 const fetch=(url)=>{
  if(String(url).includes('scrapergo-recover')){
   calls.recovered++;
   return Promise.resolve({ok:recover,json:async()=>({token:valid})});
  }
  if(url==='/api/suite/identity'){calls.checked++;return identityResponse;}
  throw Error('unexpected fetch '+url);
 };
 const context={
  document,fetch,
  window:{addEventListener(){}},
  location:{hash,pathname:'/',search:''},
  history:{replaceState(){}},
  sessionStorage:{getItem:k=>data.get(k)||null,setItem:(k,v)=>data.set(k,v),removeItem:k=>data.delete(k)},
  URLSearchParams,
  SuiteBannerStatus:{signedIn(){}}
 };
 vm.runInNewContext(source,context,{filename:'suite_auth.js'});
 assert.equal(calls.opened,0,'startup must not display authentication dialog');
 await settle();
 if(calls.checked===0){
  assert.equal(calls.opened,1,'failed recovery must produce a persistent modal');
  return calls;
 }
 assert.equal(calls.opened,0,'verification pending must not display the dialog');
 finishIdentity({ok:identity===200,status:identity,json:async()=>({state:'signed-in',user:'Tester',role:'user'})});
 await settle();
 assert.equal(calls.opened,identity===200?0:1,'dialog must match final verification result');
 assert.equal(dialog.open,identity!==200,'failure dialog must remain visible');
 return calls;
}
(async()=>{
 const good=await scenario({hash:'#suite_token='+valid,identity:200});
 assert.equal(good.checked,1);
 await scenario({hash:'#suite_token='+valid,identity:503});
 const restored=await scenario({recover:true});
 assert.equal(restored.recovered,1);
 const absent=await scenario({});
 assert.equal(absent.recovered,1);
 console.log('ScraperGo sign-in modal lifecycle regressions passed');
})().catch(e=>{console.error(e);process.exitCode=1});
