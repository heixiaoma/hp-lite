<template>
  <div class="app">
    <!-- 导航栏 -->
    <nav class="navbar" :class="{ 'scrolled': isScrolled }">
      <div class="container">
        <div class="navbar-logo" @click="scrollToTop">
          <img src="/logo-back.png" alt="HP-Lite Logo">
          <span class="navbar-logo-text">HP-Lite</span>
        </div>

        <div class="navbar-menu">
          <ul class="navbar-links">
            <li><a href="javascript:void(0)" data="features" @click="scrollToSection('features')">特色</a></li>
            <li><a href="javascript:void(0)" data="policy" @click="scrollToSection('policy')">声明</a></li>
            <li><a href="javascript:void(0)" data="contact" @click="scrollToSection('contact')">联系</a></li>
          </ul>
        </div>

        <div class="navbar-actions">
          <a-button type="dashed" size="large" ghost>
            <a href="/#/home/login">登录</a>
          </a-button>
        </div>

        <div class="navbar-mobile-menu">
          <a-button type="dashed" size="large" ghost>
            <a href="/#/home/login">登录</a>
          </a-button>
        </div>
      </div>
    </nav>

    <!-- 英雄区 -->
    <section class="hero" id="hero">
      <div class="container">
        <div class="hero-content">
          <div class="hero-text">
            <h1 class="hero-title">HP-Lite内网穿透</h1>
            <p class="hero-subtitle">
              无需公网IP、无需路由器端口映射，让内网应用随时可以
              通过域名进行外网访问
            </p>
            <div class="hero-actions">
              <a-button size="large" ghost>
                <a target="_blank" href="https://gitee.com/HServer/hp-lite">开源地址</a>
              </a-button>
            </div>
          </div>

          <div class="hero-qrcode">
            <div class="qrcode-card">
              <img src="/wx.png" alt="微信打赏二维码">
              <p class="qrcode-text">扫码支持我们</p>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- 特色区 -->
    <section class="features" id="features">
      <div class="container">
        <h2 class="section-title">核心特色</h2>
        <div class="features-grid">
          <div v-for="item in serverList" :key="item.title" class="feature-card">
            <div class="feature-icon">
              <CheckCircleOutlined />
            </div>
            <div class="feature-content">
              <h3 class="feature-title">{{ item.title }}</h3>
              <p class="feature-description">{{ item.content }}</p>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- 声明区 -->
    <section class="policy" id="policy">
      <div class="container">
        <h2 class="section-title">项目声明</h2>
        <div class="policy-content">
          <p>亲爱的用户，</p>
          <p>为了帮助您实现内网穿透的需求，我们提供一个开源的内网穿透项目，用于个人搭建。在使用该项目之前，请您仔细阅读以下声明：</p>
          <ol>
            <li><strong>项目用途限制</strong>：该开源项目旨在帮助个人实现内网穿透，方便访问自己的本地网络资源，和学习目的。请确保您在合法范围内使用该项目，并遵守法律法规，禁止搭建一切违反国家法律法规的网站，出现问题自行负责，与作者无关。</li>
            <li><strong>安全性责任</strong>：我们将尽力确保该项目的安全性和稳定性，但无法完全排除潜在的风险。因此，在使用该项目时，请做好相关的安全措施并承担个人风险。</li>
            <li><strong>防火墙设置</strong>：为了使内网穿透正常工作，您可能需要调整您的防火墙设置。请确保在进行任何更改之前，充分了解您所做的操作，并采取适当的安全措施。</li>
            <li><strong>账户与权限</strong>：为了提供更好的服务体验，您可能需要创建一个个人账户并获取相应的权限。请妥善保管您的账户信息，不要与他人共享，以免造成不必要的风险。</li>
            <li><strong>技术支持</strong>：我们提供有限的技术支持，以协助解决您在使用该项目时可能遇到的问题。请参考项目文档和相关社区资源，并理解我们的技术支持范围。</li>
            <li><strong>免责声明</strong>：对于因使用该开源项目而导致的任何直接或间接损失，包括但不限于数据丢失、网络故障、安全漏洞等，我们概不负责。使用该项目即表示您已经阅读并同意自行承担相关风险。</li>
            <li><strong>合作与反馈</strong>：我们欢迎用户对该项目提出宝贵的建议和反馈，以帮助我们改进和提供更好的服务。如果您对该项目感兴趣或遇到任何问题，请及时与我们联系。</li>
          </ol>
          <p>请注意，本声明仅作为指导性内容，最终用户应根据自己的实际情况和需求来判断和操作。使用该开源项目即代表您已经阅读并同意遵守以上声明。</p>
          <p>谢谢您的支持和理解</p>
        </div>
      </div>
    </section>

    <!-- 联系区 -->
    <section class="contact" id="contact">
      <div class="container">
        <h2 class="section-title">联系我们</h2>
        <div class="contact-content">
          <div class="contact-info">
            <span>QQ群：1065301527</span>
          </div>
          <div class="contact-note">
            本站默认用户都具有互联网基础知识，和阅读文档能力，如果您阅读文档后仍对本产品有使用上的疑问，
            我们虽有用户交流群，但不代表一定会有人工客服提供解答，因为本站规模较小，没有资金和能力提供专业的解答服务。
          </div>
        </div>
      </div>
    </section>

    <!-- 页脚 -->
    <footer class="footer">
      <div class="container">
        <p>&copy; 2025 HP-Lite 内网穿透项目</p>
      </div>
    </footer>
  </div>
