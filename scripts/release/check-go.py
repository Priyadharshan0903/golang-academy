from pathlib import Path
import concurrent.futures, json, os, subprocess, time

root = Path(__file__).resolve().parents[2]
packs = {r['edition']: Path(r['extractedDirectory']) for r in json.loads((root/'work/separation/archive-verification.json').read_text())}
cache = root/'work/go-cache'
cache.mkdir(parents=True, exist_ok=True)
environment = dict(os.environ, GOCACHE=str(cache), GOTOOLCHAIN='local', GOWORK='off', GOPROXY='off', GO111MODULE='on')

def command(label, cwd, args, expected=None, intentionally_failing=False):
    started = time.monotonic()
    try:
        result = subprocess.run(args, cwd=cwd, env=environment, capture_output=True, text=True, timeout=180)
        passed = result.returncode == 0 and (expected is None or result.stdout == expected)
        if intentionally_failing:
            passed = result.returncode != 0 and 'FAIL' in result.stdout and 'build failed' not in result.stderr
        return dict(check=label, passed=passed, intentionallyFailing=intentionally_failing, exitCode=result.returncode,
                    seconds=round(time.monotonic()-started, 2), stdout=result.stdout[-6000:], stderr=result.stderr[-3000:])
    except Exception as error:
        return dict(check=label, passed=False, error=str(error))

examples = sorted((packs['v1']/'examples').rglob('main.go'))
with concurrent.futures.ThreadPoolExecutor(max_workers=4) as executor:
    results = list(executor.map(lambda p: command('v1 example '+p.parent.name, p.parent, ['go','run','main.go'], expected=(p.parent/'expected.txt').read_text()), examples))
for label, directory, args, failing in [
    ('v1 website-checker solution', packs['v1']/'projects/website-checker/solution', ['go','test','-race','./...'], False),
    ('v1 website-checker starter behavior', packs['v1']/'projects/website-checker/starter', ['go','test','./...'], True),
    ('v2 root-module examples/projects/server', packs['v2'], ['go','test','./examples/...','./projects/...','./deploy/server'], False),
    ('v2 exercise reference solutions', packs['v2']/'exercises', ['go','test','-tags','solution','./...'], False),
]:
    results.append(command(label,directory,args,intentionally_failing=failing))
report = {'toolchain': subprocess.check_output(['go','version'],text=True).strip(), 'checks':results,
          'limits':['Go 1.27-only build-tagged code not claimed as verified with Go 1.26.5.',
                    'Nested third-party, PostgreSQL/Redis and Kubernetes modules require external dependencies/services and are not included.']}
(root/'work/separation/go-verification.json').write_text(json.dumps(report,indent=2)+'\n')
print(json.dumps({'checks':len(results),'passed':sum(r['passed'] for r in results),'failures':[r for r in results if not r['passed']],'limits':report['limits']},indent=2))
raise SystemExit(0 if all(r['passed'] for r in results) else 1)
