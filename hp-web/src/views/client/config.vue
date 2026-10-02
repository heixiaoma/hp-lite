<template>
  <div class="hp-page">
    <div class="hp-toolbar">
      <t-button theme="primary" @click="addConfigModal">
        <template #icon><add-icon/></template>
        添加穿透
      </t-button>
      <t-button variant="outline" theme="primary" @click="loadData">
        <template #icon><refresh-icon/></template>
        刷新列表
      </t-button>
      <div class="hp-toolbar__grow"></div>
      <t-input
          v-model="pagination.keyword"
          class="hp-search-input"
          clearable
          placeholder="关键字查询"
          @enter="loadData"
      >
        <template #prefix-icon><search-icon/></template>
      </t-input>
      <t-button theme="primary" @click="loadData">查询</t-button>
    </div>

    <t-table
        class="hp-table"
        row-key="id"
        :data="currentConfigList || []"
        :columns="viewColumns"
        :loading="configLoading"
        :pagination="pagination"
        empty="暂无配置，添加一个试试看看"
        table-layout="auto"
        stripe
        hover
        @page-change="onPageChange"
    >
      <template #server="{ row }">
        <div class="hp-tag-cell">
          <t-tag v-for="item in openAddress(row)" :key="item" theme="primary" variant="light">{{ item }}</t-tag>
        </div>
      </template>

      <template #status="{ row }">
        <t-switch :value="!row.status || row.status === 0" @change="() => changeData(row)"/>
      </template>

      <template #tunType="{ row }">
        <t-tag variant="light" :theme="row.tunType === 'TCP' ? 'primary' : 'success'">
          {{ row.tunType === 'TCP' ? 'TCP 多路复用' : 'QUIC 多路复用' }}
        </t-tag>
      </template>

      <template #deviceKey="{ row }">
        <div>{{ userKeyByName(row.deviceKey) }}</div>
        <div v-if="isAdmin && userKeyByUserInfo(row.deviceKey).username" class="hp-sub-line">
          归属用户：{{ userKeyByUserInfo(row.deviceKey).username }}
        </div>
        <div v-if="isAdmin && userKeyByUserInfo(row.deviceKey).userDesc" class="hp-sub-line">
          归属用户备注：{{ userKeyByUserInfo(row.deviceKey).userDesc }}
        </div>
      </template>

      <template #action="{ row }">
        <div class="hp-actions">
          <t-button size="small" variant="outline" theme="primary" @click="editConfigData(row)">编辑</t-button>
          <t-button size="small" variant="outline" theme="warning" @click="refConfigData(row)">重连配置</t-button>
          <t-popconfirm content="确定要删除该穿透配置？" theme="danger" @confirm="removeConfigData(row)">
            <t-button size="small" variant="outline" theme="danger">删除</t-button>
          </t-popconfirm>
        </div>
      </template>

      <template #expandedRow="{ row }">
        <div class="config-detail">
          <div>备注：{{ row.remarks }}</div>
          <div v-if="row.statusMsg">
            最近一条穿透服务日志：
            <t-tag theme="primary" variant="light">{{ row.statusMsg }}</t-tag>
          </div>
        </div>
      </template>
    </t-table>

    <t-dialog
        v-model:visible="addConfigVisible"
        :header="formState.id ? '编辑内网穿透配置' : '添加内网穿透配置'"
        confirm-btn="确定"
        cancel-btn="取消"
        width="720px"
        @confirm="addConfigOk"
    >
      <div class="hp-dialog-body">
        <t-form :data="formState" ref="formTable" :rules="formRules" layout="vertical" label-align="top">
          <t-form-item label="穿透设备" name="deviceKey">
            <t-select v-model="formState.deviceKey" :options="currentUserKeyList" placeholder="请选择穿透设备"/>
          </t-form-item>

          <t-form-item label="穿透备注" name="remarks">
            <t-input v-model="formState.remarks" clearable placeholder="备注如：个人博客"/>
          </t-form-item>

          <t-divider>端口映射配置</t-divider>

          <div class="form-grid">
            <t-form-item label="外网端口" name="remotePort">
              <t-input-number
                  v-model="formState.remotePort"
                  theme="normal"
                  placeholder="8084"
                  style="width: 100%"
              />
            </t-form-item>
          </div>

          <!-- 说明入口收成输入框内的前缀图标：不额外占一行的高度，
               确实不知道怎么填才点图标，展开后是各协议的完整填写示例 -->
          <t-form-item label="内网地址" name="localAddress">
            <t-input v-model="formState.localAddress" clearable placeholder="http://127.0.0.1:8084">
              <template #prefixIcon>
                <t-popup
                    placement="bottom-left"
                    trigger="click"
                    destroy-on-close
                    :overlay-inner-style="{ width: '520px', maxHeight: '380px', overflow: 'auto', padding: '14px 16px' }"
                >
                  <t-icon name="help-circle" class="hp-help-icon"/>
                  <template #content>
                    <div class="doc-list">
                      <div v-for="doc in protocolDocs" :key="doc.value" class="doc-item">
                        <div class="doc-item__title">{{ doc.header }}</div>
                        <div class="hp-tag-cell">
                          <t-tag
                              v-for="tag in doc.tags"
                              :key="tag"
                              class="doc-tag"
                              theme="primary"
                              variant="light-outline"
                              size="small"
                          >{{ tag }}</t-tag>
                        </div>
                        <p v-if="doc.desc" class="doc-item__desc">{{ doc.desc }}</p>
                      </div>
                    </div>
                  </template>
                </t-popup>
              </template>
            </t-input>
          </t-form-item>

          <t-form-item v-if="showInput.proxyVersion" label="代理协议" name="proxyVersion">
            <t-select v-model="formState.proxyVersion" :options="proxyOptions"/>
          </t-form-item>

          <t-form-item v-if="showInput.domain" label="绑定访问域名" name="domain">
            <t-select v-model="formState.domain" :options="domainOptions" placeholder="选择一个域名" clearable/>
          </t-form-item>

          <t-form-item
              v-if="showInput.domain && formState.domain"
              label="防火墙模式"
              name="safeType"
          >
            <t-select v-model="formState.safeType" :options="safeTypeOptions"/>
          </t-form-item>

          <t-form-item
              v-if="showInput.domain && formState.domain && formState.safeType !== 0"
              label="防火墙规则"
              name="safeId"
          >
            <t-select v-model="formState.safeId" :options="safeOptions" placeholder="选择一个规则"/>
          </t-form-item>

          <t-divider>其他选项配置</t-divider>

          <div class="form-grid">
            <t-form-item label="配置有效" name="status">
              <t-select v-model="formState.status" :options="statusOptions"/>
            </t-form-item>

            <t-form-item label="隧道模式" name="tunType">
              <t-select v-model="formState.tunType" :options="tunTypeOptions"/>
            </t-form-item>
          </div>
        </t-form>
      </div>
    </t-dialog>
  </div>