</template>

<script setup>
import { onMounted, onUnmounted, ref } from 'vue';
import { CheckCircleOutlined } from '@ant-design/icons-vue';

// 特色数据
const serverList = ref([
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
  },
  {
    title: "轻量级",
    content: "我们把所有功能集成在一个文件里针对个人用户只需简单的启动这个文件包就能进行穿透访问，做得快速部署。"
  }
]);

// 导航栏状态
const isScrolled = ref(false);
// 滚动处理
const handleScroll = () => {
  isScrolled.value = window.scrollY > 50;

  // 更新导航高亮
  const sections = document.querySelectorAll('section');
  const navLinks = document.querySelectorAll('.navbar-links a');

  let current = '';

  sections.forEach(section => {
    const sectionTop = section.offsetTop - 80;
    const sectionHeight = section.clientHeight;

    if (window.scrollY >= sectionTop && window.scrollY < sectionTop + sectionHeight) {
      current = section.getAttribute('id');
    }
  });

  navLinks.forEach(link => {
    link.classList.remove('active');
    if (link.getAttribute('data') === `${current}`) {
      link.classList.add('active');
    }
  });
};

// 平滑滚动
const scrollToSection = (sectionId) => {
  const section = document.getElementById(sectionId);
  if (section) {
    section.scrollIntoView({ behavior: 'smooth' });
  }
};

const scrollToTop = () => {
  window.scrollTo({ top: 0, behavior: 'smooth' });
};

// 生命周期钩子
onMounted(() => {
  window.addEventListener('scroll', handleScroll);
});

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll);
});
</script>
<style scoped lang="less">
/* 全局样式 */
.app {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
  line-height: 1.6;
  color: #333;
}

.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 20px;
}

.section-title {
  font-size: 35px;
  text-align: center;
  margin-bottom: 50px;
  font-weight: 600;
}

