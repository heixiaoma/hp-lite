<template>
  <div class="hp-page">
    <div class="hp-toolbar">
      <t-button theme="primary" @click="addDeviceModal">
        <template #icon><add-icon/></template>
        添加设备
      </t-button>
      <t-button variant="outline" theme="primary" @click="loadData">
        <template #icon><refresh-icon/></template>
        刷新列表
      </t-button>
    </div>

    <t-table
        class="hp-table"
        row-key="deviceId"
        :data="deviceList || []"
        :columns="columns"
        :loading="listLoading"
        :pagination="pagination"
        empty="暂无数据，添加一个试试看看"
        table-layout="auto"
        stripe
        hover
        @page-change="onPageChange"
    >
      <template #desc="{ row }">
        <div class="device-name">{{ row.desc }}</div>
      </template>

      <template #online="{ row }">
        <span class="hp-status" :class="row.online ? 'hp-status--online' : 'hp-status--offline'">
          <i class="hp-status__dot"></i>
          {{ row.online ? '在线中' : '未在线' }}
        </span>
      </template>

      <template #action="{ row }">
        <div class="hp-actions">
          <t-button size="small" variant="outline" theme="primary" @click="showQr(row)">连接码</t-button>
          <t-button size="small" variant="outline" theme="primary" @click="edit(row)">编辑</t-button>
        </div>
      </template>

      <template #expandedRow="{ row }">
        <div class="device-detail">
          <div class="detail-card">
            <section class="detail-section">
              <h4 class="detail-section__title">设备信息</h4>
              <div class="detail-row">
                <span class="detail-row__label">设备 ID</span>
                <div class="detail-row__value is-code">
                  <code class="hp-code" :title="row.deviceId">{{ row.deviceId }}</code>
                  <t-button size="small" variant="text" theme="primary" @click="copyText(row.deviceId)">
                    <template #icon><copy-icon/></template>
                    复制
                  </t-button>
                </div>
              </div>
              <div class="detail-row">
                <span class="detail-row__label">连接码</span>
                <div class="detail-row__value is-code">
                  <code class="hp-code" :title="row.connectKey">{{ row.connectKey }}</code>
                  <t-button size="small" variant="text" theme="primary" @click="copyText(row.connectKey)">
                    <template #icon><copy-icon/></template>
                    复制
                  </t-button>
                </div>
              </div>
            </section>

            <section v-if="row.memoryInfo" class="detail-section">
              <h4 class="detail-section__title">性能监控</h4>

              <div class="metric-grid">
                <div class="hp-metric">
                  <div class="hp-metric__head">
                    <span class="hp-metric__label">内存使用率</span>
                    <span class="hp-metric__value">{{ memRate(row) }}%</span>
                  </div>
                  <t-progress theme="line" :percentage="memRate(row)" :label="false"/>
                  <div class="hp-metric__foot">
                    <span>已用 {{ (row.memoryInfo.useMem / 1024 / 1024).toFixed(2) }} MB</span>
                    <span>共 {{ (row.memoryInfo.total / 1024 / 1024).toFixed(2) }} MB</span>
                  </div>
                </div>

                <div class="hp-metric">
                  <div class="hp-metric__head">
                    <span class="hp-metric__label">CPU 使用率</span>
                    <span class="hp-metric__value">{{ row.memoryInfo.cpuRate.toFixed(1) }}%</span>
                  </div>
                  <t-progress theme="line" color="#f97316" :percentage="row.memoryInfo.cpuRate" :label="false"/>
                </div>
              </div>

              <div class="detail-row">
                <span class="detail-row__label">HP 占用内存</span>
                <span class="detail-row__value">{{ (row.memoryInfo.hpTotalMem / 1024 / 1024).toFixed(2) }} MB</span>
              </div>
              <div class="detail-row">
                <span class="detail-row__label">HP 实际使用</span>
                <span class="detail-row__value">{{ (row.memoryInfo.hpUseMem / 1024 / 1024).toFixed(2) }} MB</span>
              </div>
            </section>

            <section v-if="isAdmin" class="detail-section">
              <h4 class="detail-section__title">归属信息</h4>
              <div class="detail-row">
                <span class="detail-row__label">归属用户</span>
                <span class="detail-row__value">{{ row.username || '无' }}</span>
              </div>
              <div class="detail-row">
                <span class="detail-row__label">用户备注</span>
                <span class="detail-row__value">{{ row.userDesc || '无' }}</span>
              </div>
            </section>
          </div>

          <div class="device-detail__actions">
            <t-popconfirm content="确定要删除该设备？" theme="danger" @confirm="removeData(row)">
              <t-button size="small" variant="outline" theme="danger">删除</t-button>
            </t-popconfirm>
            <t-popconfirm content="确定要强制停止程序？" theme="warning" @confirm="stopData(row)">
              <t-button size="small" variant="outline" theme="warning">强制停止</t-button>
            </t-popconfirm>
          </div>
        </div>
      </template>
    </t-table>

    <!-- 连接码 -->
    <t-dialog
        v-model:visible="qrModalVisible"
        header="设备二维码"
        :footer="false"
        width="560px"
        destroy-on-close
    >
      <qr v-if="deviceId" :text="deviceId"/>
      <div class="qr-footer">
        <t-button theme="primary" block @click="closeQr">我已知晓</t-button>
      </div>
    </t-dialog>

    <!-- 新增设备 -->
    <t-dialog
        v-model:visible="addDeviceModalVisible"
        header="添加设备"
        confirm-btn="确定"
        cancel-btn="取消"
        width="560px"
        @confirm="addDeviceOk"
    >
      <t-form :data="formState" ref="formTable" :rules="formRules" layout="vertical" label-align="top">
        <t-form-item label="设备编号" name="deviceId">
          <t-input v-model="formState.deviceId" clearable placeholder="设备ID：32位">
            <template #suffix>
              <t-link theme="primary" hover="color" @click="guid">自动生成</t-link>
            </template>
          </t-input>
        </t-form-item>
        <t-form-item label="设备备注" name="desc">
          <t-input v-model="formState.desc" clearable placeholder="备注如：nas中的HP"/>
        </t-form-item>
      </t-form>
    </t-dialog>

    <!-- 编辑设备 -->
    <t-dialog
        v-model:visible="updateDeviceModalVisible"
        header="编辑设备"
        confirm-btn="确定"
        cancel-btn="取消"
        width="560px"
        @confirm="updateDeviceOk"
    >
      <t-form :data="formState" ref="formTable" :rules="formRules" layout="vertical" label-align="top">
        <t-form-item label="设备编号" name="deviceId">
          <t-input v-model="formState.deviceId" disabled placeholder="设备ID：32位"/>
        </t-form-item>
        <t-form-item label="设备备注" name="desc">
          <t-input v-model="formState.desc" clearable placeholder="备注如：nas中的HP"/>
        </t-form-item>
      </t-form>
    </t-dialog>
  </div>