</template>

<script setup>
import {computed, onMounted, reactive, ref, watch} from "vue";
import {addConfig, changeStatus, getConfigList, getDeviceKey, refConfig, removeConfig} from "../../api/client/config";
import {useRoute} from 'vue-router'
import userInfo from "../../data/userInfo";
import {queryDomain} from "../../api/client/domain.js";
import {querySafe} from "../../api/client/safe.js";
import {AddIcon, RefreshIcon, SearchIcon} from 'tdesign-icons-vue-next';
import {useResponsiveColumns} from '../../utils/responsive';

const route = useRoute()

const domainOptions = ref([]);
const safeOptions = ref([]);
const currentConfigList = ref([]);
const currentUserKeyList = ref([]);

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
  status: 0,
  safeType: 0,
  // 用 undefined 而不是 0：safeOptions 里没有 value=0 的项，
  // 初始为 0 时 t-select 匹配不到 label，就会把数字 0 直接显示出来
  safeId: undefined,
  tunType: "TCP",
})

const formRules = {
  deviceKey: [{required: true, message: '穿透设备必填', type: 'error'}],
  remarks: [{required: true, message: '穿透备注必填', type: 'error'}],
  remotePort: [{required: true, message: '外网端口必填', type: 'error'}],
  localAddress: [{required: true, message: '内网地址必填', type: 'error'}],
  status: [{required: true, message: '当前配置是否有效', type: 'error'}],
  tunType: [{required: true, message: '选择隧道模式', type: 'error'}],
  safeId: [{
    validator: (val) => {
      if (!showInput.domain || !formState.domain || formState.safeType === 0) return true
      return val !== undefined && val !== null && val !== ''
    },
    message: '防火墙规则必选，没有就去创建一个',
    trigger: 'change'
  }],
};

