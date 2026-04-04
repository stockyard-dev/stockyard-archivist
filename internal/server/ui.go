package server

import "net/http"

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(dashHTML))
}

const dashHTML = `<!DOCTYPE html><html><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"><title>Archivist</title>
<link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;700&display=swap" rel="stylesheet">
<style>
:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#e8753a;--leather:#a0845c;--cream:#f0e6d3;--cd:#bfb5a3;--cm:#7a7060;--gold:#d4a843;--green:#4a9e5c;--red:#c94444;--blue:#5b8dd9;--mono:'JetBrains Mono',monospace}
*{margin:0;padding:0;box-sizing:border-box}body{background:var(--bg);color:var(--cream);font-family:var(--mono);line-height:1.5}
.hdr{padding:1rem 1.5rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between;align-items:center}.hdr h1{font-size:.9rem;letter-spacing:2px}.hdr h1 span{color:var(--rust)}
.main{padding:1.5rem;max-width:960px;margin:0 auto}
.stats{display:grid;grid-template-columns:repeat(3,1fr);gap:.5rem;margin-bottom:1rem}
.st{background:var(--bg2);border:1px solid var(--bg3);padding:.6rem;text-align:center}
.st-v{font-size:1.2rem;font-weight:700}.st-l{font-size:.5rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px;margin-top:.15rem}
.toolbar{display:flex;gap:.5rem;margin-bottom:1rem;align-items:center;flex-wrap:wrap}
.search{flex:1;min-width:180px;padding:.4rem .6rem;background:var(--bg2);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.7rem}
.search:focus{outline:none;border-color:var(--leather)}
.filter-sel{padding:.4rem .5rem;background:var(--bg2);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.65rem}
.doc{background:var(--bg2);border:1px solid var(--bg3);padding:.8rem 1rem;margin-bottom:.5rem;transition:border-color .2s}
.doc:hover{border-color:var(--leather)}
.doc-top{display:flex;justify-content:space-between;align-items:flex-start;gap:.5rem}
.doc-title{font-size:.85rem;font-weight:700}
.doc-file{font-size:.6rem;color:var(--cd);margin-top:.1rem}
.doc-meta{font-size:.55rem;color:var(--cm);margin-top:.3rem;display:flex;gap:.5rem;flex-wrap:wrap;align-items:center}
.doc-notes{font-size:.65rem;color:var(--cm);margin-top:.3rem;font-style:italic;padding:.3rem .5rem;border-left:2px solid var(--bg3)}
.doc-actions{display:flex;gap:.3rem;flex-shrink:0}
.tag{font-size:.45rem;padding:.1rem .25rem;background:var(--bg3);color:var(--cd)}
.folder-badge{font-size:.5rem;padding:.1rem .3rem;background:var(--bg3);color:var(--gold)}
.mime-badge{font-size:.45rem;padding:.1rem .3rem;border:1px solid var(--bg3);color:var(--cm)}
.btn{font-size:.6rem;padding:.25rem .5rem;cursor:pointer;border:1px solid var(--bg3);background:var(--bg);color:var(--cd);transition:all .2s}
.btn:hover{border-color:var(--leather);color:var(--cream)}.btn-p{background:var(--rust);border-color:var(--rust);color:#fff}
.btn-sm{font-size:.55rem;padding:.2rem .4rem}
.modal-bg{display:none;position:fixed;inset:0;background:rgba(0,0,0,.65);z-index:100;align-items:center;justify-content:center}.modal-bg.open{display:flex}
.modal{background:var(--bg2);border:1px solid var(--bg3);padding:1.5rem;width:460px;max-width:92vw}
.modal h2{font-size:.8rem;margin-bottom:1rem;color:var(--rust);letter-spacing:1px}
.fr{margin-bottom:.6rem}.fr label{display:block;font-size:.55rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px;margin-bottom:.2rem}
.fr input,.fr select{width:100%;padding:.4rem .5rem;background:var(--bg);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.7rem}
.fr input:focus,.fr select:focus{outline:none;border-color:var(--leather)}
.row2{display:grid;grid-template-columns:1fr 1fr;gap:.5rem}
.acts{display:flex;gap:.4rem;justify-content:flex-end;margin-top:1rem}
.empty{text-align:center;padding:3rem;color:var(--cm);font-style:italic;font-size:.75rem}
</style></head><body>
<div class="hdr"><h1><span>&#9670;</span> ARCHIVIST</h1><button class="btn btn-p" onclick="openForm()">+ Add Document</button></div>
<div class="main">
<div class="stats" id="stats"></div>
<div class="toolbar">
<input class="search" id="search" placeholder="Search documents..." oninput="render()">
<select class="filter-sel" id="folder-filter" onchange="render()"><option value="">All Folders</option></select>
</div>
<div id="docs"></div>
</div>
<div class="modal-bg" id="mbg" onclick="if(event.target===this)closeModal()"><div class="modal" id="mdl"></div></div>
<script>
var A='/api',docs=[],editId=null;

async function load(){var r=await fetch(A+'/documents').then(function(r){return r.json()});docs=r.documents||[];renderStats();buildFolderFilter();render();}

function fmtSize(b){if(!b)return'0 B';if(b<1024)return b+' B';if(b<1048576)return(b/1024).toFixed(1)+' KB';return(b/1048576).toFixed(1)+' MB';}

function renderStats(){
var total=docs.length;
var totalSize=docs.reduce(function(s,d){return s+(d.size_bytes||0)},0);
var folders={};docs.forEach(function(d){if(d.folder)folders[d.folder]=true});
document.getElementById('stats').innerHTML=[
{l:'Documents',v:total},{l:'Total Size',v:fmtSize(totalSize)},{l:'Folders',v:Object.keys(folders).length}
].map(function(x){return '<div class="st"><div class="st-v">'+x.v+'</div><div class="st-l">'+x.l+'</div></div>'}).join('');
}

function buildFolderFilter(){
var folders={};docs.forEach(function(d){if(d.folder)folders[d.folder]=true});
var sel=document.getElementById('folder-filter');var cur=sel.value;
sel.innerHTML='<option value="">All Folders</option>';
Object.keys(folders).sort().forEach(function(f){sel.innerHTML+='<option value="'+esc(f)+'"'+(cur===f?' selected':'')+'>'+esc(f)+'</option>';});
}

function render(){
var q=(document.getElementById('search').value||'').toLowerCase();
var ff=document.getElementById('folder-filter').value;
var f=docs;
if(ff)f=f.filter(function(d){return d.folder===ff});
if(q)f=f.filter(function(d){return(d.title||'').toLowerCase().includes(q)||(d.filename||'').toLowerCase().includes(q)||(d.tags||'').toLowerCase().includes(q)});
if(!f.length){document.getElementById('docs').innerHTML='<div class="empty">No documents found.</div>';return;}
var h='';f.forEach(function(d){
h+='<div class="doc"><div class="doc-top"><div style="flex:1">';
h+='<div class="doc-title">'+esc(d.title||d.filename)+'</div>';
if(d.filename&&d.title)h+='<div class="doc-file">'+esc(d.filename)+'</div>';
h+='</div><div class="doc-actions">';
h+='<button class="btn btn-sm" onclick="openEdit(''+d.id+'')">Edit</button>';
h+='<button class="btn btn-sm" onclick="del(''+d.id+'')" style="color:var(--red)">&#10005;</button>';
h+='</div></div>';
h+='<div class="doc-meta">';
if(d.folder)h+='<span class="folder-badge">&#128193; '+esc(d.folder)+'</span>';
if(d.mime_type)h+='<span class="mime-badge">'+esc(d.mime_type)+'</span>';
h+='<span>'+fmtSize(d.size_bytes)+'</span>';
if(d.tags){d.tags.split(',').forEach(function(t){t=t.trim();if(t)h+='<span class="tag">#'+esc(t)+'</span>';});}
h+='<span>'+ft(d.created_at)+'</span>';
h+='</div>';
if(d.notes)h+='<div class="doc-notes">'+esc(d.notes)+'</div>';
h+='</div>';
});
document.getElementById('docs').innerHTML=h;
}

async function del(id){if(!confirm('Delete?'))return;await fetch(A+'/documents/'+id,{method:'DELETE'});load();}

function formHTML(doc){
var i=doc||{title:'',filename:'',mime_type:'',size_bytes:0,folder:'',tags:'',notes:''};
var isEdit=!!doc;
var h='<h2>'+(isEdit?'EDIT DOCUMENT':'ADD DOCUMENT')+'</h2>';
h+='<div class="fr"><label>Title *</label><input id="f-title" value="'+esc(i.title)+'"></div>';
h+='<div class="row2"><div class="fr"><label>Filename</label><input id="f-file" value="'+esc(i.filename)+'"></div>';
h+='<div class="fr"><label>MIME Type</label><input id="f-mime" value="'+esc(i.mime_type)+'" placeholder="application/pdf"></div></div>';
h+='<div class="row2"><div class="fr"><label>Folder</label><input id="f-folder" value="'+esc(i.folder)+'" placeholder="e.g. contracts"></div>';
h+='<div class="fr"><label>Size (bytes)</label><input id="f-size" type="number" value="'+(i.size_bytes||0)+'"></div></div>';
h+='<div class="fr"><label>Tags</label><input id="f-tags" value="'+esc(i.tags)+'" placeholder="comma separated"></div>';
h+='<div class="fr"><label>Notes</label><input id="f-notes" value="'+esc(i.notes)+'"></div>';
h+='<div class="acts"><button class="btn" onclick="closeModal()">Cancel</button><button class="btn btn-p" onclick="submit()">'+(isEdit?'Save':'Add')+'</button></div>';
return h;
}

function openForm(){editId=null;document.getElementById('mdl').innerHTML=formHTML();document.getElementById('mbg').classList.add('open');document.getElementById('f-title').focus();}
function openEdit(id){var d=null;for(var j=0;j<docs.length;j++){if(docs[j].id===id){d=docs[j];break;}}if(!d)return;editId=id;document.getElementById('mdl').innerHTML=formHTML(d);document.getElementById('mbg').classList.add('open');}
function closeModal(){document.getElementById('mbg').classList.remove('open');editId=null;}

async function submit(){
var title=document.getElementById('f-title').value.trim();
if(!title){alert('Title is required');return;}
var body={title:title,filename:document.getElementById('f-file').value.trim(),mime_type:document.getElementById('f-mime').value.trim(),size_bytes:parseInt(document.getElementById('f-size').value)||0,folder:document.getElementById('f-folder').value.trim(),tags:document.getElementById('f-tags').value.trim(),notes:document.getElementById('f-notes').value.trim()};
if(editId){await fetch(A+'/documents/'+editId,{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});}
else{await fetch(A+'/documents',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});}
closeModal();load();
}

function ft(t){if(!t)return'';try{return new Date(t).toLocaleDateString('en-US',{month:'short',day:'numeric',year:'numeric'})}catch(e){return t;}}
function esc(s){if(!s)return'';var d=document.createElement('div');d.textContent=s;return d.innerHTML;}
document.addEventListener('keydown',function(e){if(e.key==='Escape')closeModal();});
load();
</script></body></html>`
