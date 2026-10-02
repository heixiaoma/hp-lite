<script setup>
import {computed, nextTick, onMounted, reactive, ref, watch} from "vue";
import axios from "axios";
import Giscus from '@giscus/vue';
import {getGithubToken} from "../../api/client/giscus.js";
import {AddIcon, ChatIcon} from 'tdesign-icons-vue-next';

const formTable = ref()
const spinning = ref(true)
const numberData = ref(null)
const visible = ref(false)
const createVisible = ref(false)
const addVisible = ref(false)

// 配置信息 - 建议在生产环境使用环境变量
const config = reactive({
  owner: "heixiaoma",
  repo: "hp-lite",
  repoId: 'R_kgDOPIQq1w',
  categoryId: 'DIC_kwDOPIQq184CuaDS',
  token: '',
  graphqlUrl: "https://api.github.com/graphql"
});

const formState = reactive({
  title: ''
})

const formRules = {
  title: [{required: true, message: '请填写话题内容', type: 'error'}],
};

const jumpAuth = () => {
  location.href = "https://giscus.app/api/oauth/authorize?redirect_uri=" + location.href
}

const addOk = async () => {
  const result = await formTable.value?.validate()
  if (result !== true) return

  createVisible.value = true;
  addVisible.value = false
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
        {query: getDiscussionsQuery, variables},
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

const filterOptions = [
  {label: '全部', value: 'ALL'},
  {label: '打开', value: 'OPEN'},
  {label: '已关闭', value: 'CLOSED'},
];

const onFilterChange = (value) => {
  const next = Array.isArray(value) ? value[value.length - 1] : value;
  if (next) changeStateFilter(next);
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

const loadToken = () => {
  const item = localStorage.getItem("giscus-session");
  if (item) {
    getGithubToken({session: JSON.parse(item)}).then(res => {
      if (res.code === 200) {
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

  nextTick(() => {
    setTimeout(function () {
      loadToken()
      spinning.value = false;
    }, 1000)
  });
});

const showDis = (item) => {
  numberData.value = item.number
  visible.value = true
}
</script>

<template>
  <div class="teach">
    <t-loading :loading="spinning" size="large" text="加载中">
      <!-- giscus 授权入口（隐藏） -->
      <div style="display:none">
        <Giscus/>
      </div>

      <!-- 未授权 -->
      <div v-if="!config.token" class="auth-empty">
        <div class="auth-empty__icon">
          <chat-icon size="28px"/>
        </div>
        <h3 class="auth-empty__title">未检查到 Github 授权</h3>
        <p class="auth-empty__desc">授权后即可参与穿透交流区的讨论</p>
        <t-button theme="primary" @click="jumpAuth">去授权</t-button>
      </div>

      <div v-else class="discussions">
        <div class="discussions__head">
          <div>
            <h2 class="discussions__title">讨论组列表</h2>
            <span class="discussions__tip">请文明交流，乱来者拉黑</span>
          </div>

          <div class="discussions__filter">
            <t-button theme="primary" @click="addVisible = true">
              <template #icon><add-icon/></template>
              创建话题
            </t-button>
            <t-check-tag-group
                :value="[state.filters.state]"
                :options="filterOptions"
                :multiple="false"
                @change="onFilterChange"
            />
          </div>
        </div>

        <!-- 错误信息 -->
        <t-alert v-if="state.status.error" theme="error" class="discussions__alert">
          <template #message>
            <div class="error-row">
              <span>⚠️ {{ state.status.error }}</span>
              <t-button size="small" theme="danger" variant="outline" @click="reloadDiscussions">重试</t-button>
            </div>
          </template>
        </t-alert>

        <!-- 加载状态 -->
        <div v-if="state.status.loading" class="state-box">
          <t-loading size="large" text="加载中..."/>
        </div>

        <!-- 空状态 -->
        <div v-else-if="discussions.length === 0 && !state.status.isInitialLoad" class="state-box">
          <t-empty description="没有找到符合条件的讨论"/>
          <t-button variant="outline" theme="primary" @click="reloadDiscussions">刷新</t-button>
        </div>

        <!-- 讨论组列表 -->
        <div v-else-if="discussions.length > 0" class="discussion-list">
          <article
              class="discussion-item"
              v-for="discussion in discussions"
              :key="discussion.id"
              @click="showDis(discussion)"
          >
            <div class="discussion-item__head">
              <div class="discussion-item__title">#{{ discussion.number }} {{ discussion.title }}</div>
              <a :href="discussion.url" target="_blank" rel="noopener noreferrer" @click.stop class="discussion-item__link">
                在 Github 打开
              </a>
            </div>

            <div v-if="discussion.labels.nodes.length" class="discussion-item__labels">
              <span
                  class="label"
                  v-for="label in discussion.labels.nodes"
                  :key="label.name"
                  :style="{
                    backgroundColor: `#${label.color}`,
                    color: getContrastColor(label.color)
                  }"
              >{{ label.name }}</span>
            </div>

            <div class="discussion-item__meta">
              <span class="author">
                <img :src="discussion.author.avatarUrl" :alt="discussion.author.login" class="avatar">
                {{ discussion.author.login }}
              </span>
              <span>创建于 {{ formatDate(discussion.createdAt) }}</span>
              <span>{{ discussion.comments.totalCount }} 条评论</span>
            </div>

            <p class="discussion-item__excerpt">
              {{ discussion.bodyText.length > 150 ? discussion.bodyText.slice(0, 150) + '...' : discussion.bodyText }}
            </p>
          </article>
        </div>

        <!-- 分页控制 -->
        <div v-if="!state.status.isInitialLoad" class="discussions__foot">
          <t-button variant="outline" :disabled="state.status.loading" @click="reloadDiscussions">刷新</t-button>
          <t-button
              theme="primary"
              variant="outline"
              :disabled="!state.pagination.hasNextPage || state.status.loading"
              @click="loadNextPage"
          >加载更多
          </t-button>
          <span class="page-info">
            第 {{ state.pagination.currentPage }} 页（共 {{ state.pagination.totalCount }} 个讨论）
          </span>
        </div>

        <!-- 话题详情 -->
        <t-drawer
            v-model:visible="visible"
            header="交流区"
            size="100%"
            destroy-on-close
            @close="reloadDiscussions"
        >
          <Giscus
              :repo="config.owner + '/' + config.repo"
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
        </t-drawer>

        <!-- 创建话题 -->
        <t-drawer
            v-model:visible="createVisible"
            header="请在下面评论一句"
            size="100%"
            destroy-on-close
            @close="reloadDiscussions"
        >
          <Giscus
              :repo="config.owner + '/' + config.repo"
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
        </t-drawer>

        <t-dialog
            v-model:visible="addVisible"
            header="创建话题"
            confirm-btn="确定"
            cancel-btn="取消"
            width="560px"
            @confirm="addOk"
        >
          <t-form :data="formState" ref="formTable" :rules="formRules" layout="vertical" label-align="top">
            <t-form-item label="话题内容" name="title">
              <t-textarea
                  v-model="formState.title"
                  :autosize="{ minRows: 4, maxRows: 8 }"
                  placeholder="请描述的问题或者话题内容"
              />
            </t-form-item>
          </t-form>
        </t-dialog>
      </div>
    </t-loading>
  </div>
</template>

<style scoped>
.teach {
  min-height: 40vh;
}

/* 未授权 */
.auth-empty {
  text-align: center;
  padding: 70px 20px;
  background: #fafbfe;
  border: 1px dashed var(--hp-border);
  border-radius: var(--hp-radius);
}

.auth-empty__icon {
  width: 58px;
  height: 58px;
  margin: 0 auto 16px;
  border-radius: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--td-brand-color-1);
  color: var(--td-brand-color);
}

.auth-empty__title {
  margin: 0 0 8px;
  font-size: 18px;
}

.auth-empty__desc {
  margin: 0 0 20px;
  color: var(--hp-text-2);
  font-size: 14px;
}

/* 列表头部 */
.discussions__head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  margin-bottom: 18px;
}

.discussions__title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.discussions__tip {
  font-size: 12px;
  color: var(--hp-text-3);
}

.discussions__filter {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px;
}

.discussions__alert {
  margin-bottom: 16px;
}

.error-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
}

.state-box {
  text-align: center;
  padding: 46px 20px;
  color: var(--hp-text-2);
}

.state-box :deep(.t-button) {
  margin-top: 14px;
}

.discussion-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
  margin-bottom: 24px;
}

