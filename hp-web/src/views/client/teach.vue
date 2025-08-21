<script setup>
import {onMounted, reactive, ref, watch, computed, nextTick} from "vue";
import axios from "axios";
import Giscus from '@giscus/vue';
import {getGithubToken} from "../../api/client/giscus.js";

const formTable =ref()
const spinning =ref(true)
const numberData =ref(null)
const visible = ref(false)
const createVisible = ref(false)
const addVisible = ref(false)
// 配置信息 - 建议在生产环境使用环境变量
const config = reactive({
  owner: "heixiaoma",
  repo: "hp-lite",
  repoId:'R_kgDOPIQq1w',
  categoryId:'DIC_kwDOPIQq184CuaDS',
  token: '',
  graphqlUrl: "https://api.github.com/graphql"
});

const formState = reactive({
  title:''
})


const jumpAuth=()=>{
  location.href="https://giscus.app/api/oauth/authorize?redirect_uri="+location.href
}


const addOk=()=>{
  formTable.value.validate().then(res => {
    createVisible.value=true;
    addVisible.value=false
  })
}

// 状态管理
const state = reactive({
  pagination: {
    perPage: 50, // 每页数量，最大100
    currentPage: 1,
    endCursor: null, // 用于分页的游标
    hasNextPage: true, // 是否有下一页
    totalCount: 0 // 总讨论数
  },
  status: {
    loading: false,
    error: null,
    isInitialLoad: true // 是否是首次加载
  },
  filters: {
    state: 'ALL' // 讨论状态筛选: ALL, OPEN, CLOSED
  }
});

const discussions = ref([]); // 讨论组数据

// GraphQL查询 - 包含状态查询
const getDiscussionsQuery = `
query GetDiscussions($owner: String!, $repo: String!, $first: Int!, $after: String, $states: [DiscussionState!]) {
  repository(owner: $owner, name: $repo) {
    discussions(
      first: $first,
      after: $after,
      orderBy: { field: CREATED_AT, direction: DESC },
      states: $states
    ) {
      nodes {
        id
        number
        title
        bodyText
        author {
          login
          avatarUrl
        }
        createdAt
        updatedAt
        url
        labels(first: 5) {
          nodes {
            name
            color
          }
        }
        comments {
          totalCount
        }
      }
      pageInfo {
        endCursor
        hasNextPage
      }
      totalCount
    }
  }
}
`;

// 计算状态筛选条件
const stateFilter = computed(() => {
  return state.filters.state === 'ALL' ? null : [state.filters.state];
});

// 计算标签文字对比度
const getContrastColor = (hexColor) => {
  // 转换16进制颜色到RGB
  const r = parseInt(hexColor.substring(0, 2), 16);
  const g = parseInt(hexColor.substring(2, 4), 16);
  const b = parseInt(hexColor.substring(4, 6), 16);

  // 计算亮度
  const luminance = (0.299 * r + 0.587 * g + 0.114 * b) / 255;

  // 根据亮度返回黑或白
  return luminance > 0.5 ? '#000000' : '#ffffff';
};

// 重置讨论数据
const resetDiscussions = () => {
  discussions.value = [];
  state.pagination.endCursor = null;
  state.pagination.currentPage = 1;
  state.pagination.hasNextPage = true;
};

