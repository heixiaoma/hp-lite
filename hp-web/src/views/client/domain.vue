<template>
  <div class="hp-page">
    <div class="hp-toolbar">
      <t-button theme="primary" @click="addModal">
        <template #icon><add-icon/></template>
        添加域名
      </t-button>
      <t-button variant="outline" theme="primary" @click="loadData">
        <template #icon><refresh-icon/></template>
        刷新列表
      </t-button>
      <div class="hp-toolbar__grow"></div>
      <t-input
          v-model="pagination.keyword"
          class="hp-search-input"
          clearable
          placeholder="关键字查询"
          @enter="loadData"
      >
        <template #prefix-icon><search-icon/></template>
      </t-input>
      <t-button theme="primary" @click="loadData">查询</t-button>
    </div>

    <t-table
        class="hp-table"
        row-key="id"
        :data="listData"
        :columns="columns"
        :loading="dataLoading"
        :pagination="pagination"
        empty="暂无数据，添加一个试试看看"
        table-layout="auto"
        stripe
        hover
        @page-change="onPageChange"
    >
      <template #certificateKey="{ row }">
        <t-link v-if="row.certificateKey" theme="primary" hover="color" @click="showText('证书密钥', row.certificateKey)">
          密钥
        </t-link>
        <span v-else class="hp-muted">-</span>
      </template>

      <template #certificateContent="{ row }">
        <t-link v-if="row.certificateContent" theme="primary" hover="color" @click="showText('证书内容', row.certificateContent)">
          证书
        </t-link>
        <span v-else class="hp-muted">-</span>
      </template>

      <template #status="{ row }">
        <t-tag v-if="row.status" :theme="row.status === 'SUCCESS' ? 'success' : 'warning'" variant="light">
          {{ row.status }}
        </t-tag>
        <span v-else class="hp-muted">-</span>
      </template>

      <template #user="{ row }">
        <div v-if="!row.userDesc && !row.username">
          <t-tag variant="light" theme="primary">自用域名</t-tag>
        </div>
        <div v-else>
          <div>归属用户：{{ row.username }}</div>
          <div class="hp-sub-line">归属用户备注：{{ row.userDesc }}</div>
        </div>
      </template>

      <template #action="{ row }">
        <div class="hp-actions">
          <t-button size="small" variant="outline" theme="success" @click="getSSl(row)">获取SSL证书</t-button>
          <t-button size="small" variant="outline" theme="primary" @click="edit(row)">编辑</t-button>
          <t-popconfirm content="确定要删除该域名？" theme="danger" @confirm="removeData(row)">
            <t-button size="small" variant="outline" theme="danger">删除</t-button>
          </t-popconfirm>
        </div>
      </template>
    </t-table>

    <!-- 新增 / 编辑 -->
    <t-dialog
        v-model:visible="addVisible"
        :header="formState.id ? '编辑域名' : '添加域名'"
        confirm-btn="确定"
        cancel-btn="取消"
        width="640px"
        @confirm="addOk"
    >
      <div class="hp-dialog-body">
        <t-form :data="formState" ref="formTable" :rules="formRules" layout="vertical">
          <t-form-item label="域名" name="domain">
            <t-input v-model="formState.domain" :disabled="!isAdd" clearable placeholder="域名"/>
          </t-form-item>
          <t-form-item label="备注" name="desc">
            <t-input v-model="formState.desc" clearable placeholder="备注"/>
          </t-form-item>
          <t-form-item label="证书" name="certificateKey">
            <t-textarea
                v-model="formState.certificateKey"
                :autosize="{ minRows: 5, maxRows: 8 }"
                placeholder="-----BEGIN RSA PRIVATE KEY-----&#10;***大概是这样的证书私钥***&#10;-----END RSA PRIVATE KEY-----"
            />
          </t-form-item>
          <t-form-item label="证书内容" name="certificateContent">
            <t-textarea
                v-model="formState.certificateContent"
                :autosize="{ minRows: 5, maxRows: 8 }"
                placeholder="-----BEGIN CERTIFICATE-----&#10;***大概是这样的证书内容***&#10;-----BEGIN CERTIFICATE-----"
            />
          </t-form-item>
        </t-form>
      </div>
    </t-dialog>

    <!-- 证书内容查看 -->
    <t-dialog v-model:visible="textVisible" :header="textTitle" :footer="false" width="720px">
      <pre class="cert-view">{{ textContent }}</pre>
    </t-dialog>
  </div>
