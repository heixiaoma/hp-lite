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
          <template v-if="column.key === 'rateLimit'">
            <div v-if="record.rateLimit<=0">
              不限制
            </div>
            <div v-else>
              {{record.rateLimit}}
            </div>
          </template>
          <template v-if="column.key === 'inLimit'">
            <div v-if="record.inLimit<=0">
              不限制
            </div>
            <div v-else>
              {{record.inLimit}}
            </div>
          </template>
          <template v-if="column.key === 'outLimit'">
            <div v-if="record.outLimit<=0">
              不限制
            </div>
            <div v-else>
              {{record.outLimit}}
            </div>
          </template>

          <template v-if="column.key === 'user'">
            <div v-if="!record.userDesc&&!record.username">
              自用
            </div>
            <div v-else>
              <div>归属用户：{{record.username}}</div>
              <div>归属用户备注：{{record.userDesc}}</div>
            </div>
          </template>

          <template v-if="column.key === 'action'">
            <a-button  class="btn edit" style="margin-bottom: 5px;margin-left: 5px" @click="edit(record)">编辑</a-button>
            <a-button  class="btn delete" style="margin-bottom: 5px;margin-left: 5px" @click="removeData(record)">删除</a-button>
          </template>
        </template>
      </a-table>
  </div>

  <div>
    <a-modal  v-model:visible="addVisible" title="添加">
      <div class="config-info">
      <a-form :model="formState" ref="formTable" :layout="'vertical'" >
        <a-form-item label="规则名字" name="ruleName"  :rules="[{ required: true, message: '规则名字'}]">
          <a-input style="width: 90%" v-model:value="formState.ruleName" placeholder="规则名字"/>
        </a-form-item>
        <a-form-item label="规则" name="rule" :rules="[{ required: true, message: '规则'}]">
          <a-textarea style="width: 90%" v-model:value="formState.rule" placeholder="ModSecurity SecLang 规则集，或者 OWASP 核心规则集 v4"/>
        </a-form-item>
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
import {getSafe, removeSafe, saveSafe} from "../../api/client/safe";
import {onMounted, reactive, ref} from "vue";
import {notification} from "ant-design-vue";

const listData = ref();
const formTable = ref();
const dataLoading = ref(false);
const addVisible = ref(false);
const formState = reactive({
  ruleName:"",
  rule:"",
  id:""
})
const pagination = reactive({
  total: 0,
  current: 1,
  pageSize: 10,
});

const loadData = () => {
  dataLoading.value = true
  getSafe({
    current: pagination.current,
    pageSize: pagination.pageSize
  }).then(res => {
    dataLoading.value = false
    listData.value = res.data.records
    pagination.total = res.data.total
  })
}

const removeData = (item) => {
  removeSafe({
    id: item.id
  }).then(res => {
    notification.open({
      message: res.msg,
    })
    loadData()
  })
}



const edit = (itemOld) => {
  const item=JSON.parse(JSON.stringify(itemOld))
  formState.rule = item.rule
  formState.ruleName = item.ruleName
  formState.id = item.id
  addVisible.value=true
}

const columns = [
  {title: '编号', dataIndex: 'id', key: 'id'},
  {title: '规则名字', dataIndex: 'ruleName', key: 'ruleName'},
  {title: '规则内容', dataIndex: 'rule', key: 'rule'},
  {title: '归属', dataIndex: 'user', key: 'user'},
  {title: '操作', key: 'action'},
];

const handleTableChange = (item) => {
  pagination.current = item.current
  pagination.pageSize = item.pageSize
  pagination.total = item.total
  loadData()
}

const addModal = () => {
  formState.rule = ""
  formState.ruleName =""
  formState.id = undefined
  addVisible.value = true
}

const addOk = () => {
  formTable.value.validate().then(valid => {
    saveSafe({...formState}).then(res => {
      notification.open({
        message: res.msg,
      })
      loadData()
      addVisible.value = false
    })
  })
}

onMounted(() => {
  loadData()
})

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
