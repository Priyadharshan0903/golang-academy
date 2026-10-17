// Compatibility command: learner packs now have one exporter per edition.
import {writeFileSync} from 'node:fs';
import {join} from 'node:path';
import {allLessons} from '../content/v1/curriculum.ts';
import {exportEdition} from './downloads/export.mjs';
const report=await exportEdition('v1');
writeFileSync('work/example-manifest.json',JSON.stringify(allLessons.map(l=>({id:l.id,path:join(report.extractedDirectory,'examples',l.moduleId,l.id,'main.go'),output:l.output})),null,2)+'\n');