const pagination = reactive({
  total: 0,
  current: 1,
  pageSize: 10,
  pageSizeOptions: [10, 20, 50],
  keyword: '',
});

const isAdmin = computed(() => userInfo.getUserInfo()?.role === 'ADMIN');

const protocolDocs = [
  {
    value: 'http',
    header: 'HTTP/HTTPS 协议',
    tags: ['http://127.0.0.1', 'https://127.0.0.1', 'http://127.0.0.1:8080', 'https://192.168.55:8080', 'http://192.168.15:8080'],
    desc: 'http/https协议支持默认端口方式或者手动指定端口、当选择http协议时可以自由选择是否绑定域名'
  },
  {
    value: 'tcp',
    header: 'TCP 协议',
    tags: ['tcp://127.0.0.1:1080', 'tcp://192.168.10.1:1080'],
    desc: 'TCP级别协议、当选择tcp协议可以设置代理协议，通常情况下是不用设置，如果有获取真实IP或者对该协议熟悉的人可以选择设置'
  },
  {
    value: 'udp',
    header: 'UDP 协议',
    tags: ['udp://127.0.0.1:1080', 'udp://192.168.10.1:1080'],
    desc: ''
  },
  {
    value: 'socks5',
    header: 'SOCKS5 协议',
    tags: ['socks5://127.0.0.1', 'socks5://用户名:密码@127.0.0.1'],
    desc: 'socks5协议、会把外网的数据转移到内网通过socks5方式代理、最后实现使用内网的IP进行上网、可以选择设置密码和不设置密码'
  },
  {
    value: 'unix',
    header: 'UNIX 协议',
    tags: ['unix:///tmp/socks.sock', 'unix:///tmp/****.sock'],
    desc: 'unix协议是直接连接到文件上、请确保sock文件路径正确'
  },
  {
    value: 'tcp_udp',
    header: 'TCP+UDP 协议',
    tags: ['tcp_udp://127.0.0.1:8080', 'tcp_udp://192.168.10.1:8080'],
    desc: 'tcp+udp双协议同时监听'
  },
];

const proxyOptions = [
  {label: '不设置(小白用户请不要设置-用于TCP获取真实IP)', value: 'NONE'},
  {label: 'TCP#V1版本', value: 'V1'},
  {label: 'TCP#V2版本', value: 'V2'},
];

const safeTypeOptions = [
  {label: '无防护 (域名+外网端口都可访问)', value: 0},
  {label: '全防护（域名防护+外网端口不可访问）', value: 1},
  {label: '半防护（域名防护+外网端口可访问但无防护）', value: 2},
];

const statusOptions = [
  {label: '有效', value: 0},
  {label: '无效', value: 1},
];

const tunTypeOptions = [
  {label: 'TCP多路复用模式', value: 'TCP'},
  {label: 'QUIC多路复用模式', value: 'QUIC'},
];

