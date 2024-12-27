<template>
  <div>
    <a-button type="primary" style="margin-bottom: 10px" @click="addModal">添加用户</a-button>

    <a-table :loading="dataLoading" :columns="columns" rowKey="id" :data-source="listData"
             :locale="{emptyText: '暂无数据,添加一个试试看看'}"
             :pagination="pagination"
             @change="handleTableChange"
             :scroll="{ x: 10 }">

      <template #bodyCell="{ column ,record}">

        <template v-if="column.key === 'createTime'">
          {{new Date(record.createTime).toLocaleString()}}
        </template>

        <template v-if="column.key === 'action'">
          <a-button type="primary" style="margin-bottom: 5px;margin-left: 5px" @click="edit(record)">编辑</a-button>
          <a-button type="primary" style="margin-bottom: 5px;margin-left: 5px" @click="removeData(record)">删除</a-button>
        </template>
      </template>
    </a-table>
  </div>


  <div>
    <a-modal okText="确定" cancelText="取消" v-model:visible="addVisible" title="添加"
             @ok="addOk">
      <a-form :model="formState" ref="formTable" :layout="'vertical'" >
        <a-form-item label="用户名 ">
          <a-input v-model:value="formState.username" placeholder="用户名"/>
        </a-form-item>
        <a-form-item label="密码">
          <a-input v-model:value="formState.password" placeholder="密码"/>
        </a-form-item>
        <a-form-item label="备注">
          <a-input v-model:value="formState.desc" placeholder="备注"/>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>

</template>

<script setup>
import {getUser, removeUser, saveUser} from "../../api/client/client_user";
import {onMounted, reactive, ref} from "vue";
import {notification} from "ant-design-vue";


const listData = ref();
const dataLoading = ref(false);
const addVisible = ref(false);
const formState = reactive({
  username: "",
  password: "",
  desc:"",
  id:""
})
const pagination = reactive({
  total: 0,
  current: 1,
  pageSize: 10,
});

const loadData = () => {
  dataLoading.value = true
  getUser({
    current: pagination.current,
    pageSize: pagination.pageSize
  }).then(res => {
    dataLoading.value = false
    listData.value = res.data.records
    pagination.current = res.data.current
    pagination.total = res.data.total
  })
}

const removeData = (item) => {
  removeUser({
    id: item.id
  }).then(res => {
    notification.open({
      message: res.msg,
    })
    loadData()
  })
}

const edit = (item) => {
  formState.username = item.username
  formState.password = item.password
  formState.desc = item.desc
  formState.id = item.id
  addVisible.value=true
}

const columns = [
  {title: '编号', dataIndex: 'id', key: 'id'},
  {title: '用户名', dataIndex: 'username', key: 'username'},
  {title: '密码', dataIndex: 'password', key: 'password'},
  {title: '备注', dataIndex: 'desc', key: 'desc'},
  {title: '创建时间', dataIndex: 'createTime', key: 'createTime'},
  {title: '操作', key: 'action'},
];


const handleTableChange = (item) => {
  pagination.current = item.current
  pagination.pageSize = item.pageSize
  pagination.total = item.total
  loadData()
}

const addModal = () => {
  formState.username = ""
  formState.password = ""
  formState.desc = ""
  formState.id = undefined
  addVisible.value = true
}

const addOk = () => {
  saveUser({...formState}).then(res => {
    notification.open({
      message: res.msg,
    })
    loadData()
    addVisible.value = false
  })
}

onMounted(() => {
  loadData()
})

</script>

<style scoped>

</style>
