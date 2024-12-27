<template>
  <div>
    <a-button type="primary" style="margin-bottom: 10px;margin-left: 10px" @click="addConfigModal">添加穿透</a-button>
    <a-button type="primary" style="margin-bottom: 10px;margin-left: 10px" @click="loadData">刷新列表</a-button>
    <a-table :loading="configLoading" :columns="columns" rowKey="id" :data-source="currentConfigList"
             :locale="{emptyText: '暂无配置,添加一个试试看看'}"
             :pagination="pagination"
             @change="handleTableChange"
             :scroll="{ x: 10 }">
      <template #bodyCell="{ column ,record}">
        <template v-if="column.key === 'deviceKey'">
          <div>
            设备状态：{{ userKeyByName(record.deviceKey) }}
          </div>
          <div v-if="userInfo.getUserInfo().role==='ADMIN'">
            <div v-if="userKeyByUserInfo(record.deviceKey).username"> 归属用户：
              {{ userKeyByUserInfo(record.deviceKey).username }}
            </div>
            <div v-if="userKeyByUserInfo(record.deviceKey).userDesc">
              归属用户备注：{{ userKeyByUserInfo(record.deviceKey).userDesc }}
            </div>
          </div>
        </template>
        <template v-if="column.key === 'action'">
          <a-button type="primary" style="margin-bottom: 5px;margin-left: 5px" @click="removeConfigData(record)">删除
          </a-button>
          <a-button type="primary" style="margin-bottom: 5px;margin-left: 5px" @click="editConfigData(record)">编辑
          </a-button>
          <a-button type="primary" style="margin-bottom: 5px;margin-left: 5px" @click="refConfigData(record)">重连配置
          </a-button>
        </template>
      </template>
      <template #expandedRowRender="{ record }">

        <div class="text-detail">
          <div>备注：{{ record.remarks }}</div>


          <div v-if="record.connectType==='TCP'">
            <div v-if="record.port">外网TCP地址: <span class="text-tips">{{ record.serverIp }}:{{ record.port }}</span>
            </div>
            <div v-if="record.domain">外网HTTP地址: <span class="text-tips">http(s)://{{ record.domain }}</span></div>
          </div>

          <div v-if="record.connectType==='UDP'">
            <div v-if="record.port">外网UDP地址: <span class="text-tips">{{ record.serverIp }}:{{ record.port }}</span>
            </div>
          </div>

          <div v-if="record.connectType==='TCP_UDP'">
            <div v-if="record.port">外网UDP地址: <span class="text-tips">{{ record.serverIp }}:{{ record.port }}</span>
            </div>
            <div v-if="record.port">外网TCP地址: <span class="text-tips">{{ record.serverIp }}:{{ record.port }}</span>
            </div>
            <div v-if="record.domain">外网HTTP地址: <span class="text-tips">http(s)://{{ record.domain }}</span></div>
          </div>

          <div v-if="record.statusMsg">
            最近一条穿透服务日志：<span class="text-tips">{{ record.statusMsg }}</span>
          </div>

          <div v-if="!record.port">
            随机端口不支持TCP和UDP协议使用
          </div>
        </div>
      </template>
    </a-table>


    <div>
      <a-modal okText="确定" cancelText="取消" v-model:visible="addConfigVisible" title="添加内网穿透配置"
               @ok="addConfigOk">
        <a-form :model="formState" ref="formTable">
          <a-form-item label="穿透设备" name="deviceKey" :rules="[{ required: true, message: '穿透设备必填'}]">
            <a-select
                v-model:value="formState.deviceKey"
                :options="currentUserKeyList"
            ></a-select>
          </a-form-item>

          <a-form-item label="穿透备注" name="remarks" :rules="[{ required: true, message: '穿透备注必填'}]">
            <a-input v-model:value="formState.remarks" placeholder="备注如：个人博客"/>
          </a-form-item>
          <a-form-item label="穿透协议" name="connectType" :rules="[{ required: true, message: '穿透协议必填'}]">
            <a-select
                v-model:value="formState.connectType"
            >
              <a-select-option value="TCP">TCP</a-select-option>
              <a-select-option value="UDP">UDP</a-select-option>
              <a-select-option value="TCP_UDP">TCP_UDP</a-select-option>
            </a-select>
          </a-form-item>
          <!--    套餐选择      -->
          <a-form-item label="外网端口" name="port" :rules="[{ required: true, message: '外网端口必填'}]">
            <a-input v-model:value.number="formState.port" type="number" placeholder="8084"/>
          </a-form-item>

          <a-form-item label="内网地址" name="localIp" :rules="[{ required: true, message: '内网地址必填'}]">
            <a-input v-model:value="formState.localIp" placeholder="内网IP如：127.0.0.1"/>
          </a-form-item>

          <a-form-item label="内网端口" name="localPort" :rules="[{ required: true, message: '内网端口必填'}]">
            <a-input v-model:value.number="formState.localPort" type="number" placeholder="内网端口如：8080"/>
          </a-form-item>

          <a-form-item label="代理协议" name="proxyVersion"
                       :rules="[{ required: true, message: '用于获取真实IP，需要内网配合完成'}]">
            <a-select
                v-model:value="formState.proxyVersion"
            >
              <a-select-option value="NONE">不设置(小白用户请不要设置-用于获取真实IP)</a-select-option>
              <a-select-option value="V1">TCP#V1版本</a-select-option>
              <a-select-option value="V2">TCP#V2版本</a-select-option>
              <a-select-option value="V3">HTTP#X-Forwarded-For</a-select-option>
            </a-select>
          </a-form-item>

          <a-form-item label="&nbsp;穿透域名&nbsp;&nbsp;" name="domain">
            <!--            <a-input v-model:value="formState.domain" placeholder="xxx.com"/>-->
            <a-select
                v-model:value="formState.domain"
                show-search
                placeholder="选择一个域名"
                :options="domainOptions"
                @change="handleChange"
            ></a-select>

          </a-form-item>

          <a-form-item label="&nbsp;证书KEY&nbsp;&nbsp;" name="certificateKey"
                       :rules="[{ required: false, message: '必须填写证书.key文件'}]">
            <a-textarea  disabled="disabled" :rows="6" v-model:value="formState.certificateKey"
                        placeholder="-----BEGIN RSA PRIVATE KEY-----&#10;***大概是这样的证书私钥***&#10;-----END RSA PRIVATE KEY-----"/>
          </a-form-item>
          <a-form-item  label="&nbsp;证书内容&nbsp;&nbsp;" name="certificateContent"
                       :rules="[{ required: false, message: '映射描述必填'}]">
            <a-textarea disabled="disabled" :rows="6" v-model:value="formState.certificateContent"
                        placeholder="-----BEGIN CERTIFICATE-----&#10;***大概是这样的证书内容***&#10;-----BEGIN CERTIFICATE-----"/>
          </a-form-item>
        </a-form>
      </a-modal>
    </div>

  </div>
