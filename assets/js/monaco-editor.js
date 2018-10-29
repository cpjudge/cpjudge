import * as monaco from 'monaco-editor';

monaco.editor.create(document.getElementById('monaco-editor'), {
  value: [
    'function x() {',
    '\tconsole.log("Hello world!");',
    '}'
  ].join('\n'),
  language: 'javascript'
});
