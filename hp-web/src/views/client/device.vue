<template>
  <div>

    <div>
      <a-button class="btn edit" style="margin-bottom: 10px;" @click="addDeviceModal">添加设备</a-button>
      <a-button class="btn view" style="margin-bottom: 10px;margin-left: 10px" @click="loadData">刷新列表</a-button>
    </div>

    <a-table :columns="columns" :data-source="deviceList"  rowKey="deviceId" :loading="listLoading"
             :locale="{emptyText: '暂无数据,添加一个试试看看'}"
             :pagination="pagination"
             @change="handleTableChange"
             :scroll="{ x: 'max-content' }"
    >
      <template #bodyCell="{ column,record }">
        <template v-if="column.key === 'online'">
          <div class="panel-header">
            <div class="device-info">
              <div class="device-status">
                <span class="status-dot" :class="record.online ? 'online' : 'offline'"></span>
                <span class="status-text" :class="record.online ? 'online' : 'offline'">
                    {{ record.online ? '在线中' : '未在线' }}
                  </span>
              </div>
            </div>
          </div>
        </template>

        <template v-if="column.key === 'action'">

          <a-button class="btn view" style="margin-right: 5px" @click="showQr(record)">
            连接码
          </a-button>

          <a-button class="btn edit" @click="edit(record)">
            编辑
          </a-button>
        </template>

      </template>

      <template #expandedRowRender="{ record }">
        <div class="panel-content">
          <div class="device-details">
            <!-- 基础信息 -->
            <div class="info-section">
              <h3 class="section-title">设备信息</h3>
              <div class="info-row" style="flex-direction: column">
                <span class="info-label">设备ID:</span>
                <span class="info-value">{{ record.deviceId }}</span>
              </div>
              <div class="info-row" style="flex-direction: column">
                <span class="info-label">连接码:</span>
                <span class="info-value">{{ record.connectKey }}</span>
              </div>
            </div>

            <!-- 性能监控 -->
            <div v-if="record.memoryInfo" class="performance-section">
              <h3 class="section-title">性能监控</h3>

              <!-- 内存使用率 -->
              <div class="metric-item">
                <div class="metric-header">
                  <span class="metric-label">内存使用率</span>
                  <span class="metric-value">
                      {{ ((record.memoryInfo.useMem / record.memoryInfo.total) * 100).toFixed(1) }}%
                    </span>
                </div>
                <div class="progress-bar">
                  <div class="progress-track">
                    <div
                        class="progress-fill memory"
                        :style="{ width: ((record.memoryInfo.useMem / record.memoryInfo.total) * 100) + '%' }"
                    ></div>
                  </div>
                </div>
                <div class="memory-details">
                  <span class="memory-used">{{ (record.memoryInfo.useMem / 1024 / 1024).toFixed(2) }} MB</span>
                  <span class="memory-total">{{ (record.memoryInfo.total / 1024 / 1024).toFixed(2) }} MB</span>
                </div>
              </div>

              <!-- CPU使用率 -->
              <div class="metric-item">
                <div class="metric-header">
                  <span class="metric-label">CPU使用率</span>
                  <span class="metric-value">{{ record.memoryInfo.cpuRate.toFixed(1) }}%</span>
                </div>
                <div class="progress-bar">
                  <div class="progress-track">
                    <div
                        class="progress-fill cpu"
                        :style="{ width: record.memoryInfo.cpuRate + '%' }"
                    ></div>
                  </div>
                </div>
              </div>

              <!-- HP内存信息 -->
              <div class="hp-memory-info">
                <div class="info-row">
                  <span class="info-label">HP占用内存:</span>
                  <span class="info-value">{{ (record.memoryInfo.hpTotalMem / 1024 / 1024).toFixed(2) }} MB</span>
                </div>
                <div class="info-row">
                  <span class="info-label">HP实际使用:</span>
                  <span class="info-value">{{ (record.memoryInfo.hpUseMem / 1024 / 1024).toFixed(2) }} MB</span>
                </div>
              </div>
            </div>

            <!-- 管理员信息 -->
            <div v-if="userInfo.getUserInfo().role === 'ADMIN'" class="admin-section">
              <h3 class="section-title">归属信息</h3>
              <div class="info-row">
                <span class="info-label">归属用户:</span>
                <span class="info-value">{{ record.username || '无' }}</span>
              </div>
              <div class="info-row">
                <span class="info-label">用户备注:</span>
                <span class="info-value">{{ record.userDesc || '无' }}</span>
              </div>
            </div>
          </div>

          <!-- 操作按钮 -->
          <div class="action-buttons">
            <a-popconfirm
                title="确定要删除该设备？"
                ok-text="删除"
                cancel-text="取消"
                @confirm="() => removeData(record)"
            >
              <a-button class="btn delete">
                删除
              </a-button>
            </a-popconfirm>

            <a-popconfirm
                title="确定要强制停止程序？"
                ok-text="停止"
                cancel-text="取消"
                @confirm="() => stopData(record)"
            >
              <a-button class="btn stop">
                强制停止
              </a-button>
            </a-popconfirm>

          </div>
        </div>
      </template>
    </a-table>

    <div>
      <a-modal :destroyOnClose="true"  v-model:visible="qrModalVisible" title="设备二维码">
        <qr :text="deviceId"/>
        <template #footer>
          <a-button class="btn view" @click="closeQr">我已知晓</a-button>
        </template>
      </a-modal>
    </div>
    <div>
      <a-modal  v-model:visible="addDeviceModalVisible" title="设备信息">
        <a-form :model="formState" ref="formTable" :layout="'vertical'" >
          <a-form-item label="设备编号" name="deviceId" :rules="[{ required: true, message: '设备编号必填'}]">
            <a-input style="width: 70%" v-model:value="formState.deviceId" placeholder="设备ID：32位"/>
            <span style="padding-left: 8px;user-select: none"><a @click="guid">自动生成</a></span>
          </a-form-item>
          <a-form-item label="设备备注" name="desc" :rules="[{ required: true, message: '设备备注必填'}]">
            <a-input style="width: 70%" v-model:value="formState.desc" placeholder="备注如：nas中的HP"/>
          </a-form-item>
        </a-form>
        <template #footer>
          <a-button class="btn view" @click="addDeviceModalVisible=!addDeviceModalVisible">取消</a-button>
          <a-button class="btn edit" @click="addDeviceOk">确定</a-button>
        </template>
      </a-modal>
    </div>
    <div>
      <a-modal v-model:visible="updateDeviceModalVisible" title="设备信息">
        <a-form :model="formState" ref="formTable" :layout="'vertical'" >
          <a-form-item label="设备编号"  name="deviceId" :rules="[{ required: true, message: '设备编号必填'}]">
            <a-input style="width: 70%" disabled="disabled" v-model:value="formState.deviceId" placeholder="设备ID：32位"/>
          </a-form-item>
          <a-form-item label="设备备注" name="desc" :rules="[{ required: true, message: '设备备注必填'}]">
            <a-input style="width: 70%" v-model:value="formState.desc" placeholder="备注如：nas中的HP"/>
          </a-form-item>
        </a-form>

        <template #footer>
          <a-button class="btn view" @click="updateDeviceModalVisible=!updateDeviceModalVisible">取消</a-button>
          <a-button class="btn edit" @click="updateDeviceOk">确定</a-button>
        </template>

      </a-modal>
    </div>
  </div>