/* 窄屏留备注/内外网地址/操作：配置ID、隧道模式、部署设备让位 */
const columns = [
  {colKey: 'id', title: '配置ID', width: 90, mobile: false},
  {colKey: 'remarks', title: '备注'},
  {colKey: 'tunType', title: '隧道模式', width: 140, mobile: false},
  {colKey: 'localAddress', title: '内网服务'},
  {colKey: 'server', title: '外网服务'},
  {colKey: 'status', title: '配置有效', width: 90, align: 'center', mobile: false},
  {colKey: 'deviceKey', title: '部署设备', mobile: false},
  {colKey: 'action', title: '操作', width: 250},
];
const viewColumns = useResponsiveColumns(columns);

const showInput = reactive({
  proxyVersion: false,
  domain: false,
})

watch(() => formState.localAddress, (newVal) => {
  const val = newVal || ''
  showInput.proxyVersion = val.startsWith("tcp")
  showInput.domain = val.startsWith("http")
})

// 切回「无防护」时清掉规则：否则残留的 safeId 会被提交，
// 下次打开也会因为匹配不到选项而显示成数字
watch(() => formState.safeType, (val) => {
  if (!val) formState.safeId = undefined
})

const onPageChange = (pageInfo) => {
  pagination.current = pageInfo.current
  pagination.pageSize = pageInfo.pageSize
  loadData()
}

const changeData = (item) => {
  configLoading.value = true
  changeStatus({
    configId: item.id
  }).then(res => {
    configLoading.value = false
    if (res.code === 200) {
      item.status = (!item.status || item.status === 0) ? 1 : 0
    }
  }).catch(() => {
    configLoading.value = false
  })
}

/**
 * 统一取出列表数组。
 * http 拦截器返回的已经是 JSON body（res.data 就是接口返回的 data 字段），
 * 原先这里写的是 res.data.data，取到的是 undefined，
 * 导致域名 / 防护规则下拉选项一直是空的 —— select 匹配不到 label
 * 就只能把数字显示出来。这里对两种返回形态都兼容。
 */
const toList = (res) => {
  const d = res?.data
  if (Array.isArray(d)) return d
  if (Array.isArray(d?.data)) return d.data
  if (Array.isArray(d?.records)) return d.records
  return []
}

const loadDomains = () => {
  queryDomain({}).then(res => {
    domainOptions.value = toList(res).map(r => ({value: r.domain, label: r.domain}))
  })
}

const loadSafes = () => {
  querySafe({}).then(res => {
    safeOptions.value = toList(res).map(r => ({value: r.id, label: r.ruleName}))
  })
}

const loadDeviceKey = () => {
  getDeviceKey().then(res => {
    currentUserKeyList.value = toList(res).map(k => ({
      label: k.desc,
      value: k.key,
      userDesc: k.userDesc,
      username: k.username
    }))
  })
}

const userKeyByName = (deviceKey) => {
  try {
    return currentUserKeyList.value.filter(r => r.value === deviceKey)[0].label
  } catch (e) {
    return "设备获取错误"
  }
}

const userKeyByUserInfo = (deviceKey) => {
  try {
    return currentUserKeyList.value.filter(r => r.value === deviceKey)[0]
  } catch (e) {
    return {}
  }
}