</template>

<script setup>
import {computed, onMounted, reactive, ref} from "vue";
import {addDevice, getDeviceList, removeDevice, stopDevice, updateDevice} from "../../api/client/device";
import qr from './qr.vue';
import userInfo from "../../data/userInfo";
import {MessagePlugin} from 'tdesign-vue-next';
import {AddIcon, CopyIcon, RefreshIcon} from 'tdesign-icons-vue-next';
import {copyText} from '../../utils/clipboard';

const formTable = ref()
const deviceId = ref('')
const deviceList = ref([])
const listLoading = ref(false)
const qrModalVisible = ref(false)
const addDeviceModalVisible = ref(false)
const updateDeviceModalVisible = ref(false)

const formState = reactive({
  deviceId: "",
  desc: ""
})

const formRules = {
  deviceId: [{required: true, message: '设备编号必填', type: 'error'}],
  desc: [{required: true, message: '设备备注必填', type: 'error'}],
};

const pagination = reactive({
  total: 0,
  current: 1,
  pageSize: 10,
  pageSizeOptions: [10, 20, 50],
});

const isAdmin = computed(() => userInfo.getUserInfo()?.role === 'ADMIN');

const columns = [
  {colKey: 'desc', title: '描述'},
  {colKey: 'online', title: '在线状态', width: 130},
  {colKey: 'action', title: '操作', width: 170},
];

