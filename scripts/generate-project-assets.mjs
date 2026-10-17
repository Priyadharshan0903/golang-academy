import {mkdirSync,readdirSync,readFileSync,writeFileSync,cpSync} from 'node:fs';
import {join,resolve,relative} from 'node:path';
import {execFileSync} from 'node:child_process';
const base=resolve('content/v1/projects/website-checker');
const descriptions={
 'go.mod':'The module path connects the CLI import to the checker package. No third-party dependencies.',
 'checker/checker.go':'The result contract, input validation, worker ownership, request deadline, and resource cleanup.',
 'checker/checker_test.go':'Behavior-focused tests use injected transports to control failures, concurrency, and cancellation.',
 'cmd/sitecheck/main.go':'Process concerns: flags, bounded URL input, signals, output, and exit codes.',
 'cmd/sitecheck/main_test.go':'CLI usage, input limits, canceled runs, and human-readable help.',
 'urls.example.txt':'An example input file. Actual remote responses depend on the network.',
 'README.md':'Run instructions, architecture, contracts, limits, and suggested extensions.',
};
function files(directory){return readdirSync(directory,{withFileTypes:true}).sort((a,b)=>a.name.localeCompare(b.name)).flatMap(entry=>entry.isDirectory()?files(join(directory,entry.name)):[join(directory,entry.name)]);}
const solution=files(join(base,'solution')).map(path=>({name:relative(join(base,'solution'),path).split('\\').join('/'),content:readFileSync(path,'utf8')})).map(file=>({...file,description:descriptions[file.name]||'Project source file.'}));
const testCount=solution.filter(f=>f.name.endsWith('_test.go')).reduce((sum,file)=>sum+(file.content.match(/^func Test/gm)||[]).length,0);
writeFileSync('content/v1/project-files.json',JSON.stringify({testCount,files:solution},null,2)+'\n');
const {exportEdition}=await import('./downloads/export.mjs');
await exportEdition('v1');
console.log(`Generated two project downloads and source viewer with ${testCount} tests.`);
