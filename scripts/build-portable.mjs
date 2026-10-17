// Some desktop hosts permit project writes but restrict directory replacement.
// Build in an OS temporary directory, then copy the completed artifact back.
import {cpSync,existsSync,mkdtempSync,readFileSync,readdirSync,symlinkSync,writeFileSync} from 'node:fs';
import {tmpdir} from 'node:os';
import {join,resolve} from 'node:path';
import {spawnSync} from 'node:child_process';
const source=process.cwd();
const scratch=mkdtempSync(join(tmpdir(),'go-by-doing-build-'));
const omitted=new Set(['node_modules','.git','.vinext','.wrangler','.sites-runtime','dist','work','outputs']);
for(const entry of readdirSync(source,{withFileTypes:true})) {
 if(omitted.has(entry.name)||entry.name==='.env'||entry.name.startsWith('.env.'))continue;
 cpSync(join(source,entry.name),join(scratch,entry.name),{recursive:true});
}
symlinkSync(join(source,'node_modules'),join(scratch,'node_modules'),'dir');
const result=spawnSync(process.execPath,['/Users/priyadharshan.senthil/.codex/plugins/cache/openai-curated-remote/sites/1.0.0/scripts/build-site.mjs'],{cwd:scratch,stdio:'inherit',env:process.env});
if(result.error)throw result.error;
if(result.status!==0)process.exit(result.status||1);
cpSync(join(scratch,'dist'),join(source,'dist'),{recursive:true});
writeFileSync(resolve('work/portable-build.json'),JSON.stringify({source,scratch,completed:true},null,2));
console.log('Completed production build; artifact copied to dist.');
