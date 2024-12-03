<template>
  <div>

    <div>
      <a-button type="primary" style="margin-bottom: 10px;margin-left: 10px" @click="addDeviceModal">添加设备</a-button>
      <a-button type="primary" style="margin-bottom: 10px;margin-left: 10px" @click="loadData">刷新列表</a-button>
    </div>

    <a-list :loading="listLoading" :locale="{emptyText: '暂无设备,请添加后在客户端'}" item-layout="horizontal"
            :data-source="deviceList">
      <template #renderItem="{ item }">
        <div style="padding: 30px" :class="[!item.online?'onRead':'']">
          <a-card :title="'备注：'+item.desc" :bordered="false">
            <p style="overflow-wrap: break-word;">设备ID：{{ item.deviceId }}</p>
            <p v-if="item.memoryInfo">总内存：{{ (item.memoryInfo.total / 1024 / 1024).toFixed(2) }}/MB</p>
            <p v-if="item.memoryInfo">使用内存：{{ (item.memoryInfo.useMem / 1024 / 1024).toFixed(2) }}/MB </p>
            <p v-if="item.memoryInfo">CPU：{{ (item.memoryInfo.cpuRate).toFixed(2) }}%</p>
            <p v-if="item.memoryInfo">HP占用内存：{{ (item.memoryInfo.hpTotalMem / 1024 / 1024).toFixed(2) }}/MB </p>
            <p v-if="item.memoryInfo">HP实际使用内存：{{ (item.memoryInfo.hpUseMem / 1024 / 1024).toFixed(2) }}/MB </p>
            <p>是否在线：{{ item.online ? "在线中" : "未在线" }}</p>

            <div v-if="userInfo.getUserInfo().role==='ADMIN'">
              <p v-if="item.username">归属用户：{{item.username}}</p>
              <p v-if="item.userDesc">归属用户备注：{{item.userDesc}}</p>
            </div>

            <div class="op-btn">
              <a-popconfirm
                  title="你真的要删除？"
                  ok-text="是的我删除"
                  cancel-text="点错了"
                  @confirm="()=>{removeData(item)}"
              >
                <a-button type="primary">删除</a-button>
              </a-popconfirm>
              <a-popconfirm
                  title="你真的要强制停止程序？"
                  ok-text="是的我要停止"
                  cancel-text="点错了"
                  @confirm="()=>{stopData(item)}"
              >
                <a-button type="danger">强制停止</a-button>
              </a-popconfirm>
              <a-button type="dashed" @click="showQr(item)">查看设备码</a-button>
              <a-button type="dashed" @click="edit(item)">编辑</a-button>
              </div>
          </a-card>
        </div>
      </template>
    </a-list>


    <div>
      <a-modal :destroyOnClose="true" cancelText="取消" okText="我已知晓" v-model:visible="qrModalVisible" @ok="closeQr" title="设备二维码">
        <qr :text="deviceId"/>
        <template slot="footer">
          <a-button @click="qrModalVisible=false"></a-button>
        </template>
      </a-modal>
    </div>


    <div>
      <a-modal okText="确定" cancelText="取消" v-model:visible="addDeviceModalVisible" title="设备信息"
               @ok="addDeviceOk">
        <a-form :model="formState" ref="formTable" :layout="'vertical'" >
          <a-form-item label="设备编号" name="deviceId" :rules="[{ required: true, message: '设备编号必填'}]">
            <a-input style="width: 70%" v-model:value="formState.deviceId" placeholder="设备ID：32位"/>
            <span style="padding-left: 8px;user-select: none"><a @click="guid">自动生成</a></span>
          </a-form-item>
          <a-form-item label="设备备注" name="desc" :rules="[{ required: true, message: '设备备注必填'}]">
            <a-input style="width: 70%" v-model:value="formState.desc" placeholder="备注如：nas中的HP"/>
          </a-form-item>
        </a-form>
      </a-modal>
    </div>
    <div>
      <a-modal okText="确定" cancelText="取消" v-model:visible="updateDeviceModalVisible" title="设备信息"
               @ok="updateDeviceOk">
        <a-form :model="formState" ref="formTable" :layout="'vertical'" >
          <a-form-item label="设备编号"  name="deviceId" :rules="[{ required: true, message: '设备编号必填'}]">
            <a-input style="width: 70%" disabled="disabled" v-model:value="formState.deviceId" placeholder="设备ID：32位"/>
          </a-form-item>
          <a-form-item label="设备备注" name="desc" :rules="[{ required: true, message: '设备备注必填'}]">
            <a-input style="width: 70%" v-model:value="formState.desc" placeholder="备注如：nas中的HP"/>
          </a-form-item>
        </a-form>
      </a-modal>
    </div>

  </div>
</template>

<script setup>
import {onMounted, reactive, ref} from "vue";
import {useRouter} from "vue-router";
import {getDeviceList, addDevice, removeDevice, stopDevice, updateDevice} from "../../api/client/device";
import qr from './qr.vue';
import userInfo from "../../data/userInfo";
const qrModalVisible = ref(false)
const router = useRouter()
const formTable = ref()
const deviceId = ref()
const deviceList = ref()
const listLoading = ref(false)
const addDeviceModalVisible = ref(false)
const updateDeviceModalVisible = ref(false)
const formState = reactive({
  deviceId: "",
  desc: ""
})

const addDeviceModal = () => {
  formState.deviceId=""
  formState.desc=""
  addDeviceModalVisible.value = true;
};


const edit=(item)=>{
  formState.deviceId=item.deviceId
  formState.desc=item.desc
  updateDeviceModalVisible.value = true;
}


const showQr = (item)=>{
  qrModalVisible.value = true;
  deviceId.value = item.deviceId;
}
const closeQr = ()=>{
  qrModalVisible.value = false;
  deviceId.value = "";
}


const addDeviceOk = () => {
  formTable.value.validate().then(res => {
    addDevice({
      ...formState
    }).then(res => {
      formState.deviceId = ''
      formState.desc = ''
      loadData();
      addDeviceModalVisible.value = false;
    })
  })
};

const updateDeviceOk = () => {
  formTable.value.validate().then(res => {
    updateDevice({
      ...formState
    }).then(res => {
      formState.deviceId = ''
      formState.desc = ''
      loadData();
      updateDeviceModalVisible.value = false;
    })
  })
};
const stopData = (item) => {
  stopDevice({
    deviceId: item.deviceId
  }).then(res => {
    loadData();
  })
};

const loadData = () => {
  listLoading.value = true
  getDeviceList().then(res => {
    listLoading.value = false
    if (res.data){
      deviceList.value = res.data
    }
  }).catch(e => {
    listLoading.value = false
  })
}

const guid = () => {
  let uid = 'xxxxxxxxxxxx4xxxyxxxxxxxxxxxxxxx'.replace(/[xy]/g, function (c) {
    let r = Math.random() * 16 | 0,
        v = c === 'x' ? r : (r & 0x3 | 0x8);
    return v.toString(16);
  });

  formState.deviceId = uid

}

onMounted(() => {
  loadData()
})

const removeData = (item) => {
  removeDevice({
    deviceId: item.deviceId
  }).then(res => {
    loadData();
  })
};


</script>

<style lang="less" scoped>

.onRead {
  div {
    -webkit-filter: grayscale(100%);
    filter: progid:DXImageTransform.Microsoft.BasicImage(graysale=1);
  }
}

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