// 加载讨论组数据
const loadDiscussions = async () => {
  console.log("---")
  // 防止重复加载和无更多数据时继续加载
  if (state.status.loading || !state.pagination.hasNextPage) return;

  state.status.loading = true;
  state.status.error = null;

  try {
    const variables = {
      owner: config.owner,
      repo: config.repo,
      first: state.pagination.perPage,
      after: state.pagination.endCursor,
      states: stateFilter.value
    };

    const response = await axios.post(
        config.graphqlUrl,
        { query: getDiscussionsQuery, variables },
        {
          headers: {
            "Authorization": `Bearer ${config.token}`,
            "Content-Type": "application/json"
          }
        }
    );

    // 处理GraphQL错误
    if (response.data.errors) {
      const errorMessages = response.data.errors.map(err => err.message).join(", ");
      throw new Error(`GraphQL Error: ${errorMessages}`);
    }

    // 处理空响应
    if (!response.data.data?.repository?.discussions) {
      throw new Error("未能获取讨论数据，请检查仓库信息");
    }

    const discussionData = response.data.data.repository.discussions;

    // 更新总计数
    state.pagination.totalCount = discussionData.totalCount;

    // 第一页清空数据，后续页追加数据
    if (state.pagination.currentPage === 1) {
      discussions.value = discussionData.nodes;
    } else {
      discussions.value = [...discussions.value, ...discussionData.nodes];
    }

    // 更新分页信息
    state.pagination.endCursor = discussionData.pageInfo.endCursor;
    state.pagination.hasNextPage = discussionData.pageInfo.hasNextPage;

  } catch (err) {
    state.status.error = err.message || "加载讨论组失败";
    console.error("加载讨论组错误:", err);
  } finally {
    state.status.loading = false;
    state.status.isInitialLoad = false;
  }
};

// 加载下一页
const loadNextPage = () => {
  if (state.pagination.hasNextPage && !state.status.loading) {
    state.pagination.currentPage++;
    loadDiscussions();
  }
};

// 重新加载当前筛选条件下的讨论
const reloadDiscussions = () => {
  resetDiscussions();
  loadDiscussions();
};

// 更改状态筛选条件
const changeStateFilter = (newState) => {
  if (state.filters.state !== newState) {
    state.filters.state = newState;
    reloadDiscussions();
  }
};

// 格式化日期
const formatDate = (dateString) => {
  return new Date(dateString).toLocaleString();
};

// 监听页码变化，自动加载数据
watch(
    () => state.pagination.currentPage,
    (newPage, oldPage) => {
      if (newPage > 1 && newPage > oldPage) {
        loadDiscussions();
      }
    }
);

// 监听筛选条件变化
watch(
    () => state.filters.state,
    () => {
      // 状态变化时已在changeStateFilter中处理
    }
);


const loadToken=()=>{
 var item = localStorage.getItem("giscus-session");
 if (item){
   getGithubToken({session:JSON.parse(item)}).then(res=>{
     if (res.code===200){
       config.token = JSON.parse(res.data).token
       loadDiscussions()
     }
   })
 }
}

// 初始加载
onMounted(() => {
  window.addEventListener('message', (event) => {
    console.log(JSON.stringify(event.data));
  });

  nextTick(()=>{
    setTimeout(function () {
      loadToken()
      spinning.value=false;
    },1000)
  })



});
const showDis=(item)=>{
  numberData.value=item.number
  visible.value=true
  console.log(item)
}

</script>

