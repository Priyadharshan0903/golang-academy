import {readFileSync,writeFileSync} from 'node:fs';
import {createHash} from 'node:crypto';
import {resolve} from 'node:path';
import {build} from 'esbuild';
import {marked} from '../vendor/marked/lib/marked.esm.js';
await build({entryPoints:['content/v2/gopath.ts'],bundle:true,format:'esm',platform:'node',outfile:'work/gopath-data.mjs'});
const {course,pages,resolveLink}=await import('../work/gopath-data.mjs');
const links=[];const broken=[];
for(const page of pages){
 marked.walkTokens(marked.lexer(page.body),token=>{if(token.type==='link'){const target=resolveLink(token.href,page);links.push({page:page.id,href:token.href,target});if(target.startsWith('#/link-missing'))broken.push({page:page.id,href:token.href,target});}});
}
const snapshot=JSON.parse(readFileSync('content-source/import-manifest.json','utf8'));
const changed=[];
for(const file of snapshot.files){const path=resolve('content-source/gopath-academy',file.path);if(createHash('sha256').update(readFileSync(path)).digest('hex')!==file.sha256)changed.push(file.path);}
const report={modules:course.modules.length,lessons:course.lessons.length,exercises:Object.keys(course.exercises).length,cases:course.caseStudies.length,projects:course.projects.length,embeds:Object.keys(course.embeds).length,sourceFiles:Object.keys(course.sourceFiles).length,linksChecked:links.length,brokenLinks:broken,modifiedSnapshotFiles:changed};
writeFileSync('work/link-verification.json',JSON.stringify(report,null,2));
console.log(JSON.stringify(report,null,2));
process.exitCode=broken.length||changed.length?1:0;
