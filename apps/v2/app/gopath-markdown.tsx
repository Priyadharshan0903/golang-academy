'use client';
import {runGo} from '@/shared/playground';
import {useMemo,useState,type ReactNode} from 'react';
import {marked} from '@/vendor/marked/lib/marked.esm.js';
import {ArrowUpRight,Check,Code2,Copy,Download,FileCode2,Play,ShieldCheck,Terminal} from 'lucide-react';
import {course,pageHref,resolveLink,slug,downloadText,type CoursePage} from '@content/v2/gopath';
type Token=Record<string,any>;
const entities=(text:string)=>text.replace(/&(amp|lt|gt|quot|#39|#x[0-9a-f]+|#\d+);/gi,(all,name)=>(({amp:'&',lt:'<',gt:'>',quot:'"','#39':"'"} as Record<string,string>)[name]|| (name.startsWith('#x')?String.fromCodePoint(parseInt(name.slice(2),16)):name.startsWith('#')?String.fromCodePoint(Number(name.slice(1))):all)));

export function Markdown({body,page}:{body:string;page?:CoursePage}){
 const tokens=useMemo(()=>marked.lexer(body,{gfm:true}) as Token[],[body]);
 const inline=(items:Token[]|undefined):ReactNode=>(items||[]).map((token,i)=>{
  const children=token.tokens?inline(token.tokens):entities(token.text||'');
  switch(token.type){
   case 'strong':return <strong key={i}>{children}</strong>;
   case 'em':return <em key={i}>{children}</em>;
   case 'del':return <del key={i}>{children}</del>;
   case 'codespan':return <code key={i}>{entities(token.text)}</code>;
   case 'br':return <br key={i}/>;
   case 'link':{const target=resolveLink(token.href,page),external=/^https?:/.test(target);return <a key={i} href={target} target={external?'_blank':undefined} rel={external?'noreferrer':undefined} onClick={event=>{if(target.startsWith('#')&&!target.startsWith('#/')){event.preventDefault();document.getElementById(decodeURIComponent(target.slice(1)))?.scrollIntoView({behavior:'smooth'});}}}>{children}{external&&<ArrowUpRight size={10} className="inline-external"/>}</a>;}
   case 'html':return null;
   default:return <span key={i}>{children}</span>;
  }
 });
 const blocks=(items:Token[]):ReactNode=>items.map((token,i)=>{
  switch(token.type){
   case 'space':return null;
   case 'heading':{const Tag=`h${Math.min(6,Math.max(2,token.depth))}` as 'h2';return <Tag id={slug(token.text)} key={i}>{inline(token.tokens)}</Tag>;}
   case 'paragraph':return <p key={i}>{inline(token.tokens)}</p>;
   case 'text':return <p key={i}>{token.tokens?inline(token.tokens):entities(token.text)}</p>;
   case 'code':{
    if(token.lang==='gopath-example')return <Example key={i} id={token.text.trim()}/>;
    if(token.lang==='gopath-exercise')return <Exercise key={i} id={token.text.trim()} page={page}/>;
    if(token.lang==='gopath-callout'){const note=course.callouts[token.text.trim()];return note?<aside key={i} className={`academy-callout ${note.type}`}><ShieldCheck size={18}/><Markdown body={note.body} page={page}/></aside>:null;}
    if(token.lang==='gopath-disclosure'){const d=course.disclosures[token.text.trim()];return d?<details className="academy-disclosure" key={i}><summary>{d.title}</summary><Markdown body={d.body} page={page}/></details>:null;}
    return <CodeBlock key={i} code={token.text} language={token.lang||'text'}/>;
   }
   case 'blockquote':return <blockquote key={i}>{blocks(token.tokens||[])}</blockquote>;
   case 'list':{const Tag=token.ordered?'ol':'ul';return <Tag key={i} start={token.ordered?token.start:undefined}>{token.items.map((item:Token,j:number)=><li key={j}>{item.task&&<input type="checkbox" checked={Boolean(item.checked)} readOnly aria-label={item.text}/>} {blocks(item.tokens||[])}</li>)}</Tag>;}
   case 'table':return <div className="academy-table-wrap" key={i}><table><thead><tr>{token.header.map((cell:Token,j:number)=><th key={j}>{inline(cell.tokens)}</th>)}</tr></thead><tbody>{token.rows.map((row:Token[],j:number)=><tr key={j}>{row.map((cell,k)=><td key={k}>{inline(cell.tokens)}</td>)}</tr>)}</tbody></table></div>;
   case 'hr':return <hr key={i}/>;
   case 'html':return null;
   default:return token.tokens?<div key={i}>{blocks(token.tokens)}</div>:null;
  }
 });
 return <div className="academy-markdown">{blocks(tokens)}</div>;
}

export function CodeBlock({code,language='go',file,canRun=false,fullCode}:{code:string;language?:string;file?:string;canRun?:boolean;fullCode?:string}){
 const [copied,setCopied]=useState(false),[copyError,setCopyError]=useState(false),[editing,setEditing]=useState(false),[source,setSource]=useState(fullCode||code),[running,setRunning]=useState(false),[output,setOutput]=useState(''),[runError,setRunError]=useState(false);
 const copy=async()=>{try{await navigator.clipboard.writeText(editing?source:code);setCopied(true);setTimeout(()=>setCopied(false),1500);}catch{setCopyError(true);}};
 const run=async()=>{if(running)return;setRunning(true);setOutput('');setRunError(false);try{const {response,data}=await runGo(source);setRunError(!response.ok||Boolean(data.error)||Boolean(data.errors)||(data.events||[]).some(e=>e.Kind==='stderr'));setOutput(data.error||data.errors||(data.events||[]).map(e=>e.Message).join('')||'Program finished with no output.');}catch{setRunError(true);setOutput('The runner is unavailable. Download the course and use the local command shown below.');}finally{setRunning(false);}};
 return <div className="academy-code"><div className="academy-code-toolbar"><span><FileCode2 size={14}/>{file||language.split(' ')[0]}</span><div>{canRun&&<button onClick={()=>setEditing(v=>!v)}><Code2 size={13}/>{editing?'Read excerpt':'Edit & run'}</button>}<button aria-label="Copy code" onClick={copy}>{copied?<Check size={13}/>:<Copy size={13}/>}<span>{copied?'Copied':'Copy'}</span></button>{file&&<button aria-label="Download source file" onClick={()=>downloadText(editing?source:(fullCode||code),file.split('/').pop()||'source.go')}><Download size={13}/></button>}</div></div>{editing?<><textarea aria-label="Editable Go example" value={source} spellCheck={false} onChange={e=>setSource(e.target.value)}/><div className="academy-run-row"><span>Go Playground · full source file</span><button className="run-button" onClick={()=>void run()} disabled={running}><Play size={12}/>{running?'Running…':'Run code'}</button></div></>:<pre tabIndex={0}><code>{code}</code></pre>}{(output||running)&&<pre className={`academy-output ${runError?'is-error':''}`} role="status">{running?'Compiling and running…':output}</pre>}{copyError&&<p className="copy-feedback" role="status">Clipboard unavailable. Select the code or use Download.</p>}</div>;
}
function Example({id}:{id:string}){const example=course.embeds[id];if(!example)return <p>Example unavailable.</p>;return <div className="academy-embedded-example"><CodeBlock code={example.code} file={example.file} canRun={example.canRun} fullCode={course.sourceFiles[example.file]?.content}/><div className="academy-example-meta"><a href={`#/source/${example.file}`}>View complete file <ArrowUpRight size={12}/></a>{example.region&&<span>Region: {example.region}</span>}{!example.canRun&&<span>Local package / test</span>}</div>{example.command&&<div className="academy-local-command"><Terminal size={13}/><code>{example.command}</code></div>}</div>;}
export function Exercise({id,page,standalone=false}:{id:string;page?:CoursePage;standalone?:boolean}){const exercise=course.exercises[id];if(!exercise)return <p>Exercise unavailable.</p>;return <section className="academy-exercise"><span className="eyebrow">{standalone?'LOCAL PRACTICE':'PUT IT INTO PRACTICE'} · {id}</span><Markdown body={exercise.body} page={page}/><div className="academy-local-command"><Terminal size={14}/><code>{`cd exercises\ngo test ./${id}/`}</code></div><p className="exercise-boundary">Edit starter.go in the downloaded course. Tests intentionally fail until your implementation is complete.</p><details className="academy-disclosure"><summary>Inspect the starter</summary><CodeBlock code={exercise.starter} file={`exercises/${id}/starter.go`}/></details><details className="academy-disclosure"><summary>Reference solution — try it first</summary><CodeBlock code={exercise.solution} file={`exercises/${id}/solution.go`}/><div className="academy-local-command"><Terminal size={14}/><code>{`go test -tags solution ./${id}/`}</code></div></details><div className="exercise-links"><a href="downloads/gopath-v2-content.zip" download><Download size={14}/> Download learning materials</a>{exercise.tests.map(path=><a href={`#/source/${path}`} key={path}>View tests <ArrowUpRight size={12}/></a>)}{standalone&&<a href={`#/lesson/${exercise.lessonId}`}>Read the lesson <ArrowUpRight size={12}/></a>}</div></section>;}
