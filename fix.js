const fs=require('fs'),path=require('path');
const TARGET=String.fromCharCode(92)+String.fromCharCode(34);
const REPL=String.fromCharCode(34);
function fix(dir){
fs.readdirSync(dir).forEach(f=>{
const p=path.join(dir,f);
if(fs.statSync(p).isDirectory()){if(f!=='node_modules'&&f!=='.deepseek'&&f!=='dist'&&f!=='node_modules')fix(p)}
else if(/\.(go|json|ts|vue|css|html)$/.test(f)){
let c=fs.readFileSync(p,'utf8');
let c2=c.split(TARGET).join(REPL).replace(/\</g,'<').replace(/\>/g,','&').replace(/\\(/g,'(').replace(/\\)/g,')').replace(/\\//g,'/').replace(/\!/g,'!').replace(/:\\$/g,':');
if(c!==c2){fs.writeFileSync(p,c2);console.log('fixed:',p)}
}
})}
fix('.');console.log('done')