</template>

<script setup>
import {onMounted, reactive, ref} from "vue";
import {removeConfig, getConfigList, getDeviceKey, addConfig, refConfig} from "../../api/client/config";
import {useRoute} from 'vue-router'
import userInfo from "../../data/userInfo";
import {queryDomain} from "../../api/client/domain.js";

const domainOptions = ref([]);

const route = useRoute()
const pagination = reactive({
  total: 0,
  current: 1,
  pageSize: 10,
});

const handleTableChange = (item) => {
  pagination.current = item.current
  pagination.pageSize = item.pageSize
  pagination.total = item.total
  loadData()
}


const formTable = ref()
const addConfigVisible = ref(false)
const configLoading = ref(false)

const formState = reactive({
  id: undefined,
  deviceKey: "",
  remarks: "",
  port: undefined,
  domain: undefined,
  localIp: "",
  localPort: undefined,
  connectType: "",
  proxyVersion: "",
  certificateKey: "",
  certificateContent: "",
})


const currentConfigList = ref()

const currentUserKeyList = ref()


const loadDomains=()=>{
  queryDomain({}).then(res => {
      const result = res.data.data;
      console.log(result)
      result.forEach(r => {
        domainOptions.value.push({
          value: r.domain,
          label: r.domain,
          certificateContent: r.certificateContent,
          certificateKey: r.certificateKey,
        });
      });
  })
}


