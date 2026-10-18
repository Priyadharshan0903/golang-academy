// Build two independent client apps into one GitHub Pages artifact.
import {cpSync,existsSync,mkdirSync,mkdtempSync,readFileSync,readdirSync,symlinkSync,unlinkSync,writeFileSync} from 'node:fs';
import {tmpdir} from 'node:os';
import {basename,dirname,join,relative,resolve} from 'node:path';
import {fileURLToPath} from 'node:url';
import {spawnSync} from 'node:child_process';

const root=resolve(dirname(fileURLToPath(import.meta.url)),'..');
const args=process.argv.slice(2);
let dependencies=join(root,'node_modules'),serveAfterBuild=false,previewOnly=false,port=8080;
for(let i=0;i<args.length;i++){
 const arg=args[i];
 if(arg==='--dependencies'&&args[i+1])dependencies=resolve(args[++i]);
 else if(arg==='--serve')serveAfterBuild=true;
 else if(arg==='--preview')previewOnly=true;
 else if(arg==='--port'&&/^\d+$/.test(args[i+1]||''))port=Number(args[++i]);
 else throw new Error('Usage: node scripts/pages.mjs [--serve|--preview] [--port 8080] [--dependencies /path/to/node_modules]');
}
if(port<1024||port>65535)throw new Error('Choose a preview port between 1024 and 65535.');
if(serveAfterBuild&&previewOnly)throw new Error('Choose either --serve or --preview.');
function serve(){
 let directory=join(root,'dist/pages');
 const reportPath=join(root,'work/pages/build.json');
 if(!existsSync(join(directory,'index.html'))||!existsSync(reportPath))throw new Error('Run npm run build before starting the preview.');
 const report=JSON.parse(readFileSync(reportPath,'utf8'));
 if(report.base!=='/'){
  directory=mkdtempSync(join(tmpdir(),'go-pages-preview-'));
  const mount=join(directory,...report.base.split('/').filter(Boolean));
  mkdirSync(dirname(mount),{recursive:true});
  symlinkSync(join(root,'dist/pages'),mount,'dir');
 }
 const result=spawnSync('python3',[join(root,'scripts/preview.py'),'--port',String(port),'--directory',directory,'--base',report.base],{stdio:'inherit'});
 if(result.error)throw result.error;
 process.exit(result.status??1);
}
if(previewOnly)serve();
const repository=process.env.GITHUB_REPOSITORY?.split('/')[1];
const requestedBase=process.env.PAGES_BASE_PATH??(repository&&!repository.toLowerCase().endsWith('.github.io')?`/${repository}/`:'/');
const base='/'+requestedBase.split('/').filter(Boolean).join('/')+(requestedBase.split('/').filter(Boolean).length?'/':'');
if(!/^\/(?:[A-Za-z0-9._~-]+\/)*$/.test(base)||base.includes('/../'))throw new Error('Invalid Pages base path.');
const artifact=mkdtempSync(join(tmpdir(),'go-pages-artifact-'));
const reports=[];
const run=(command,cwd)=>{
 const result=spawnSync(process.execPath,command,{cwd,stdio:'inherit',env:{...process.env,PATH:dirname(process.execPath)+':'+process.env.PATH}});
 if(result.error)throw result.error;
 if(result.status!==0)throw new Error('Pages build failed.');
};

for(const edition of ['v1','v2']){
 run(['scripts/downloads/export.mjs',edition],root);
 const directory=mkdtempSync(join(tmpdir(),`go-pages-${edition}-`));
 const copy=(from,to=from)=>{mkdirSync(dirname(join(directory,to)),{recursive:true});cpSync(join(root,from),join(directory,to),{recursive:true,filter:path=>!basename(path).startsWith('.env')&&!['node_modules','.git'].includes(basename(path))});};
 // Only frontend modules, this edition's content, and learner downloads enter
 // the static build. Hosting credentials and server routes are not inputs.
 mkdirSync(join(directory,'app'));
 for(const entry of readdirSync(join(root,'apps',edition,'app'),{withFileTypes:true})){
  if(entry.isFile()&&entry.name.endsWith('.tsx')&&!['page.tsx','layout.tsx'].includes(entry.name))copy(`apps/${edition}/app/${entry.name}`,`app/${entry.name}`);
 }
 copy(`content/${edition}`);
 copy('shared/styles');copy('shared/editions.ts');copy('shared/playground.ts');copy('shared/repository.ts');
 if(edition==='v2')copy('vendor/marked');
 copy(`apps/${edition}/public`,'public');copy('public/favicon.svg','public/favicon.svg');
 copy('postcss.config.mjs');
 symlinkSync(dependencies,join(directory,'node_modules'),'dir');
 writeFileSync(join(directory,'package.json'),'{"type":"module"}\n');
 const title=edition==='v1'?'Go by Doing — Original Edition':'Go by Doing — GoPath Edition V2';
 const description=edition==='v1'?'48 practical Go lessons, exercises, quizzes, and a guided project.':'116 GoPath lessons, 27 exercises, project sources, and company case studies.';
 writeFileSync(join(directory,'index.html'),`<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>${title}</title><meta name="description" content="${description}"><link rel="icon" href="${base}${edition}/favicon.svg"></head><body><div id="root"></div><noscript>Enable JavaScript to explore the interactive course. You can also <a href="downloads/${edition==='v1'?'go-by-doing-v1-content':'gopath-v2-content'}.zip">download the learning materials</a>.</noscript><script type="module" src="/main.tsx"></script></body></html>\n`);
 writeFileSync(join(directory,'main.tsx'),`import {createRoot} from 'react-dom/client';\nimport App from './app/${edition==='v1'?'learning-app':'gopath-app'}';\nimport './shared/styles/globals.css';\ncreateRoot(document.getElementById('root')!).render(<App/>);\n`);
 writeFileSync(join(directory,'vite.config.mjs'),`import {defineConfig} from 'vite';\nimport react from '@vitejs/plugin-react';\nimport {fileURLToPath} from 'node:url';\nexport default defineConfig({envDir:false,base:${JSON.stringify(base+edition+'/')},plugins:[react()],resolve:{alias:{'@content':fileURLToPath(new URL('./content',import.meta.url)),'@':fileURLToPath(new URL('./',import.meta.url))}}});\n`);
 run([join(dependencies,'vite/bin/vite.js'),'build'],directory);
 const output=join(directory,'dist');
 const javascript=readdirSync(join(output,'assets')).filter(name=>name.endsWith('.js')).map(name=>readFileSync(join(output,'assets',name),'utf8')).join('\n');
 if(javascript.includes(edition==='v1'?'425 resolved source embeds':'A LITTLE CODE. A LOT OF GO.'))throw new Error('Cross-edition content in Pages build.');
 if(javascript.includes('/api/run'))throw new Error('Static build still requires the API server.');
 cpSync(output,join(artifact,edition),{recursive:true});
 reports.push({edition,path:base+edition+'/',directory,output,built:true});
}

