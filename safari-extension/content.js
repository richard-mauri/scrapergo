// ScraperGo Safari content script. Only reads content visible in the page.
function visible(n){const s=getComputedStyle(n);return s.display!=='none'&&s.visibility!=='hidden'&&n.getClientRects().length>0&&!n.closest('aside,nav,footer,[aria-hidden="true"],[class*="paywall"],[class*="subscribe"]')}
function publicArticle(){
 if(document.querySelector('[class*="paywall"], [id*="paywall"]')&&/subscribe|subscription required/i.test(document.body.innerText.slice(0,18000)))return null;
 const root=document.querySelector('article')||document.querySelector('main');if(!root)return null;
 const blocks=[...root.querySelectorAll('p,h2,h3')].filter(visible).map(n=>({kind:n.tagName.toLowerCase(),text:(n.innerText||'').trim()})).filter(b=>b.text.length>=35&& !/subscribe|sign up|advertisement|related articles/i.test(b.text)).slice(0,100);
 const wc=blocks.reduce((n,b)=>n+b.text.split(/\s+/).length,0);if(wc<80)return null;
 return {url:location.href,title:document.querySelector('h1')?.innerText||document.title,site:location.hostname,blocks,wordCount:wc};
}
browser.runtime.onMessage.addListener((m)=>{
 if(m?.type==='SCRAPERGO_EXTRACT')return Promise.resolve(publicArticle());
 if(m?.type==='SCRAPERGO_DELIVER'&&location.origin===new URL(m.reader).origin){window.postMessage({type:'SCRAPERGO_BROWSER_ARTICLE',article:m.article},location.origin);return Promise.resolve(true)}
});
