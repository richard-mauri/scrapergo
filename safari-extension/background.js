// Open article links directly in ScraperGo. No transient publisher tabs.
const MENU_ID = 'scrapergo-open-link';
const DEFAULT_READER = 'https://scrapergo.themauris.org/';
function validURL(address) {
  try { return ['http:', 'https:'].includes(new URL(address).protocol); }
  catch { return false; }
}
async function readerOrigin() {
  const value = (await browser.storage.local.get('readerURL')).readerURL || DEFAULT_READER;
  try {
    const u = new URL(value);
    if (!['http:', 'https:'].includes(u.protocol)) throw Error();
    return u.origin + '/';
  } catch { return DEFAULT_READER; }
}
// Serialize overlapping context-menu clicks so they cannot create two reader tabs.
let pendingOpen = Promise.resolve();
async function openReader(articleURL) {
  const base = await readerOrigin();
  const destination = base + '?url=' + encodeURIComponent(articleURL);
  // Query all tabs rather than a single window; a reader may be open elsewhere.
  const tabs = await browser.tabs.query({});
  const reader = tabs.find(tab => {
    try { return new URL(tab.url).origin === new URL(base).origin; }
    catch { return false; }
  });
  if (reader) {
    try {
      await browser.tabs.update(reader.id, {url: destination, active: true});
      // Bring the reader window forward where Safari permits this API.
      if (typeof reader.windowId === 'number' && browser.windows?.update) {
        try { await browser.windows.update(reader.windowId, {focused: true}); }
        catch (err) { console.debug('ScraperGo could not focus reader window:', err); }
      }
      return;
    } catch (err) {
      // The reader might have closed between query and update.
      console.debug('ScraperGo reader tab no longer available:', err);
    }
  }
  await browser.tabs.create({url: destination, active: true});
}
function installMenu() {
  browser.contextMenus.removeAll()
    .then(() => browser.contextMenus.create({id: MENU_ID, title: 'Open Link in ScraperGo', contexts: ['link']}))
    .catch(console.error);
}
browser.runtime.onInstalled.addListener(installMenu);
browser.runtime.onStartup.addListener(installMenu);
installMenu();
browser.contextMenus.onClicked.addListener(async info => {
  if (info.menuItemId !== MENU_ID || !validURL(info.linkUrl)) return;
  pendingOpen = pendingOpen.catch(() => {}).then(() => openReader(info.linkUrl));
  try { await pendingOpen; }
  catch (err) { console.error('ScraperGo could not open reader:', err); }
});
