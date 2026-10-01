<template>
  <div class="hp-page">
    <div class="hp-toolbar">
      <t-button theme="primary" @click="addModal">
        <template #icon><add-icon/></template>
        添加反向代理
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
        :columns="columns"
        :loading="dataLoading"
        :pagination="pagination"
        empty="暂无数据，添加一个试试看看"
        table-layout="auto"
        stripe
        hover
        @page-change="onPageChange"
    >
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
          <t-popconfirm content="确定要删除该反向代理？" theme="danger" @confirm="removeData(row)">
            <t-button size="small" variant="outline" theme="danger">删除</t-button>
          </t-popconfirm>
        </div>
      </template>
    </t-table>

    <t-dialog
        v-model:visible="addVisible"
        :header="formState.id ? '编辑反向代理' : '添加反向代理'"
        confirm-btn="确定"
        cancel-btn="取消"
        width="560px"
        @confirm="addOk"
    >
      <t-form :data="formState" ref="formTable" :rules="formRules" layout="vertical">
        <t-form-item label="域名" name="domain">
          <t-select
              v-model="formState.domain"
              filterable
              clearable
              placeholder="选择一个域名"
              :options="domainOptions"
          />
        </t-form-item>
        <t-form-item label="地址" name="address">
          <t-input v-model="formState.address" clearable placeholder="http://127.0.0.1:9090"/>
        </t-form-item>
        <t-form-item label="备注" name="desc">
          <t-input v-model="formState.desc" clearable placeholder="备注"/>
        </t-form-item>
        <t-form-item label="防火墙规则" name="safeId">
          <t-select v-model="formState.safeId" clearable placeholder="选择一个规则" :options="safeOptions"/>
        </t-form-item>
      </t-form>
    </t-dialog>
  </div>
</template>

<script setup>
import {getReverse, removeReverse, saveReverse} from "../../api/client/reverse.js";
import {onMounted, reactive, ref} from "vue";
import {MessagePlugin} from "tdesign-vue-next";
import {queryDomain} from "../../api/client/domain.js";
import {querySafe} from "../../api/client/safe.js";
import {AddIcon, RefreshIcon} from 'tdesign-icons-vue-next';

const formTable = ref();
const listData = ref([]);
const dataLoading = ref(false);
const addVisible = ref(false);
const domainOptions = ref([]);
const safeOptions = ref([]);

const formState = reactive({
  address: "",
  domain: "",
  desc: "",
  safeId: undefined,
  id: undefined
})

const formRules = {
  domain: [{required: true, message: '必选域名', type: 'error'}],
  address: [{required: true, message: '必填地址', type: 'error'}],
  desc: [{required: true, message: '必填备注', type: 'error'}],
};

const pagination = reactive({
  total: 0,
  current: 1,
  pageSize: 10,
  pageSizeOptions: [10, 20, 50],
});

const loadSafes = () => {
  querySafe({}).then(res => {
    const result = res.data.data || [];
    safeOptions.value = result.map(r => ({value: r.id, label: r.ruleName}))
  })
}

const loadDomains = () => {
  queryDomain({}).then(res => {
    const result = res.data.data || [];
    domainOptions.value = result.map(r => ({value: r.domain, label: r.domain}))
  })
}

const loadData = () => {
  loadSafes()
  loadDomains()
  dataLoading.value = true
  getReverse({
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
  removeReverse({
    id: item.id
  }).then(res => {
    MessagePlugin.success(res.msg)
    loadData()
  })
}

const edit = (item) => {
  formState.domain = item.domain
  formState.address = item.address
  formState.desc = item.desc
  formState.safeId = item.safeId
  formState.id = item.id
  addVisible.value = true
}

const columns = [
  {colKey: 'id', title: '编号', width: 90},
  {colKey: 'domain', title: '域名'},
  {colKey: 'address', title: '地址'},
  {colKey: 'desc', title: '备注'},
  {colKey: 'user', title: '归属', width: 200},
  {colKey: 'action', title: '操作', width: 150},
];

const onPageChange = (pageInfo) => {
  pagination.current = pageInfo.current
  pagination.pageSize = pageInfo.pageSize
  loadData()
}

const addModal = () => {
  formState.domain = ""
  formState.address = ""
  formState.desc = ""
  formState.safeId = undefined
  formState.id = undefined
  addVisible.value = true
}

const addOk = async () => {
  const result = await formTable.value?.validate()
  if (result !== true) return

  saveReverse({...formState}).then(res => {
    MessagePlugin.success(res.msg)
    loadData()
    addVisible.value = false
  })
}

onMounted(() => {
  loadData()
})
</script>
