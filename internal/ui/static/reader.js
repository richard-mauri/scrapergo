
let current=null;const $=id=>document.getElementById(id);const form=$('form');function put(tag,text,parent){const e=document.createElement(tag);e.textContent=text;parent.appendChild(e);return e}function articleLinkHref(u){return /\.(?:pdf|docx?|xlsx?|zip)$/i.test(u.pathname)?u.href:'/?url='+encodeURIComponent(u.href)}function putLinked(tag,b,parent){const el=document.createElement(tag);if(b.kind==='li'){el.className='bullet';el.appendChild(document.createTextNode('• '))}const text=b.kind==='li'?b.text:b.text;if(Array.isArray(b.parts)&&b.parts.length){for(const part of b.parts){let e=document.createElement(part.bold?'strong':part.italic?'em':'span');e.textContent=part.text||'';if(part.url){try{const u=new URL(part.url);if(['http:','https:'].includes(u.protocol)&&!u.username&&!u.password){const a=document.createElement('a');a.href=articleLinkHref(u);if(a.href===u.href){a.target='_blank';a.rel='noopener noreferrer'}a.appendChild(e);el.appendChild(a);continue}}catch(_){}}el.appendChild(e)}parent.appendChild(el);return el;}const links=Array.isArray(b.links)?b.links:[];let at=0;for(const ref of links){if(!ref||typeof ref.text!=='string'||!ref.text||typeof ref.url!=='string')continue;let dest;try{dest=new URL(ref.url);if(!['http:','https:'].includes(dest.protocol)||dest.username||dest.password)continue}catch(_){continue}const pos=text.indexOf(ref.text,at);if(pos<0)continue;el.appendChild(document.createTextNode(text.slice(at,pos)));const a=document.createElement('a');a.textContent=ref.text;a.href=/\.(?:pdf|docx?|xlsx?|zip)$/i.test(dest.pathname)?dest.href:'/?url='+encodeURIComponent(dest.href);if(a.href===dest.href){a.target='_blank';a.rel='noopener noreferrer'}a.title='Open linked document or page';el.appendChild(a);at=pos+ref.text.length}el.appendChild(document.createTextNode(text.slice(at)));parent.appendChild(el);return el}

function renderTable(b,parent){
 const frame=document.createElement('div');frame.className='table-scroll';
 const table=document.createElement('table');const body=document.createElement('tbody');
 for(const row of b.rows||[]){const tr=document.createElement('tr');for(const cell of row.cells||[]){
 const td=document.createElement(cell.header?'th':'td');if(cell.header)td.scope='col';
 if(cell.colspan>1)td.colSpan=Math.min(100,cell.colspan);
 if(cell.rowspan>1)td.rowSpan=Math.min(100,cell.rowspan);
 putLinkedCell(td,cell);tr.appendChild(td)}body.appendChild(tr)}
 table.appendChild(body);frame.appendChild(table);parent.appendChild(frame)
}
function putLinkedCell(td,c){if(!Array.isArray(c.parts)||!c.parts.length){td.textContent=c.text||'';return}
 for(const p of c.parts){const e=document.createElement(p.bold?'strong':p.italic?'em':'span');e.textContent=p.text||'';
 if(p.url){try{const u=new URL(p.url);if(['http:','https:'].includes(u.protocol)&&!u.username&&!u.password){const a=document.createElement('a');a.href=articleLinkHref(u);a.appendChild(e);td.appendChild(a);continue}}catch(_){}}
 td.appendChild(e)}
}
function exportTable(b,md){const rows=(b.rows||[]).map(r=>(r.cells||[]).map(c=>String(c.text||'').replace(/\s+/g,' ').trim()));if(!rows.length)return '';
 if(md){const escape=x=>x.replace(/\|/g,'\\|');const widths=Math.max(...rows.map(r=>r.length));const all=rows.map(r=>'| '+Array.from({length:widths},(_,i)=>escape(r[i]||'')).join(' | ')+' |');all.splice(1,0,'| '+Array(widths).fill('---').join(' | ')+' |');return all.join('\n')}
 return rows.map(r=>r.join('\t')).join('\n')
}

