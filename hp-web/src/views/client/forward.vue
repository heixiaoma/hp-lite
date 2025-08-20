<template>
  <div>
    <a-button  style="margin-bottom: 10px" class="btn edit" @click="addModal">创建代理服务器</a-button>
    <a-button class="btn view" style="margin-bottom: 10px;margin-left: 5px" @click="loadData">刷新列表</a-button>

    <a-table :loading="dataLoading" :columns="columns" rowKey="id" :data-source="listData"
             :locale="{emptyText: '暂无数据,添加一个试试看看'}"
             :pagination="pagination"
             @change="handleTableChange"
             :scroll="{ x: 'max-content' }">

      <template #bodyCell="{ column ,record}">
        <template v-if="column.key === 'action'">
          <a-button  class="btn edit" style="margin-bottom: 5px;margin-left: 5px" @click="edit(record)">编辑</a-button>
          <a-button  class="btn delete" style="margin-bottom: 5px;margin-left: 5px" @click="removeData(record)">删除</a-button>
        </template>

        <template v-if="column.key === 'user'">
          <div v-if="!record.userDesc&&!record.username">
            自用
          </div>
          <div v-else>
            <div>归属用户：{{record.username}}</div>
            <div>归属用户备注：{{record.userDesc}}</div>
          </div>
        </template>

        <template v-if="column.key === 'auth'">
          <div v-if="record.user&&record.pwd">
            <div>认证用户：{{record.user}}</div>
            <div>认证密码：{{record.pwd}}</div>
          </div>
          <div v-else>
            无认证
          </div>
        </template>

        <template v-if="column.key === 'type'">
          <div v-if="record.type==='1'">
              http/https
          </div>
          <div v-else>
            socks5
          </div>
        </template>

        <template v-if="column.key === 'status'">
          <div v-if="record.status==='1'">
            启用
          </div>
          <div v-else>
            未启用
          </div>
        </template>

      </template>
    </a-table>
  </div>


  <div>
    <a-modal  v-model:visible="addVisible" title="添加"
    >
      <a-form :model="formState" ref="formTable" :layout="'vertical'" >
        <a-form-item label="端口" name="port"   :rules="[{ required: true, message: '端口'}]">
          <a-input v-model:value="formState.port" placeholder="端口"/>
        </a-form-item>
        <a-form-item label="用户">
          <a-input v-model:value="formState.user" placeholder="认证用户名"/>
        </a-form-item>
        <a-form-item label="密码">
          <a-input v-model:value="formState.pwd" placeholder="认证密码"/>
        </a-form-item>
        <a-form-item label="类型" name="type"
                     :rules="[{ required: true, message: '代理类型'}]">
          <a-select
              v-model:value="formState.type"
          >
            <a-select-option value="1">HTTP/HTTPS</a-select-option>
            <a-select-option value="2">Socks5</a-select-option>
          </a-select>
        </a-form-item>

        <a-form-item label="启用状态" name="status"
                     :rules="[{ required: true, message: '启用状态'}]">
          <a-select
              v-model:value="formState.status"
          >
            <a-select-option value="1">开启</a-select-option>
            <a-select-option value="0">关闭</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="备注" name="desc" :rules="[{ required: true, message: '备注'}]">
          <a-input v-model:value="formState.desc" placeholder="备注"/>
        </a-form-item>
      </a-form>
      <template #footer>
        <a-button class="btn view" @click="addVisible=!addVisible">取消</a-button>
        <a-button class="btn edit" @click="addOk">确定</a-button>
      </template>
    </a-modal>
  </div>

</template>

<script setup>
import {getForward, removeForward, saveForward} from "../../api/client/forward.js";
import {onMounted, reactive, ref} from "vue";
import {notification} from "ant-design-vue";


const formTable = ref();
const listData = ref();
const dataLoading = ref(false);
const addVisible = ref(false);

const formState = reactive({
  port: "",
  user: "",
  pwd: "",
  type:'',
  status:"",
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
  getForward({
    current: pagination.current,
    pageSize: pagination.pageSize
  }).then(res => {
    dataLoading.value = false
    listData.value = res.data.records
    pagination.total = res.data.total
  })
}

const removeData = (item) => {
  removeForward({
    id: item.id
  }).then(res => {
    notification.open({
      message: res.msg,
    })
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
  addVisible.value=true
}

const columns = [
  {title: '编号', dataIndex: 'id', key: 'id'},
  {title: '端口', dataIndex: 'port', key: 'port'},
  {title: '认证信息', dataIndex: 'auth', key: 'auth'},
  {title: '类型', dataIndex: 'type', key: 'type'},
  {title: '启用', dataIndex: 'status', key: 'status'},
  {title: '服务状态', dataIndex: 'tips', key: 'tips'},
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
  formState.port = ''
  formState.user = ''
  formState.pwd = ''
  formState.type = ''
  formState.status = ''
  formState.desc = ""
  formState.id = undefined
  addVisible.value = true
}

const addOk = () => {
  formTable.value.validate().then(res => {
    saveForward({...formState}).then(res => {
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

</style>
