<template>
  <div class="hp-page">
    <div class="hp-toolbar">
      <t-button theme="primary" @click="addModal">
        <template #icon><add-icon/></template>
        创建代理服务器
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
      <template #auth="{ row }">
        <div v-if="row.user && row.pwd">
          <div>认证用户：{{ row.user }}</div>
          <div class="hp-sub-line">认证密码：{{ row.pwd }}</div>
        </div>
        <t-tag v-else>无认证</t-tag>
      </template>

      <template #type="{ row }">
        <t-tag variant="light" :theme="row.type === '1' ? 'primary' : 'success'">
          {{ row.type === '1' ? 'http/https' : 'socks5' }}
        </t-tag>
      </template>

      <template #status="{ row }">
        <t-tag variant="light" :theme="row.status === '1' ? 'success' : 'default'">
          {{ row.status === '1' ? '启用' : '未启用' }}
        </t-tag>
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
          <t-popconfirm content="确定要删除该代理服务器？" theme="danger" @confirm="removeData(row)">
            <t-button size="small" variant="outline" theme="danger">删除</t-button>
          </t-popconfirm>
        </div>
      </template>
    </t-table>

    <t-dialog
        v-model:visible="addVisible"
        :header="formState.id ? '编辑代理服务器' : '创建代理服务器'"
        confirm-btn="确定"
        cancel-btn="取消"
        width="560px"
        @confirm="addOk"
    >
      <t-form :data="formState" ref="formTable" :rules="formRules" layout="vertical">
        <t-form-item label="端口" name="port">
          <t-input v-model="formState.port" clearable placeholder="端口"/>
        </t-form-item>
        <t-form-item label="认证用户名">
          <t-input v-model="formState.user" clearable placeholder="认证用户名"/>
        </t-form-item>
        <t-form-item label="认证密码">
          <t-input v-model="formState.pwd" clearable placeholder="认证密码"/>
        </t-form-item>
        <t-form-item label="类型" name="type">
          <t-select v-model="formState.type" :options="typeOptions" placeholder="请选择代理类型"/>
        </t-form-item>
        <t-form-item label="启用状态" name="status">
          <t-select v-model="formState.status" :options="statusOptions" placeholder="请选择启用状态"/>
        </t-form-item>
        <t-form-item label="备注" name="desc">
          <t-input v-model="formState.desc" clearable placeholder="备注"/>
        </t-form-item>
      </t-form>
    </t-dialog>
  </div>
</template>

<script setup>
import {getForward, removeForward, saveForward} from "../../api/client/forward.js";
import {onMounted, reactive, ref} from "vue";
import {MessagePlugin} from "tdesign-vue-next";
import {AddIcon, RefreshIcon} from 'tdesign-icons-vue-next';

const formTable = ref();
const listData = ref([]);
const dataLoading = ref(false);
const addVisible = ref(false);

const formState = reactive({
  port: "",
  user: "",
  pwd: "",
  type: '',
  status: "",
  desc: "",
  id: undefined
})

const formRules = {
  port: [{required: true, message: '端口必填', type: 'error'}],
  type: [{required: true, message: '代理类型必选', type: 'error'}],
  status: [{required: true, message: '启用状态必选', type: 'error'}],
  desc: [{required: true, message: '备注必填', type: 'error'}],
};

const typeOptions = [
  {label: 'HTTP/HTTPS', value: '1'},
  {label: 'Socks5', value: '2'},
];

const statusOptions = [
  {label: '开启', value: '1'},
  {label: '关闭', value: '0'},
];

const pagination = reactive({
  total: 0,
  current: 1,
  pageSize: 10,
  pageSizeOptions: [10, 20, 50],
});

const loadData = () => {
  dataLoading.value = true
  getForward({
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
  removeForward({
    id: item.id
  }).then(res => {
    MessagePlugin.success(res.msg)
    loadData()
  })
}

const edit = (item) => {
  formState.port = item.port
  formState.user = item.user
  formState.pwd = item.pwd
  formState.type = item.type
  formState.status = item.status
  formState.desc = item.desc
  formState.id = item.id
  addVisible.value = true
}

const columns = [
  {colKey: 'id', title: '编号', width: 90},
  {colKey: 'port', title: '端口', width: 110},
  {colKey: 'auth', title: '认证信息', width: 200},
  {colKey: 'type', title: '类型', width: 120},
  {colKey: 'status', title: '启用', width: 100},
  {colKey: 'tips', title: '服务状态'},
  {colKey: 'user', title: '归属', width: 200},
  {colKey: 'action', title: '操作', width: 150},
];

const onPageChange = (pageInfo) => {
  pagination.current = pageInfo.current
  pagination.pageSize = pageInfo.pageSize
  loadData()
}

const addModal = () => {
  formState.port = ''
  formState.user = ''
  formState.pwd = ''
  formState.type = ''
  formState.status = ''
  formState.desc = ""
  formState.id = undefined
  addVisible.value = true
}

const addOk = async () => {
  const result = await formTable.value?.validate()
  if (result !== true) return

  saveForward({...formState}).then(res => {
    MessagePlugin.success(res.msg)
    loadData()
    addVisible.value = false
  })
}

onMounted(() => {
  loadData()
})
</script>
