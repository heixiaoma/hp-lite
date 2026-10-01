<template>
  <div class="hp-page">
    <div class="hp-toolbar">
      <t-button theme="primary" @click="addModal">
        <template #icon><add-icon/></template>
        添加规则
      </t-button>
      <t-button variant="outline" theme="primary" @click="loadData">
        <template #icon><refresh-icon/></template>
        刷新列表
      </t-button>
    </div>

    <t-table
        class="hp-table"
        row-key="id"
        :data="listData || []"
        :columns="viewColumns"
        :loading="dataLoading"
        :pagination="pagination"
        empty="暂无数据，添加一个试试看看"
        table-layout="auto"
        stripe
        hover
        @page-change="onPageChange"
    >
      <template #allowedIps="{ row }">
        <div class="hp-tag-cell">
          <template v-if="row.allowedIps && row.allowedIps.length > 0">
            <t-tag v-for="ip in row.allowedIps" :key="ip" theme="success" variant="light">{{ ip }}</t-tag>
          </template>
          <t-tag v-else>未启用</t-tag>
        </div>
      </template>

      <template #blockedIps="{ row }">
        <div class="hp-tag-cell">
          <template v-if="row.blockedIps && row.blockedIps.length > 0">
            <t-tag v-for="ip in row.blockedIps" :key="ip" theme="danger" variant="light">{{ ip }}</t-tag>
          </template>
          <t-tag v-else>未启用</t-tag>
        </div>
      </template>

      <template #inLimit="{ row }">
        <span v-if="row.inLimit <= 0" class="hp-muted">不限制</span>
        <span v-else>{{ row.inLimit }}</span>
      </template>

      <template #outLimit="{ row }">
        <span v-if="row.outLimit <= 0" class="hp-muted">不限制</span>
        <span v-else>{{ row.outLimit }}</span>
      </template>

      <template #rateLimit="{ row }">
        <span v-if="row.rateLimit <= 0" class="hp-muted">不限制</span>
        <span v-else>{{ row.rateLimit }}</span>
      </template>

      <template #user="{ row }">
        <t-tag v-if="!row.userDesc && !row.username" variant="light" theme="primary">自用</t-tag>
        <div v-else>
          <div>归属用户：{{ row.username }}</div>
          <div class="hp-sub-line">归属用户备注：{{ row.userDesc }}</div>
        </div>
      </template>

      <template #action="{ row }">
        <div class="hp-actions">
          <t-button size="small" variant="outline" theme="primary" @click="edit(row)">编辑</t-button>
          <t-button size="small" variant="outline" theme="warning" @click="refConfigData(row)">刷新规则</t-button>
          <t-popconfirm content="确定要删除该规则？" theme="danger" @confirm="removeData(row)">
            <t-button size="small" variant="outline" theme="danger">删除</t-button>
          </t-popconfirm>
        </div>
      </template>
    </t-table>

    <t-dialog
        v-model:visible="addVisible"
        :header="formState.id ? '编辑限制规则' : '添加限制规则'"
        confirm-btn="确定"
        cancel-btn="取消"
        width="640px"
        @confirm="addOk"
    >
      <div class="hp-dialog-body">
        <t-form
            :data="formState"
            ref="formTable"
            :rules="formRules"
            layout="vertical"
            label-align="top"
        >
          <t-form-item label="穿透配置" name="configId">
            <t-select
                v-model="formState.configId"
                filterable
                clearable
                placeholder="穿透配置备注关键字"
                :options="data"
                :loading="searchLoading"
                @search="handleSearch"
            />
          </t-form-item>

          <!-- 三个限制值：后端用 <=0 表示「不限制」，把 -1 摆给用户填没有意义。
               统一留空（靠 placeholder 说明），提交时再回落成 -1；填了数字就以数字为准 -->
          <div class="form-grid">
            <t-form-item label="并发限制" name="rateLimit">
              <t-input v-model="formState.rateLimit" clearable placeholder="留空不限制"/>
            </t-form-item>

            <t-form-item label="上传限制" name="inLimit">
              <t-input v-model="formState.inLimit" clearable placeholder="留空不限制"/>
            </t-form-item>

            <t-form-item label="下载限制" name="outLimit">
              <t-input v-model="formState.outLimit" clearable placeholder="留空不限制"/>
            </t-form-item>
          </div>
          <p class="form-tip">
            并发单位为每分钟连接数，上传 / 下载单位为字节 / 秒。三项都留空即不做限制。
          </p>

          <!-- IP 规则显式三选一。
               原来靠「另一个数组为空」来互斥显示：初始两个数组都为空，两块同时出现；
               点一次「添加」后另一种模式被直接锁死，且首行减号还是禁用的，退不回去 -->
          <t-form-item label="IP 规则">
            <t-radio-group v-model="ipMode" variant="default-filled">
              <t-radio-button value="none">不限制</t-radio-button>
              <t-radio-button value="allow">仅允许以下 IP</t-radio-button>
              <t-radio-button value="block">禁止以下 IP</t-radio-button>
            </t-radio-group>
          </t-form-item>

          <div v-if="ipMode === 'allow'" class="ip-block">
            <t-form-item
                v-for="(ip, index) in formState.allowedIps"
                :key="'allow-' + index"
                :label="index === 0 ? '允许IP(CIDR地址)' : ''"
            >
              <div class="ip-row">
                <t-input v-model="formState.allowedIps[index]" placeholder="请输入IP规则:0.0.0.0/0"/>
                <t-button variant="outline" theme="danger" shape="square" @click="removeAllowedIps(index)">
                  <template #icon><remove-icon/></template>
                </t-button>
              </div>
            </t-form-item>
            <t-button variant="dashed" block @click="addAllowedIps">
              <template #icon><add-icon/></template>
              添加一行
            </t-button>
          </div>

          <div v-if="ipMode === 'block'" class="ip-block">
            <t-form-item
                v-for="(ip, index) in formState.blockedIps"
                :key="'block-' + index"
                :label="index === 0 ? '禁止IP(CIDR地址)' : ''"
            >
              <div class="ip-row">
                <t-input v-model="formState.blockedIps[index]" placeholder="请输入IP规则:127.0.0.1/0"/>
                <t-button variant="outline" theme="danger" shape="square" @click="removeBlockedIps(index)">
                  <template #icon><remove-icon/></template>
                </t-button>
              </div>
            </t-form-item>
            <t-button variant="dashed" block @click="addBlockedIps">
              <template #icon><add-icon/></template>
              添加一行
            </t-button>
          </div>
        </t-form>
      </div>
    </t-dialog>
  </div>
