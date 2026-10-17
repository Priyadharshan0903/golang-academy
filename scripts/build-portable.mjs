// Build the default edition in an isolated directory, then copy its artifact.
import {cpSync,mkdirSync,writeFileSync} from 'node:fs';
import {join} from 'node:path';
import {buildEdition,root} from './editions.mjs';

const build=buildEdition('v2');
cpSync(build.dist,join(root,'dist'),{recursive:true});
mkdirSync(join(root,'work'),{recursive:true});
writeFileSync(join(root,'work/portable-build.json'),JSON.stringify({...build,completed:true},null,2)+'\n');
console.log('Completed production build; artifact copied to dist.');
