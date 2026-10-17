// Supply the short-lived private-site bypass token over hidden stdin only.
import {readFileSync} from 'node:fs';
import {createHash} from 'node:crypto';

console.log('Ready for hosted verification JSON on stdin (input is hidden).');
const input=await new Promise((resolve,reject)=>{
 let data='';
 if(process.stdin.isTTY)process.stdin.setRawMode(true);
 process.stdin.setEncoding('utf8');
 process.stdin.on('data',chunk=>{
  data+=chunk;
  if(!data.includes('\n')&&!data.includes('\r'))return;
  if(process.stdin.isTTY)process.stdin.setRawMode(false);
  process.stdin.pause();
  try{resolve(JSON.parse(data.trim()));}catch{reject(new Error('Invalid verification JSON.'));}
 });
 process.stdin.resume();
});
const headers={'OAI-Sites-Authorization':`Bearer ${input.token}`};
const request=path=>fetch(new URL(path,input.url),{headers,signal:AbortSignal.timeout(30000)});
const response=await request('/');
if(!response.ok)throw new Error(`Hosted overview HTTP ${response.status}`);
const html=await response.text();
const path=input.edition==='v1'?'/downloads/go-by-doing-v1-content.zip':'/downloads/gopath-v2-content.zip';
if(!html.includes(path))throw new Error('Hosted overview points to a different learner archive.');
const archive=await request(path);
if(!archive.ok)throw new Error(`Hosted archive HTTP ${archive.status}`);
const bytes=Buffer.from(await archive.arrayBuffer());
const sha256=value=>createHash('sha256').update(value).digest('hex');
const expected=sha256(readFileSync(input.archivePath));
if(sha256(bytes)!==expected)throw new Error('Hosted learner archive differs from the verified local release.');
const legacy=input.edition==='v1'?'/downloads/go-by-doing-examples.zip':'/downloads/gopath-course-v2.zip';
const old=await request(legacy);
if(!old.ok||sha256(Buffer.from(await old.arrayBuffer()))!==expected)throw new Error('Legacy download is not compatible.');
console.log(JSON.stringify({edition:input.edition,url:input.url,overviewHTTP:response.status,archiveHTTP:archive.status,archiveSha256:expected,legacyDownload:'passed'}));
