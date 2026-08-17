window.MathJax = {
  tex: { inlineMath: [['\\(', '\\)']], displayMath: [['\\[', '\\]']] },
  chtml: {
    displayAlign: 'left',
    displayIndent: '0',
    displayOverflow: 'linebreak',
    linebreaks: { inline: true, width: '100%' }
  },
  options: { skipHtmlTags: ['script', 'noscript', 'style', 'textarea'] },
  startup: {
    pageReady: function () {
      return MathJax.startup.defaultPageReady().then(function () {
        document.documentElement.classList.add('math-ready');
      });
    }
  }
};

(function () {
  var s = document.createElement('script');
  s.async = true;
  s.src = 'https://cdn.jsdelivr.net/npm/mathjax@3/es5/tex-mml-chtml.js';
  document.head.appendChild(s);

  window.addEventListener('load', function () {
    setTimeout(function () {
      if (document.documentElement.classList.contains('math-ready')) return;
      var warn = document.getElementById('mathwarn');
      if (warn) warn.hidden = false;
    }, 4000);
  });
})();
