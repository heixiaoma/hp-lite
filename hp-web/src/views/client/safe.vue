<template>
  <div class="hp-page">
    <div class="hp-toolbar">
      <t-button theme="primary" @click="addModal">
        <template #icon><add-icon/></template>
        添加规则
      </t-button>
      <t-button variant="outline" theme="primary" @click="loadData">
        <template #icon><refresh-icon/></template>
        刷新列表
      </t-button>
    </div>

    <t-table
        class="hp-table"
        row-key="id"
        :data="listData || []"
        :columns="viewColumns"
        :loading="dataLoading"
        :pagination="pagination"
        empty="暂无数据，添加一个试试看看"
        table-layout="auto"
        stripe
        hover
        @page-change="onPageChange"
    >
      <template #rule="{ row }">
        <t-popup placement="top-left" :overlay-inner-style="{ maxWidth: '640px', maxHeight: '320px', overflow: 'auto' }">
          <span class="rule-ellipsis">{{ row.rule }}</span>
          <template #content>
            <pre class="rule-pop">{{ row.rule }}</pre>
          </template>
        </t-popup>
      </template>

      <template #user="{ row }">
        <t-tag v-if="!row.userDesc && !row.username" variant="light" theme="primary">自用</t-tag>
        <div v-else>
          <div>归属用户：{{ row.username }}</div>
          <div class="hp-sub-line">归属用户备注：{{ row.userDesc }}</div>
        </div>
      </template>

      <template #action="{ row }">
        <div class="hp-actions">
          <t-button size="small" variant="outline" theme="primary" @click="edit(row)">编辑</t-button>
          <t-popconfirm content="确定要删除该规则？" theme="danger" @confirm="removeData(row)">
            <t-button size="small" variant="outline" theme="danger">删除</t-button>
          </t-popconfirm>
        </div>
      </template>
    </t-table>

    <t-dialog
        v-model:visible="addVisible"
        :header="formState.id ? '编辑规则' : '添加规则'"
        confirm-btn="确定"
        cancel-btn="取消"
        width="80%"
        @confirm="addOk"
    >
      <div class="hp-dialog-body">
        <!-- label-align="top" 才是「标签在上、控件在下」。
             layout="vertical" 只负责表单项之间上下堆叠，
             label 默认仍贴左侧占 100px，会把编辑器挤窄 -->
        <t-form
            :data="formState"
            ref="formTable"
            :rules="formRules"
            layout="vertical"
            label-align="top"
        >
          <t-form-item label="规则名字" name="ruleName">
            <t-input v-model="formState.ruleName" clearable placeholder="规则名字"/>
          </t-form-item>

          <t-form-item label="规则" name="rule">
            <div class="monaco-container">
              <MonacoEditor
                  v-model:value="formState.rule"
                  language="seclang"
                  :options="editorOptions"
                  @mounted="handleEditorMounted"
              />
            </div>
          </t-form-item>
        </t-form>
      </div>
    </t-dialog>
  </div>
</template>

<script setup>
import {getSafe, removeSafe, saveSafe} from "../../api/client/safe";
import {nextTick, onMounted, reactive, ref, watch} from "vue";
import {MessagePlugin} from "tdesign-vue-next";
import {AddIcon, RefreshIcon} from 'tdesign-icons-vue-next';
import {useResponsiveColumns} from '../../utils/responsive';

import MonacoEditor from 'monaco-editor-vue3'
import * as monaco from 'monaco-editor'
import 'monaco-editor/min/vs/editor/editor.main.css'

// 编辑器配置
const editorOptions = {
  fontSize: 14,
  lineNumbers: 'on',
  lineWrapping: true,
  tabSize: 2,
  // 容器尺寸变化时自动重排，避免弹窗里初始化时算成 0 宽
  automaticLayout: true,
  minimap: {enabled: true},
  scrollBeyondLastLine: false,
  placeholder: '请输入 ModSecurity SecLang 规则（兼容 OWASP CRS v4）...',
  theme: 'seclang-theme'
}

