<template>
  <div id="qr">
    <canvas id="canvas"></canvas>
    <a-tabs v-if="showText" >
      <a-tab-pane class="info-panel" key="1" tab="连接码">{{text}}</a-tab-pane>
      <a-tab-pane class="info-panel" key="2" tab="Docker">
        <a-tag color="#87d068">官方源</a-tag>
        <p>docker run --name hp-lite --restart=always -d -e  c={{text}} heixiaoma/hp-lite:latest</p>
        <a-tag color="#2db7f5">阿里源</a-tag>
        <p>docker run --name hp-lite --restart=always -d -e  c={{text}} registry.cn-shenzhen.aliyuncs.com/heixiaoma/hp-lite:latest</p>
      </a-tab-pane>
      <a-tab-pane class="info-panel" key="3" tab="Win命令行">
        <a-divider>临时运行</a-divider>
        <a-steps :current="-1" direction="vertical">
          <a-step title="安装" :description="'hp-lite.exe -c '+text+''" />
        </a-steps>
        <a-divider>后台运行</a-divider>
        <a-steps :current="-1" direction="vertical">
          <a-step title="安装" :description="'hp-lite.exe -c '+text+' -action install'" />
          <a-step title="启动" description="hp-lite.exe -action start" />
          <a-step title="停止" description="hp-lite.exe -action stop" />
          <a-step title="状态查看" description="hp-lite.exe -action status" />
          <a-step title="卸载" description="hp-lite.exe -action uninstall" />
        </a-steps>
      </a-tab-pane>
      <a-tab-pane class="info-panel" key="4" tab="X86命令行">
        <a-divider>临时运行</a-divider>
        <a-steps :current="-1" direction="vertical">
          <a-step title="安装" :description="' hp-lite-amd64 -c '+text+''" />
        </a-steps>
        <a-divider>后台运行</a-divider>
        <a-steps :current="-1" direction="vertical">
          <a-step title="安装" :description="'hp-lite-amd64 -c '+text+' -action install'" />
          <a-step title="启动" description="hp-lite-amd64 -action start" />
          <a-step title="停止" description="hp-lite-amd64 -action stop" />
          <a-step title="状态查看" description="hp-lite-amd64 -action status" />
          <a-step title="卸载" description="hp-lite-amd64 -action uninstall" />
        </a-steps>
      </a-tab-pane>
      <a-tab-pane class="info-panel" key="5" tab="Arm64命令行">
        <a-divider>临时运行</a-divider>
        <a-steps :current="-1" direction="vertical">
          <a-step title="安装" :description="'hp-lite-arm64 -c '+text+''" />
        </a-steps>
        <a-divider>后台运行</a-divider>
        <a-steps :current="-1" direction="vertical">
          <a-step title="安装" :description="'hp-lite-arm64 -c '+text+' -action install'" />
          <a-step title="启动" description="hp-lite-arm64 -action start" />
          <a-step title="停止" description="hp-lite-arm64 -action stop" />
          <a-step title="状态查看" description="hp-lite-arm64 -action status" />
          <a-step title="卸载" description="hp-lite-arm64 -action uninstall" />
        </a-steps>
      </a-tab-pane>
    </a-tabs>
  </div>
</template>

<script>
import QRCode from 'qrcode';

export default {
  props: {
    text: {
      type: String,
      required: true
    } ,
    showText: {
      type: Boolean,
      default:true,
      required: false
    }
  },
  mounted() {
    let canvas = document.getElementById('canvas')
    QRCode.toCanvas(canvas, this.text, function (error) {
      if (error) console.error(error)
      console.log('success!');
    })
  }
}
</script>
<style>
#qr {
  text-align: center;
}
.info-panel{
  text-align: left;
  max-height: 35vh;
  overflow-y: scroll;
  /* Firefox：设置滚动条为极简样式 + 窄宽度 */
  scrollbar-width: thin;
  /* Firefox 可选：设置滚动条颜色（轨道/滑块） */
  scrollbar-color: #ccc #f5f5f5;
}
/* WebKit 内核浏览器（Chrome/Safari/Edge）自定义极简滚动条 */
.info-panel::-webkit-scrollbar {
  /* 滚动条宽度（纵向是width，横向是height） */
  width: 4px; /* 越小越极简，建议3-6px */
}

/* 滚动条轨道（背景） */
.info-panel::-webkit-scrollbar-track {
  background: #f5f5f5;
  border-radius: 2px; /* 圆角更美观 */
}

/* 滚动条滑块（可拖动的部分） */
.info-panel::-webkit-scrollbar-thumb {
  background: #ccc;
  border-radius: 2px; /* 圆角匹配轨道 */
  /* 鼠标悬停时的滑块样式（可选，提升交互） */
}
.info-panel::-webkit-scrollbar-thumb:hover {
  background: #999;
}
</style>
