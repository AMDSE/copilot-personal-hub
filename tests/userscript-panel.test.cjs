const test = require('node:test');
const assert = require('node:assert/strict');
const vm = require('node:vm');
const fs = require('node:fs');
const path = require('node:path');
const source = fs.readFileSync(path.join(__dirname, '../internal/web/web/consumer-capture.user.js'), 'utf8');

function harness(hostname) {
  const menus = new Map();
  const alerts = [];
  const listeners = [];
  let copied = '';
  class Element {
    constructor(tag) {this.tagName=tag;this.children=[];this.style={};this.events={};this.parent=null;this.textContent='';}
    set innerHTML(value) {throw new TypeError('TrustedHTML required');}
    appendChild(child){child.parent=this;this.children.push(child);return child;}
    remove(){if(this.parent)this.parent.children=this.parent.children.filter(child=>child!==this);}
    attachShadow(){this.shadowRoot=new Element('shadow-root');return this.shadowRoot;}
    addEventListener(name,handler){this.events[name]=handler;}
    setAttribute(name,value){this[name]=value;}
    click(){if(this.events.click)return this.events.click({});}
    focus(){}
    select(){}
  }
  const html = new Element('html');
  const find = (root,id) => root.id===id ? root : root.children.map(child=>find(child,id)).find(Boolean);
  const storage={getItem:()=>null,setItem(){},key:()=>null,length:0};
  function Socket(){this.send=()=>{};this.addEventListener=()=>{};}
  Object.assign(Socket,{CONNECTING:0,OPEN:1,CLOSING:2,CLOSED:3});
  const context={console,URL,Blob,TextDecoder,Uint8Array,Date,Set,Map,JSON,Promise,decodeURIComponent,encodeURIComponent,
    location:{hostname,href:'https://'+hostname+'/chat'},localStorage:storage,sessionStorage:storage,
    navigator:{language:'zh-CN',clipboard:{writeText:async value=>{copied=value;}}},
    document:{readyState:'complete',documentElement:html,body:html,createElement:tag=>new Element(tag),getElementById:id=>find(html,id),addEventListener:(...args)=>listeners.push(args)},
    GM_getValue:(_key,fallback)=>fallback,GM_setValue(){},GM_registerMenuCommand:(label,handler)=>menus.set(label,handler),
    alert:value=>alerts.push(value),setTimeout:()=>0,clearTimeout(){},setInterval:()=>0,clearInterval(){},
    fetch:async()=>({}),WebSocket:Socket,addEventListener(){},atob:value=>Buffer.from(value,'base64').toString('binary'),
  };
  context.window=context;context.self=context;context.top=context;context.unsafeWindow=context;
  vm.runInNewContext(source,context,{timeout:2000});
  return {context,menus,alerts,html,find,copied:()=>copied,listeners};
}

for (const hostname of ['copilot.com','www.copilot.com','copilot.microsoft.com']) {
  test('personal panel opens under Trusted Types on '+hostname,()=>{
    const app=harness(hostname);
    assert.ok(app.find(app.html,'copilot-export-launcher'));
    const entry=[...app.menus].find(([label])=>label.startsWith('打开个人版'));
    assert.ok(entry);
    entry[1]();
    assert.ok(app.find(app.html,'m365-token-panel').shadowRoot);
    assert.deepEqual(app.alerts,[]);
    assert.equal(app.listeners.filter(([name])=>name==='keydown').length,1);
  });
}

test('new Substrate path gives credential-free diagnostics, not fake legacy export',async()=>{
  const app=harness('copilot.com');
  new app.context.WebSocket('wss://substrate.office.com/m365Copilot/Chathub/private-account?secret=DO_NOT_DISCLOSE');
  await [...app.menus].find(([label])=>label.startsWith('复制安全诊断'))[1]();
  const info=JSON.parse(app.copied());
  assert.deepEqual(info.observedTransportKinds,['substrate-chathub']);
  assert.equal(info.legacyChatAITokenCaptured,false);
  assert.equal(info.scriptVersion,'1.1.0');
  assert.ok(!app.copied().includes('DO_NOT_DISCLOSE'));
  assert.ok(!app.copied().includes('private-account'));
});

test('legacy ChatAI capture remains available',async()=>{
  const app=harness('copilot.microsoft.com');
  new app.context.WebSocket('wss://copilot.microsoft.com/c/api/chat?accessToken=FAKE_TEST_TOKEN');
  await [...app.menus].find(([label])=>label.startsWith('复制安全诊断'))[1]();
  assert.equal(JSON.parse(app.copied()).legacyChatAITokenCaptured,true);
  assert.ok(!app.copied().includes('FAKE_TEST_TOKEN'));
  assert.deepEqual(app.alerts,[]);
});