const loadData = () => {
  loadDomains()
  loadSafes()
  currentConfigList.value = []
  configLoading.value = true
  getConfigList(pagination).then(res => {
    configLoading.value = false
    currentConfigList.value = res.data.records
    pagination.total = res.data.total
  }).catch(() => {
    configLoading.value = false
  })
}

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
  formState.proxyVersion = item.proxyVersion || 'NONE'
  formState.tunType = item.tunType || "QUIC"
  formState.status = item.status
  // 后端用 0 表示「未选规则」，转成 undefined，否则 select 会显示 0
  formState.safeId = item.safeId || undefined
  formState.safeType = item.safeType
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
  }).catch(() => {
    configLoading.value = false
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
  formState.tunType = 'TCP'
  formState.status = 0
  formState.safeType = 0
  formState.safeId = undefined
  addConfigVisible.value = true;
};

const addConfigOk = async () => {
  const result = await formTable.value?.validate()
  if (result !== true) return

  addConfig({
    packageId: route.query.packageId,
    ...formState,
    // undefined 在 JSON 序列化时会被整个丢掉，后端仍需要一个数字
    safeId: formState.safeId ?? 0,
  }).then(() => {
    loadData()
    addConfigVisible.value = false;
  })
};

const openAddress = (item) => {
  const address = []

  if (item.localAddress.startsWith("tcp") || item.localAddress.startsWith("unix") || item.localAddress.startsWith("tcp_udp")) {
    address.push("tcp://" + item.serverIp + ":" + item.remotePort)
  }

  if (item.localAddress.startsWith("udp") || item.localAddress.startsWith("tcp_udp")) {
    address.push("udp://" + item.serverIp + ":" + item.remotePort)
  }
  if (item.localAddress.startsWith("socks5")) {
    address.push("socks5://" + item.serverIp + ":" + item.remotePort)
  }

  if (item.localAddress.startsWith("http")) {
    address.push("http://" + item.serverIp + ":" + item.remotePort)
    if (item.domain) {
      address.push("http://" + item.domain)
      address.push("https://" + item.domain)
      switch (item.safeType) {
        case 0:
          address.push("无防护")
          break;
        case 1:
          address.push("全防护")
          break;
        case 2:
          address.push("半防护")
          break;
      }
    }
  }

  return address
}

onMounted(() => {
  loadDeviceKey();
  loadData()
})
</script>

<style scoped>
.config-detail {
  padding: 14px 18px;
  background: #fbfcfe;
  border-radius: 12px;
  line-height: 1.9;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 0 20px;
}

/* TDesign 会把「最后一个」form-item 的 margin-bottom 归零，
   而外网端口被 .form-grid 包了一层、配置说明是自定义块，都打断了这个判定，
   实测相邻两块间距是 0（贴死）。这里按 24px 的节奏补回来。
   排在末尾的分组（其他选项配置）不需要底部间距，用 :not(:last-child) 跳过 */
.hp-dialog-body .form-grid:not(:last-child) {
  margin-bottom: 24px;
}

/* 说明入口现在是输入框内的前缀图标：平时很淡不抢视线，hover 变品牌色提示「可以点」。
   注意 global.css 里有 `.t-input__prefix-icon .t-icon` 给它定了颜色，
   这里要借 .t-input 抬高一级特异性才盖得住 */
.t-input .hp-help-icon {
  color: var(--hp-text-3);
  font-size: 16px;
  cursor: pointer;
  transition: color .18s ease;
}

.t-input .hp-help-icon:hover {
  color: #2f5bef;
}

/* 图标和输入文字之间留一点呼吸空间 */
:deep(.t-input__prefix-icon) {
  margin-right: 8px;
}

/* 协议之间用虚线分隔，比每个协议套一个绿色警示框轻得多 */
.doc-item + .doc-item {
  margin-top: 14px;
  padding-top: 14px;
  border-top: 1px dashed #e8edf6;
}

.doc-item__title {
  margin-bottom: 8px;
  font-size: 13px;
  font-weight: 600;
  color: var(--hp-text);
}

.doc-item__desc {
  margin: 8px 0 0;
  font-size: 12px;
  line-height: 1.75;
  /* 用次级文字色而非 --hp-text-3：12px 配 #94a3b8 对比度不足，读起来费劲 */
  color: var(--hp-text-2);
}

/* 示例地址统一用等宽字体 + 单一描边色，替换原来的多彩实心标签 */
.doc-tag {
  font-family: 'SFMono-Regular', Consolas, Menlo, monospace;
  font-size: 12px;
}
</style>
