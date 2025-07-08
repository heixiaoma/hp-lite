<template>
  <div>
    <a-button  style="margin-bottom: 10px" class="btn edit" @click="addModal">添加规则</a-button>
    <a-button class="btn view" style="margin-bottom: 10px;margin-left: 5px" @click="loadData">刷新列表</a-button>

      <a-table :loading="dataLoading" :columns="columns" rowKey="id" :data-source="listData"
               :locale="{emptyText: '暂无数据,添加一个试试看看'}"
               :pagination="pagination"
               @change="handleTableChange"
               :scroll="{ x: 'max-content' }">

        <template #bodyCell="{ column ,record}">
          <template v-if="column.key === 'allowedIps'">
            <template v-if="column.key === 'allowedIps'">
              <div v-if="record.allowedIps.length>0" v-for="(item,index) in record.allowedIps">
                <a-tag color="#87d068">{{item}}</a-tag>
              </div>
              <div v-else>
                <a-tag color="#87d068">未启用</a-tag>
              </div>
            </template>
          </template>
          <template v-if="column.key === 'blockedIps'">
            <div v-if="record.blockedIps.length>0" v-for="(item,index) in record.blockedIps">
              <a-tag color="#f50">{{item}}</a-tag>
            </div>
            <div v-else>
              <a-tag color="#f50">未启用</a-tag>
            </div>
          </template>
          <template v-if="column.key === 'action'">
            <a-button  class="btn edit" style="margin-bottom: 5px;margin-left: 5px" @click="edit(record)">编辑</a-button>
            <a-button class="btn view" style="margin-bottom: 5px;margin-left: 5px" @click="refConfigData(record)">刷新规则</a-button>
            <a-button  class="btn delete" style="margin-bottom: 5px;margin-left: 5px" @click="removeData(record)">删除</a-button>
          </template>
        </template>
      </a-table>
  </div>

  <div>
    <a-modal  v-model:visible="addVisible" title="添加">
      <div class="config-info">
      <a-form :model="formState" ref="formTable" :layout="'vertical'" >
        <a-form-item label="穿透配置 ">
          <a-select
              v-model:value="formState.configId"
              show-search
              placeholder="穿透配置备注关键字"
              style="width: 90%"
              :default-active-first-option="false"
              :show-arrow="false"
              :filter-option="false"
              :not-found-content="null"
              :options="data"
              @search="handleSearch"
              @change="handleChange"
          ></a-select>
        </a-form-item>
        <a-form-item label="并发限制">

          <a-input style="width: 90%" v-model:value="formState.rateLimit" placeholder="每分钟最大连接数(-1不限制)"/>
        </a-form-item>
        <a-form-item label="上传限制(字节)">
          <a-input style="width: 90%" v-model:value="formState.inLimit" placeholder="上传限制字节单位(-1不限制)"/>
        </a-form-item>
        <a-form-item label="下载限制(字节)">
          <a-input style="width: 90%" v-model:value="formState.outLimit" placeholder="下载限制字节单位(-1不限制)"/>
        </a-form-item>

        <div v-if="formState.blockedIps.length===0">

        <a-form-item
            label="允许IP(CIDR地址)"
            v-for="(ip, index) in formState.allowedIps"
        >
          <a-input
              v-model:value="formState.allowedIps[index]"
              placeholder="请输入IP规则:0.0.0.0/0"
              style="width: 90%; margin-right: 5px"
          />
          <MinusCircleOutlined
              v-if="formState.allowedIps.length > 0"
              :disabled="formState.allowedIps.length === 1"
              @click="removeAllowedIps(ip)"
          />
        </a-form-item>
        <a-form-item >
          <a-button type="dashed" style="width: 90%" @click="addAllowedIps">
            <PlusOutlined />
            添加一行允许IP规则
          </a-button>
        </a-form-item>
        </div>

        <div v-if="formState.allowedIps.length===0">
        <a-form-item
            label="禁止IP(CIDR地址)"
            v-for="(ip, index) in formState.blockedIps"
        >
          <a-input
              v-model:value="formState.blockedIps[index]"
              placeholder="请输入IP规则:127.0.0.1/0"
              style="width: 90%; margin-right: 5px"
          />
          <MinusCircleOutlined
              v-if="formState.blockedIps.length > 0"
              :disabled="formState.blockedIps.length === 1"
              @click="removeBlockedIps(ip)"
          />
        </a-form-item>
        <a-form-item >
          <a-button type="dashed" style="width: 90%" @click="addBlockedIps">
            <PlusOutlined />
            添加一行禁用IP规则
          </a-button>
        </a-form-item>
        </div>
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
import {getWaf, removeWaf, saveWaf} from "../../api/client/waf";
import {onMounted, reactive, ref} from "vue";
import {notification} from "ant-design-vue";
import { MinusCircleOutlined, PlusOutlined } from '@ant-design/icons-vue';
import {getConfigByKeyword, refConfig} from "../../api/client/config.js";

