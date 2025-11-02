<template>
  <div>
    <a-button class="btn edit" style="margin-bottom: 10px;" @click="addConfigModal">添加穿透</a-button>
    <a-button class="btn view" style="margin-bottom: 10px;margin-left: 10px" @click="loadData">刷新列表</a-button>
      <a-input v-model:value="pagination.keyword" allow-clear placeholder="关键字查询" style="width: 150px;margin-bottom: 10px;margin-left: 10px"/>
      <a-button class="btn view" style="margin-bottom: 10px;margin-left: 10px" type="primary" @click="loadData">查询</a-button>
    <a-table :loading="configLoading" :columns="columns" rowKey="id" :data-source="currentConfigList"
             :locale="{emptyText: '暂无配置,添加一个试试看看'}"
             :pagination="pagination"
             @change="handleTableChange"
             :scroll="{ x: 'max-content'}">
      <template #bodyCell="{ column ,record}">

        <template v-if="column.key === 'server'">
          <div v-for="(item,index) in openAddress(record)">
            <a-tag color="pink">{{item}}</a-tag>
          </div>
        </template>

        <template v-if="column.key==='status'">
          <a-switch :checked="!record.status||record.status==0"  @click="changeData(record)"/>
        </template>
        <template v-if="column.key==='tunType'">
          <div v-if="record.tunType&&record.tunType==='TCP'">
            TCP多路复用
          </div>
          <div v-else-if="record.tunType&&record.tunType==='QUIC'">
            QUIC多路复用
          </div>
          <div v-else>
            QUIC多路复用
          </div>
        </template>

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
          <a-button class="btn delete" style="margin-bottom: 5px;margin-left: 5px" @click="removeConfigData(record)">删除
          </a-button>
          <a-button class="btn edit" style="margin-bottom: 5px;margin-left: 5px" @click="editConfigData(record)">编辑
          </a-button>
          <a-button class="btn view" style="margin-bottom: 5px;margin-left: 5px" @click="refConfigData(record)">重连配置
          </a-button>
        </template>
      </template>
      <template #expandedRowRender="{ record }">
        <div class="text-detail">
          <div>备注：{{ record.remarks }}</div>
          <div v-if="record.statusMsg">
            最近一条穿透服务日志：<span class="text-tips">{{ record.statusMsg }}</span>
          </div>
        </div>
      </template>
    </a-table>


    <div>
      <a-modal v-model:visible="addConfigVisible" title="添加内网穿透配置">
        <div class="config-info">
        <a-form :model="formState" layout="vertical" ref="formTable">
          <a-form-item label="穿透设备" name="deviceKey" :rules="[{ required: true, message: '穿透设备必填'}]">
            <a-select
                v-model:value="formState.deviceKey"
                :options="currentUserKeyList"
            ></a-select>
          </a-form-item>

          <a-form-item label="穿透备注" name="remarks" :rules="[{ required: true, message: '穿透备注必填'}]">
            <a-input v-model:value="formState.remarks" allow-clear placeholder="备注如：个人博客"/>
          </a-form-item>

          <a-divider>端口映射配置</a-divider>

          <a-form-item label="外网端口" name="remotePort" :rules="[{ required: true, message: '外网端口必填'}]">
            <a-input v-model:value.number="formState.remotePort" allow-clear type="number" placeholder="8084"/>
          </a-form-item>

          <a-form-item label="内网地址" name="localAddress" :rules="[{ required: true, message: '内网地址必填'}]">






            <a-collapse>
              <a-collapse-panel  header="配置说明">
                <a-collapse accordion>
                  <a-collapse-panel key="1" header="HTTP/HTTPS协议">
                    <a-alert style="margin: 10px 5px" type="success" >
                      <template #message>
                        <div>
                          <a-tag color="pink">http://127.0.0.1</a-tag>
                          <a-tag color="red">https://127.0.0.1</a-tag>
                          <a-tag color="orange">http://127.0.0.1:8080</a-tag>
                          <a-tag color="green">https://192.168.55:8080</a-tag>
                          <a-tag color="cyan">http://192.168.15:8080</a-tag>
                          <p>http/https协议支持默认端口方式或者手动指定端口、当选择http协议时可以自由选择是否绑定域名</p>
                        </div>
                      </template>
                    </a-alert>
                  </a-collapse-panel>
                  <a-collapse-panel key="2" header="TCP协议">
                    <a-alert style="margin: 10px 5px" type="success" >
                      <template #message>
                        <a-tag color="pink">tcp://127.0.0.1:1080</a-tag>
                        <a-tag color="red">tcp://192.168.10.1:1080</a-tag>
                        <p>TCP级别协议、当选择tcp协议可以设置代理协议，通常情况下是不用设置，如果有获取真实IP或者对该协议熟悉的人可以选择设置</p>
                      </template>
                    </a-alert>
                  </a-collapse-panel>
                  <a-collapse-panel key="3" header="UDP协议">
                    <a-alert style="margin: 10px 5px" type="success" >
                      <template #message>
                        <a-tag color="pink">udp://127.0.0.1:1080</a-tag>
                        <a-tag color="red">udp://192.168.10.1:1080</a-tag>
                      </template>
                    </a-alert>
                  </a-collapse-panel>
                  <a-collapse-panel key="4" header="SOCKS5协议">
                    <a-alert style="margin: 10px 5px" type="success" >
                      <template #message>
                        <a-tag color="pink">socks5://127.0.0.1:1080</a-tag>
                        <a-tag color="red">socks5://用户名:密码@127.0.0.1:1080</a-tag>
                        <p>socks5协议会在内网创建创建一个socks5服务，然后暴露到公网 可以选择设置密码和不设置密码</p>
                      </template>
                    </a-alert>
                  </a-collapse-panel>
                  <a-collapse-panel key="5" header="UNIX协议">
                    <a-alert style="margin: 10px 5px" type="success" >
                      <template #message>
                        <a-tag color="pink">unix:///tmp/socks.sock</a-tag>
                        <a-tag color="red">unix:///tmp/****.sock</a-tag>
                        <p>unix协议是直接连接到文件上，请确保sock文件路径正确</p>
                      </template>
                    </a-alert>
                  </a-collapse-panel>
                </a-collapse>
              </a-collapse-panel>
            </a-collapse>


            <a-input style="margin-top: 5px" allow-clear v-model:value="formState.localAddress" placeholder="http://127.0.0.1:8084"/>

          </a-form-item>


          <a-form-item label="代理协议" name="proxyVersion" v-if="showInput.proxyVersion"
                       :rules="[{message: '用于获取真实IP，需要内网配合完成'}]">
            <a-select
                v-model:value="formState.proxyVersion"
            >
              <a-select-option value="NONE">不设置(小白用户请不要设置-用于TCP获取真实IP)</a-select-option>
              <a-select-option value="V1">TCP#V1版本</a-select-option>
              <a-select-option value="V2">TCP#V2版本</a-select-option>
            </a-select>
          </a-form-item>
          <a-form-item label="&nbsp;绑定访问域名&nbsp;&nbsp;" name="domain" v-if="showInput.domain">
            <a-select
                v-model:value="formState.domain"
                show-search
                placeholder="选择一个域名"
                :options="domainOptions"
            ></a-select>
          </a-form-item>

          <a-divider>其他选项配置</a-divider>

          <a-form-item label="配置有效" name="status"
                       :rules="[{ required: true, message: '当前配置是否有效'}]">
            <a-select
                v-model:value="formState.status"
            >
              <a-select-option value="0">有效</a-select-option>
              <a-select-option value="1">无效</a-select-option>
            </a-select>
          </a-form-item>

          <a-form-item label="隧道模式" name="tunType"
                       :rules="[{ required: true, message: '选择隧道模式'}]">
            <a-select
                v-model:value="formState.tunType"
            >
              <a-select-option value="TCP">TCP多路复用模式</a-select-option>
              <a-select-option value="QUIC">QUIC多路复用模式</a-select-option>
            </a-select>
          </a-form-item>
        </a-form>
        </div>
        <template #footer>
          <a-button class="btn view" @click="addConfigVisible=!addConfigVisible">取消</a-button>
          <a-button class="btn edit" @click="addConfigOk">确定</a-button>
        </template>

      </a-modal>
    </div>

  </div>
