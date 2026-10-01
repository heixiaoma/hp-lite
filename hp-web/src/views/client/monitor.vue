<template>
  <div class="hp-page">
    <div class="hp-toolbar">
      <t-button variant="outline" theme="primary" @click="loadData">
        <template #icon><refresh-icon/></template>
        刷新列表
      </t-button>
    </div>

    <t-table
        class="hp-table"
        row-key="id"
        :data="monitorData"
        :columns="columns"
        :loading="dataLoading"
        empty="暂无数据"
        table-layout="auto"
        stripe
        hover
    >
      <template #server="{ row }">
        <div class="hp-tag-cell">
          <t-tag v-for="item in openAddress(row)" :key="item" theme="primary" variant="light">{{ item }}</t-tag>
        </div>
      </template>

      <template #tunType="{ row }">
        <t-tag variant="light" :theme="row.tunType === 'TCP' ? 'primary' : 'success'">
          {{ row.tunType === 'TCP' ? 'TCP 多路复用' : 'QUIC 多路复用' }}
        </t-tag>
      </template>

      <template #action="{ row }">
        <t-button size="small" variant="outline" theme="primary" @click="show(row)">查看统计</t-button>
      </template>
    </t-table>

    <t-dialog
        v-model:visible="open"
        :header="currentData.name ? currentData.name + ' - 统计图' : '统计图'"
        :footer="false"
        width="90%"
        destroy-on-close
    >
      <monitor-chart v-if="currentData.value" :value="currentData.value" :name="currentData.name"/>
    </t-dialog>
  </div>
</template>

<script setup>
import {onMounted, reactive, ref} from "vue";
import {monitorDetail, monitorList} from "../../api/client/monitor.js";
import {MessagePlugin} from "tdesign-vue-next";
import MonitorChart from "./monitor_chart.vue";
import {RefreshIcon} from 'tdesign-icons-vue-next';

const monitorData = ref([]);
const dataLoading = ref(false);

const loadData = async () => {
  dataLoading.value = true
  let data = await monitorList()
  monitorData.value = data.data
  dataLoading.value = false
}

const columns = [
  {colKey: 'remarks', title: '备注'},
  {colKey: 'domain', title: '域名'},
  {colKey: 'localAddress', title: '内网'},
  {colKey: 'server', title: '外网'},
  {colKey: 'tunType', title: '隧道类型', width: 140},
  {colKey: 'action', title: '操作', width: 120},
];

const open = ref(false)
const currentData = reactive({
  name: null,
  value: null
})

const show = (record) => {
  monitorDetail({id: record.id}).then(res => {
    if (res.data) {
      currentData.value = res.data
      currentData.name = record.remarks
      open.value = true
    } else {
      MessagePlugin.warning("暂无数据")
    }
  })
}

onMounted(async () => {
  await loadData()
})

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
    }
  }

  return address
}
</script>