<template>


  <div>
    <a-spin :spinning="spinning">
    <div style="display:none">
      <Giscus/>
    </div>
    <div v-if="!config.token">
      <a-result status="warning" title="未检查到Github授权,授权后进行讨论">
        <template #extra>
          <a-button key="console" class="btn view"  @click="jumpAuth">去授权</a-button>
        </template>
      </a-result>
    </div>


  <div class="discussions-container" v-else>
    <div class="header-section">
      <h2>讨论组列表  <span style="font-size: 12px">请文明交流,乱来者拉黑</span> </h2>
      <div class="filter-controls">
        <button class="btn edit"
            @click="addVisible=true"
        >
          创建话题
        </button>

        <span class="filter-label">状态筛选:</span>
        <button
            @click="changeStateFilter('ALL')"
            :class="{ active: state.filters.state === 'ALL' }"
        >
          全部
        </button>
        <button
            @click="changeStateFilter('OPEN')"
            :class="{ active: state.filters.state === 'OPEN' }"
        >
          打开
        </button>
        <button
            @click="changeStateFilter('CLOSED')"
            :class="{ active: state.filters.state === 'CLOSED' }"
        >
          已关闭
        </button>
      </div>
    </div>

    <!-- 错误信息 -->
    <div v-if="state.status.error" class="error-message">
      ⚠️ {{ state.status.error }}
      <button class="retry-button" @click="reloadDiscussions">重试</button>
    </div>

    <!-- 加载状态 -->
    <div v-if="state.status.loading" class="loading-indicator">
      <div class="spinner"></div>
      <p>加载中...</p>
    </div>

    <!-- 空状态 -->
    <div v-if="!state.status.loading && !state.status.error && discussions.length === 0 && !state.status.isInitialLoad" class="empty-state">
      <p>没有找到符合条件的讨论</p>
      <button @click="reloadDiscussions">刷新</button>
    </div>

    <!-- 讨论组列表 -->
    <div class="discussions-list" v-else-if="discussions.length > 0">
      <div
          class="discussion-item"
          v-for="discussion in discussions"
          :key="discussion.id"
          @click="showDis(discussion)"
      >
        <div class="discussion-header">
          <h3>
            <a :href="discussion.url" target="_blank" rel="noopener noreferrer">
              #{{ discussion.number }} {{ discussion.title }}
            </a>
          </h3>
          <div class="labels">
            <span
                class="label"
                v-for="label in discussion.labels.nodes"
                :key="label.name"
                :style="{
                backgroundColor: `#${label.color}`,
                color: getContrastColor(label.color)
              }"
            >
              {{ label.name }}
            </span>
          </div>
        </div>

        <div class="discussion-meta">
          <div class="author">
            <img
                :src="discussion.author.avatarUrl"
                :alt="discussion.author.login"
                class="avatar"
            >
            <span>{{ discussion.author.login }}</span>
          </div>
          <span class="date">创建于 {{ formatDate(discussion.createdAt) }}</span>
          <span class="comments">{{ discussion.comments.totalCount }} 条评论</span>
        </div>

        <div class="discussion-excerpt">
          {{ discussion.bodyText.length > 150 ? discussion.bodyText.slice(0, 150) + '...' : discussion.bodyText }}
        </div>
      </div>
    </div>

    <!-- 分页控制 -->
    <div class="pagination-controls" v-if="!state.status.isInitialLoad">
      <button
          @click="reloadDiscussions"
          :disabled="state.status.loading"
          class="refresh-btn"
      >
        刷新
      </button>
      <button
          @click="loadNextPage"
          :disabled="!state.pagination.hasNextPage || state.status.loading"
          class="load-more-btn"
      >
        加载更多
      </button>
      <span class="page-info">
        第 {{ state.pagination.currentPage }} 页 (共 {{ state.pagination.totalCount }} 个讨论)
      </span>
    </div>



    <a-drawer
        width="100vw"
        @close="reloadDiscussions"
        v-model:visible="visible"
        class="custom-class"
        title="交流区"
    >
      <Giscus
          :repo="config.owner+'/'+config.repo"
          :repo-id="config.repoId"
          category="General"
          :category-id="config.categoryId"
          mapping="number"
          :term="numberData"
          strict="1"
          reactions-enabled="0"
          emit-metadata="0"
          input-position="bottom"
          theme="preferred_color_scheme"
          lang="zh-CN"
      />
    </a-drawer>


    <a-drawer
        width="100vw"
        @close="reloadDiscussions"
        v-model:visible="createVisible"
        class="custom-class"
        title="请在下面评论一句"
    >
      <Giscus
          :repo="config.owner+'/'+config.repo"
          :repo-id="config.repoId"
          category="General"
          :category-id="config.categoryId"
          mapping="specific"
          :term="formState.title"
          strict="1"
          reactions-enabled="0"
          emit-metadata="0"
          input-position="bottom"
          theme="preferred_color_scheme"
          lang="zh-CN"
      />
    </a-drawer>

    <div>
      <a-modal  v-model:visible="addVisible" title="添加"
      >
        <a-form :model="formState" ref="formTable" :layout="'vertical'" >
          <a-form-item label="话题内容 " name="title"  :rules="[{ required: true, message: 'title'}]">
            <a-textarea v-model:value="formState.title" placeholder="请描述的问题或者话题内容"/>
          </a-form-item>
        </a-form>
        <template #footer>
          <a-button class="btn view" @click="addVisible=!addVisible">取消</a-button>
          <a-button class="btn edit" @click="addOk">确定</a-button>
        </template>
      </a-modal>
    </div>


  </div>
    </a-spin>
  </div>