onMounted(async () => {
  try {
    // 1. 注册 seclang 语言（极简配置，避免解析器校验）
    // Monaco 没有 languages.unregister —— 语言一旦注册就无法注销，
    // 所以要做幂等判断：注册过就跳过。弹窗 destroy-on-close 会让这里被反复执行，
    // 重复 register 同一 id 会堆积冗余语言项，但不跳过也不会崩。
    const langRegistered = monaco.languages.getLanguages().some(l => l.id === 'seclang')
    if (!langRegistered) {
      monaco.languages.register({id: 'seclang'})
    }

    // 2. 重构 tokenizer 规则（核心：避开 rx 解析陷阱）
    monaco.languages.setMonarchTokensProvider('seclang', {
      defaultToken: 'text',

      keywords: [
        'SecRule', 'SecAction', 'SecMarker', 'SecRequestBodyAccess',
        'SecResponseBodyAccess', 'SecRuleEngine', 'SecDebugLogLevel',
        'SecAuditEngine', 'SecAuditLogParts', 'SecAuditLog',
        'SecDataDir', 'SecTmpDir', 'SecPcreMatchLimit'
      ],

      actions: [
        'id', 'phase', 't', 'msg', 'log', 'nolog', 'pass', 'deny',
        'allow', 'status', 'setvar', 'expirevar', 'chain',
        'skip', 'skipAfter', 'capture', 'ctl', 'severity',
        'tag', 'ver', 'rev', 'accuracy', 'maturity'
      ],

      operators: [
        '@rx', '@pm', '@pmFromFile', '@eq', '@ne', '@gt', '@ge',
        '@lt', '@le', '@streq', '@contains', '@beginsWith',
        '@endsWith', '@detectSQLi', '@detectXSS', '@validateByteRange',
        '@validateUrlEncoding', '@validateUtf8Encoding',
        '@within'
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
        'MATCHED_VAR_NAME', 'MATCHED_VARS_NAMES'
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
          [/[(),]/, 'delimiter']
        ],

        string_double: [
          [/[^\\"]+/, 'string'],
          [/\\./, 'string.escape'],
          [/"/, {token: 'string.quote', next: '@pop'}]
        ],

        string_single: [
          [/[^\\']+/, 'string'],
          [/\\./, 'string.escape'],
          [/'/, {token: 'string.quote', next: '@pop'}]
        ]
      },

      comments: {
        lineComment: '#'
      }
    })

    // 3. 注册主题（仅用基础 token 类型，无自定义属性）
    monaco.editor.defineTheme('seclang-theme', {
      base: 'vs-dark',
      inherit: true,
      rules: [
        {token: 'comment', foreground: '808080', fontStyle: 'italic'},
        {token: 'keyword', foreground: '569CD6', fontStyle: 'bold'},
        {token: 'attribute', foreground: '9CDCFE'},
        {token: 'type', foreground: 'DCDCAA'},
        {token: 'regexp', foreground: 'B5CEA8'},
        {token: 'string', foreground: 'CE9178'},
        {token: 'variable', foreground: '9CDCFE'}
      ],
      colors: {
        'editor.background': '#1E1E1E',
        'editor.lineHighlightBackground': '#2A2A2A',
        'editor.foreground': '#D4D4D4'
      }
    })
    // 4. 强制应用主题和语言
    monaco.editor.setTheme('seclang-theme')
  } catch (e) {
    console.error('Monaco 初始化失败：', e)
  }
})

let editorInstance = null

// 编辑器挂载后兜底（确保语言生效）
const handleEditorMounted = (editor) => {
  editorInstance = editor
  // 直接设置模型语言，跳过解析器的规则校验
  const model = editor.getModel()
  if (model) {
    monaco.editor.setModelLanguage(model, 'seclang')
  }
  monaco.editor.setTheme('seclang-theme')
}

const listData = ref([]);
const formTable = ref();
const dataLoading = ref(false);
const addVisible = ref(false);

// 弹窗每次打开都让编辑器按容器的真实尺寸重排一次：
// Monaco 初始化时若容器还没完成布局，宽度会被算成 0 而显示不出来
watch(addVisible, (visible) => {
  if (!visible) return
  nextTick(() => editorInstance?.layout())
})

const formState = reactive({
  ruleName: "",
  rule: "",
  id: undefined
})

const formRules = {
  ruleName: [{required: true, message: '规则名字必填', type: 'error'}],
  rule: [{required: true, message: '规则必填', type: 'error'}],
};

const pagination = reactive({
  total: 0,
  current: 1,
  pageSize: 10,
  pageSizeOptions: [10, 20, 50],
});

const loadData = () => {
  dataLoading.value = true
  getSafe({
    current: pagination.current,
    pageSize: pagination.pageSize
  }).then(res => {
    dataLoading.value = false
    listData.value = res.data.records
    pagination.total = res.data.total
  }).catch(() => {
    dataLoading.value = false
  })
}

const removeData = (item) => {
  removeSafe({
    id: item.id
  }).then(res => {
    MessagePlugin.success(res.msg)
    loadData()
  })
}

const edit = (itemOld) => {
  const item = JSON.parse(JSON.stringify(itemOld))
  formState.rule = item.rule
  formState.ruleName = item.ruleName
  formState.id = item.id
  addVisible.value = true
}

/* 窄屏留规则名 + 内容 + 操作：编号和归属让位 */
const columns = [
  {colKey: 'id', title: '编号', width: 90, mobile: false},
  {colKey: 'ruleName', title: '规则名字', width: 200},
  {colKey: 'rule', title: '规则内容', ellipsis: true},
  {colKey: 'user', title: '归属', width: 220, mobile: false},
  {colKey: 'action', title: '操作', width: 150},
];
const viewColumns = useResponsiveColumns(columns);

const onPageChange = (pageInfo) => {
  pagination.current = pageInfo.current
  pagination.pageSize = pageInfo.pageSize
  loadData()
}

const addModal = () => {
  formState.rule = ""
  formState.ruleName = ""
  formState.id = undefined
  addVisible.value = true
}

const addOk = async () => {
  const result = await formTable.value?.validate()
  if (result !== true) return

  saveSafe({...formState}).then(res => {
    MessagePlugin.success(res.msg)
    loadData()
    addVisible.value = false
  })
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.rule-ellipsis {
  display: block;
  max-width: 420px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  cursor: default;
}

.rule-pop {
  margin: 0;
  font-size: 12px;
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-all;
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
}

.monaco-container {
  /* 必须显式给宽度：父级 .t-form__controls-content 是 flex 容器，
     flex 子项的宽度由内容决定，而 Monaco 内部不撑宽，会被压到只剩 7px */
  width: 100%;
  min-width: 0;
  height: 400px;
  border: 1px solid var(--hp-border);
  border-radius: 12px;
  overflow: hidden;
  /* 对齐 Monaco 深色主题，避免加载完成前先闪一下白底 */
  background: #1e1e1e;
}
</style>