const listData = ref();
const dataLoading = ref(false);
const addVisible = ref(false);
const formState = reactive({
  configId: "",
  allowedIps: [""],
  blockedIps:[""],
  rateLimit:"",
  outLimit:"",
  inLimit:"",
  id:""
})
const pagination = reactive({
  total: 0,
  current: 1,
  pageSize: 10,
});

const loadData = () => {
  dataLoading.value = true
  getWaf({
    current: pagination.current,
    pageSize: pagination.pageSize
  }).then(res => {
    dataLoading.value = false
    listData.value = res.data.records
    pagination.total = res.data.total
  })
}

const removeData = (item) => {
  removeWaf({
    id: item.id
  }).then(res => {
    notification.open({
      message: res.msg,
    })
    loadData()
  })
}

const refConfigData = (item) => {
  refConfig({
    configId: item.configId
  }).then(res => {
    if (res.data) {
      loadData()
    }
  })
}


const edit = (itemOld) => {
  const item=JSON.parse(JSON.stringify(itemOld))
  formState.allowedIps = item.allowedIps
  formState.blockedIps = item.blockedIps
  formState.rateLimit = item.rateLimit
  formState.inLimit = item.inLimit
  formState.outLimit = item.outLimit
  formState.id = item.id
  formState.configId = item.configId
  addVisible.value=true
  handleSearch(item.configId)
}

const columns = [
  {title: '编号', dataIndex: 'id', key: 'id'},
  {title: '配置ID', dataIndex: 'configId', key: 'configId'},
  {title: '配置描述', dataIndex: 'configDesc', key: 'configDesc'},
  {title: '允许IP', dataIndex: 'allowedIps', key: 'allowedIps'},
  {title: '禁止IP', dataIndex: 'blockedIps', key: 'blockedIps'},
  {title: '上传速率(byte)', dataIndex: 'inLimit', key: 'inLimit'},
  {title: '下载速率(byte)', dataIndex: 'outLimit', key: 'outLimit'},
  {title: '并发限制', dataIndex: 'rateLimit', key: 'rateLimit'},
  {title: '操作', key: 'action'},
];

const handleTableChange = (item) => {
  pagination.current = item.current
  pagination.pageSize = item.pageSize
  pagination.total = item.total
  loadData()
}

const addModal = () => {
  formState.allowedIps = []
  formState.blockedIps = []
  formState.rateLimit = -1
  formState.inLimit = -1
  formState.outLimit = -1
  formState.configId = 0
  formState.id = undefined
  addVisible.value = true
}

const addOk = () => {
  formState.rateLimit=parseInt(formState.rateLimit)
  formState.inLimit=parseInt(formState.inLimit)
  formState.outLimit=parseInt(formState.outLimit)
  formState.configId=parseInt(formState.configId)

  saveWaf({...formState}).then(res => {
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
const addAllowedIps = () => {
  formState.allowedIps.push("");
};

const removeAllowedIps = (item) => {
    let index = formState.allowedIps.indexOf(item);
    if (index !== -1) {
      formState.allowedIps.splice(index, 1);
    }
};
const addBlockedIps = () => {
  formState.blockedIps.push("");
};

const removeBlockedIps = (item) => {
    let index = formState.blockedIps.indexOf(item);
    if (index !== -1) {
      formState.blockedIps.splice(index, 1);
    }
};


const data = ref([]);
const value = ref();
const handleSearch = val => {
  fetch(val, d => data.value = d);
};
const handleChange = val => {
  console.log(val);
  value.value = val;
  fetch(val, d => data.value = d);
};

let timeout;
let currentValue = '';
function fetch(value, callback) {
  if (timeout) {
    clearTimeout(timeout);
    timeout = null;
  }
  currentValue = value;
  function fake() {
    getConfigByKeyword(value).then(res=>{
      if (currentValue === value) {
        console.log(res)
        if (res.data){
          const data = [];
          res.data.forEach(r => {
            data.push({
              value: r.id,
              label: r.remarks,
            });
          });
          callback(data);
        }
      }
    })
  }
  timeout = setTimeout(fake, 300);
}


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
