import assert from 'node:assert/strict';
import {existsSync,readFileSync,readdirSync,writeFileSync} from 'node:fs';
import {join,resolve,posix} from 'node:path';
import {execFileSync} from 'node:child_process';
import {marked} from '../../vendor/marked/lib/marked.esm.js';
import {root} from '../editions.mjs';
import {exportEdition} from '../downloads/export.mjs';

process.chdir(root);
const readJSON=path=>JSON.parse(readFileSync(path,'utf8'));
const walk=dir=>readdirSync(dir,{withFileTypes:true}).flatMap(e=>e.isDirectory()?walk(join(dir,e.name)):[join(dir,e.name)]);
const results=[];
for(const edition of ['v1','v2']){
 const previous=readJSON(`work/separation/${edition}-download.json`);
 const current=await exportEdition(edition);
 assert.equal(previous.archiveSha256,current.archiveSha256,`${edition} download must be deterministic`);
 const build=readJSON(`work/separation/${edition}-build.json`);
 const other=edition==='v1'?'v2':'v1';
 assert.ok(!existsSync(join(build.directory,'content',other)));
 assert.ok(!walk(join(build.directory,'public')).some(path=>path.includes(other==='v1'?'go-by-doing-v1-content':'gopath-v2-content')));
 const broken=[];
 for(const file of walk(current.extractedDirectory).filter(p=>p.endsWith('.md'))){
  const markdown=readFileSync(file,'utf8');
  marked.walkTokens(marked.lexer(markdown),token=>{
   if(!['link','image'].includes(token.type)||!token.href||/^(https?:|mailto:|#)/i.test(token.href))return;
   const path=decodeURIComponent(token.href.split('#')[0]);
   if(path&&!existsSync(resolve(file,'..',path)))broken.push({file:posix.relative(current.extractedDirectory,file),href:token.href});
  });
 }
 assert.deepEqual(broken,[],`${edition} exported Markdown has broken local links`);
 results.push({edition,reproducible:true,isolatedContent:true,localMarkdownLinks:'passed',archiveSha256:current.archiveSha256});
}
execFileSync('python3',['scripts/downloads/check.py'],{stdio:'inherit'});
writeFileSync('work/separation/verification.json',JSON.stringify({checks:results},null,2)+'\n');
console.log('PASS: separate app inputs, repeatable archives, Markdown links, ZIP integrity, and file manifests.');
