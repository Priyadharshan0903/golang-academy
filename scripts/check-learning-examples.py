import concurrent.futures,json,os,pathlib,subprocess,time
root=pathlib.Path.cwd()
examples=json.loads((root/'work/example-manifest.json').read_text())
cache=root/'work/go-cache';cache.mkdir(parents=True,exist_ok=True)
env=dict(os.environ,GOCACHE=str(cache),GOTOOLCHAIN='local',GOWORK='off',GOPROXY='off',GO111MODULE='on')
def check(item):
    start=time.monotonic()
    try:
        result=subprocess.run(['go','run',item['path']],capture_output=True,text=True,env=env,cwd=str(pathlib.Path(item['path']).parent),timeout=120)
        passed=result.returncode==0 and result.stdout==item['output']
        return dict(id=item['id'],passed=passed,seconds=round(time.monotonic()-start,2),error=result.stderr if result.returncode else '' if passed else f"expected {item['output']!r}; got {result.stdout!r}")
    except Exception as exc:return dict(id=item['id'],passed=False,error=str(exc))
with concurrent.futures.ThreadPoolExecutor(max_workers=4) as pool:results=list(pool.map(check,examples))
report=dict(toolchain=subprocess.check_output(['go','version'],text=True).strip(),total=len(results),passed=sum(x['passed'] for x in results),results=results)
(root/'work/example-verification.json').write_text(json.dumps(report,indent=2))
for item in results:
    if not item['passed']:print(item['id'],item['error'])
print(f"Go examples: {report['passed']}/{report['total']} passed with exact expected output")
raise SystemExit(report['passed']!=report['total'])
