  (function () {
    var btns = document.querySelectorAll('.tab-btn');
    var panels = document.querySelectorAll('.panel');

    btns.forEach(function (btn) {
      btn.addEventListener('click', function () {
        var target = btn.getAttribute('data-panel');

        btns.forEach(function (b) {
          b.classList.remove('active');
          b.setAttribute('aria-selected', 'false');
        });
        panels.forEach(function (p) { p.classList.remove('active'); });

        btn.classList.add('active');
        btn.setAttribute('aria-selected', 'true');
        document.getElementById('panel-' + target).classList.add('active');
      });
    });
  })();

  function chkKey(btn){if(btn.dataset.chk)return btn.dataset.chk; const p=btn.closest('.panel'); const pid=p?p.id:'nopanel'; const r=btn.closest('tr'); const n=r&&r.querySelector('.enum')?r.querySelector('.enum').textContent.trim():''; return pid+'#'+n;}
  function saveChk(){const s={}; document.querySelectorAll('.chk-btn').forEach(b=>{if(b.classList.contains('checked'))s[chkKey(b)]=1;}); try{localStorage.setItem('chk-go-clock',JSON.stringify(s));}catch(e){}}
  function restoreChk(){let s={}; try{s=JSON.parse(localStorage.getItem('chk-go-clock')||'{}');}catch(e){} document.querySelectorAll('.chk-btn').forEach(b=>{const on=!!s[chkKey(b)]; b.classList.toggle('checked',on); b.textContent=on?'☑':'☐'; b.setAttribute('aria-pressed',on?'true':'false');});}

  document.addEventListener('click', function (e) {
    var btn = e.target;
    if (!btn || !btn.classList || !btn.classList.contains('chk-btn')) return;
    var checked = btn.classList.toggle('checked');
    btn.textContent = checked ? '☑' : '☐';
    btn.setAttribute('aria-pressed', checked ? 'true' : 'false');
    saveChk();
  });

  window.addEventListener('DOMContentLoaded', restoreChk);