const memRate = (row) => {
  if (!row.memoryInfo || !row.memoryInfo.total) return 0
  return Number(((row.memoryInfo.useMem / row.memoryInfo.total) * 100).toFixed(1))
}

/* 设备 ID / 连接码是 32 位长串，界面上被省略号截断，只能靠复制拿完整值。
   提示和「复制失败时弹手动复制框」都封装在 utils/clipboard 里了 */


const onPageChange = (pageInfo) => {
  pagination.current = pageInfo.current
  pagination.pageSize = pageInfo.pageSize
  loadData()
}

const addDeviceModal = () => {
  formState.deviceId = ""
  formState.desc = ""
  addDeviceModalVisible.value = true;
};

const edit = (item) => {
  formState.deviceId = item.deviceId
  formState.desc = item.desc
  updateDeviceModalVisible.value = true;
}

const showQr = (item) => {
  qrModalVisible.value = true;
  deviceId.value = item.connectKey;
}

const closeQr = () => {
  qrModalVisible.value = false;
  deviceId.value = "";
}

const addDeviceOk = async () => {
  const result = await formTable.value?.validate()
  if (result !== true) return

  addDevice({...formState}).then(() => {
    formState.deviceId = ''
    formState.desc = ''
    loadData();
    addDeviceModalVisible.value = false;
  })
};

const updateDeviceOk = async () => {
  const result = await formTable.value?.validate()
  if (result !== true) return

  updateDevice({...formState}).then(() => {
    formState.deviceId = ''
    formState.desc = ''
    loadData();
    updateDeviceModalVisible.value = false;
  })
};

const stopData = (item) => {
  stopDevice({
    deviceId: item.deviceId
  }).then(() => {
    loadData();
  });
};

const loadData = () => {
  listLoading.value = true
  getDeviceList(pagination).then(res => {
    listLoading.value = false
    if (res.data) {
      deviceList.value = res.data.records
      pagination.total = res.data.total
    }
  }).catch(() => {
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

const removeData = (item) => {
  removeDevice({
    deviceId: item.deviceId
  }).then(() => {
    loadData();
  });
};

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.device-name {
  font-weight: 600;
}

.device-detail {
  padding: 14px 16px;
  background: #fbfcfe;
  border-radius: 12px;
}

/* 三个分区合并进一个 card，内部改用分隔线区分，不再各自套一层白底边框 */
.detail-card {
  background: #fff;
  border: 1px solid var(--hp-border);
  border-radius: 12px;
  padding: 0 18px 16px;
}

.detail-section + .detail-section {
  padding-top: 14px;
  border-top: 1px solid #eef2f8;
}

.detail-section__title {
  margin: 0 0 12px;
  padding-top: 14px;
  font-size: 13px;
  font-weight: 600;
  color: var(--hp-text);
}

.detail-row {
  display: flex;
  align-items: baseline;
  gap: 10px;
  margin-bottom: 8px;
  font-size: 13px;
}

.detail-row__label {
  flex: none;
  width: 88px;
  color: var(--hp-text-3);
}

.detail-row__value {
  flex: 1;
  min-width: 0;
  color: var(--hp-text);
  word-break: break-all;
}

/* 设备 ID / 连接码这一行：长串单行省略，右侧留出复制按钮 */
.detail-row__value.is-code {
  display: flex;
  align-items: center;
  gap: 6px;
  word-break: normal;
}

/* 32 位长串原来 break-all 会折成好几行，撑得又高又乱。
   改成单行省略号，完整值靠 title 悬浮和「复制」按钮拿到 */
.hp-code {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-family: 'SFMono-Regular', Consolas, Menlo, monospace;
  font-size: 12px;
}

/* 内存 / CPU 上下排列：指标各占满一整行，进度条更长，百分比与用量数字也更好读。
   并排时每个指标只有半行宽，进度条被压短，两个百分比容易看串行 */
.metric-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 14px;
  margin-bottom: 12px;
}

.device-detail__actions {
  margin-top: 14px;
  display: flex;
  gap: 10px;
}

.qr-footer {
  margin-top: 16px;
}
</style>
