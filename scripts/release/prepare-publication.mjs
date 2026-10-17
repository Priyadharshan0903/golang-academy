import {cpSync,existsSync,lstatSync,mkdirSync,readFileSync,readdirSync,unlinkSync,writeFileSync} from 'node:fs';
import {execFileSync} from 'node:child_process';
import {join,relative,resolve} from 'node:path';
import {root} from '../editions.mjs';

const edition=process.argv[2];
if(!['v1','v2'].includes(edition))throw new Error('Choose v1 or v2.');
const build=JSON.parse(readFileSync(join(root,`work/separation/${edition}-build.json`),'utf8'));
const destination=join(root,'work/publish',edition);
if(!existsSync(join(destination,'.git')))throw new Error('Open the existing Site source with the Sites source helper first.');
const blocked=new Set(['.git','node_modules','dist','.sites-runtime','.wrangler','.vinext','.next']);
const inputFiles=[];
function copyInputs(source,target){
 mkdirSync(target,{recursive:true});
 for(const e of readdirSync(source,{withFileTypes:true})){
  if(blocked.has(e.name)||e.name==='.env'||e.name.startsWith('.env.')||e.name.endsWith('.tsbuildinfo'))continue;
  const from=join(source,e.name),to=join(target,e.name);
  if(lstatSync(from).isSymbolicLink())throw new Error('Unexpected publication input symlink.');
  if(e.isDirectory())copyInputs(from,to);
  else{cpSync(from,to);inputFiles.push(relative(destination,to));}
 }
}
copyInputs(build.directory,destination);
const wanted=new Set(inputFiles);
// Remove obsolete tracked source files only. Preserve Git metadata and local state.
const tracked=execFileSync('git',['ls-files','-z'],{cwd:destination,encoding:'utf8'}).split('\0').filter(Boolean);
for(const file of tracked){
 const parts=file.split('/');
 if(wanted.has(file)||blocked.has(parts[0])||parts.some(p=>p==='.env'||p.startsWith('.env.')))continue;
 const path=resolve(destination,file);
 if(!path.startsWith(destination+'/'))throw new Error('Invalid tracked publication path.');
 if(existsSync(path)&&lstatSync(path).isFile())unlinkSync(path);
}
const previousDist=join(destination,'dist');
// Dist is generated and ignored. Stale files from a previous package are excluded
// by using the freshly built tree; each publication checkout is used once here.
if(existsSync(previousDist))throw new Error('Publication output already exists; inspect before replacing it.');
cpSync(join(build.directory,'dist'),previousDist,{recursive:true});
writeFileSync(join(destination,'README.md'),`# Go by Doing ${edition.toUpperCase()} publication\n\nThis is an isolated, generated publication checkout for ${edition.toUpperCase()}. Source preparation, learner exports, and edition separation are maintained in the main Go by Doing workspace.\n\nUse Node.js 22.13 or newer. Install with npm ci and build with npm run build. The checked-in learner ZIPs correspond to the included content revision; regenerate them in the main workspace when curriculum sources change.\n`);
console.log(JSON.stringify({edition,destination,files:inputFiles.length,project_id:build.project_id,build:build.directory}));
