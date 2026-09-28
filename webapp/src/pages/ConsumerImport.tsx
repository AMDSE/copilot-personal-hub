import { useState } from "react";
import { api } from "../api";

export function ConsumerImport({push,onImported}:{push:(message:string,kind?:"success"|"error"|"info")=>void;onImported:()=>Promise<void>}) {
  const [expanded,setExpanded]=useState(false);
  const [snapshot,setSnapshot]=useState("");
  const [busy,setBusy]=useState(false);
  return <div className="card" style={{marginBottom:16}}>
    <div className="card-head"><span>个人账号接入</span><button className="btn btn-sm primary" onClick={()=>setExpanded(!expanded)}>{expanded?"收起":"导入 / 更新个人账号"}</button></div>
    {expanded&&<div className="card-body">
      <p style={{color:"var(--muted)",fontSize:13,marginTop:0}}>使用个人版 Copilot 凭据，不走企业 / 学校账号授权。账号、调度、API 密钥和调用统计统一由当前面板管理。</p>
      <ol style={{fontSize:13,lineHeight:1.9,paddingLeft:20}}><li>安装 Tampermonkey BETA，再安装 <a href="/api/accounts/consumer-script">个人版凭据导出脚本</a>。</li><li>在你自己的浏览器打开 copilot.microsoft.com，登录并发送一条消息。</li><li>脚本面板点击“导出个人版凭据”，下载 JSON 文件；在这里选择文件或粘贴内容。</li><li>导入后在“模型测试”页面验证。仅导入凭据不代表上游已验证可用。</li></ol>
      <input className="form-input" aria-label="个人凭据 JSON 文件" type="file" accept=".json,application/json" onChange={async event=>{const file=event.target.files?.[0];if(!file)return;if(file.size>2*1024*1024){push("文件不能超过 2MB","error");return;}setSnapshot(await file.text());event.target.value="";}}/>
      <textarea className="form-input" aria-label="个人凭据 JSON" autoComplete="off" spellCheck={false} value={snapshot} onChange={event=>setSnapshot(event.target.value)} placeholder="粘贴导出的个人账号凭据 JSON（包含敏感信息，不要公开）" style={{width:"100%",minHeight:110,marginTop:10}}/>
      <div style={{display:"flex",gap:10,alignItems:"center",marginTop:12}}><button className="btn primary" disabled={busy||!snapshot.trim()} onClick={async()=>{setBusy(true);try{const credentials=JSON.parse(snapshot);await api("/api/accounts/consumer",{method:"POST",body:JSON.stringify(credentials)});setSnapshot("");await onImported();push("个人凭据已导入，请到模型测试验证","success");}catch(error){push(error instanceof Error?error.message:"导入失败","error");}finally{setBusy(false);}}}>{busy?"正在保存…":"保存个人账号"}</button><button className="btn" onClick={()=>setSnapshot("")}>清空内容</button></div>
      <p style={{fontSize:12,color:"var(--muted)",marginBottom:0}}>凭据加密保存；只导入你有权使用的账号。支持匹配账号的刷新令牌自动续期，Cookie 失效或微软拒绝时需重新导出。个人版模型不等同于企业 GPT / Claude 型号，企业云端会话与记忆接口不适用个人版。</p>
    </div>}
  </div>;
}