</template>

<script setup>
import {onMounted, reactive, ref} from "vue";
import {useRouter} from "vue-router";
import {addDevice, getDeviceList, removeDevice, stopDevice, updateDevice} from "../../api/client/device";
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
const pagination = reactive({
  total: 0,
  current: 1,
  pageSize: 10,
});

const columns = [
  {title: '描述', dataIndex: 'desc', key: 'desc'},
  {title: '在线状态', dataIndex: 'online', key: 'online'},
  {title: '操作', dataIndex: 'action',key: 'action'},
];

const handleTableChange = (item) => {
  pagination.current = item.current
  pagination.pageSize = item.pageSize
  pagination.total = item.total
  loadData()
}

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
  deviceId.value = item.connectKey;
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
  getDeviceList(pagination).then(res => {
    listLoading.value = false
    if (res.data){
      deviceList.value = res.data.records
      pagination.total = res.data.total
    }
  }).catch(e => {
    listLoading.value = false
  })
}

const guid = () => {
  formState.deviceId = 'xxxxxxxxxxxx4xxxyxxxxxxxxxxxxxxx'.replace(/[xy]/g, function (c) {
    let r = Math.random() * 16 | 0,
        v = c === 'x' ? r : (r & 0x3 | 0x8);
    return v.toString(16);
  })

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

/* 整体容器 */
.device-list-container {
  padding: 20px;
  background-color: #f8fafc;
  min-height: 100vh;
}

.ant-collapse{
  border: none;
}

/* 折叠面板样式 */
.custom-collapse .ant-collapse-item {
  margin-bottom: 16px;
  border-radius: 10px;
  overflow: hidden;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  border: none;
}

/* 面板头部 */
.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.device-info {
  display: flex;
  flex-direction: column;
}

.device-name {
  font-size: 16px;
  font-weight: 500;
  color: #1f2937;
  margin-bottom: 4px;
}

.device-status {
  display: flex;
  align-items: center;
}

.status-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  margin-right: 6px;
}