</template>

<style scoped>

.discussions-container {
  max-width: 1000px;
  margin: 0 auto;
  padding: 20px;
}

.header-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-wrap: wrap;
  gap: 15px;
}

.filter-controls {
  display: flex;
  align-items: center;
  gap: 10px;
}

.filter-label {
  color: #6a737d;
  font-size: 14px;
}

.filter-controls button {
  padding: 4px 10px;
  font-size: 13px;
}

.filter-controls button.active {
  background-color: #0366d6;
  color: white;
  border-color: #0366d6;
}

.error-message {
  color: #dc3545;
  background-color: #f8d7da;
  padding: 12px 15px;
  border-radius: 6px;
  margin-bottom: 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.retry-button {
  padding: 4px 10px;
  background-color: #dc3545;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.retry-button:hover {
  background-color: #c82333;
}

.loading-indicator {
  text-align: center;
  padding: 40px 20px;
  color: #666;
}

.spinner {
  width: 40px;
  height: 40px;
  margin: 0 auto 15px;
  border: 4px solid #f3f3f3;
  border-top: 4px solid #0366d6;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: #6a737d;
  border: 1px dashed #e1e4e8;
  border-radius: 6px;
  margin-bottom: 20px;
}

.empty-state button {
  margin-top: 15px;
}

.discussions-list {
  display: flex;
  flex-direction: column;
  gap: 20px;
  margin-bottom: 30px;
}

.discussion-item {
  overflow: auto;
  border: 1px solid #e1e4e8;
  border-radius: 6px;
  padding: 15px;
  transition: box-shadow 0.2s, border-color 0.2s;
  cursor: pointer;
}

.discussion-item:hover {
  box-shadow: 0 3px 10px rgba(0, 0, 0, 0.1);
  border-color: #d1d5da;
}

.discussion-header {
  margin-bottom: 10px;
}

.discussion-header h3 {
  margin: 0 0 10px 0;
  font-size: 18px;
}

.discussion-header a {
  color: #0366d6;
  text-decoration: none;
  transition: color 0.2s;
}

.discussion-header a:hover {
  text-decoration: underline;
  color: #005cc5;
}

.state-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  margin-bottom: 8px;
}

.state-badge.open {
  background-color: #dcffe4;
  color: #137333;
}

.state-badge.closed {
  background-color: #ffebe9;
  color: #d73a49;
}

.labels {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 10px;
}

.label {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.discussion-meta {
  display: flex;
  align-items: center;
  gap: 15px;
  color: #6a737d;
  font-size: 14px;
  margin-bottom: 10px;
  flex-wrap: wrap;
}

.author {
  display: flex;
  align-items: center;
  gap: 5px;
}

.avatar {
  width: 20px;
  height: 20px;
  border-radius: 50%;
}

.discussion-excerpt {
  color: #24292e;
  font-size: 14px;
  line-height: 1.5;
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px solid #f1f1f1;
}

.pagination-controls {
  display: flex;
  align-items: center;
  gap: 15px;
  justify-content: center;
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid #e1e4e8;
}

button {
  padding: 6px 12px;
  border: 1px solid #e1e4e8;
  border-radius: 4px;
  background-color: #fff;
  cursor: pointer;
  font-size: 14px;
  transition: background-color 0.2s;
}

button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

button:hover:not(:disabled) {
  background-color: #f6f8fa;
}

.page-info {
  color: #6a737d;
  font-size: 14px;
}

/* 响应式调整 */
@media (max-width: 768px) {
  .header-section {
    flex-direction: column;
    align-items: flex-start;
  }

  .discussion-meta {
    flex-direction: column;
    align-items: flex-start;
    gap: 5px;
  }

  .pagination-controls {
    flex-wrap: wrap;
  }
}
</style>
