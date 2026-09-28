import { useEffect, useState } from "react";
import { LayoutDashboard, Users, KeyRound, FlaskConical, BookOpen, LogOut } from "lucide-react";

type Account = { id:string; name:string; provider:string; has_consumer_token:boolean; token_status?:{valid:boolean} };
type ApiKey = { id:string; key:string; name:string; account_id:string; enabled:boolean };
const navigation = [{id:"overview", label:"总览", icon:LayoutDashboard},{id:"accounts", label:"账号管理", icon:Users},{id:"keys", label:"API 密钥", icon:KeyRound},{id:"test", label:"对话测试", icon:FlaskConical},{id:"guide", label:"个人账号接入", icon:BookOpen}];
async function request(path:string, method="GET", body?:unknown, bearer?:string) {
  const response = await fetch(path, {method, credentials:"same-origin", cache:"no-store", headers:{"Content-Type":"application/json", ...(bearer ? {Authorization:"Bearer "+bearer}: {})}, ...(body !== undefined ? {body:JSON.stringify(body)} : {})});
  const data = await response.json();
  if (!response.ok) throw new Error(data.error?.message || data.detail || "请求失败："+response.status);
  return data;
}
export default function App() {
  const [authenticated,setAuthenticated] = useState(false);
  const [loading,setLoading] = useState(true);
  const [password,setPassword] = useState("");
  const [page,setPage] = useState("overview");
  const [accounts,setAccounts] = useState<Account[]>([]);
  const [keys,setKeys] = useState<ApiKey[]>([]);
  const [notice,setNotice] = useState("");
  const [busy,setBusy] = useState(false);
  const [name,setName] = useState("");
  const [accountId,setAccountId] = useState("");
  const [testKey,setTestKey] = useState("");
  const [model,setModel] = useState("copilot");
  const [prompt,setPrompt] = useState("你好，请简单介绍一下自己。");
  const [answer,setAnswer] = useState("");
  const refresh = async () => {
    const [accountData,keyData] = await Promise.all([request("/admin/accounts"),request("/admin/keys")]);
    setAccounts(accountData.accounts); setKeys(keyData.keys); setAuthenticated(true);
  };
  useEffect(() => { refresh().catch(()=>setAuthenticated(false)).finally(()=>setLoading(false)); },[]);
  const action = async (task:()=>Promise<void>) => {
    setBusy(true); setNotice("");
    try { await task(); } catch(error) { setNotice(error instanceof Error ? error.message : "操作失败"); }
    finally { setBusy(false); }
  };
  const copy = (value:string) => action(async()=>{await navigator.clipboard.writeText(value);setNotice("已复制，请妥善保管，不要公开密钥。");});
  if (loading) return <main className="login">正在检查登录状态…</main>;
  if (!authenticated) return <main className="login card"><h1>Copilot Personal Hub</h1><p className="muted">个人账号核心 · 统一管理控制台</p><form onSubmit={event=>{event.preventDefault();void action(async()=>{await request("/admin/login","POST",{password});setPassword("");await refresh();});}}><label>管理员密码<input type="password" autoComplete="current-password" value={password} onChange={event=>setPassword(event.target.value)} required /></label><button className="primary" disabled={busy}>登录控制台</button></form><p role="alert">{notice}</p><p className="muted">此处不是微软登录页。个人微软账号请在登录后按接入指南操作。</p><a href="https://github.com/AMDSE/copilot-personal-hub">源代码与许可证</a></main>;
return <div className="hub"><nav><strong>Copilot Personal Hub</strong>{navigation.map(item=><button className={page===item.id?"active":""} key={item.id} onClick={()=>{setPage(item.id);setNotice("");}}><item.icon size={17}/>{item.label}</button>)}<button disabled={busy} onClick={()=>void action(async()=>{await request("/admin/logout","POST");setAuthenticated(false);setAccounts([]);setKeys([]);setTestKey("");setAnswer("");})}><LogOut size={17}/>退出登录</button><a href="/admin">高级管理</a><a href="/self-service">用户自助页</a><a href="https://github.com/AMDSE/copilot-personal-hub">源代码与许可证</a></nav><main><div className="row"><h1>{navigation.find(item=>item.id===page)?.label}</h1><button disabled={busy} onClick={()=>void action(refresh)}>刷新状态</button></div>{notice&&<p className="notice" role="status">{notice}</p>}
    {page==="overview"&&<><div className="grid"><section className="card"><span className="muted">账号总数</span><div className="metric">{accounts.length}</div></section><section className="card"><span className="muted">已导入个人凭据</span><div className="metric">{accounts.filter(account=>account.provider==="consumer"&&account.has_consumer_token).length}</div></section><section className="card"><span className="muted">已启用密钥</span><div className="metric">{keys.filter(key=>key.enabled).length}</div></section></div><section className="card"><h2>先接入，再测试</h2><p>创建一个空账号 → 创建并绑定 API 密钥 → 在自己的浏览器登录 Copilot 并推送个人凭据 → 发起对话测试。</p><p>“已导入凭据”不等于微软已接受请求。实际可用性请以对话测试为准。</p><button className="primary" onClick={()=>setPage("guide")}>查看接入步骤</button></section><section className="card"><h2>兼容接口</h2><pre>{location.origin+"/v1"}</pre><p>支持的协议继承自 Ciallo 核心。工具调用与模型效果取决于上游，不保证等同于官方 API。</p></section></>}
    {page==="accounts"&&<><form className="card" onSubmit={event=>{event.preventDefault();void action(async()=>{await request("/admin/accounts","POST",{name});setName("");await refresh();setNotice("空账号已创建，请创建绑定密钥后推送个人版凭据。");});}}><h2>添加账号容器</h2><label>备注名称<input value={name} onChange={event=>setName(event.target.value)} placeholder="例如：我的个人 Copilot" required maxLength={80}/></label><p className="muted">这里只创建记录，不填写微软密码；推送个人凭据后才会切换为 Consumer。</p><button disabled={busy} className="primary">创建空账号</button></form>{accounts.map(account=><section className="card" key={account.id}><div className="account-name">{account.name||account.id}</div><p>{account.provider==="consumer"?"个人版 Consumer":"空账号 / M365"} · {account.has_consumer_token?"个人凭据已导入":"等待凭据"}</p><div className="row"><button onClick={()=>{setAccountId(account.id);setPage("keys");}}>创建绑定密钥</button><button className="danger" disabled={busy} onClick={()=>{if(confirm("删除这个账号及其凭据？绑定密钥将失去对应账号。"))void action(async()=>{await request("/admin/accounts/"+account.id,"DELETE");await refresh();});}}>删除账号</button></div></section>)}</>}
    {page==="keys"&&<><form className="card" onSubmit={event=>{event.preventDefault();void action(async()=>{await request("/admin/keys","POST",{name,account_id:accountId});setName("");await refresh();setNotice("密钥已创建，点击复制后用于个人凭据推送或客户端调用。");});}}><h2>创建 API 密钥</h2><label>名称<input value={name} onChange={event=>setName(event.target.value)} required maxLength={80}/></label><label>绑定账号<select value={accountId} onChange={event=>setAccountId(event.target.value)} required><option value="">请选择账号</option>{accounts.map(account=><option value={account.id} key={account.id}>{account.name||account.id}</option>)}</select></label><button disabled={busy} className="primary">创建密钥</button></form>{keys.map(key=><section className="card" key={key.id}><div className="account-name">{key.name||key.id}</div><p>{key.enabled?"已启用":"已停用"} · {accounts.find(account=>account.id===key.account_id)?.name||"未绑定"}</p><code>{key.key.slice(0,7)}••••••••</code><div className="row"><button onClick={()=>void copy(key.key)}>复制密钥</button><button onClick={()=>{setTestKey(key.key);setPage("test");}}>测试</button><button disabled={busy} onClick={()=>void action(async()=>{await request("/admin/keys/"+key.id,"POST",{enabled:!key.enabled});await refresh();})}>{key.enabled?"停用":"启用"}</button><button disabled={busy} className="danger" onClick={()=>{if(confirm("永久删除此密钥？使用它的客户端将无法访问。"))void action(async()=>{await request("/admin/keys/"+key.id,"DELETE");await refresh();});}}>删除</button></div></section>)}</>}
    {page==="test"&&<form className="card" onSubmit={event=>{event.preventDefault();void action(async()=>{setAnswer("请求中，请等待上游响应…");try{const result=await request("/v1/chat/completions","POST",{model,messages:[{role:"user",content:prompt}],stream:false},testKey);setAnswer(result.choices?.[0]?.message?.content||JSON.stringify(result,null,2));}catch(error){setAnswer("");throw error;}});}}><h2>真实请求测试</h2><p className="muted">此操作会调用你的微软账号，不是本地模拟。</p><label>API 密钥<select value={testKey} onChange={event=>setTestKey(event.target.value)} required><option value="">选择已启用密钥</option>{keys.filter(key=>key.enabled).map(key=><option value={key.key} key={key.id}>{key.name||key.id}</option>)}</select></label><label>模型<input value={model} onChange={event=>setModel(event.target.value)} required/></label><label>消息<textarea value={prompt} onChange={event=>setPrompt(event.target.value)} required/></label><button disabled={busy} className="primary">发送测试消息</button><pre aria-live="polite">{answer}</pre></form>}
    {page==="guide"&&<section className="card"><h2>个人微软账号怎么接入？</h2><ol><li>在“账号管理”创建空账号，再到“API 密钥”创建并绑定密钥。</li><li>在你自己的浏览器安装 Tampermonkey 扩展，然后安装本项目的 <a href="/hub-assets/get_token.user.js">凭据推送脚本</a>。先阅读脚本，确认信任后安装。</li><li>打开 <a href="https://copilot.microsoft.com" target="_blank" rel="noreferrer">微软个人版 Copilot</a>，登录你自己的账号并发送一条消息。</li><li>在脚本面板填入代理地址 <code>{location.origin}</code> 和刚复制的 API 密钥，点击“一键推送个人版”。不要选择仅 M365 的授权入口。</li><li>回到本控制台刷新状态，在“对话测试”中选择该密钥发送消息。</li></ol><h3>凭据与隐私</h3><p>Cookie 和 Token 相当于登录凭据，只推送到你自己信任的服务。不要把微软密码、Cookie、Token 或 API 密钥放进 GitHub。</p><h3>凭据到期怎么办？</h3><p>当前轻量部署未安装 Camoufox 浏览器。需要续期时重新在 Copilot 发送消息并推送个人凭据。高级设置仍保留在原生管理页面。</p><p>仅用于你有权使用的账号及非商业用途；上游协议、地区限制和账号风控可能影响可用性。</p></section>}
  </main></div>;
}
