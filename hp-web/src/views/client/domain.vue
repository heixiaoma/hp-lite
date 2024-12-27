<template>
  <div>
    <a-button type="primary" style="margin-bottom: 10px" @click="addModal">添加域名</a-button>
    <a-button type="primary" style="margin-bottom: 10px;margin-left: 5px" @click="loadData">刷新</a-button>

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
          <a-button type="primary" style="margin-bottom: 5px;margin-left: 5px" @click="getSSl(record)">获取SSL证书</a-button>
          <a-button type="primary" style="margin-bottom: 5px;margin-left: 5px" @click="edit(record)">编辑</a-button>
          <a-button type="primary" style="margin-bottom: 5px;margin-left: 5px" @click="removeData(record)">删除</a-button>
        </template>
      </template>
    </a-table>
  </div>


  <div>
    <a-modal okText="确定" cancelText="取消" v-model:visible="addVisible" title="信息"
             @ok="addOk">
      <a-form :model="formState" ref="formTable" :layout="'vertical'" >
        <a-form-item label="域名 " >
          <a-input :disabled="!isAdd" v-model:value="formState.domain" placeholder="域名"/>
        </a-form-item>
        <a-form-item label="备注">
          <a-input v-model:value="formState.desc" placeholder="备注"/>
        </a-form-item>
        <a-form-item label="证书" name="certificateKey"
                     :rules="[{ required: false, message: '必须填写证书.key文件'}]">
          <a-textarea :rows="6" v-model:value="formState.certificateKey"
                       placeholder="-----BEGIN RSA PRIVATE KEY-----&#10;***大概是这样的证书私钥***&#10;-----END RSA PRIVATE KEY-----"/>
        </a-form-item>
        <a-form-item  label="证书内容" name="certificateContent"
                      :rules="[{ required: false, message: '映射描述必填'}]">
          <a-textarea  :rows="6" v-model:value="formState.certificateContent"
                      placeholder="-----BEGIN CERTIFICATE-----&#10;***大概是这样的证书内容***&#10;-----BEGIN CERTIFICATE-----"/>
        </a-form-item>

      </a-form>
    </a-modal>
  </div>



</template>

<script setup>
import {getDomain, removeDomain, addDomain,genSSL} from "../../api/client/domain.js";
import {onMounted, reactive, ref} from "vue";
import {notification} from "ant-design-vue";


const listData = ref();
const dataLoading = ref(false);
const addVisible = ref(false);
const isAdd = ref(false);

const formState = reactive({
  domain: "",
  desc: "",
  id:""
})
const pagination = reactive({
  total: 0,
  current: 1,
  pageSize: 10,
});

const loadData = () => {
  dataLoading.value = true
  getDomain({
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
  removeDomain({
    id: item.id
  }).then(res => {
    notification.open({
      message: res.msg,
    })
    loadData()
  })
}



const edit = (item) => {
  isAdd.value=false
  formState.desc = item.desc
  formState.id = item.id
  formState.domain = item.domain
  formState.certificateKey = item.certificateKey.trim()
  formState.certificateContent = item.certificateContent.trim()
  addVisible.value = true
}
const getSSl = (item) => {
  genSSL({
    id: item.id
  }).then(res => {
    notification.open({
      message: "任务已经提交，请稍等几分钟刷新列表",
    })
    loadData()
  })
}

const columns = [
  {title: '编号', dataIndex: 'id', key: 'id'},
  {title: '域名', dataIndex: 'domain', key: 'domain'},
  {title: '备注', dataIndex: 'desc', key: 'desc'},
  {title: '证书密钥', dataIndex: 'certificateKey', key: 'certificateKey', ellipsis: true,},
  {title: '证书内容', dataIndex: 'certificateContent', key: 'certificateContent' ,ellipsis: true,},
  {title: '状态', dataIndex: 'status', key: 'status'},
  {title: '提示', dataIndex: 'tips', key: 'tips'},
  {title: '操作', key: 'action'},
];


const handleTableChange = (item) => {
  pagination.current = item.current
  pagination.pageSize = item.pageSize
  pagination.total = item.total
  loadData()
}

const addModal = () => {
  isAdd.value=true
  formState.domain = ""
  formState.desc = ""
  formState.certificateKey = ''
  formState.certificateContent = ''
  formState.id = undefined
  addVisible.value = true
}

const addOk = () => {
  addDomain({...formState}).then(res => {
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
