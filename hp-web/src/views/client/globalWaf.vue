<template>
  <div class="hp-page">
    <!-- 和「防护规则」最容易混的地方写在这里：那份要绑到穿透上才生效，这份是系统级的 -->
    <div class="hp-tip">
      <b>全局生效</b>
      <span>
        这里的规则在 HTTP / HTTPS 入口统一执行，对所有域名、所有用户的进站流量生效，
        不用绑定到某条穿透或反向代理。保存后即时重建，不需要重启服务；
        把「启用」关掉即可临时摘掉一条规则，规则内容仍留在库里。
        仅管理员可增删改。
      </span>
    </div>

    <div class="hp-toolbar">
      <t-button v-if="isAdmin" theme="primary" @click="addModal">
        <template #icon><add-icon/></template>
        添加规则
      </t-button>
      <t-button variant="outline" theme="primary" @click="loadData">
        <template #icon><refresh-icon/></template>
        刷新列表
      </t-button>
      <span v-if="!isAdmin" class="hp-toolbar__note">当前账号为子用户，只能查看</span>
    </div>

    <t-table
        class="hp-table"
        row-key="id"
        :data="listData || []"
        :columns="showColumns"
        :loading="dataLoading"
        :pagination="pagination"
        empty="暂无全局规则，添加一条试试看看"
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

      <template #remark="{ row }">
        <span v-if="row.remark">{{ row.remark }}</span>
        <span v-else class="hp-sub-line">—</span>
      </template>

      <template #enabled="{ row }">
        <t-switch v-if="isAdmin" :value="row.enabled" @change="(v) => toggle(row, v)">
          <template #label="slotProps">{{ slotProps.value ? '启用' : '停用' }}</template>
        </t-switch>
        <t-tag v-else :theme="row.enabled ? 'success' : 'default'" variant="light">
          {{ row.enabled ? '启用' : '停用' }}
        </t-tag>
      </template>

      <template #updatedAt="{ row }">
        <span class="hp-sub-line">{{ fmtTime(row.updatedAt) }}</span>
      </template>

      <template #action="{ row }">
        <div class="hp-actions">
          <t-button size="small" variant="outline" theme="primary" @click="edit(row)">编辑</t-button>
          <t-popconfirm content="确定要删除该全局规则？" theme="danger" @confirm="removeData(row)">
            <t-button size="small" variant="outline" theme="danger">删除</t-button>
          </t-popconfirm>
        </div>
      </template>
    </t-table>

    <t-dialog
        v-model:visible="addVisible"
        :header="formState.id ? '编辑全局规则' : '添加全局规则'"
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

          <t-form-item label="备注" name="remark">
            <t-input v-model="formState.remark" clearable placeholder="这条规则挡的是什么，给以后看的"/>
          </t-form-item>

          <t-form-item label="是否启用" name="enabled">
            <t-switch v-model="formState.enabled">
              <template #label="slotProps">{{ slotProps.value ? '启用' : '停用' }}</template>
            </t-switch>
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
import {getGlobalWaf, removeGlobalWaf, saveGlobalWaf} from "../../api/client/globalWaf";
import {computed, nextTick, onMounted, reactive, ref, watch} from "vue";
import {MessagePlugin} from "tdesign-vue-next";
import {AddIcon, RefreshIcon} from 'tdesign-icons-vue-next';
import {useResponsiveColumns} from '../../utils/responsive';
import userInfoStore from '../../data/userInfo';

import MonacoEditor from 'monaco-editor-vue3'
// SecLang 的语言 / 主题定义和「防护规则」共用一份，见 utils/seclang.js
import {applySeclangModel, seclangEditorOptions as editorOptions, setupSeclangMonaco} from '../../utils/seclang';

// 只有管理员能改：后端接口也是 Admin 校验，这里只是先把按钮收起来
const isAdmin = computed(() => userInfoStore.getUserInfo()?.role === 'ADMIN')

onMounted(() => {
  try {
    setupSeclangMonaco()
  } catch (e) {
    console.error('Monaco 初始化失败：', e)
  }
})

let editorInstance = null

const handleEditorMounted = (editor) => {
  editorInstance = editor
  applySeclangModel(editor)
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
  id: undefined,
  ruleName: "",
  remark: "",
  enabled: true,
  rule: "",
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
  getGlobalWaf({
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
  removeGlobalWaf({id: item.id}).then(res => {
    MessagePlugin.success(res.msg)
    loadData()
  })
}

// 列表上的开关：直接整条保存，后端会重新编译整套规则
const toggle = (row, value) => {
  saveGlobalWaf({
    id: row.id,
    ruleName: row.ruleName,
    rule: row.rule,
    remark: row.remark,
    enabled: value,
  }).then(res => {
    MessagePlugin.success(res.msg)
    loadData()
  }).catch(() => {
    // 保存失败（例如这套规则编译不过）时把开关拨回原状态
    loadData()
  })
}

const edit = (itemOld) => {
  const item = JSON.parse(JSON.stringify(itemOld))
  formState.id = item.id
  formState.ruleName = item.ruleName
  formState.remark = item.remark || ""
  formState.enabled = !!item.enabled
  formState.rule = item.rule
  addVisible.value = true
}

/* 窄屏留规则名 + 内容 + 状态 + 操作：编号、备注、时间让位 */
const columns = [
  {colKey: 'id', title: '编号', width: 90, mobile: false},
  {colKey: 'ruleName', title: '规则名字', width: 180},
  {colKey: 'rule', title: '规则内容', ellipsis: true},
  {colKey: 'remark', title: '备注', width: 180, mobile: false},
  {colKey: 'enabled', title: '状态', width: 120},
  {colKey: 'updatedAt', title: '更新时间', width: 160, mobile: false},
  {colKey: 'action', title: '操作', width: 150},
];
const viewColumns = useResponsiveColumns(columns);
// 子用户看得到清单，但不给编辑入口
const showColumns = computed(() => {
  if (isAdmin.value) return viewColumns.value
  return viewColumns.value.filter((c) => c.colKey !== 'action')
})

const fmtTime = (v) => {
  if (!v) return '—'
  const d = new Date(v)
  if (Number.isNaN(d.getTime())) return '—'
  return d.toLocaleString('zh-CN', {hour12: false})
}

const onPageChange = (pageInfo) => {
  pagination.current = pageInfo.current
  pagination.pageSize = pageInfo.pageSize
  loadData()
}

const addModal = () => {
  formState.id = undefined
  formState.ruleName = ""
  formState.remark = ""
  formState.enabled = true
  formState.rule = ""
  addVisible.value = true
}

const addOk = async () => {
  const result = await formTable.value?.validate()
  if (result !== true) return

  // 规则编译不过时后端会带错误回来，http.js 的响应拦截器已经弹了提示
  saveGlobalWaf({...formState}).then(res => {
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
.hp-tip {
  display: flex;
  gap: 8px;
  padding: 12px 16px;
  font-size: 13px;
  line-height: 1.7;
  color: var(--hp-text-2);
  background: #f4f8ff;
  border: 1px solid var(--hp-border);
  border-radius: var(--hp-radius-sm);
}

.hp-tip b {
  flex: none;
  color: var(--hp-text);
}

.hp-toolbar__note {
  font-size: 13px;
  color: var(--hp-text-3);
}

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