</template>

<script setup>
import {addDomain, genSSL, getDomain, removeDomain} from "../../api/client/domain.js";
import {onMounted, reactive, ref} from "vue";
import {MessagePlugin} from "tdesign-vue-next";
import {AddIcon, RefreshIcon, SearchIcon} from 'tdesign-icons-vue-next';

const formTable = ref();
const listData = ref([]);
const dataLoading = ref(false);
const addVisible = ref(false);
const isAdd = ref(false);

const textVisible = ref(false);
const textTitle = ref('');
const textContent = ref('');

const formState = reactive({
  domain: "",
  desc: "",
  id: undefined,
  certificateKey: '',
  certificateContent: '',
})

const formRules = {
  domain: [{required: true, message: '必选域名', type: 'error'}],
  desc: [{required: true, message: '必选备注', type: 'error'}],
};

const pagination = reactive({
  total: 0,
  current: 1,
  pageSize: 10,
  pageSizeOptions: [10, 20, 50],
  keyword: ''
});

const loadData = () => {
  dataLoading.value = true
  getDomain({
    current: pagination.current,
    pageSize: pagination.pageSize,
    keyword: pagination.keyword,
  }).then(res => {
    dataLoading.value = false
    listData.value = res.data.records
    pagination.total = res.data.total
  }).catch(() => {
    dataLoading.value = false
  })
}

const removeData = (item) => {
  removeDomain({
    id: item.id
  }).then(res => {
    MessagePlugin.success(res.msg)
    loadData()
  })
}

const edit = (item) => {
  isAdd.value = false
  formState.desc = item.desc
  formState.id = item.id
  formState.domain = item.domain
  formState.certificateKey = (item.certificateKey || '').trim()
  formState.certificateContent = (item.certificateContent || '').trim()
  addVisible.value = true
}

const getSSl = (item) => {
  genSSL({
    id: item.id
  }).then(() => {
    MessagePlugin.success("任务已经提交，请稍等几分钟刷新列表")
    loadData()
  })
}

const showText = (title, content) => {
  textTitle.value = title
  textContent.value = content
  textVisible.value = true
}

const columns = [
  {colKey: 'id', title: '编号', width: 90},
  {colKey: 'domain', title: '域名'},
  {colKey: 'desc', title: '备注'},
  {colKey: 'certificateKey', title: '证书密钥', width: 100, align: 'center'},
  {colKey: 'certificateContent', title: '证书内容', width: 100, align: 'center'},
  {colKey: 'status', title: '状态', width: 120},
  {colKey: 'tips', title: '提示'},
  {colKey: 'user', title: '归属'},
  {colKey: 'action', title: '操作', width: 250},
];

const onPageChange = (pageInfo) => {
  pagination.current = pageInfo.current
  pagination.pageSize = pageInfo.pageSize
  loadData()
}

const addModal = () => {
  isAdd.value = true
  formState.domain = ""
  formState.desc = ""
  formState.certificateKey = ''
  formState.certificateContent = ''
  formState.id = undefined
  addVisible.value = true
}

const addOk = async () => {
  const result = await formTable.value?.validate()
  if (result !== true) return

  addDomain({...formState}).then(res => {
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
.cert-view {
  margin: 0;
  padding: 14px;
  background: #f8fafc;
  border: 1px solid var(--hp-border);
  border-radius: 10px;
  font-size: 12.5px;
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 52vh;
  overflow: auto;
}
</style>