const handleChange = val => {
  const data = domainOptions.value.filter((item)=>{return item.value===val})
  console.log(data)
  if (data.length===1){
    formState.certificateKey=data[0].certificateKey.trim()
    formState.certificateContent=data[0].certificateContent.trim()
  }
};


const loadDeviceKey = () => {
  getDeviceKey().then(res => {
    let data = []
    for (let k of res.data) {
      data.push({"label": k.desc, "value": k.key, userDesc: k.userDesc, username: k.username})
    }
    currentUserKeyList.value = data
  })
}

const userKeyByName = (deviceKey) => {
  try {
    return currentUserKeyList.value.filter(r => {
      return r.value === deviceKey
    })[0].label
  } catch (e) {
    return "设备获取错误"
  }
}

const userKeyByUserInfo = (deviceKey) => {
  try {
    return currentUserKeyList.value.filter(r => {
      return r.value === deviceKey
    })[0]
  } catch (e) {
    return ""
  }
}


const loadData = () => {
  currentConfigList.value = []
  configLoading.value = true
  getConfigList(pagination).then(res => {
    configLoading.value = false
    currentConfigList.value = res.data.records
    pagination.current = res.data.current
    pagination.total = res.data.total
  }).catch(e => {
    configLoading.value = false
  })
}

onMounted(() => {
  loadDeviceKey();
  loadDomains();
  loadData()
})


const removeConfigData = (item) => {
  removeConfig({
    configId: item.id
  }).then(res => {
    if (res.data) {
      loadData()
    }
  })
}

const editConfigData = (item) => {
  formState.id = item.id
  formState.deviceKey = item.deviceKey
  formState.remarks = item.remarks
  formState.port = item.port
  formState.domain = item.domain
  formState.localIp = item.localIp
  formState.localPort = item.localPort
  formState.connectType = item.connectType
  formState.proxyVersion = item.proxyVersion
  formState.certificateKey = item.certificateKey
  formState.certificateContent = item.certificateContent
  addConfigVisible.value = true;
}
const refConfigData = (item) => {
  configLoading.value = true
  refConfig({
    configId: item.id
  }).then(res => {
    configLoading.value = false
    if (res.data) {
      loadData()
    }
  })
}


const addConfigModal = () => {
  formState.id = undefined
  formState.deviceKey = ""
  formState.remarks = ""
  formState.port = undefined
  formState.domain = undefined
  formState.localIp = ""
  formState.localPort = undefined
  formState.connectType = ""
  formState.proxyVersion = ""
  formState.certificateKey = ""
  formState.certificateContent = ""
  addConfigVisible.value = true;
};
const addConfigOk = () => {
  formTable.value.validate().then(res => {
    console.log("添加配置表单", formState)
    addConfig(
        {
          packageId: route.query.packageId,
          ...formState
        }
    ).then(res => {
      loadData()
      addConfigVisible.value = false;
    })
  })
};


const columns = [
  {title: '配置ID', dataIndex: 'id', key: 'id'},
  {title: '备注', dataIndex: 'remarks', key: 'remarks'},
  {title: '内网IP', dataIndex: 'localIp', key: 'localIp'},
  {title: '内网端口', dataIndex: 'localPort', key: 'localPort'},
  {title: '外网端口', dataIndex: 'port', key: 'port'},
  {title: '穿透类型', dataIndex: 'connectType', key: 'connectType'},
  {title: '域名', dataIndex: 'domain', key: 'domain'},
  {title: '部署设备', dataIndex: 'deviceKey', key: 'deviceKey'},
  {title: '操作', key: 'action'},
];

</script>

<style lang="less">

.op-btn button {
  text-align: center;
  margin: 5px;
}

.text-tips {
  margin-top: 10px;
  background-color: #4b6ff6;
  color: #ffffff;
  padding: 2px 10px;
  border-radius: 10px;
}

.text-detail div {
  margin-bottom: 10px;
}

.ant-card-body {
  overflow: hidden;
}

.full-modal {
  .ant-modal {
    max-width: 100%;
    top: 0;
    padding-bottom: 0;
    margin: 0;
  }

  .ant-modal-content {
    box-shadow: none;
    display: flex;
    flex-direction: column;
    height: calc(100vh);
  }

  .ant-modal-body {
    flex: 1;
  }
}

</style>
