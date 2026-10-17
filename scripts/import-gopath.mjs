import {readFileSync,writeFileSync,readdirSync,mkdirSync,statSync} from 'node:fs';
import {resolve,join,relative,extname} from 'node:path';
import {createRequire} from 'node:module';
import {createHash} from 'node:crypto';
import {execFileSync} from 'node:child_process';
import {repositoryUrl} from '../shared/repository.ts';
const require=createRequire(import.meta.url),yaml=require('js-yaml');
const root=resolve('content-source/gopath-academy');
const contentRoot=join(root,'site/content');
const files=dir=>readdirSync(dir,{withFileTypes:true}).sort((a,b)=>a.name.localeCompare(b.name)).flatMap(x=>x.isDirectory()?files(join(dir,x.name)):[join(dir,x.name)]);
const read=path=>readFileSync(path,'utf8');
const parse=text=>{const found=text.match(/^---\r?\n([\s\S]*?)\r?\n---\r?\n?/);return {meta:found?yaml.load(found[1])||{}:{},body:found?text.slice(found[0].length):text};};
const pageFiles=files(contentRoot).filter(x=>x.endsWith('.md'));
const sourceFiles={};
for(const dir of ['examples','exercises','projects','deploy'])for(const path of files(join(root,dir))){
 if(!['.go','.mod','.sum','.md','.sql','.txt','.yaml','.yml','.json','.sh'].includes(extname(path))&&!['Makefile','Dockerfile'].includes(path.split('/').pop()))continue;
 if(statSync(path).size>300000)continue;
 sourceFiles[relative(root,path)]={path:relative(root,path),content:read(path)};
}
for(const name of ['go.mod','LICENSE','LICENSE-CONTENT.md'])sourceFiles[name]={path:name,content:read(join(root,name))};
const embeds={},callouts={},disclosures={},exercises={},errors=[],unhandled=[];let embedCount=0;
const attrs=text=>Object.fromEntries([...text.matchAll(/([\w-]+)="([^"]*)"/g)].map(m=>[m[1],m[2]]));
const fence=(kind,id)=>`\n\n\`\`\`gopath-${kind}\n${id}\n\`\`\`\n\n`;
function regionCode(file,region){
 const all=sourceFiles[file];if(!all){errors.push(`Missing source ${file}`);return '';}
 let inside=!region,found=!region;const lines=[];
 for(const line of all.content.split('\n')){
  const start=line.trim().match(/^\/\/ \[start (.+)\]$/),end=line.trim().match(/^\/\/ \[end (.+)\]$/);
  if(start){if(region===start[1])inside=found=true;continue;}
  if(end){if(region===end[1])inside=false;continue;}
  if(inside)lines.push(line);
 }
 if(!found)errors.push(`Missing region ${region} in ${file}`);
 return lines.join('\n').trim();
}
function convert(markdown,page){
 let output=markdown;
 output=output.replace(/<details[^>]*>\s*<summary[^>]*>([\s\S]*?)<\/summary>([\s\S]*?)<\/details>/g,(_,title,body)=>{const id=`disclosure-${Object.keys(disclosures).length}`;disclosures[id]={title:title.replace(/<[^>]+>/g,'').trim(),body:convert(body,page)};return fence('disclosure',id);});
 output=output.replace(/\{\{<\s*exercise\s+([^>]*?)>\}\}([\s\S]*?)\{\{<\s*\/exercise\s*>\}\}/g,(_,attributes,body)=>{
  const {dir}=attrs(attributes);const starter=`exercises/${dir}/starter.go`,solution=`exercises/${dir}/solution.go`;
  if(!sourceFiles[starter]||!sourceFiles[solution])errors.push(`Missing exercise ${dir}`);
  exercises[dir]={id:dir,moduleId:page.moduleId,lessonId:page.id,title:page.title,body:convert(body,page),starter:sourceFiles[starter]?.content||'',solution:sourceFiles[solution]?.content||'',tests:Object.keys(sourceFiles).filter(x=>x.startsWith(`exercises/${dir}/`)&&x.endsWith('_test.go'))};
  return fence('exercise',dir);
 });
 output=output.replace(/\{\{<\s*callout\s*([^>]*?)>\}\}([\s\S]*?)\{\{<\s*\/callout\s*>\}\}/g,(_,attributes,body)=>{const id=`callout-${Object.keys(callouts).length}`;callouts[id]={type:attrs(attributes).type||'info',body:convert(body,page)};return fence('callout',id);});
 output=output.replace(/\{\{<\s*example\s+([^>]*?)>\}\}/g,(_,attributes)=>{
  const {file,region,run}=attrs(attributes),id=`example-${embedCount++}`;
  const content=regionCode(file,region);
  const full=sourceFiles[file]?.content||'';
  const imports=[...full.matchAll(/"([^"\n]+)"/g)].map(m=>m[1]);
  const canRun=file?.endsWith('.go')&&!file.endsWith('_test.go')&&/\bpackage main\b/.test(full)&&/func main\s*\(/.test(full)&&!full.includes('//go:build')&&!imports.some(x=>/^github\.|^golang\.|^example\./.test(x));
  embeds[id]={id,file,region:region||'',command:run||'',code:content,canRun,lessonId:page.id};
  return fence('example',id);
 });
 output=output.replace(/\{\{<\s*card\s+([^>]*?)>\}\}/g,(_,a)=>{const x=attrs(a);return `\n- [${x.title||x.link}](${x.link})${x.subtitle?' — '+x.subtitle:''}\n`;});
 output=output.replace(/\{\{<\s*tab\s+([^>]*?)>\}\}/g,(_,a)=>`\n\n### ${attrs(a).name||'Option'}\n\n`);
 output=output.replace(/\{\{<\s*filetree\/(folder|file)\s+([^>]*?)>\}\}/g,(_,kind,a)=>`\n- ${kind==='folder'?'📁':'📄'} ${attrs(a).name||'File'}\n`);
 output=output.replace(/\{\{<\s*\/?(?:cards|tabs|tab|steps|filetree\/container|filetree\/folder)\s*>\}\}/g,'\n');
 output=output.replace(/\{\{%\s*\/?steps\s*%\}\}/g,'\n');
 output=output.replace(/<sup>(.*?)<\/sup>/g,'^$1').replace(/<br\s*\/?>/g,'\n').replace(/<\/?div\b[^>]*>/g,'\n');
 const leftover=output.match(/\{\{[<%][\s\S]*?[>%]\}\}/g)||[];
 if(leftover.length)unhandled.push({page:page.id,shortcodes:leftover});
 return output.trim();
}
const modules=pageFiles.filter(p=>/\/learn\/[^/]+\/_index\.md$/.test(p)).map(path=>{const {meta,body}=parse(read(path)),id=relative(contentRoot,path).split('/')[1];return {id,title:meta.title,shortTitle:meta.title.replace(/^\d+\.\s*/,''),description:meta.description||'',weight:Number(meta.weight),intro:body,lessons:[]};}).sort((a,b)=>a.weight-b.weight);
for(const mod of modules){mod.intro=convert(mod.intro.replace(/\{\{<\s*cards\s*>\}\}[\s\S]*?\{\{<\s*\/cards\s*>\}\}/g,''),{id:mod.id,title:mod.title,moduleId:mod.id});}
const lessons=[];const caseStudies=[];const pages=[];
for(const path of pageFiles){
 const rel=relative(contentRoot,path);if(rel.endsWith('_index.md'))continue;
 const {meta,body}=parse(read(path));
 const parts=rel.replace(/\.md$/,'').split('/');const isLesson=parts[0]==='learn';
 const id=isLesson?parts.slice(1).join('/'):parts.slice(1).join('/');
 const page={id,kind:isLesson?'lesson':'case',sourceFile:'site/content/'+rel,canonical:'/'+parts.join('/')+'/',title:meta.title||id,shortTitle:meta.linkTitle||meta.title||id,description:meta.description||'',weight:Number(meta.weight||0),moduleId:isLesson?parts[1]:'',goVersion:meta.goVersion||'',body:'',minutes:Math.max(5,Math.ceil(body.split(/\s+/).length/190)),headings:[],searchText:''};
 page.body=convert(body,page);
 page.headings=[...page.body.matchAll(/^##\s+(.+)$/gm)].map(x=>x[1]);
 page.searchText=(page.title+' '+page.description+' '+page.body.replace(/```[\s\S]*?```/g,'').replace(/[#*_`]/g,'')).toLowerCase();
 if(isLesson){lessons.push(page);modules.find(m=>m.id===page.moduleId)?.lessons.push(page.id);}else caseStudies.push(page);
 pages.push(page);
}
lessons.sort((a,b)=>modules.findIndex(m=>m.id===a.moduleId)-modules.findIndex(m=>m.id===b.moduleId)||a.weight-b.weight);
for(const mod of modules)mod.lessons=lessons.filter(x=>x.moduleId===mod.id).map(x=>x.id);
caseStudies.sort((a,b)=>a.weight-b.weight);
const projects=[];
for(const path of files(join(root,'projects')).filter(p=>p.endsWith('/README.md')&&relative(join(root,'projects'),p).split('/').length===2)){
 const id=relative(join(root,'projects'),path).split('/')[0],body=read(path);
 const page={id,kind:'project',sourceFile:relative(root,path),canonical:'/projects/'+id+'/',title:({greet:'Greet CLI',wordfreq:'Word frequency counter',crawler:'Concurrent web crawler',ratelimit:'Token-bucket rate limiter',miniredis:'Mini Redis server',shortener:'URL shortener',jobqueue:'PostgreSQL job queue','shortlink-operator':'ShortLink Kubernetes operator'})[id]||id,shortTitle:id,description:body.split('\n').find(x=>x.trim()&&!x.startsWith('#')&&!x.startsWith('```'))||'',weight:0,moduleId:'',goVersion:'',body:'',minutes:0,headings:[],searchText:'',files:Object.keys(sourceFiles).filter(f=>f.startsWith('projects/'+id+'/'))};
 page.body=convert(body.replace(/^# .+\n/,''),page);page.searchText=(page.title+' '+page.body).toLowerCase();projects.push(page);pages.push(page);
}
const routes={};
for(const m of modules)routes['/learn/'+m.id]=`#/module/${m.id}`;
for(const p of pages)routes[p.canonical.replace(/\/$/,'')]=`#/${p.kind}/${p.id}`;
routes['/learn']='#/curriculum';routes['/projects']='#/projects';routes['/case-studies']='#/cases';routes['/']='#/';
for(const path of Object.keys(sourceFiles))routes['/'+path]=`#/source/${path}`;
const data={title:'Go by Doing — GoPath Edition',version:'V2',sourceTitle:'GoPath Academy',sourceRepository:repositoryUrl,contentLicense:'CC BY-SA 4.0',codeLicense:'MIT',targetGo:'1.27',verifiedToolchain:'1.26.5',modules,lessons,caseStudies,projects,embeds,callouts,disclosures,exercises,sourceFiles,routes};
mkdirSync('content/v2',{recursive:true});writeFileSync('content/v2/gopath.json',JSON.stringify(data));
const report={contentPagesFound:pageFiles.length,modules:modules.length,lessons:lessons.length,exercises:Object.keys(exercises).length,caseStudies:caseStudies.length,projects:projects.length,sourceEmbeds:embedCount,sourceFiles:Object.keys(sourceFiles).length,runnableMainEmbeds:Object.values(embeds).filter(x=>x.canRun).length,unhandled,errors,snapshotHash:createHash('sha256').update(read('content-source/import-manifest.json')).digest('hex')};
writeFileSync('work/import-report.json',JSON.stringify(report,null,2));
if(errors.length||unhandled.length){console.error(JSON.stringify(report,null,2));process.exit(1);}
console.log(JSON.stringify(report,null,2));
