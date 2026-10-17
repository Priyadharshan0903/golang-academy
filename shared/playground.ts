type PlaygroundEvent = {Kind:string;Message:string};
type PlaygroundOutput = {error?:string;errors:string;events:PlaygroundEvent[];vet:string};

// The official Playground allows cross-origin form requests, so static hosts
// can run examples directly without an API server or account credentials.
export async function runGo(code:string):Promise<{response:Response;data:PlaygroundOutput}> {
 if(!code.trim()||new TextEncoder().encode(code).length>30000)throw new Error('Enter a Go program between 1 byte and 30 KB.');
 const response=await fetch('https://play.golang.org/compile',{
  method:'POST',credentials:'omit',
  headers:{'Content-Type':'application/x-www-form-urlencoded'},
  body:new URLSearchParams({version:'2',body:code,withVet:'true'}),
  signal:AbortSignal.timeout(25000),
 });
 if(!response.ok)return {response,data:{error:'The Go Playground is busy. Try again or download and run locally.',errors:'',events:[],vet:''}};
 const result=await response.json() as {Errors?:unknown;Events?:unknown;VetErrors?:unknown};
 if(!result||(result.Errors!==undefined&&typeof result.Errors!=='string')||(result.Events!==undefined&&!Array.isArray(result.Events)))throw new Error('Invalid Playground response.');
 const events=(result.Events||[]) as PlaygroundEvent[];
 if(events.some(event=>!event||typeof event.Kind!=='string'||typeof event.Message!=='string'))throw new Error('Invalid Playground events.');
 return {response,data:{errors:typeof result.Errors==='string'?result.Errors:'',events,vet:typeof result.VetErrors==='string'?result.VetErrors:''}};
}
