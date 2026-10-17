from pathlib import Path
import json, urllib.request, urllib.error, re, hashlib
root=Path(__file__).resolve().parents[2]
reports=[]
for edition,port,name,legacy,other in [('v1',4183,'go-by-doing-v1-content.zip','go-by-doing-examples.zip','gopath-v2-content.zip'),('v2',4184,'gopath-v2-content.zip','gopath-course-v2.zip','go-by-doing-v1-content.zip')]:
 base=f'http://127.0.0.1:{port}'
 response=urllib.request.urlopen(base+'/',timeout=15)
 html=response.read().decode();assert response.status==200
 links=set(re.findall(r'href="(/downloads/[^"#]+)"',html))
 assert '/downloads/'+name in links,links
 for path in links:
  r=urllib.request.urlopen(base+path,timeout=15);assert r.status==200
  assert r.read()[:2]==b'PK'
 downloaded=urllib.request.urlopen(base+'/downloads/'+name,timeout=15).read()
 assert downloaded==(root/'apps'/edition/'public/downloads'/name).read_bytes()
 assert urllib.request.urlopen(base+'/downloads/'+legacy,timeout=15).read()==downloaded
 try:
  r=urllib.request.urlopen(base+'/downloads/'+other,timeout=15)
  assert r.status==404
 except urllib.error.HTTPError as e:assert e.code==404
 request=urllib.request.Request(base+'/api/run',data=b'{',headers={'Content-Type':'application/json'})
 try:urllib.request.urlopen(request,timeout=15);raise AssertionError('Runner accepted malformed JSON')
 except urllib.error.HTTPError as e:assert e.code==400
 reports.append({'edition':edition,'rootHTTP':200,'visibleDownloads':sorted(links),'archiveSha256':hashlib.sha256(downloaded).hexdigest(),'legacyArchiveHTTP':200,'otherEditionArchiveHTTP':404,'runnerMalformedJSONHTTP':400})
(root/'work/separation/http-verification.json').write_text(json.dumps(reports,indent=2)+'\n')
print(json.dumps(reports,indent=2))