</template>

<script setup>
import {onMounted, reactive, ref, watch} from "vue";
import {removeConfig, getConfigList, getDeviceKey, addConfig, refConfig, changeStatus} from "../../api/client/config";
import {useRoute} from 'vue-router'
import userInfo from "../../data/userInfo";
import {queryDomain} from "../../api/client/domain.js";

const domainOptions = ref([]);

const route = useRoute()
const pagination = reactive({
  total: 0,
  current: 1,
  pageSize: 10,
  keyword:'',
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
  remotePort: undefined,
  domain: undefined,
  localAddress: "",
  proxyVersion: "NONE",
  status: '0',
  tunType:"TCP",
})


const currentConfigList = ref()

const currentUserKeyList = ref()

const changeData = (item)=>{
  configLoading.value = true
  changeStatus({
    configId: item.id
  }).then(res => {
    configLoading.value = false
    if (res.code===200) {
      if (!item.status||item.status==0){
        item.status=1
      }else {
        item.status=0
      }
    }
  })

}

const loadDomains=()=>{
  queryDomain({}).then(res => {
      const result = res.data.data;
      console.log(result)
    domainOptions.value=[]
      result.forEach(r => {
        domainOptions.value.push({
          value: r.domain,
          label: r.domain,
        });
      });
  })
}

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
  loadDomains()
  currentConfigList.value = []
  configLoading.value = true
  getConfigList(pagination).then(res => {
    configLoading.value = false
    currentConfigList.value = res.data.records
    pagination.total = res.data.total
  }).catch(e => {
    configLoading.value = false
  })
}

