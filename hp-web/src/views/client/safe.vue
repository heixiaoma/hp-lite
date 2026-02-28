<template>
  <div>
    <a-button  style="margin-bottom: 10px" class="btn edit" @click="addModal">添加规则</a-button>
    <a-button class="btn view" style="margin-bottom: 10px;margin-left: 5px" @click="loadData">刷新列表</a-button>

      <a-table :loading="dataLoading" :columns="columns" rowKey="id" :data-source="listData"
               :locale="{emptyText: '暂无数据,添加一个试试看看'}"
               :pagination="pagination"
               @change="handleTableChange">

        <template #bodyCell="{ column ,record}">
          <template v-if="column.key === 'user'">
            <div v-if="!record.userDesc&&!record.username">
              自用
            </div>
            <div v-else>
              <div>归属用户：{{record.username}}</div>
              <div>归属用户备注：{{record.userDesc}}</div>
            </div>
          </template>

          <template v-if="column.key === 'action'">
            <a-button  class="btn edit" style="margin-bottom: 5px;margin-left: 5px" @click="edit(record)">编辑</a-button>
            <a-button  class="btn delete" style="margin-bottom: 5px;margin-left: 5px" @click="removeData(record)">删除</a-button>
          </template>
        </template>
      </a-table>
  </div>

  <div>
    <a-modal  v-model:visible="addVisible" title="添加" width="80%">
      <div class="config-info">
      <a-form :model="formState" ref="formTable" :layout="'vertical'" >
        <a-form-item label="规则名字" name="ruleName"  :rules="[{ required: true, message: '规则名字'}]">
          <a-input style="width: 90%" v-model:value="formState.ruleName" placeholder="规则名字"/>
        </a-form-item>
        <a-form-item label="规则" name="rule" :rules="[{ required: true, message: '规则'}]">
          <div class="monaco-container" style="height: 400px; border: 1px solid #e5e7eb;">
          <MonacoEditor
              v-model:value="formState.rule"
              language="seclang"
              :options="editorOptions"
              @mounted="handleEditorMounted"
          />
          </div>
        </a-form-item>
      </a-form>
      </div>

      <template #footer>
        <a-button class="btn view" @click="addVisible=!addVisible">取消</a-button>
        <a-button class="btn edit" @click="addOk">确定</a-button>
      </template>
    </a-modal>

  </div>

</template>

<script setup>
import {getSafe, removeSafe, saveSafe} from "../../api/client/safe";
import {onMounted, reactive, ref} from "vue";
import {notification} from "ant-design-vue";

import MonacoEditor from 'monaco-editor-vue3'
import * as monaco from 'monaco-editor'
import 'monaco-editor/min/vs/editor/editor.main.css'


// 编辑器配置
const editorOptions = {
  fontSize: 14,
  lineNumbers: 'on',
  lineWrapping: true,
  tabSize: 2,
  minimap: { enabled: true },
  scrollBeyondLastLine: false,
  placeholder: '请输入 ModSecurity SecLang 规则（兼容 OWASP CRS v4）...',
  theme: 'seclang-theme'
}

