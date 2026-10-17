import {editions} from '@/shared/editions';
import raw from './gopath.json';
export type CoursePage={id:string;kind:string;sourceFile:string;canonical:string;title:string;shortTitle:string;description:string;weight:number;moduleId:string;goVersion:string;body:string;minutes:number;headings?:string[];searchText:string;files?:string[]};
export const course=raw as unknown as {
 title:string;version:string;sourceTitle:string;sourceRepository:string;contentLicense:string;codeLicense:string;targetGo:string;verifiedToolchain:string;
 modules:{id:string;title:string;shortTitle:string;description:string;weight:number;intro:string;lessons:string[]}[];
 lessons:CoursePage[];caseStudies:CoursePage[];projects:CoursePage[];
 embeds:Record<string,{id:string;file:string;region:string;command:string;code:string;canRun:boolean;lessonId:string}>;
 callouts:Record<string,{type:string;body:string}>;
 disclosures:Record<string,{title:string;body:string}>;
 exercises:Record<string,{id:string;moduleId:string;lessonId:string;title:string;body:string;starter:string;solution:string;tests:string[]}>;
 sourceFiles:Record<string,{path:string;content:string}>;routes:Record<string,string>;
};
export const pages=[...course.lessons,...course.caseStudies,...course.projects];
export const originalHosted=editions.v1.url;
export function pageHref(page:CoursePage){return `#/${page.kind}/${page.id}`;}
export function slug(text:string){return text.toLowerCase().replace(/<[^>]*>/g,'').replace(/[^\p{L}\p{N} _-]/gu,'').trim().replace(/[ _]+/g,'-');}
export function resolveLink(href:string,page?:CoursePage):string {
 if(!href||/^(?:javascript|data|vbscript|file):/i.test(href))return '#';
 if(/^https?:\/\//i.test(href)||href.startsWith('mailto:'))return href;
 if(href.startsWith('#/'))return href;
 if(href.startsWith('#'))return href;
 try {
  let path:string,hash:string;
  if(page?.kind==='project'&&!href.startsWith('/')){
   const url=new URL(href,'https://content.local/'+page.sourceFile);path=url.pathname;hash=url.hash;
  }else{const url=new URL(href,'https://content.local/'+(page?.canonical?.replace(/^\//,'')||''));path=url.pathname;hash=url.hash;}
  path=decodeURIComponent(path).replace(/\/index\.html$/,'').replace(/\/$/,'').replace(/\.md$/,'');
  const direct=course.routes[path]||course.routes[path+'.md'];
  if(direct)return direct+(hash?`?section=${encodeURIComponent(hash.slice(1))}`:'');
  const sibling=page?.kind==='lesson'?'/learn/'+page.moduleId+'/'+href.replace(/^\.\//,'').replace(/\.md$/,'').replace(/\/$/,''):'';
  if(course.routes[sibling])return course.routes[sibling];
  return `#/link-missing?path=${encodeURIComponent(href)}`;
 }catch{return '#';}
}
export function downloadText(content:string,name:string){const url=URL.createObjectURL(new Blob([content],{type:'text/plain;charset=utf-8'}));const a=document.createElement('a');a.href=url;a.download=name;a.click();setTimeout(()=>URL.revokeObjectURL(url),1000);}
