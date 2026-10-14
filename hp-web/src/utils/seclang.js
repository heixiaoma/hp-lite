import * as monaco from 'monaco-editor'
import 'monaco-editor/min/vs/editor/editor.main.css'

/* SecLang（ModSecurity / coraza 规则）的 Monaco 语言定义。
   抽成一份是因为「防护规则」和「Web 安全」两个页面编辑的是同一套语法，
   各写一份 tokenizer 的话，以后加个关键字就得改两遍。 */

export const SEC_LANG = 'seclang'
export const SEC_THEME = 'seclang-theme'

// 编辑器配置
export const seclangEditorOptions = {
  fontSize: 14,
  lineNumbers: 'on',
  lineWrapping: true,
  tabSize: 2,
  // 容器尺寸变化时自动重排，避免弹窗里初始化时算成 0 宽
  automaticLayout: true,
  minimap: {enabled: true},
  scrollBeyondLastLine: false,
  placeholder: '请输入 ModSecurity SecLang 规则（兼容 OWASP CRS v4）...',
  theme: SEC_THEME,
}

/* 注册语言 + tokenizer + 主题。
   Monaco 没有 languages.unregister —— 语言一旦注册就无法注销，
   所以要做幂等判断：注册过就跳过。弹窗 destroy-on-close 会让这里被反复执行，
   重复 register 同一 id 会堆积冗余语言项，但不跳过也不会崩。 */
export function setupSeclangMonaco() {
  const langRegistered = monaco.languages.getLanguages().some((l) => l.id === SEC_LANG)
  if (!langRegistered) {
    monaco.languages.register({id: SEC_LANG})
  }

  // 重构 tokenizer 规则（核心：避开 rx 解析陷阱）
  monaco.languages.setMonarchTokensProvider(SEC_LANG, {
    defaultToken: 'text',

    keywords: [
      'SecRule', 'SecAction', 'SecMarker', 'SecRequestBodyAccess',
      'SecResponseBodyAccess', 'SecRuleEngine', 'SecDebugLogLevel',
      'SecAuditEngine', 'SecAuditLogParts', 'SecAuditLog',
      'SecDataDir', 'SecTmpDir', 'SecPcreMatchLimit',
    ],

    actions: [
      'id', 'phase', 't', 'msg', 'log', 'nolog', 'pass', 'deny',
      'allow', 'status', 'setvar', 'expirevar', 'chain',
      'skip', 'skipAfter', 'capture', 'ctl', 'severity',
      'tag', 'ver', 'rev', 'accuracy', 'maturity',
    ],

    operators: [
      '@rx', '@pm', '@pmFromFile', '@eq', '@ne', '@gt', '@ge',
      '@lt', '@le', '@streq', '@contains', '@beginsWith',
      '@endsWith', '@detectSQLi', '@detectXSS', '@validateByteRange',
      '@validateUrlEncoding', '@validateUtf8Encoding',
      '@within',
    ],

    variables: [
      'ARGS', 'ARGS_NAMES', 'REQUEST_URI', 'REQUEST_METHOD',
      'REQUEST_HEADERS', 'REQUEST_HEADERS_NAMES',
      'REQUEST_BODY', 'REQUEST_COOKIES',
      'REQUEST_COOKIES_NAMES',
      'RESPONSE_BODY', 'RESPONSE_STATUS',
      'TX', 'IP', 'SESSION', 'GLOBAL',
      'FILES', 'FILES_TMPNAMES', 'FILES_NAMES',
      'MATCHED_VAR', 'MATCHED_VARS',
      'MATCHED_VAR_NAME', 'MATCHED_VARS_NAMES',
    ],

    tokenizer: {
      root: [
        // ---------------- 注释 ----------------
        [/#.*/, 'comment'],

        // ---------------- 核心指令 ----------------
        [/\b(SecRule|SecAction|SecMarker)\b/, 'keyword'],

        // ---------------- 变量集合 ----------------
        [/\b(ARGS|TX|IP|SESSION|GLOBAL|REQUEST_\w+|RESPONSE_\w+|FILES\w*)\b/, 'variable'],

        // 带子键的集合 ARGS:username
        [/\b(ARGS|TX|IP|SESSION|GLOBAL|REQUEST_\w+|FILES\w*):[A-Za-z0-9_\-]+/, 'variable'],

        // ---------------- 操作符 ----------------
        [/@[a-zA-Z]+/, 'operator'],

        // ---------------- Action key ----------------
        [/\b(id|phase|t|msg|log|nolog|pass|deny|allow|status|setvar|expirevar|chain|skip|skipAfter|capture|ctl|severity|tag|ver|rev|accuracy|maturity)\b(?=:)/, 'attribute'],

        // ---------------- 数字 ----------------
        [/\b\d+\b/, 'number'],

        // ---------------- 正则内容 ----------------
        [/\/.*?\//, 'regexp'],      // /regex/
        [/\^.*$/, 'regexp'],       // ^regex

        // ---------------- 字符串 ----------------
        [/"/, {token: 'string.quote', next: '@string_double'}],
        [/'/, {token: 'string.quote', next: '@string_single'}],

        // ---------------- 运算符 ----------------
        [/[!~<>]=?/, 'operator'],
        [/[=|&]/, 'operator'],

        // ---------------- 分隔符 ----------------
        [/[(),]/, 'delimiter'],
      ],

      string_double: [
        [/[^\\"]+/, 'string'],
        [/\\./, 'string.escape'],
        [/"/, {token: 'string.quote', next: '@pop'}],
      ],

      string_single: [
        [/[^\\']+/, 'string'],
        [/\\./, 'string.escape'],
        [/'/, {token: 'string.quote', next: '@pop'}],
      ],
    },

    comments: {
      lineComment: '#',
    },
  })

  // 注册主题（仅用基础 token 类型，无自定义属性）
  monaco.editor.defineTheme(SEC_THEME, {
    base: 'vs-dark',
    inherit: true,
    rules: [
      {token: 'comment', foreground: '808080', fontStyle: 'italic'},
      {token: 'keyword', foreground: '569CD6', fontStyle: 'bold'},
      {token: 'attribute', foreground: '9CDCFE'},
      {token: 'type', foreground: 'DCDCAA'},
      {token: 'regexp', foreground: 'B5CEA8'},
      {token: 'string', foreground: 'CE9178'},
      {token: 'variable', foreground: '9CDCFE'},
    ],
    colors: {
      'editor.background': '#1E1E1E',
      'editor.lineHighlightBackground': '#2A2A2A',
      'editor.foreground': '#D4D4D4',
    },
  })
  // 强制应用主题和语言
  monaco.editor.setTheme(SEC_THEME)
}

/* 编辑器挂载后的兜底：确保语言生效。
   直接设置模型语言，跳过解析器的规则校验 */
export function applySeclangModel(editor) {
  const model = editor.getModel()
  if (model) {
    monaco.editor.setModelLanguage(model, SEC_LANG)
  }
  monaco.editor.setTheme(SEC_THEME)
}
