<template>
  <div class="hp-page">
    <div class="hp-toolbar">
      <t-button theme="primary" @click="addModal">
        <template #icon><add-icon/></template>
        添加用户
      </t-button>
      <t-button variant="outline" theme="primary" @click="loadData">
        <template #icon><refresh-icon/></template>
        刷新列表
      </t-button>
    </div>

    <t-table
        class="hp-table"
        row-key="id"
        :data="listData"
        :columns="viewColumns"
        :loading="dataLoading"
        :pagination="pagination"
        empty="暂无数据，添加一个试试看看"
        table-layout="auto"
        stripe
        hover
        @page-change="onPageChange"
    >
      <template #createTime="{ row }">
        {{ new Date(row.createTime).toLocaleString() }}
      </template>

      <template #action="{ row }">
        <div class="hp-actions">
          <t-button size="small" variant="outline" theme="primary" @click="edit(row)">编辑</t-button>
          <t-popconfirm content="确定要删除该用户？" theme="danger" @confirm="removeData(row)">
            <t-button size="small" variant="outline" theme="danger">删除</t-button>
          </t-popconfirm>
        </div>
      </template>
    </t-table>

    <t-dialog
        v-model:visible="addVisible"
        :header="formState.id ? '编辑用户' : '添加用户'"
        confirm-btn="确定"
        cancel-btn="取消"
        width="560px"
        @confirm="addOk"
    >
      <t-form :data="formState" ref="formTable" :rules="formRules" layout="vertical" label-align="top">
        <t-form-item label="用户名" name="username">
          <t-input v-model="formState.username" clearable placeholder="用户名"/>
        </t-form-item>
        <t-form-item label="密码" name="password">
          <t-input v-model="formState.password" clearable placeholder="密码"/>
        </t-form-item>
        <t-form-item label="备注" name="desc">
          <t-input v-model="formState.desc" clearable placeholder="备注"/>
        </t-form-item>
      </t-form>
    </t-dialog>
  </div>
</template>

<script setup>
import {getUser, removeUser, saveUser} from "../../api/client/client_user";
import {onMounted, reactive, ref} from "vue";
import {MessagePlugin} from "tdesign-vue-next";
import {AddIcon, RefreshIcon} from 'tdesign-icons-vue-next';
import {useResponsiveColumns} from '../../utils/responsive';

const formTable = ref();
const listData = ref([]);
const dataLoading = ref(false);
const addVisible = ref(false);
const formState = reactive({
  username: "",
  password: "",
  desc: "",
  id: undefined
});

const formRules = {
  username: [{required: true, message: '必填用户名', type: 'error'}],
  password: [
    {required: true, message: '必填密码', type: 'error'},
    {min: 6, message: '密码长度不能少于6位', type: 'error', trigger: 'blur'},
  ],
  desc: [{required: true, message: '必填备注', type: 'error'}],
};

const pagination = reactive({
  total: 0,
  current: 1,
  pageSize: 10,
  pageSizeOptions: [10, 20, 50],
});

const loadData = () => {
  dataLoading.value = true
  getUser({
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
  removeUser({
    id: item.id
  }).then(res => {
    MessagePlugin.success(res.msg)
    loadData()
  })
}

const edit = (item) => {
  formState.username = item.username
  formState.password = item.password
  formState.desc = item.desc
  formState.id = item.id
  addVisible.value = true
}

/* 窄屏留用户名 + 备注 + 操作：编号/密码/创建时间属于次要信息 */
const columns = [
  {colKey: 'id', title: '编号', width: 90, mobile: false},
  {colKey: 'username', title: '用户名'},
  {colKey: 'password', title: '密码', mobile: false},
  {colKey: 'desc', title: '备注'},
  {colKey: 'createTime', title: '创建时间', width: 200, mobile: false},
  {colKey: 'action', title: '操作', width: 150},
];
const viewColumns = useResponsiveColumns(columns);

const onPageChange = (pageInfo) => {
  pagination.current = pageInfo.current
  pagination.pageSize = pageInfo.pageSize
  loadData()
}

const addModal = () => {
  formState.username = ""
  formState.password = ""
  formState.desc = ""
  formState.id = undefined
  addVisible.value = true
}

const addOk = async () => {
  const result = await formTable.value?.validate()
  if (result !== true) return

  saveUser({...formState}).then(res => {
    MessagePlugin.success(res.msg)
    loadData()
    addVisible.value = false
  })
}

onMounted(() => {
  loadData()
})
</script>
