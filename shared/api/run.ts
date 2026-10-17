export const dynamic = 'force-dynamic';
const maxBodyBytes = 35000;
const json = (body: unknown, status = 200) => Response.json(body, {status, headers:{'Cache-Control':'no-store'}});

async function readBody(request: Request): Promise<string | null> {
  if (!request.body) return '';
  const reader = request.body.getReader();
  const chunks: Uint8Array[] = [];
  let length = 0;
  try {
    while (true) {
      const chunk = await reader.read();
      if (chunk.done) break;
      length += chunk.value.byteLength;
      if (length > maxBodyBytes) { await reader.cancel(); return null; }
      chunks.push(chunk.value);
    }
  } finally { reader.releaseLock(); }
  const bytes = new Uint8Array(length);
  let offset = 0;
  for (const chunk of chunks) { bytes.set(chunk, offset); offset += chunk.byteLength; }
  return new TextDecoder('utf-8', {fatal:true}).decode(bytes);
}

export async function POST(request: Request) {
  if (Number(request.headers.get('content-length') || 0) > maxBodyBytes) return json({error:'Please keep examples under 30 KB.'}, 413);
  const origin = request.headers.get('origin');
  if (origin && origin !== new URL(request.url).origin) return json({error:'Cross-origin execution is not supported.'}, 403);
  let input: {code?: unknown};
  try {
    const body = await readBody(request);
    if (body === null) return json({error:'Please keep examples under 30 KB.'}, 413);
    input = JSON.parse(body);
  } catch { return json({error:'Send valid JSON with a code field.'}, 400); }
  if (!input || typeof input.code !== 'string' || !input.code.trim() || new TextEncoder().encode(input.code).length > 30000) {
    return json({error:'Enter a Go program between 1 byte and 30 KB.'}, 400);
  }
  try {
    const upstream = await fetch('https://go.dev/_/compile', {
      method:'POST', headers:{'Content-Type':'application/x-www-form-urlencoded'},
      body:new URLSearchParams({version:'2', body:input.code, withVet:'true'}),
      signal:AbortSignal.timeout(20000),
    });
    if (!upstream.ok) return json({error:'The Go sandbox is busy. Try again or download and run locally.'}, 502);
    const result = await upstream.json() as {Errors?:unknown; Events?:unknown; VetErrors?:unknown};
    if (!result || (result.Errors !== undefined && typeof result.Errors !== 'string') || (result.Events !== undefined && !Array.isArray(result.Events))) throw new Error('invalid sandbox response');
    const events = (result.Events || []) as {Message?:unknown;Kind?:unknown}[];
    if (events.some(event=>!event || typeof event.Message !== 'string' || typeof event.Kind !== 'string')) throw new Error('invalid sandbox event');
    return json({errors:result.Errors || '', events, vet:typeof result.VetErrors==='string'?result.VetErrors:''});
  } catch {
    return json({error:'The Go sandbox did not respond in time. Try again or run the downloaded file locally.'}, 503);
  }
}
