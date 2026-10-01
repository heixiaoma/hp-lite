<template>
  <div class="qr">
    <div class="qr__canvas">
      <canvas ref="canvasRef"></canvas>
    </div>

    <t-tabs v-if="showText" v-model="tab">
      <t-tab-panel value="key" label="连接码">
        <!-- 连接码是长串，用 t-tag 承载会被截断看不到完整值。
             改成与 Docker 命令一致的代码块：等宽字体整段换行，右侧放复制按钮 -->
        <div class="qr__key">
          <div class="qr__key-value">{{ text }}</div>
          <t-button size="small" variant="outline" theme="primary" class="qr__key-copy" @click="copy">
            <template #icon><copy-icon/></template>
            复制
          </t-button>
        </div>
      </t-tab-panel>

      <t-tab-panel value="docker" label="Docker">
        <p class="qr__label">
          <t-tag theme="success" variant="light">官方源</t-tag>
        </p>
        <pre class="qr__code">docker run --name hp-lite --restart=always -d -e  c={{ text }} heixiaoma/hp-lite:latest</pre>
        <p class="qr__label">
          <t-tag theme="primary" variant="light">阿里源</t-tag>
        </p>
        <pre class="qr__code">docker run --name hp-lite --restart=always -d -e  c={{ text }} registry.cn-shenzhen.aliyuncs.com/heixiaoma/hp-lite:latest</pre>
      </t-tab-panel>

      <t-tab-panel value="win" label="Win命令行">
        <t-divider align="left">临时运行</t-divider>
        <t-steps :current="-1" layout="vertical">
          <t-step-item title="启动" :content="'hp-lite.exe -c ' + text"/>
        </t-steps>
        <t-divider align="left">后台运行</t-divider>
        <t-steps :current="-1" layout="vertical">
          <t-step-item title="安装" :content="'hp-lite.exe -c ' + text + ' -action install'"/>
          <t-step-item title="启动" content="hp-lite.exe -action start"/>
          <t-step-item title="停止" content="hp-lite.exe -action stop"/>
          <t-step-item title="状态查看" content="hp-lite.exe -action status"/>
          <t-step-item title="卸载" content="hp-lite.exe -action uninstall"/>
        </t-steps>
      </t-tab-panel>

      <t-tab-panel value="x86" label="X86命令行">
        <t-divider align="left">临时运行</t-divider>
        <t-steps :current="-1" layout="vertical">
          <t-step-item title="启动" :content="' hp-lite-amd64 -c ' + text"/>
        </t-steps>
        <t-divider align="left">后台运行</t-divider>
        <t-steps :current="-1" layout="vertical">
          <t-step-item title="安装" :content="'hp-lite-amd64 -c ' + text + ' -action install'"/>
          <t-step-item title="启动" content="hp-lite-amd64 -action start"/>
          <t-step-item title="停止" content="hp-lite-amd64 -action stop"/>
          <t-step-item title="状态查看" content="hp-lite-amd64 -action status"/>
          <t-step-item title="卸载" content="hp-lite-amd64 -action uninstall"/>
        </t-steps>
      </t-tab-panel>

      <t-tab-panel value="arm64" label="Arm64命令行">
        <t-divider align="left">临时运行</t-divider>
        <t-steps :current="-1" layout="vertical">
          <t-step-item title="启动" :content="'hp-lite-arm64 -c ' + text"/>
        </t-steps>
        <t-divider align="left">后台运行</t-divider>
        <t-steps :current="-1" layout="vertical">
          <t-step-item title="安装" :content="'hp-lite-arm64 -c ' + text + ' -action install'"/>
          <t-step-item title="启动" content="hp-lite-arm64 -action start"/>
          <t-step-item title="停止" content="hp-lite-arm64 -action stop"/>
          <t-step-item title="状态查看" content="hp-lite-arm64 -action status"/>
          <t-step-item title="卸载" content="hp-lite-arm64 -action uninstall"/>
        </t-steps>
      </t-tab-panel>
    </t-tabs>
  </div>
</template>

<script setup>
import {onMounted, ref} from 'vue';
import QRCode from 'qrcode';
import {MessagePlugin} from 'tdesign-vue-next';
import {CopyIcon} from 'tdesign-icons-vue-next';
import {copyText} from '../../utils/clipboard';

const props = defineProps({
  text: {
    type: String,
    required: true
  },
  showText: {
    type: Boolean,
    default: true,
    required: false
  }
});

const canvasRef = ref(null);
const tab = ref('key');

const copy = async () => {
  const ok = await copyText(props.text)
  ok ? MessagePlugin.success('连接码已复制') : MessagePlugin.error('复制失败，请手动选中复制')
}

onMounted(() => {
  if (!canvasRef.value || !props.text) return;
  QRCode.toCanvas(canvasRef.value, props.text, function (error) {
    if (error) console.error(error)
  })
})
</script>

<style scoped>
.qr {
  text-align: center;
}

.qr__canvas {
  display: flex;
  justify-content: center;
  padding-bottom: 16px;
}

/* 连接码是长串，原来用 t-tag 承载会被截断。
   改成与下方 Docker 命令一致的代码块：等宽字体整段换行，右侧放复制按钮 */
.qr__key {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 10px 12px;
  background: #f8fafc;
  border: 1px solid var(--hp-border);
  border-radius: 10px;
}

.qr__key-value {
  flex: 1;
  min-width: 0;
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
  font-size: 12.5px;
  line-height: 1.6;
  text-align: left;
  word-break: break-all;
}

.qr__key-copy {
  flex: none;
}

.qr__label {
  margin: 0 0 6px;
  text-align: left;
}

.qr__code {
  margin: 0 0 14px;
  padding: 10px 12px;
  background: #f8fafc;
  border: 1px solid var(--hp-border);
  border-radius: 10px;
  font-size: 12.5px;
  line-height: 1.6;
  text-align: left;
  white-space: pre-wrap;
  word-break: break-all;
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
}

.qr :deep(.t-tabs__content) {
  max-height: 38vh;
  overflow-y: auto;
  text-align: left;
  padding-right: 4px;
}
</style>
