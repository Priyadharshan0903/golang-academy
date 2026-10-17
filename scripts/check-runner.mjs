import assert from 'node:assert/strict';
import {writeFileSync} from 'node:fs';
import {POST} from '../shared/api/run.ts';
const originalFetch=globalThis.fetch;
const checks=[];
const request=(body,headers={})=>new Request('https://learn.example/api/run',{method:'POST',headers:{'Content-Type':'application/json',...headers},body:typeof body==='string'?body:JSON.stringify(body)});
async function check(name,fn){await fn();checks.push(name);console.log('PASS',name);}
try {
 await check('Reject invalid JSON',async()=>assert.equal((await POST(request('{'))).status,400));
 for(const input of [null,{}, {code:''},{code:123},{code:' '.repeat(10)},{code:'界'.repeat(10001)}])await check(`Reject invalid code ${checks.length}`,async()=>assert.equal((await POST(request(input))).status,400));
 await check('Reject oversized Content-Length',async()=>assert.equal((await POST(request({code:'x'},{'Content-Length':'35001'}))).status,413));
 await check('Bound streamed body without trusting Content-Length',async()=>assert.equal((await POST(request('x'.repeat(35001)))).status,413));
 await check('Reject cross-origin request',async()=>assert.equal((await POST(request({code:'x'},{Origin:'https://elsewhere.example'}))).status,403));
 await check('Forward source only to the fixed official sandbox',async()=>{
  globalThis.fetch=async(url,options)=>{assert.equal(url,'https://go.dev/_/compile');assert.equal(options.body.get('body'),'package main\nfunc main(){}');assert.equal(options.body.get('version'),'2');return Response.json({Errors:'',Events:[{Message:'ok\n',Kind:'stdout'}],VetErrors:''});};
  const response=await POST(request({code:'package main\nfunc main(){}'}));assert.equal(response.status,200);assert.equal((await response.json()).events[0].Message,'ok\n');assert.equal(response.headers.get('Cache-Control'),'no-store');
 });
 await check('Preserve compiler errors and runtime stderr',async()=>{globalThis.fetch=async()=>Response.json({Errors:'syntax problem',Events:[{Message:'panic: sample',Kind:'stderr'}]});const data=await(await POST(request({code:'x'}))).json();assert.equal(data.errors,'syntax problem');assert.equal(data.events[0].Kind,'stderr');});
 await check('Handle an upstream busy response',async()=>{globalThis.fetch=async()=>new Response('busy',{status:429});const response=await POST(request({code:'x'}));assert.equal(response.status,502);assert.match((await response.json()).error,/busy/);});
 await check('Handle a network failure',async()=>{globalThis.fetch=async()=>{throw new TypeError('network unavailable');};assert.equal((await POST(request({code:'x'}))).status,503);});
 await check('Handle a timeout',async()=>{globalThis.fetch=async()=>{throw new DOMException('timed out','TimeoutError');};assert.equal((await POST(request({code:'x'}))).status,503);});
 await check('Reject malformed sandbox events',async()=>{globalThis.fetch=async()=>Response.json({Events:[{Message:123,Kind:'stdout'}]});assert.equal((await POST(request({code:'x'}))).status,503);});
}finally{globalThis.fetch=originalFetch;}
writeFileSync('work/runner-verification.json',JSON.stringify({passed:checks.length,checks},null,2));
console.log(`${checks.length} runner checks passed.`);