</template>

<script setup>
import {getWaf, removeWaf, saveWaf} from "../../api/client/waf";
import {onMounted, reactive, ref} from "vue";
import {MessagePlugin} from "tdesign-vue-next";
import {AddIcon, RefreshIcon, RemoveIcon} from 'tdesign-icons-vue-next';
import {getConfigByKeyword, refConfig} from "../../api/client/config.js";
import {useResponsiveColumns} from '../../utils/responsive';

const listData = ref([]);
const formTable = ref();
const dataLoading = ref(false);
const addVisible = ref(false);
const searchLoading = ref(false);

const formState = reactive({
  configId: "",
  allowedIps: [],
  blockedIps: [],
  rateLimit: "",
  outLimit: "",
  inLimit: "",
  id: undefined
})

// IP 规则三选一：none 不限制 / allow 白名单 / block 黑名单
const ipMode = ref('none')

/* 后端用 <=0 表示「不限制」。表单里留空更直观，提交时再回落成 -1；
   打开表单时反过来，把 -1 还原成空串，用户看到的是 placeholder 而不是没意义的 -1 */
const toLimitValue = (v) => {
  if (v === '' || v === null || v === undefined) return -1
  const n = Number(v)
  return Number.isNaN(n) ? -1 : n
}

const toLimitInput = (v) => {
  if (v === '' || v === null || v === undefined) return ''
  const n = Number(v)
  return (Number.isNaN(n) || n <= 0) ? '' : String(n)
}

// 过滤掉加了行却没填内容的空串，避免把空规则提交上去
const cleanIps = (arr) => (arr || []).map(s => String(s).trim()).filter(Boolean)

// 三项限制都允许留空，所以只有穿透配置是必填
const formRules = {
  configId: [{required: true, message: '选择穿透配置', type: 'error'}],
};

const pagination = reactive({
  total: 0,
  current: 1,
  pageSize: 10,
  pageSizeOptions: [10, 20, 50],
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
  }).catch(() => {
    dataLoading.value = false
  })
}