/* 导航栏 */
.navbar {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  background-color: transparent;
  padding: 16px 0;
  z-index: 999;
  transition: all 0.3s ease;

  &.scrolled {
    background-color: rgba(75, 111, 246, 0.95);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    padding: 10px 0;
  }

  .container {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .navbar-logo {
    display: flex;
    align-items: center;
    cursor: pointer;

    img {
      width: 45px;
      height: auto;
      margin-right: 10px;
      transition: transform 0.3s ease;
    }

    &:hover img {
      transform: scale(1.05);
    }

    .navbar-logo-text {
      color: white;
      font-size: 18px;
      font-weight: 600;
    }
  }

  .navbar-menu {
    display: none;

    @media (min-width: 768px) {
      display: block;
    }

    .navbar-links {
      list-style: none;
      display: flex;
      gap: 30px;

      a {
        color: white;
        font-size: 16px;
        transition: all 0.3s ease;
        position: relative;
        text-decoration: none;

        &:hover {
          color: #e0e0e0;
        }

        &::after {
          content: '';
          position: absolute;
          width: 0;
          height: 2px;
          bottom: -4px;
          left: 0;
          background-color: white;
          transition: width 0.3s ease;
        }

        &:hover::after, &.active::after {
          width: 100%;
        }
      }
    }
  }

  .navbar-actions {
    display: none;

    @media (min-width: 768px) {
      display: flex;
      align-items: center;
    }
  }

  .navbar-mobile-menu {
    display: block;
    color: white;
    font-size: 24px;
    cursor: pointer;

    @media (min-width: 768px) {
      display: none;
    }
  }
}

/* 英雄区 */
.hero {
  background-color: #4b6ff6;
  color: white;
  padding: 160px 0 100px;

  .hero-content {
    display: flex;
    align-items: center;
    gap: 60px;
  }

  .hero-text {
    flex: 1;

    .hero-title {
      color: #ffffff;
      font-size: 44px;
      margin-bottom: 20px;
      font-weight: 700;
    }

    .hero-subtitle {
      font-size: 18px;
      margin-bottom: 30px;
      line-height: 1.6;
    }

    .hero-actions {
      display: flex;
      justify-content: flex-start;
    }
  }

  .hero-qrcode {
    .qrcode-card {
      background-color: white;
      padding: 20px;
      border-radius: 12px;
      box-shadow: 0 8px 24px rgba(0, 0, 0, 0.1);
      text-align: center;
      max-width: 240px;
      transition: transform 0.3s ease;

      &:hover {
        transform: translateY(-5px);
      }

      img {
        width: 200px;
        height: 200px;
        margin-bottom: 15px;
        border-radius: 8px;
      }

      .qrcode-text {
        color: #333;
        font-size: 16px;
        font-weight: 500;
      }
    }
  }
}

/* 特色区 */
.features {
  padding: 100px 0;

  .features-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: 30px;

    .feature-card {
      background-color: white;
      border-radius: 12px;
      padding: 30px;
      box-shadow:
          0 2px 4px rgba(0, 0, 0, 0.05),
          0 4px 8px rgba(0, 0, 0, 0.05),
          0 8px 16px rgba(0, 0, 0, 0.05);
      transition: all 0.3s ease;
      position: relative;
      overflow: hidden;

      &::before {
        content: '';
        position: absolute;
        top: 0;
        left: 0;
        width: 4px;
        height: 100%;
        background-color: #4b6ff6;
        transition: width 0.3s ease;
      }

      &:hover {
        transform: translateY(-5px);
        box-shadow:
            0 4px 8px rgba(0, 0, 0, 0.08),
            0 8px 16px rgba(0, 0, 0, 0.08),
            0 16px 32px rgba(0, 0, 0, 0.08);
      }

      &:hover::before {
        width: 8px;
      }

      .feature-icon {
        width: 50px;
        height: 50px;
        background-color: rgba(75, 111, 246, 0.1);
        border-radius: 50%;
        display: flex;
        align-items: center;
        justify-content: center;
        margin-bottom: 20px;
        color: #4b6ff6;
        font-size: 24px;
      }

      .feature-title {
        font-size: 18px;
        font-weight: 600;
        margin-bottom: 12px;
        color: #333;
      }

      .feature-description {
        color: #555;
        line-height: 1.6;
        font-size: 15px;
      }
    }
  }
}

/* 声明区 */
.policy {
  padding: 100px 0;
  background-color: #f9f9f9;

  .policy-content {
    max-width: 800px;
    margin: 0 auto;

    p {
      margin-bottom: 15px;
      line-height: 1.8;
    }

    ol {
      margin-left: 20px;
      margin-bottom: 20px;

      li {
        margin-bottom: 10px;
        line-height: 1.8;
      }
    }
  }
}

/* 联系区 */
.contact {
  padding: 100px 0;

  .contact-content {
    max-width: 800px;
    margin: 0 auto;
    text-align: center;

    .contact-info {
      margin-bottom: 20px;
      font-size: 18px;
    }

    .contact-note {
      font-size: 14px;
      color: #555;
    }
  }
}

/* 页脚 */
.footer {
  background-color: #4b6ff6;
  color: white;
  padding: 20px 0;
  text-align: center;
}

/* 响应式调整 */
@media (max-width: 768px) {
  .hero {
    .hero-content {
      flex-direction: column;
    }

    .hero-text {
      text-align: center;

      .hero-actions {
        justify-content: center;
      }
    }
  }

  .features-grid {
    grid-template-columns: 1fr;
  }

  .section-title {
    font-size: 28px;
  }

  .hero-title {
    font-size: 36px;
  }

  .hero-subtitle {
    font-size: 16px;
  }
}
</style>