onMounted(() => {
  loadDeviceKey();
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
  formState.remotePort = item.remotePort
  formState.localAddress = item.localAddress
  formState.domain = item.domain
  formState.proxyVersion = item.proxyVersion
  if (!item.tunType){
    item.tunType="QUIC"
  }
  formState.tunType = item.tunType
  formState.status = item.status+""
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
  formState.remotePort = undefined
  formState.localAddress = ''
  formState.domain = undefined
  formState.proxyVersion = "NONE"
  formState.tunType='TCP'
  formState.status = "0"
  addConfigVisible.value = true;
};
const addConfigOk = () => {
  formTable.value.validate().then(res => {
    console.log("添加配置表单", formState)
    formState.status=parseInt(formState.status)
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
  {title: '隧道模式', dataIndex: 'tunType', key: 'tunType'},
  {title: '内网服务', dataIndex: 'localAddress', key: 'localAddress'},
  {title: '外网服务', dataIndex: 'server', key: 'server'},
  {title: '配置有效', dataIndex: 'status', key: 'status'},
  {title: '部署设备', dataIndex: 'deviceKey', key: 'deviceKey'},
  {title: '操作', key: 'action'},
];

const showInput=reactive({
  proxyVersion:false,
  domain:false,
})

watch(() => formState.localAddress, (newVal) => {
  if (newVal.startsWith("tcp")){
    showInput.proxyVersion=true
  }else {
    showInput.proxyVersion=false
  }
  if (newVal.startsWith("http")){
    showInput.domain=true
  }else {
    showInput.domain=false
  }
})

const openAddress = (item) => {
  const address=[]

  if (item.localAddress.startsWith("tcp")||item.localAddress.startsWith("unix")){
    address.push("tcp://"+item.serverIp+":"+item.remotePort)
  }

  if (item.localAddress.startsWith("udp")){
    address.push("udp://"+item.serverIp+":"+item.remotePort)
  }
  if (item.localAddress.startsWith("socks5")){
    address.push("socks5://"+item.serverIp+":"+item.remotePort)
  }

  if (item.localAddress.startsWith("http")){
    address.push("http://"+item.serverIp+":"+item.remotePort)
    if (item.domain){
    address.push("http://"+item.domain)
    address.push("https://"+item.domain)
    }
  }

  return address
}



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

.config-info{
  max-height: 60vh;
  overflow-y: scroll;

}
/* 滚动条整体样式 */
.config-info::-webkit-scrollbar {
  width: 0px; /* 滚动条宽度 */
  height: 0px;
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