cpSync(join(root,'public/favicon.svg'),join(artifact,'favicon.svg'));
writeFileSync(join(artifact,'.nojekyll'),'');
writeFileSync(join(artifact,'index.html'),`<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Go by Doing — Choose your course</title><meta name="description" content="Learn Go with practical lessons, exercises, and projects. Choose the Original or GoPath edition."><link rel="icon" href="./favicon.svg"><style>
:root{color-scheme:light dark;font-family:system-ui,sans-serif;color:#203b32;background:#f4f7f2}*{box-sizing:border-box}body{margin:0;min-height:100svh;display:grid;place-items:center;padding:32px 20px}main{width:min(900px,100%)}.brand{font-size:30px;font-weight:750;color:#216e65}h1{font-size:clamp(30px,5vw,48px);letter-spacing:-1.5px;line-height:1.15}p{font-size:16px;line-height:1.7;color:#637168}.courses{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:20px;margin-top:32px}a{display:block;text-decoration:none;color:inherit;background:#fff;border:1px solid #d8e3d8;border-radius:16px;padding:28px}a:hover,a:focus-visible{border-color:#216e65;outline:2px solid #216e65;outline-offset:3px}small{font-size:14px;color:#216e65}h2{font-size:24px;margin:14px 0}.open{font-weight:600;color:#216e65}footer{font-size:14px;color:#637168;margin-top:24px}@media(max-width:600px){.courses{grid-template-columns:1fr}}@media(prefers-color-scheme:dark){:root{color:#dae4dc;background:#131e18}a{background:#1b2c23;border-color:#344a3d}p,footer{color:#a3b4a5}.brand,small,.open{color:#9cd1bd}}
</style></head><body><main><div class="brand">go.</div><h1>Small steps. Deeper Go.</h1><p>Pick your course. Both editions are free, with progress saved in your browser and learning materials to take with you.</p><div class="courses"><a href="./v1/"><small>ORIGINAL EDITION · V1</small><h2>Build your foundations.</h2><p>48 focused lessons, hands-on challenges, quizzes, and a guided website-checker project.</p><span class="open">Start the original course →</span></a><a href="./v2/"><small>GOPATH EDITION · V2</small><h2>Take the deeper path.</h2><p>116 lessons, 27 exercises, company case studies, and projects from GoPath Academy.</p><span class="open">Explore GoPath →</span></a></div><footer>No signup. Learn at your pace.</footer></main></body></html>\n`);
const destination=join(root,'dist/pages');
// Synchronize files without replacing directories: desktop filesystem profiles
// can permit file writes while rejecting directory removal. Prune stale assets
// so previous bundles and downloads do not survive in the next publication.
const walk=directory=>readdirSync(directory,{withFileTypes:true}).flatMap(entry=>entry.isDirectory()?walk(join(directory,entry.name)):[join(directory,entry.name)]);
const wanted=new Set(walk(artifact).map(file=>relative(artifact,file)));
if(existsSync(destination))for(const file of walk(destination))if(!wanted.has(relative(destination,file)))unlinkSync(file);
mkdirSync(destination,{recursive:true});
cpSync(artifact,destination,{recursive:true});
mkdirSync(join(root,'work/pages'),{recursive:true});
writeFileSync(join(root,'work/pages/build.json'),JSON.stringify({base,destination,editions:reports},null,2)+'\n');
console.log(`GitHub Pages artifact: ${destination}\nV1: ${base}v1/\nV2: ${base}v2/`);
if(serveAfterBuild)serve();
