<template>
  <div class="menu-head">
    <div class="data menu-data">
      <div style="text-align: left;cursor: pointer;line-height: 2.5rem">
        <img src="/logo-back.png" style="width:45px">
      </div>
      <div style="text-align: right;display: flex;justify-content: right">
        <div class="menu-btn">
          <a-button type="dashed" size="large" ghost><a href="/#/home/login">登陆</a></a-button>
        </div>
      </div>
    </div>
  </div>
  <div class="one" style="  height: 100%;
  width: 100%;">
    <div style="text-align: center;color: #ffffff;margin-top: 5rem">
      <div style="font-size: 44px">HP-Lite内网穿透</div>
      <div style="font-size: 18px;margin-top: 1rem">
        无需公网IP、无需路由器端口映射，让内网应用随时可以
        通过域名进行外网访问
      </div>
      <div style="margin-top: 3rem">
        <a-button size="large" ghost><a target="_blank" href="https://gitee.com/HServer/hp-lite">下载客服端</a></a-button>
      </div>
    </div>
  </div>

  <div class="bock two">

    <div>
      <div class="data" style="margin: 30px auto">
        <div style="font-size: 35px;text-align: center;margin-bottom: 50px">HP-Lite内网穿透开源项目针对个人搭建的声明</div>
        <div class="sm">
          <p>亲爱的用户，</p>
          <p>
            为了帮助您实现内网穿透的需求，我们提供一个开源的内网穿透项目，用于个人搭建。在使用该项目之前，请您仔细阅读以下声明：</p>
          <ol>
            <li><strong>项目用途限制</strong>：该开源项目旨在帮助个人实现内网穿透，方便访问自己的本地网络资源，和学习目的。请确保您在合法范围内使用该项目，并遵守法律法规，禁止搭建一切违反国家法律法规的网站，出现问题自行负责，与作者无关。
            </li>
            <li><strong>安全性责任</strong>：我们将尽力确保该项目的安全性和稳定性，但无法完全排除潜在的风险。因此，在使用该项目时，请做好相关的安全措施并承担个人风险。
            </li>
            <li><strong>防火墙设置</strong>：为了使内网穿透正常工作，您可能需要调整您的防火墙设置。请确保在进行任何更改之前，充分了解您所做的操作，并采取适当的安全措施。
            </li>
            <li><strong>账户与权限</strong>：为了提供更好的服务体验，您可能需要创建一个个人账户并获取相应的权限。请妥善保管您的账户信息，不要与他人共享，以免造成不必要的风险。
            </li>
            <li><strong>技术支持</strong>：我们提供有限的技术支持，以协助解决您在使用该项目时可能遇到的问题。请参考项目文档和相关社区资源，并理解我们的技术支持范围。
            </li>
            <li><strong>免责声明</strong>：对于因使用该开源项目而导致的任何直接或间接损失，包括但不限于数据丢失、网络故障、安全漏洞等，我们概不负责。使用该项目即表示您已经阅读并同意自行承担相关风险。
            </li>
            <li><strong>合作与反馈</strong>：我们欢迎用户对该项目提出宝贵的建议和反馈，以帮助我们改进和提供更好的服务。如果您对该项目感兴趣或遇到任何问题，请及时与我们联系。
            </li>
          </ol>
          <p>
            请注意，本声明仅作为指导性内容，最终用户应根据自己的实际情况和需求来判断和操作。使用该开源项目即代表您已经阅读并同意遵守以上声明。</p>
          <p>谢谢您的支持和理解</p>
        </div>
      </div>
    </div>
    <div>
      <div class="data" style="margin: 30px auto">
        <div style="font-size: 35px;text-align: center;margin-bottom: 50px">特色</div>
        <a-list
            :grid="{ gutter: 16, xs: 1, sm: 2, md: 2, lg: 3, xl: 3, xxl: 3, xxxl: 3 }" item-layout="horizontal"
            :data-source="serverList2">
          <template #renderItem="{ item }">
            <a-list-item>
              <a-card style="height: 12rem;box-shadow:
    0 2px 2px rgba(0, 0, 0, 0.1),
    0 4px 4px rgba(0, 0, 0, 0.1),
    0 6px 6px rgba(0, 0, 0, 0.1),
    0 8px 8px rgba(0, 0, 0, 0.1); " :title="item.title">{{ item.content }}
              </a-card>
            </a-list-item>
          </template>
        </a-list>
      </div>
    </div>
    <div style="font-size: 35px;text-align: center;margin-bottom: 50px">
      <img style="width: 40vw;" src="/teach/img.png">
    </div>
    <div class="foot">
      <div style="text-align: center;">
          <span>QQ群：1065301527</span>
      </div>
      <div style="text-align: center;font-size: 10px;margin: 10px auto;">
        本站默认用户都具有互联网基础知识，和阅读文档能力，如果您阅读文档后仍对本产品有使用上的疑问，
        我们虽有用户交流群，但不代表一定会有人工客服提供解答，因为本站规模较小，没有资金和能力提供专业的解答服务。
      </div>
    </div>
  </div>

</template>

<script setup>

import {ref} from "vue";

const serverList2 = ref([
  {
    title: "多端支持",
    content: "Android端、Win、Linux、Mac、NAS、Docker等环境 X86、ARM、等CPU架构。"

  },
  {
    title: "固定端口",
    content: "我们支持动态端口与固定端口模式，穿透FTP MYSQL Redis MQ等服务需要。"
  },
  {
    title: "自定义域名",
    content: "我们支持用户自定义域名,只需要域名名字或者二级名字和账号名字一样就可以解析。"
  },
  {
    title: "HTTPS支持",
    content: "支持用户自定义域名的同时添加证书，或者使用平台的一键获取功能获取ssl证书"
  },
  {
    title: "高性能",
    content: "我们使用Golang做为后台基石，在http协议上做了大量优化操作，通过SNI协议解析或者明文HOST解析。"
  }, {
    title: "轻量级",
    content: "我们把所有功能集成在一个文件里针对个人用户只需简单的启动这个文件包就能进行穿透访问，做得快速部署。"
  },

])

</script>

<style scoped lang="less">

.bock {
  height: 100vh;
}

.one {
  padding: 5rem 2rem;
  background-color: #4b6ff6;
}

.two {
  padding: 5rem 2rem;
  background-color: #ffffff;
}


.data {
  max-width: 1024px;
  min-width: 300px;
}

.menu-head {
  width: 100%;
  z-index: 999;
  position: fixed;
  top: 0;
  right: 0;
  left: 0;
  font-size: 1.2rem;
  background-color: #4b6ff6;
  color: #fff;
}

.menu-data {
  padding: 1rem 2rem;
  margin: 0 auto;
  text-align: center;
  display: flex;
  justify-content: space-between;
}

.menu-btn {
  margin: 0 10px;
}

.foot {
  padding: 2rem;
  height: auto;
  position: absolute;
  left: 0;
  width: 100%;
  background-color: #4b6ff6;
  color: #ffffff;
}

.sm {
  h1 {
    color: #333;
    font-size: 24px;
    margin-bottom: 10px;
  }

  p {
    color: #555;
    font-size: 16px;
    line-height: 1.5;
    margin-bottom: 15px;
  }

  ol {
    margin-left: 30px;
    margin-bottom: 15px;
  }

  li {
    color: #555;
    font-size: 16px;
    line-height: 1.5;
  }
}

</style>