// Remove failed publisher images rather than leaving a large broken frame or
// repeating its alt text. Image delivery may fail independently of extraction.
function removeFailedImage(img,source,article){
 const discard=()=>{const figure=img.closest('figure');if(figure){figure.remove()}else{img.remove()}};
 let retried=false;
 const retry=()=>{
  if(retried||!article||!article.url||!source){discard();return}
  retried=true;
  img.referrerPolicy='same-origin';
  img.src='/api/article-image?url='+encodeURIComponent(article.url)+'&image='+encodeURIComponent(source);
 };
 img.addEventListener('error',retry);
 img.addEventListener('load',()=>{if(img.naturalWidth<=1||img.naturalHeight<=1)retry()});
}
function editorialCaption(caption,title){const text=(caption||'').trim();return text&&text.toLowerCase()!==(title||'').trim().toLowerCase()?text:''}
function asText(md){const a=current;let s=a.title+'\n'+[a.byline,a.published,a.site].filter(Boolean).join(' · ')+'\nSource: '+a.url+'\n\n';for(const b of a.blocks){if(b.kind==='table'){s+=exportTable(b,md)+'\n\n';continue}let t=b.kind==='img'?(md?'![ '+(b.text||'Article image')+' ]('+b.url+')':'Image: '+b.url+(b.text?' — '+b.text:'')):b.text;if(md&&Array.isArray(b.parts)&&b.parts.length){t=b.parts.map(p=>p.url?'['+p.text.trim()+']('+p.url+')':p.text).join('').trim()}if(md){if(/^h[234]$/.test(b.kind))t='### '+t;else if(b.kind==='li')t='- '+t;else if(b.kind==='blockquote')t='> '+t};s+=t+'\n\n'}return s}function save(ext,md){const blob=new Blob([asText(md)],{type:'text/plain;charset=utf-8'});const u=URL.createObjectURL(blob);const link=document.createElement('a');link.href=u;link.download='article.'+ext;link.click();setTimeout(()=>URL.revokeObjectURL(u),1000)}$('txt').onclick=()=>save('txt',false);$('md').onclick=()=>save('md',true);$('print').onclick=()=>window.print();form.onsubmit=async ev=>{ev.preventDefault();$('read').disabled=true;$('status').className='status';$('status').textContent='Fetching and cleaning article…';$('paper').style.display='none';$('fallback').style.display='none';$('paynotice').style.display='none';try{const response=await fetch('/api/read',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({url:$('url').value})});const data=await response.json();if(!response.ok){const error=Error(data.error||'Could not read article');error.kind=data.kind;error.publisher=data.publisher;error.evidence=data.evidence;error.originalUrl=data.url;throw error;}current=data;$('accessnote').textContent=data.accessNote||'';$('accessnote').style.display=data.accessNote?'block':'none';$('site').textContent=data.site||new URL(data.url).hostname;$('title').textContent=data.title;$('meta').textContent=[data.byline,data.published,data.wordCount+' words',data.method||'Go extraction'].filter(Boolean).join(' · ');$('content').replaceChildren();$('hero').replaceChildren();for(let bi=0;bi<data.blocks.length;bi++){const b=data.blocks[bi];if(b.kind==='topic-title'){const sec=document.createElement('section');sec.className='topic-index';put('h3',b.text,sec);const links=document.createElement('div');links.className='topic-links';let j=bi+1;for(;j<data.blocks.length&&data.blocks[j].kind==='topic-link';j++){const item=data.blocks[j];try{const u=new URL(item.url);if(!['http:','https:'].includes(u.protocol)||u.username||u.password)continue;const a=document.createElement('a');a.href='/?url='+encodeURIComponent(u.href);a.textContent=item.text;links.appendChild(a)}catch(_){}}sec.appendChild(links);$('content').appendChild(sec);bi=j-1;continue;}if(b.kind==='topic-link')continue;if(b.kind==='table'){renderTable(b,$('content'));continue;}
// Image-first and heading-first identity groups share the same generic renderer.
const first=b.kind==='img'?b:data.blocks[bi+1];
const head=b.kind==='img'?data.blocks[bi+1]:b;
const isHeading=head&&/^h[234]$/.test(head.kind);
const tidy=s=>(s||'').toLowerCase().replace(/^\s*\d+[.)]\s*/, '').replace(/[^a-z0-9]+/g,' ').trim();
const imageLabel=tidy(first?.text),headingLabel=tidy(head?.text);
const isIdentity=isHeading&&first?.kind==='img'&&imageLabel.length>=4&&headingLabel.length>=4&&(headingLabel===imageLabel||headingLabel.includes(imageLabel)||imageLabel.includes(headingLabel));
if(isIdentity){
 const fields=[];
 for(let off=bi+2;off<Math.min(data.blocks.length,bi+6);off++){
  const x=data.blocks[off];
  if(!x||x.kind!=='p'||!(x.text||'').trim()||x.text.length>150||!x.text.split('\n').every(line=>/^\s*[^:]{2,45}:/.test(line)))break;
  fields.push(x);
 }
 if(fields.length){
  const group=document.createElement('div');group.className='compact-entry';
  const pic=document.createElement('img');pic.src=first.url;pic.alt=first.text||head.text;pic.loading='eager';pic.referrerPolicy='no-referrer';removeFailedImage(pic,first.url,data);group.appendChild(pic);
  const info=document.createElement('div');info.className='compact-info';putLinked('h3',head,info);
  for(const field of fields)putLinked('p',field,info);
  group.appendChild(info);$('content').appendChild(group);bi+=1+fields.length;continue;
 }
}
const logo=data.blocks[bi+1];const heading=(b.text||'').replace(/^\s*\d+[.)]\s*/, '').trim();const label=(logo?.text||'').trim();const normalizeLabel=s=>s.toLowerCase().replace(/[^a-z0-9]+/g,' ').trim();const headingKey=normalizeLabel(heading),logoKey=normalizeLabel(label);const identifyingImage=logo?.kind==='img'&&headingKey.length>=4&&logoKey.length>=4&&(headingKey===logoKey||headingKey.includes(logoKey)||logoKey.includes(headingKey));const metadata=[];if(identifyingImage){for(let off=bi+2;off<Math.min(data.blocks.length,bi+5);off++){const x=data.blocks[off];if(!x||x.kind!=='p'||!(x.text||'').trim()||x.text.length>150||!(/^\s*[^\n:]{2,45}:/.test(x.text)||x.text.split('\n').every(line=>/^\s*[^:]{2,45}:/.test(line))))break;metadata.push(x)}}if(/^h[234]$/.test(b.kind)&&identifyingImage&&metadata.length>=1){const group=document.createElement('div');group.className='compact-entry';const pic=document.createElement('img');pic.src=logo.url;pic.alt=logo.text||heading;pic.loading='eager';pic.referrerPolicy='no-referrer';removeFailedImage(pic,logo.url,data);group.appendChild(pic);const info=document.createElement('div');info.className='compact-info';putLinked('h3',b,info);for(const m of metadata)putLinked('p',m,info);group.appendChild(info);$('content').appendChild(group);bi+=1+metadata.length;continue}if(b.kind==='img'&&b.gallery){const members=[];let j=bi;while(j<data.blocks.length&&data.blocks[j].kind==='img'&&data.blocks[j].gallery===b.gallery){members.push(data.blocks[j]);j++}if(members.length>1){const group=document.createElement('div');group.className='source-gallery';for(const item of members){const fig=document.createElement('figure');const im=document.createElement('img');im.src=item.url;im.alt=item.text||'Gallery photograph';im.loading='eager';im.referrerPolicy='no-referrer';removeFailedImage(im,item.url,data);fig.appendChild(im);if(editorialCaption(item.text,data.title))put('figcaption',item.text,fig);group.appendChild(fig)}$('content').appendChild(group);bi=j-1;continue;}}if(b.kind==='img'&&/^https?:\/\//.test(b.url||'')){const fig=document.createElement('figure');if(b.role==='lead')fig.className='lead-photo';else if(b.role==='author')fig.className='author-photo';const im=document.createElement('img');im.src=b.url;im.alt=b.text||'Article image';im.loading='eager';im.referrerPolicy='no-referrer';removeFailedImage(im,b.url,data);fig.appendChild(im);if(editorialCaption(b.text,data.title))put('figcaption',b.text,fig);if(b.role!=='lead'&&b.role!=='author'&&b.text&&data.blocks.some((x,j)=>x===b&&j>0&&/^h[234]$/.test(data.blocks[j-1].kind))){fig.className='compact-mark'}$('content').appendChild(fig);continue;}const tag=/^h[234]$/.test(b.kind)?'h3':b.kind==='blockquote'?'blockquote':'p';putLinked(tag,b,$('content'))}const link=$('source');link.href=data.url;link.textContent=data.url;$('paper').style.display='block';$('status').textContent=data.accessNote?'Publicly available article preview extracted.':'Article extracted successfully.'}catch(err){$('status').className='status error';$('status').textContent=err.message; if(err.kind==='page_index'){$('status').className='status';$('status').textContent=err.message;try{const original=new URL(err.originalUrl||$('url').value);if(['http:','https:'].includes(original.protocol)){$('original').href=original.href;$('original').textContent='Open original website ↗';$('fallback').style.display='block'}}catch(_){}return;}if(['paywall','registration','restricted'].includes(err.kind)) { const titles={paywall:'Paid Subscription Required',registration:'Free Registration Required',restricted:'Access Restricted'}; $('paytitle').textContent=titles[err.kind]; $('paymessage').textContent=err.message; $('paydiagnostic').textContent=[err.publisher?'Publisher: '+err.publisher:'',err.evidence?'Detection: '+err.evidence:''].filter(Boolean).join(' · '); $('paynotice').style.display='block'; $('status').textContent=''; return; } try{const original=new URL($('url').value);if(original.protocol==='http:'||original.protocol==='https:'){$('original').href=original.href;$('fallback').style.display='block'}}catch(_){}}finally{$('read').disabled=false}};

// Browser-assisted reading: extension passes ONLY visible article paragraphs. No raw HTML.
window.addEventListener('message',ev=>{
 if(ev.source!==window||ev.origin!==location.origin||!ev.data||ev.data.type!=='SCRAPERGO_BROWSER_ARTICLE')return;
 const a=ev.data.article;
 if(!a||!Array.isArray(a.blocks)||a.blocks.length<2||typeof a.url!=='string')return;
 try{const u=new URL(a.url);if(!['https:','http:'].includes(u.protocol))return;}catch(_){return;}
 current=a;$('url').value=a.url;$('paper').style.display='block';$('paynotice').style.display='none';$('fallback').style.display='none';$('status').className='status';$('status').textContent='Article read from visible Safari page.';
 $('site').textContent=a.site||new URL(a.url).hostname;$('title').textContent=a.title||'Article';$('meta').textContent=a.wordCount+' words · Safari';$('content').replaceChildren();
 for(const b of a.blocks){if(b.kind==='table'){renderTable(b,$('content'));continue}put(/^h[234]$/.test(b.kind)?'h3':'p',b.text,$('content'))}
 $('source').href=a.url;$('source').textContent=a.url;
});

// Safari bookmarklet supplies ?url=...; keep manual entry available.
const incoming=new URLSearchParams(window.location.search).get('url');
if(incoming){$('url').value=incoming;if(new URLSearchParams(location.search).get('browser')!=='1')form.requestSubmit();else {$('status').textContent='Waiting for Safari article text…';setTimeout(()=>{if(!$('paper').style.display||$('paper').style.display==='none')form.requestSubmit()},2800)}}