.status-dot.online {
  background-color: #10b981;
  animation: pulse 2s infinite;
}

.status-dot.offline {
  background-color: #ef4444;
}

.status-text {
  font-size: 13px;
  color: #6b7280;
}

.status-text.online {
  color: #10b981;
}

.status-text.offline {
  color: #ef4444;
}

.more-button {
  background-color: #f3f4f6;
  color: #6b7280;
  border: none;
  transition: all 0.2s ease;
}

.more-button:hover {
  background-color: #e5e7eb;
  color: #4b5563;
}

/* 面板内容 */
.panel-content {
  padding: 20px;
  background-color: #ffffff;
  border-top: 1px solid #e5e7eb;
}

.device-details {
  display: grid;
  grid-template-columns: 1fr;
  gap: 20px;
}

@media (min-width: 768px) {
  .device-details {
    grid-template-columns: 1fr 1fr;
  }
}

.section-title {
  font-size: 14px;
  font-weight: 500;
  color: #1f2937;
  margin-bottom: 12px;
  padding-bottom: 6px;
  border-bottom: 1px solid #e5e7eb;
}

.info-row {
  display: flex;
  margin-bottom: 10px;
}

.info-label {
  width: 120px;
  color: #6b7280;
  font-size: 14px;
}

.info-value {
  flex: 1;
  color: #1f2937;
  font-size: 14px;
  word-break: break-all;
}

.status.online {
  color: #10b981;
  font-weight: 500;
}

.status.offline {
  color: #ef4444;
  font-weight: 500;
}

.info-section{
  grid-column: 1 / -1;
}

/* 性能监控部分 */
.performance-section {
  grid-column: 1 / -1;
}

.metric-item {
  margin-bottom: 16px;
}

.metric-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 6px;
}

.metric-label {
  color: #6b7280;
  font-size: 14px;
}

.metric-value {
  color: #1f2937;
  font-weight: 500;
}

.progress-bar {
  height: 8px;
  border-radius: 4px;
  background-color: #f3f4f6;
  overflow: hidden;
}

.progress-track {
  height: 100%;
  position: relative;
}

.progress-fill {
  height: 100%;
  position: absolute;
  transition: width 0.5s ease;
}

.progress-fill.memory {
  background-color: #3b82f6;
}

.progress-fill.cpu {
  background-color: #f97316;
}

.memory-details {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #6b7280;
  margin-top: 4px;
}

.hp-memory-info {
  margin-top: 16px;
}

/* 管理员信息 */
.admin-section {
  grid-column: 1 / -1;
}

/* 操作按钮 */
.action-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 20px;
}

/* 动画效果 */
@keyframes pulse {
  0% {
    box-shadow: 0 0 0 0 rgba(16, 185, 129, 0.4);
  }
  70% {
    box-shadow: 0 0 0 8px rgba(16, 185, 129, 0);
  }
  100% {
    box-shadow: 0 0 0 0 rgba(16, 185, 129, 0);
  }
}

</style>
