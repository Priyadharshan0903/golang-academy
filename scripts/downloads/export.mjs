import {cpSync,existsSync,lstatSync,mkdirSync,mkdtempSync,readFileSync,readdirSync,writeFileSync} from 'node:fs';
import {createHash} from 'node:crypto';
import {execFileSync} from 'node:child_process';
import {dirname,extname,join,posix,relative,resolve} from 'node:path';
import {tmpdir} from 'node:os';
import {fileURLToPath} from 'node:url';
import {repositoryLink} from '../../shared/repository.ts';

const root=resolve(dirname(fileURLToPath(import.meta.url)),'../..');
const hash=value=>createHash('sha256').update(value).digest('hex');
const walk=directory=>readdirSync(directory,{withFileTypes:true}).sort((a,b)=>a.name.localeCompare(b.name,'en')).flatMap(e=>e.isDirectory()?walk(join(directory,e.name)):[join(directory,e.name)]);
const read=path=>readFileSync(path,'utf8');
const write=(directory,name,text)=>{const path=join(directory,name);mkdirSync(dirname(path),{recursive:true});writeFileSync(path,text);};
const fence=(text,language='')=>{const size=Math.max(3,...[...text.matchAll(/`+/g)].map(m=>m[0].length+1));const delimiter='`'.repeat(size);return `${delimiter}${language}\n${text.trimEnd()}\n${delimiter}`;};
const blocked=new Set(['node_modules','.git','.github','public','resources','dist','work','outputs']);

function copyLearningTree(source,target){
 for(const entry of readdirSync(source,{withFileTypes:true}).sort((a,b)=>a.name.localeCompare(b.name,'en'))){
  if(entry.name.startsWith('.')||blocked.has(entry.name))continue;
  const from=join(source,entry.name),to=join(target,entry.name);
  if(lstatSync(from).isSymbolicLink())throw new Error(`Refusing symlink in learner sources: ${from}`);
  if(entry.isDirectory())copyLearningTree(from,to);
  else{mkdirSync(dirname(to),{recursive:true});cpSync(from,to);}
 }
}

async function exportV1(directory){
 const {allLessons,modules}=await import('../../content/v1/curriculum.ts');
 const {cases}=await import('../../content/v1/cases.ts');
 write(directory,'go.mod','module example.com/go-by-doing\n\ngo 1.26.0\n');
 copyLearningTree(join(root,'content/v1/projects'),join(directory,'projects'));
 for(const l of allLessons){
  const example=`examples/${l.moduleId}/${l.id}`;
  const code=execFileSync('gofmt',[],{input:l.code,encoding:'utf8'});
  write(directory,`${example}/main.go`,code);
  write(directory,`${example}/expected.txt`,l.output);
  write(directory,`lessons/${l.moduleId}/${l.id}.md`,[
   `# ${l.title}`,l.summary,'## Understand',l.explain.join('\n\n'),'## Run',
   `From the pack root: \`go run ${example}/main.go\`.`,fence(code,'go'),'## Expected output',fence(l.output,'text'),
   '## Watch out',l.watch,'## Practice',l.task,'### Hint',l.hint,'### Explained solution',l.solution,
   '## Knowledge check',l.quiz.question,l.quiz.options.map((v,i)=>`${i+1}. ${v}`).join('\n'),
   `Answer: ${l.quiz.correct+1}. ${l.quiz.why}`,'## Official reference',l.reference,
  ].join('\n\n')+'\n');
 }
 for(const c of cases)write(directory,`case-studies/${c.id}.md`,[
  `# ${c.title}`,'Original teaching scenario and build brief.',c.scenario,'## Constraints',c.constraints.map(x=>`- ${x}`).join('\n'),
  '## Engineering decisions',c.decisions.map(([h,t])=>`### ${h}\n\n${t}`).join('\n\n'),'## Failure mode',c.failure,'## Build brief',c.task,
  '## Acceptance criteria',c.acceptance.map(x=>`- [ ] ${x}`).join('\n'),
 ].join('\n\n')+'\n');
 write(directory,'INDEX.md','# Go by Doing V1 learning materials\n\n'+modules.map(m=>`## ${m.title}\n\n`+m.lessons.map(l=>`- [${l.title}](lessons/${m.id}/${l.id}.md)`).join('\n')).join('\n\n')+'\n\n## Case studies\n\n'+cases.map(c=>`- [${c.title}](case-studies/${c.id}.md)`).join('\n')+'\n\n## Guided project\n\n- [Website checker starter](projects/website-checker/starter/README.md)\n- [Reference solution](projects/website-checker/solution/README.md)\n');
 write(directory,'README.md',`# Go by Doing V1 learner pack\n\nStart with [the curriculum index](INDEX.md). Includes ${allLessons.length} lessons in ${modules.length} modules, runnable examples, original case-study briefs, and the website-checker starter and solution. Read Markdown in any text editor; no website tooling is required.\n\n## Run examples\n\nInstall Go. Examples were originally checked with Go 1.26.5. From this pack root, run \`go run examples/getting-started/hello-go/main.go\`. Each example is independent; its expected.txt records expected output. The website-checker project has its own Go modules. Run \`go test ./...\` inside its solution directory; starter tests intentionally fail until implemented.\n\n## Answers and licenses\n\nHints, explained answers, and reference solutions are clearly labeled. Try the work before reading them. Original course text: CC BY 4.0, attribution Go by Doing. Code: MIT; see LICENSE and LICENSE-CONTENT.md.\n`);
 write(directory,'LICENSE',read(join(root,'LICENSE')));
 write(directory,'LICENSE-CONTENT.md','# Original Go by Doing content\n\nOriginal lesson text, quizzes, project briefs, and explanations: Go by Doing contributors, CC BY 4.0.\nhttps://creativecommons.org/licenses/by/4.0/\n\nCode is MIT licensed; see LICENSE. Third-party linked documentation retains its own terms.\n');
 return {modules:modules.length,lessons:allLessons.length,cases:cases.length,projects:1,examples:allLessons.length};
}

function exportV2(directory){
 const course=JSON.parse(read(join(root,'content/v2/gopath.json')));
 const source=join(root,'content-source/gopath-academy');
 for(const folder of ['examples','exercises','projects'])copyLearningTree(join(source,folder),join(directory,folder));
 // These server and Kubernetes examples are explicitly discussed by the course.
 // The Hugo-building Dockerfile and website publication tooling are excluded.
 for(const folder of ['deploy/server','deploy/k8s'])copyLearningTree(join(source,folder),join(directory,folder));
 for(const name of ['go.mod','go.sum','LICENSE','LICENSE-CONTENT.md'])if(existsSync(join(source,name)))write(directory,name,read(join(source,name)));
 const pages=[...course.lessons,...course.caseStudies,...course.projects];
 const output=p=>p.kind==='lesson'?`lessons/${p.id}.md`:p.kind==='case'?`case-studies/${p.id}.md`:`projects/${p.id}/README.md`;
 const routeFiles=new Map(pages.map(p=>[`#/${p.kind}/${p.id}`,output(p)]));
 routeFiles.set('#/curriculum','INDEX.md');
 for(const route of ['#/cases','#/projects','#/'])routeFiles.set(route,'INDEX.md');
 for(const m of course.modules)routeFiles.set(`#/module/${m.id}`,`lessons/${m.id}/README.md`);
 const unresolved=[];
 function localLink(href,page,from){
  if(/^(?:https?:|mailto:)/i.test(href)||href.startsWith('#')&&!href.startsWith('#/'))return repositoryLink(href);
  let route,anchor='';
  if(href.startsWith('#/'))route=href;
  else{
   const base=page.kind==='project'?page.sourceFile:page.canonical.replace(/^\//,'');
   const url=new URL(href,'https://content.local/'+base);
   const path=decodeURIComponent(url.pathname).replace(/\/index\.html$/,'').replace(/\/$/,'').replace(/\.md$/,'');
   anchor=url.hash;
   route=course.routes[path]||course.routes[path+'.md'];
   if(!route&&page.kind==='lesson')route=course.routes['/learn/'+page.moduleId+'/'+href.replace(/^\.\//,'').replace(/\.md$/,'').replace(/\/$/,'')];
   if(!route&&['/learn','/case-studies','/projects'].includes(path))route='#/curriculum';
  }
  let destination=routeFiles.get(route);
  if(route?.startsWith('#/source/'))destination=decodeURIComponent(route.slice(9));
  if(!destination){unresolved.push({page:page.id,href});return href;}
  return posix.relative(posix.dirname(from),destination)+anchor;
 }
 function convert(body,page,from,depth=0){
  if(depth>15)throw new Error('Recursive content shortcode.');
  body=body.replace(/^```gopath-(example|exercise|callout|disclosure)\r?\n([^\n]+)\r?\n```/gm,(_,kind,id)=>{
   id=id.trim();
   if(kind==='example'){
    const item=course.embeds[id];if(!item)throw new Error(`Missing code embed ${id}`);
    const language=extname(item.file).slice(1).replace('yml','yaml');
    return `Source: \`${item.file}\`\n\n${fence(item.code,language)}${item.command?'\n\nLocal command:\n\n'+fence(item.command,'sh'):''}`;
   }
   if(kind==='exercise'){
    const item=course.exercises[id];if(!item)throw new Error(`Missing exercise ${id}`);
    return `## Exercise: ${item.title}\n\n${convert(item.body,page,from,depth+1)}\n\n### Starter\n\n${fence(item.starter,'go')}\n\n${fence(`cd exercises\ngo test ./${id}/`,'sh')}\n\n### Reference solution — try the exercise first\n\n${fence(item.solution,'go')}\n\n${fence(`go test -tags solution ./${id}/`,'sh')}`;
   }
   const item=(kind==='callout'?course.callouts:course.disclosures)[id];if(!item)throw new Error(`Missing ${kind} ${id}`);
   const text=convert(item.body,page,from,depth+1);
   return kind==='callout'?text.split('\n').map(line=>'> '+line).join('\n'):`### ${item.title}\n\n${text}`;
  });
  // Lesson commands must address the exported Markdown tree.
  body=body.replaceAll('../../site/content/learn/','../../lessons/');
  let inFence=false;
  return body.split('\n').map(line=>{
   if(/^\s*`{3,}/.test(line)){inFence=!inFence;return line;}
   if(inFence)return line;
   return line.split(/(`+[^`]*`+)/).map(part=>part.startsWith('`')?part:part.replace(/(!?\[[^\]]*\]\()([^\s)]+)(\))/g,(_,before,href,after)=>before+localLink(href,page,from)+after).replace(/https?:\/\/[^\s<>)\]]+/g,repositoryLink)).join('');
  }).join('\n');
 }
 for(const p of pages){const name=output(p);write(directory,name,`# ${p.title}\n\n${p.description||''}\n\n${convert(p.body,p,name)}\n`);}
 for(const m of course.modules){
  const name=`lessons/${m.id}/README.md`;
  const page={id:m.id,kind:'module',canonical:`/learn/${m.id}/`,sourceFile:''};
  write(directory,name,`# ${m.title}\n\n${convert(m.intro,page,name)}\n\n`+course.lessons.filter(l=>l.moduleId===m.id).map(l=>`- [${l.title}](${posix.relative(posix.dirname(name),output(l))})`).join('\n')+'\n');
 }
 if(unresolved.length)throw new Error('Unresolved learner links: '+JSON.stringify(unresolved,null,2));
 write(directory,'INDEX.md','# GoPath V2 learning materials\n\n'+course.modules.map(m=>`- [${m.title}](lessons/${m.id}/README.md)`).join('\n')+'\n\n## Case studies\n\n'+course.caseStudies.map(p=>`- [${p.title}](${output(p)})`).join('\n')+'\n\n## Projects\n\n'+course.projects.map(p=>`- [${p.title}](${output(p)})`).join('\n')+'\n');
 write(directory,'README.md',`# GoPath V2 learner pack\n\nStart with [the curriculum index](INDEX.md). This pack includes ${course.lessons.length} lessons, ${Object.keys(course.exercises).length} exercises, ${course.caseStudies.length} case studies, and ${course.projects.length} projects. Markdown is readable without React, Node.js, Vinext, or Hugo.\n\n## Requirements and commands\n\nThe imported course targets Go ${course.targetGo}; compatibility checks were originally scoped to Go ${course.verifiedToolchain}. Read each lesson's Go-version notes and build tags. The root and nested modules preserve their original go.mod and go.sum files. Database, Redis, third-party, and Kubernetes examples require their documented services or dependencies.\n\nRun an example using its displayed local command. Exercise tests deliberately fail for unfinished starters. From exercises/, run \`go test -tags solution ./...\` to check reference solutions. For the word-frequency checkpoint, exported lesson text lives under lessons/, not site/content/.\n\n## Learning assets and answers\n\nThe deploy/server and deploy/k8s directories are teaching examples referenced in the Kubernetes lesson. Website publishing Dockerfiles, CI, Hugo templates, and internal plans are excluded. Reference solutions are included and labeled; attempt the exercise before opening them. Code embeds and course callouts have been converted into ordinary Markdown.\n\n## Attribution\n\nAdapted from GoPath Academy by Priyadharshan Senthil and contributors: ${course.sourceRepository}. Written content and adaptations retain CC BY-SA 4.0; code retains MIT. See LICENSE and LICENSE-CONTENT.md. Changes include Markdown export, resolved source embeds, local links, and export-specific commands.\n`);
 return {modules:course.modules.length,lessons:course.lessons.length,exercises:Object.keys(course.exercises).length,cases:course.caseStudies.length,projects:course.projects.length,embeds:Object.keys(course.embeds).length,sourceFiles:Object.keys(course.sourceFiles).length};
}

function archive(directory,destination,prefix){execFileSync('python3',[join(root,'scripts/downloads/archive.py'),directory,destination,prefix],{stdio:'inherit'});}

export async function exportEdition(edition){
 if(!['v1','v2'].includes(edition))throw new Error('Choose v1 or v2.');
 const directory=mkdtempSync(join(tmpdir(),`go-learning-${edition}-`));
 const inventory=edition==='v1'?await exportV1(directory):exportV2(directory);
 const files=walk(directory).map(path=>({path:relative(directory,path).split('\\').join('/'),sha256:hash(readFileSync(path)),bytes:readFileSync(path).byteLength}));
 const revision=hash(JSON.stringify(files));
 const manifest={format:1,edition,release:`${edition}-${revision.slice(0,12)}`,contentRevision:revision,inventory,files};
 write(directory,'manifest.json',JSON.stringify(manifest,null,2)+'\n');
 const output=join(root,'apps',edition,'public/downloads');mkdirSync(output,{recursive:true});
 const name=edition==='v1'?'go-by-doing-v1-content':'gopath-v2-content';
 const destination=join(output,name+'.zip');archive(directory,destination,name);
 const legacy=edition==='v1'?'go-by-doing-examples.zip':'gopath-course-v2.zip';
 cpSync(destination,join(output,legacy));
 if(edition==='v1')for(const kind of ['starter','solution'])archive(join(root,`content/v1/projects/website-checker/${kind}`),join(output,`website-checker-${kind}.zip`),`website-checker-${kind}`);
 else{mkdirSync(join(root,'public/downloads'),{recursive:true});cpSync(destination,join(root,'public/downloads',name+'.zip'));cpSync(destination,join(root,'public/downloads',legacy));}
 const report={...manifest,files:files.length,archive:relative(root,destination),archiveSha256:hash(readFileSync(destination)),bytes:readFileSync(destination).byteLength,extractedDirectory:directory};
 mkdirSync(join(root,'work/separation'),{recursive:true});
 writeFileSync(join(root,`work/separation/${edition}-download.json`),JSON.stringify(report,null,2)+'\n');
 console.log(JSON.stringify(report,null,2));return report;
}

if(process.argv[1]&&resolve(process.argv[1])===fileURLToPath(import.meta.url)){
 const edition=process.argv[2]||'all';
 for(const item of edition==='all'?['v1','v2']:[edition])await exportEdition(item);
}
