import {cpSync,existsSync,mkdirSync,mkdtempSync,readFileSync,readdirSync,symlinkSync,writeFileSync} from 'node:fs';
import {tmpdir} from 'node:os';
import {basename,dirname,join,resolve} from 'node:path';
import {fileURLToPath} from 'node:url';
import {spawnSync} from 'node:child_process';

export const root=resolve(dirname(fileURLToPath(import.meta.url)),'..');
const editions=['v1','v2'];
const commonFiles=['package.json','package-lock.json','tsconfig.json','next.config.ts','postcss.config.mjs','cloudflare-env.d.ts','LICENSE'];
const commonScripts=['run-framework.mjs','execution-profile.mjs','sites-env.mjs','sites-env.sh','install-ci.mjs','npm-install.mjs','pnpm-install.mjs','install-ci.sh','install-pnpm.sh','build-verified.sh'];
const readJSON=path=>JSON.parse(readFileSync(path,'utf8'));
const run=(args,cwd=root)=>{
 const result=spawnSync(process.execPath,args,{cwd,stdio:'inherit',env:{...process.env,CHOKIDAR_USEPOLLING:'true',PATH:dirname(process.execPath)+':'+process.env.PATH}});
 if(result.error)throw result.error;
 if(result.status!==0)throw new Error(`Command failed (${result.status}): ${args.join(' ')}`);
};

export function prepareEdition(edition,{target,dependencies=true,dependencyDirectory=join(root,'node_modules')}={}){
 if(!editions.includes(edition))throw new Error('Choose v1 or v2.');
 const directory=target||mkdtempSync(join(tmpdir(),`go-by-doing-${edition}-`));
 mkdirSync(directory,{recursive:true});
 const copy=(source,destination=source)=>{
  const from=join(root,source),to=join(directory,destination);
  if(!existsSync(from))throw new Error(`Missing edition input: ${source}`);
  mkdirSync(dirname(to),{recursive:true});cpSync(from,to,{recursive:true,filter:source=>{const name=basename(source);return name!=='.env'&&!name.startsWith('.env.')&&name!=='node_modules'&&name!=='.git';}});
 };
 for(const file of commonFiles)copy(file);
 for(const file of commonScripts)copy('scripts/'+file);
 for(const folder of ['build','lib','shared'])copy(folder);
 copy('apps/'+edition+'/app','app');
 if(existsSync(join(root,'apps',edition,'.openai')))copy('apps/'+edition+'/.openai','.openai');
 copy('content/'+edition,'content/'+edition);
 copy('public/favicon.svg');
 copy('apps/'+edition+'/public/downloads','public/downloads');
 if(edition==='v2')copy('vendor/marked');
 // Only the selected edition and its own canonical content enter the build root.
 const pkg=readJSON(join(directory,'package.json'));
 pkg.scripts={dev:'node scripts/run-framework.mjs dev',build:'node scripts/run-framework.mjs build',start:'node --import ./scripts/sites-env.mjs ./node_modules/wrangler/bin/wrangler.js dev --config dist/server/wrangler.json --local --ip 127.0.0.1 --inspector-port 0'};
 writeFileSync(join(directory,'package.json'),JSON.stringify(pkg,null,2)+'\n');
 copy('vite.config.ts');
 writeFileSync(join(directory,'.gitignore'),'node_modules/\ndist/\n.vinext/\n.next/\n.wrangler/\n.sites-runtime/\n.env*\n*.tsbuildinfo\n');
 if(dependencies&&!existsSync(join(directory,'node_modules')))symlinkSync(dependencyDirectory,join(directory,'node_modules'),'dir');
 return directory;
}

export function buildEdition(edition,options={}){
 run(['scripts/downloads/export.mjs',edition]);
 const directory=prepareEdition(edition,options);
 run(['scripts/run-framework.mjs','build'],directory);
 const client=join(directory,'dist/client');
 const walk=dir=>readdirSync(dir,{withFileTypes:true}).flatMap(e=>e.isDirectory()?walk(join(dir,e.name)):[join(dir,e.name)]);
 const javascript=walk(client).filter(p=>p.endsWith('.js')).map(p=>readFileSync(p,'utf8')).join('\n');
 const forbidden=edition==='v1'?'425 resolved source embeds':'A LITTLE CODE. A LOT OF GO.';
 if(javascript.includes(forbidden))throw new Error(`Cross-edition content detected in ${edition}.`);
 const hosting=join(directory,'.openai/hosting.json');
 const report={edition,directory,dist:join(directory,'dist'),project_id:existsSync(hosting)?readJSON(hosting).project_id:undefined,javascriptBytes:Buffer.byteLength(javascript),built:true};
 mkdirSync(join(root,'work/separation'),{recursive:true});
 writeFileSync(join(root,`work/separation/${edition}-build.json`),JSON.stringify(report,null,2)+'\n');
 console.log(JSON.stringify(report,null,2));
 return report;
}

if(process.argv[1]&&resolve(process.argv[1])===fileURLToPath(import.meta.url)){
 process.chdir(root);
 const [command,edition,previewPort]=process.argv.slice(2);
 if(previewPort&&(!/^\d+$/.test(previewPort)||Number(previewPort)<1024||Number(previewPort)>65535))throw new Error("Invalid preview port.");
 if(!editions.includes(edition))throw new Error('Usage: node scripts/editions.mjs <prepare|build|dev|start> <v1|v2>');
 if(command==='build')buildEdition(edition);
 else if(command==='prepare'){
  run(['scripts/downloads/export.mjs',edition]);console.log(prepareEdition(edition));
 }else if(command==='dev'){
  run(['scripts/downloads/export.mjs',edition]);
  const directory=prepareEdition(edition),site=readJSON(join(root,'apps',edition,'site.json'));
  run(['scripts/run-framework.mjs','dev','--port',String(previewPort||site.port)],directory);
 }else if(command==='start'){
  const report=readJSON(join(root,`work/separation/${edition}-build.json`));
  const site=readJSON(join(root,'apps',edition,'site.json'));
  run(['--import','./scripts/sites-env.mjs','./node_modules/wrangler/bin/wrangler.js','dev','--config','dist/server/wrangler.json','--local','--ip','127.0.0.1','--port',String(previewPort||site.port),'--inspector-port','0'],report.directory);
 }else throw new Error('Choose prepare, build, dev, or start.');
}