.discussion-item {
  border: 1px solid var(--hp-border);
  border-radius: 14px;
  padding: 16px 18px;
  cursor: pointer;
  transition: box-shadow .22s ease, border-color .22s ease, transform .22s ease;
  background: #fff;
}

.discussion-item:hover {
  border-color: #c7d6ff;
  box-shadow: var(--hp-shadow);
  transform: translateY(-2px);
}

.discussion-item__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 14px;
}

.discussion-item__title {
  font-size: 16px;
  font-weight: 600;
  color: var(--td-brand-color);
}

.discussion-item__link {
  flex: none;
  font-size: 13px;
  color: var(--hp-text-3);
}

.discussion-item__link:hover {
  color: var(--td-brand-color);
}

.discussion-item__labels {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin: 10px 0;
}

.label {
  padding: 2px 9px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
}

.discussion-item__meta {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
  color: var(--hp-text-3);
  font-size: 13px;
}

.author {
  display: flex;
  align-items: center;
  gap: 6px;
}

.avatar {
  width: 20px;
  height: 20px;
  border-radius: 50%;
}

.discussion-item__excerpt {
  margin: 12px 0 0;
  padding-top: 12px;
  border-top: 1px solid #f1f3f8;
  color: var(--hp-text-2);
  font-size: 14px;
  line-height: 1.7;
}

.discussions__foot {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
  padding-top: 18px;
  border-top: 1px solid var(--hp-border);
}

.page-info {
  color: var(--hp-text-3);
  font-size: 13px;
}
</style>
