export const repositoryUrl='https://github.com/Priyadharshan0903/golang-academy';
const snapshotRoot=`${repositoryUrl}/blob/master/content-source/gopath-academy/`;

export function sourceFileUrl(path:string){
 return snapshotRoot+path.split('/').map(encodeURIComponent).join('/');
}

// Keep imported material unchanged while pointing its GitHub links at the
// preserved snapshot in this repository.
export function repositoryLink(href:string){
 const original='https://github.com/Priyadharshan0903/gopath-academy';
 if(href===original||href===original+'/')return repositoryUrl;
 for(const kind of ['blob','tree']){
  const prefix=`${original}/${kind}/main/`;
  if(href.startsWith(prefix))return `${repositoryUrl}/${kind}/master/content-source/gopath-academy/${href.slice(prefix.length)}`;
 }
 return href;
}
