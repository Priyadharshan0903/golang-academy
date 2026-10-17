import assert from 'node:assert/strict';
import {mkdirSync,writeFileSync} from 'node:fs';
import {runGo} from '../shared/playground.ts';
const originalFetch=globalThis.fetch;
const checks=[];
async function check(name,fn){await fn();checks.push(name);console.log('PASS',name);}
try {
 for(const code of ['', '   ', '界'.repeat(10001)])await check('Reject empty or oversized code',async()=>{
  globalThis.fetch=async()=>{throw new Error('Invalid source must not be sent.');};
  await assert.rejects(runGo(code),/between 1 byte and 30 KB/);
 });
 await check('Use the fixed official CORS endpoint without credentials',async()=>{
  globalThis.fetch=async(url,options)=>{
   assert.equal(url,'https://play.golang.org/compile');
   assert.equal(options.method,'POST');assert.equal(options.credentials,'omit');
   assert.equal(options.headers['Content-Type'],'application/x-www-form-urlencoded');
   assert.equal(options.body.get('version'),'2');assert.equal(options.body.get('withVet'),'true');
   assert.equal(options.body.get('body'),'package main\nfunc main(){}');
   assert.ok(options.signal instanceof AbortSignal);
   return Response.json({Errors:'',Events:[{Message:'ok\n',Kind:'stdout'}],VetErrors:''});
  };
  const result=await runGo('package main\nfunc main(){}');assert.equal(result.data.events[0].Message,'ok\n');
 });
 await check('Preserve compiler errors, stderr, and vet output',async()=>{
  globalThis.fetch=async()=>Response.json({Errors:'syntax problem',Events:[{Kind:'stderr',Message:'panic: sample'}],VetErrors:'vet warning'});
  const {data}=await runGo('x');assert.equal(data.errors,'syntax problem');assert.equal(data.events[0].Kind,'stderr');assert.equal(data.vet,'vet warning');
 });
 await check('Normalize a program with no output',async()=>{
  globalThis.fetch=async()=>Response.json({});const {data}=await runGo('x');assert.deepEqual(data,{errors:'',events:[],vet:''});
 });
 await check('Handle an upstream busy response',async()=>{
  globalThis.fetch=async()=>new Response('busy',{status:429});const {response,data}=await runGo('x');assert.equal(response.ok,false);assert.match(data.error,/busy/);
 });
 for(const payload of [{Errors:42},{Events:{}},{Events:[{Message:42,Kind:'stdout'}]},null])await check('Reject malformed sandbox responses',async()=>{
  globalThis.fetch=async()=>Response.json(payload);await assert.rejects(runGo('x'),/Invalid Playground/);
 });
 await check('Propagate network failures to the editor',async()=>{
  globalThis.fetch=async()=>{throw new TypeError('network unavailable');};await assert.rejects(runGo('x'),/network unavailable/);
 });
}finally{globalThis.fetch=originalFetch;}
mkdirSync('work',{recursive:true});
writeFileSync('work/runner-verification.json',JSON.stringify({passed:checks.length,checks},null,2)+'\n');
console.log(`${checks.length} browser runner checks passed.`);