const removeData = (item) => {
  removeWaf({
    id: item.id
  }).then(res => {
    MessagePlugin.success(res.msg)
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
  const item = JSON.parse(JSON.stringify(itemOld))
  formState.allowedIps = item.allowedIps || []
  formState.blockedIps = item.blockedIps || []
  formState.rateLimit = toLimitInput(item.rateLimit)
  formState.inLimit = toLimitInput(item.inLimit)
  formState.outLimit = toLimitInput(item.outLimit)
  formState.id = item.id
  formState.configId = item.configId
  ipMode.value = formState.allowedIps.length > 0
      ? 'allow'
      : (formState.blockedIps.length > 0 ? 'block' : 'none')
  addVisible.value = true
  handleSearch(item.configId)
}

/* 窄屏只留「认得出是哪条规则 + 能操作」，速率/并发/归属这些次要列让位 */
const columns = [
  {colKey: 'id', title: '编号', width: 90, mobile: false},
  {colKey: 'configId', title: '配置ID', width: 90, mobile: false},
  {colKey: 'configDesc', title: '配置描述'},
  {colKey: 'allowedIps', title: '允许IP', width: 200},
  {colKey: 'blockedIps', title: '禁止IP', width: 200},
  {colKey: 'inLimit', title: '上传速率(byte)', width: 140, mobile: false},
  {colKey: 'outLimit', title: '下载速率(byte)', width: 140, mobile: false},
  {colKey: 'rateLimit', title: '并发连接限制', width: 130, mobile: false},
  {colKey: 'user', title: '归属', width: 200, mobile: false},
  {colKey: 'action', title: '操作', width: 230},
];
const viewColumns = useResponsiveColumns(columns);

const onPageChange = (pageInfo) => {
  pagination.current = pageInfo.current
  pagination.pageSize = pageInfo.pageSize
  loadData()
}

const addModal = () => {
  formState.allowedIps = []
  formState.blockedIps = []
  formState.rateLimit = ''
  formState.inLimit = ''
  formState.outLimit = ''
  formState.configId = ''
  formState.id = undefined
  ipMode.value = 'none'
  addVisible.value = true
}

const addOk = async () => {
  const result = await formTable.value?.validate()
  if (result !== true) return

  // 留空回落 -1，填了数字就以数字为准
  formState.rateLimit = toLimitValue(formState.rateLimit)
  formState.inLimit = toLimitValue(formState.inLimit)
  formState.outLimit = toLimitValue(formState.outLimit)
  formState.configId = parseInt(formState.configId)

  // 只下发当前选中的那种规则，另一种置空，避免白名单和黑名单同时生效
  formState.allowedIps = ipMode.value === 'allow' ? cleanIps(formState.allowedIps) : []
  formState.blockedIps = ipMode.value === 'block' ? cleanIps(formState.blockedIps) : []

  saveWaf({...formState}).then(res => {
    MessagePlugin.success(res.msg)
    loadData()
    addVisible.value = false
  })
}

const addAllowedIps = () => {
  formState.allowedIps.push("");
};

// 按索引删。原来用 indexOf 找值，两行填了相同 IP 时会删错行；
// 而且不再限制「至少留一行」——删空了就回到只有添加按钮的状态
const removeAllowedIps = (index) => {
  formState.allowedIps.splice(index, 1);
};

const addBlockedIps = () => {
  formState.blockedIps.push("");
};

const removeBlockedIps = (index) => {
  formState.blockedIps.splice(index, 1);
};

const data = ref([]);

let timeout;
let currentValue = '';
const handleSearch = (value) => {
  if (timeout) {
    clearTimeout(timeout);
    timeout = null;
  }
  currentValue = value;

  const fake = () => {
    searchLoading.value = true
    getConfigByKeyword(value).then(res => {
      searchLoading.value = false
      if (currentValue === value && res.data) {
        data.value = res.data.map(r => ({value: r.id, label: r.remarks}))
      }
    }).catch(() => {
      searchLoading.value = false
    })
  }
  timeout = setTimeout(fake, 300);
};

onMounted(() => {
  loadData()
})
</script>

<style scoped>
/* TDesign 只把「直接子级」的 form-item 当字段。form-grid / form-tip 是自定义块，
   于是它前面的「穿透配置」被判成最后一项（t-form__item--last）、margin-bottom 归零，
   实测 select 到「并发限制」label 的间距是 0。这里按 24px 的节奏补回来 */
.form-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0 16px;
  margin-top: 24px;
}

/* 弹窗在手机上只有 ~366px，三列会把每个输入框压到 110px，数字根本输不进去 */
@media (max-width: 768px) {
  .form-grid {
    grid-template-columns: 1fr;
  }
}

/* label 在 top 模式下的 min-height 是 32px，比文字本身（22px 行高）高出一截，
   纯白占垂直空间。压到和文字一致，整个弹窗能矮 40px 左右 */
.hp-dialog-body :deep(.t-form__label--top) {
  min-height: 0;
  line-height: 22px;
}

/* 限制值的单位说明。原来塞在 placeholder 里，列一窄就被截断了。
   -8px 是抵消网格里 form-item 的 24px 底边距，收成 16px 贴着自己的字段组；
   底部留 24px，否则会和下面的「IP 规则」贴死 */
.form-tip {
  margin: -8px 0 24px;
  font-size: 12px;
  line-height: 1.6;
  color: var(--hp-text-3);
}

/* TDesign 的弹窗按 20vh 的顶部偏移往下排，而全局的 .hp-dialog-body 上限是 60vh，
   两者相加在内容多时（IP 规则加几行）会把底部的「确定 / 取消」顶出视口。
   IP 行数不固定，这里收紧上限，超出时让 body 内部滚动 */
.hp-dialog-body {
  max-height: 52vh;
}

.ip-block {
  padding: 14px;
  margin-top: 4px;
  background: #fafbfe;
  border: 1px dashed var(--hp-border);
  border-radius: 12px;
}

.ip-row {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
}
</style>
