<template>
  <div>
    <a-button type="primary" style="margin-bottom: 10px;" @click="loadData">刷新列表</a-button>
    <a-table :loading="dataLoading" :columns="columns" rowKey="id" :data-source="monitorData"
             :locale="{emptyText: '暂无数据'}"
             :scroll="{ x: 'max-content' }">

      <template #bodyCell="{ column ,record}">
        <template v-if="column.key === 'local'">
        {{record.localIp}}:{{record.localPort}}
        </template>
        <template v-if="column.key === 'remote'">
          {{record.serverIp}}:{{record.serverPort}}
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
        <template v-if="column.key === 'action'">
          <a-button type="primary" style="margin-bottom: 5px;margin-left: 5px" @click="show(record)">查看统计</a-button>
        </template>
      </template>
    </a-table>

    <a-modal
        v-model:visible="open"
        title="统计图"
        :footer="null"
        width="100%"
        wrap-class-name="full-modal"
    >
      <monitor_chart v-if="currentData.value" :value="currentData.value" :name="currentData.name"></monitor_chart>
    </a-modal>
  </div>
</template>

<script setup>

import {onMounted, reactive, ref} from "vue";
import {monitorDetail, monitorList} from "../../api/client/monitor.js";
import {message} from "ant-design-vue";
import Monitor_chart from "./monitor_chart.vue";

const monitorData = ref([]);
const dataLoading = ref(false);

const loadData = async () => {
  dataLoading.value=true
  let data = await monitorList()
  monitorData.value = data.data
  dataLoading.value=false
}

const columns = [
  {title: '备注', dataIndex: 'remarks', key: 'remarks'},
  {title: '域名', dataIndex: 'domain', key: 'domain'},
  {title: '内网', dataIndex: 'local', key: 'local'},
  {title: '外网', dataIndex: 'remote', key: 'remote'},
  {title: '隧道类型', dataIndex: 'tunType', key: 'tunType'},
  {title: '操作', key: 'action'},
];

const open=ref(false)
const currentData=reactive({
  name:null,
  value:null
})

const show = (record) => {
  monitorDetail({id:record.id}).then(res=>{
    if (res.data){
      currentData.value=res.data
      currentData.name=record.remarks
      open.value=true
    }else {
      message.warn("暂无数据")
    }
  })
}


onMounted(async () => {
  await loadData()
})


</script>

<style scoped>

</style>