onMounted(async () => {
  try {
    // 1. 先注销已存在的 seclang 语言（避免重复注册）
    const existingLang = monaco.languages.getLanguages().find(l => l.id === 'seclang')
    if (existingLang) {
      monaco.languages.unregister(existingLang.id)
    }

    // 2. 注册 seclang 语言（极简配置，避免解析器校验）
    monaco.languages.register({ id: 'seclang' })

    // 3. 重构 tokenizer 规则（核心：避开 rx 解析陷阱）
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
          [/"/, { token: 'string.quote', next: '@string_double' }],
          [/'/, { token: 'string.quote', next: '@string_single' }],

          // ---------------- 运算符 ----------------
          [/[!~<>]=?/, 'operator'],
          [/[=|&]/, 'operator'],

          // ---------------- 分隔符 ----------------
          [/[(),]/, 'delimiter']
        ],

        string_double: [
          [/[^\\"]+/, 'string'],
          [/\\./, 'string.escape'],
          [/"/, { token: 'string.quote', next: '@pop' }]
        ],

        string_single: [
          [/[^\\']+/, 'string'],
          [/\\./, 'string.escape'],
          [/'/, { token: 'string.quote', next: '@pop' }]
        ]
      },

      comments: {
        lineComment: '#'
      }
    })

    // 4. 注册主题（仅用基础 token 类型，无自定义属性）
    monaco.editor.defineTheme('seclang-theme', {
      base: 'vs-dark',
      inherit: true,
      rules: [
        { token: 'comment', foreground: '808080', fontStyle: 'italic' },
        { token: 'keyword', foreground: '569CD6', fontStyle: 'bold' },
        { token: 'attribute', foreground: '9CDCFE' },
        { token: 'type', foreground: 'DCDCAA' },
        { token: 'regexp', foreground: 'B5CEA8' },
        { token: 'string', foreground: 'CE9178' },
        { token: 'variable', foreground: '9CDCFE' }
      ],
      colors: {
        'editor.background': '#1E1E1E',
        'editor.lineHighlightBackground': '#2A2A2A',
        'editor.foreground': '#D4D4D4'
      }
    })
    // 5. 强制应用主题和语言
    monaco.editor.setTheme('seclang-theme')
  } catch (e) {
    console.error('Monaco 初始化失败：', e)
  }
})

// 编辑器挂载后兜底（确保语言生效）
const handleEditorMounted = (editor) => {
  // 直接设置模型语言，跳过解析器的规则校验
  const model = editor.getModel()
  if (model) {
    monaco.editor.setModelLanguage(model, 'seclang')
  }
  monaco.editor.setTheme('seclang-theme')
  console.log('已注册语言：', monaco.languages.getLanguages().map(l => l.id))
}

const listData = ref();
const formTable = ref();
const dataLoading = ref(false);
const addVisible = ref(false);
const formState = reactive({
  ruleName:"",
  rule:"",
  id:""
})
const pagination = reactive({
  total: 0,
  current: 1,
  pageSize: 10,
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
  })
}

const removeData = (item) => {
  removeSafe({
    id: item.id
  }).then(res => {
    notification.open({
      message: res.msg,
    })
    loadData()
  })
}



const edit = (itemOld) => {
  const item=JSON.parse(JSON.stringify(itemOld))
  formState.rule = item.rule
  formState.ruleName = item.ruleName
  formState.id = item.id
  addVisible.value=true
}

const columns = [
  {title: '编号', dataIndex: 'id', key: 'id'},
  {title: '规则名字', dataIndex: 'ruleName', key: 'ruleName'},
  {title: '规则内容', dataIndex: 'rule', key: 'rule',ellipsis: true},
  {title: '归属', dataIndex: 'user', key: 'user'},
  {title: '操作', key: 'action'},
];

const handleTableChange = (item) => {
  pagination.current = item.current
  pagination.pageSize = item.pageSize
  pagination.total = item.total
  loadData()
}

const addModal = () => {
  formState.rule = ""
  formState.ruleName =""
  formState.id = undefined
  addVisible.value = true
}

const addOk = () => {
  formTable.value.validate().then(valid => {
    saveSafe({...formState}).then(res => {
      notification.open({
        message: res.msg,
      })
      loadData()
      addVisible.value = false
    })
  })
}

onMounted(() => {
  loadData()
})

</script>

<style scoped>

.config-info{
  height: 60vh;
  overflow-y: scroll;

}
/* 滚动条整体样式 */
.config-info::-webkit-scrollbar {
  width: 2px; /* 滚动条宽度 */
  height: 2px;
}

/* 滚动条轨道 */
.config-info::-webkit-scrollbar-track {
  background: #f1f1f1; /* 轨道背景色 */
  border-radius: 1px;
}

/* 滚动条滑块 */
.config-info::-webkit-scrollbar-thumb {
  background: #888; /* 滑块颜色 */
  border-radius: 1px; /* 滑块圆角 */
}

/* 滑块悬停效果 */
.config-info::-webkit-scrollbar-thumb:hover {
  background: #555; /* 悬停时滑块颜色 */
}
</style>
