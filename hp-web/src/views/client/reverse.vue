<template>
  <div>
    <a-button  style="margin-bottom: 10px" class="btn edit" @click="addModal">添加反向代理</a-button>
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
      </template>
    </a-table>
  </div>


  <div>
    <a-modal  v-model:visible="addVisible" title="添加"
             >
      <a-form :model="formState" ref="formTable" :layout="'vertical'" >
        <a-form-item label="&nbsp;域名&nbsp;&nbsp;" name="domain"  :rules="[{ required: true, message: '必选域名'}]">
          <a-select
              v-model:value="formState.domain"
              show-search
              placeholder="选择一个域名"
              :options="domainOptions"
          ></a-select>
        </a-form-item>
        <a-form-item label="地址" name="address"  :rules="[{ required: true, message: '必填地址'}]">
          <a-input v-model:value="formState.address" placeholder="http://127.0.0.1:9090"/>
        </a-form-item>
        <a-form-item label="备注" name="desc"  :rules="[{ required: true, message: '必填备注'}]">
          <a-input v-model:value="formState.desc" placeholder="备注"/>
        </a-form-item>

        <a-form-item label="防火墙规则" name="safeId" :rules="[{required: false,message: '防火墙规则比选，没有就去创建一个' }]">
          <a-select
              v-model:value="formState.safeId"
              :options="safeOptions"
          >
          </a-select>
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
import {getReverse, removeReverse, saveReverse} from "../../api/client/reverse.js";
import {onMounted, reactive, ref} from "vue";
import {notification} from "ant-design-vue";
import {queryDomain} from "../../api/client/domain.js";
import {querySafe} from "../../api/client/safe.js";

const safeOptions = ref([]);

const formTable = ref();
const listData = ref();
const dataLoading = ref(false);
const addVisible = ref(false);
const formState = reactive({
  address: "",
  domain: "",
  desc:"",
  id:""
})
const domainOptions = ref([]);

const pagination = reactive({
  total: 0,
  current: 1,
  pageSize: 10,
});


const loadSafes=()=>{
  querySafe({}).then(res => {
    const result = res.data.data;
    console.log(result)
    safeOptions.value=[]
    result.forEach(r => {
      safeOptions.value.push({
        value: r.id,
        label: r.ruleName,
      });
    });
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
  })
}

const removeData = (item) => {
  removeReverse({
    id: item.id
  }).then(res => {
    notification.open({
      message: res.msg,
    })
    loadData()
  })
}

const edit = (item) => {
  formState.domain = item.domain
  formState.address = item.address
  formState.desc = item.desc
  formState.safeId = item.safeId
  formState.id = item.id
  addVisible.value=true
}

const columns = [
  {title: '编号', dataIndex: 'id', key: 'id'},
  {title: '域名', dataIndex: 'domain', key: 'domain'},
  {title: '地址', dataIndex: 'address', key: 'address'},
  {title: '备注', dataIndex: 'desc', key: 'desc'},
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
  formState.domain = ""
  formState.address = ""
  formState.desc = ""
  formState.id = undefined
  addVisible.value = true
}

const addOk = () => {
  formTable.value.validate().then(res => {
    saveReverse({...formState}).then(res => {
      notification.open({
        message: res.msg,
      })
      loadData()
      addVisible.value = false
    })
  })
}

const loadDomains=()=>{
  queryDomain({}).then(res => {
    const result = res.data.data;
    domainOptions.value=[]
    console.log(result)
    result.forEach(r => {
      domainOptions.value.push({
        value: r.domain,
        label: r.domain,
      });
    });
  })
}

onMounted(() => {
  loadData()
})

</script>

<style scoped>

</style>
