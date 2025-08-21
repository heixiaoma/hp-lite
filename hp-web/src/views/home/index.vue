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
            <a href="/home/login">登录</a>
          </a-button>
        </div>

        <div class="navbar-mobile-menu">
          <a-button type="dashed" size="large" ghost>
            <a href="/home/login">登录</a>
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
        <h2 class="section-title">项目许可声明（MIT License）</h2>
        <div class="policy-content">
          <p>版权所有 (c) [2025] [黑小马]</p>
          <p>本内网穿透项目采用 <a href="https://opensource.org/licenses/MIT" target="_blank" rel="noopener noreferrer">MIT开源许可证</a> 授权。在使用本项目前，请您仔细阅读以下条款：</p>
          <ol>
            <li><strong>授权范围</strong>：您有权对本项目进行复制、使用、修改、合并、发布、分发、再授权及销售副本，可用于个人搭建、学习研究及符合法律法规的商业用途。但在所有副本或重要衍生部分中，必须保留本版权声明及许可条款。</li>
            <li><strong>合法使用义务</strong>：本项目旨在帮助用户实现内网穿透以访问自身本地网络资源。您必须在合法范围内使用，严格遵守国家法律法规及相关规定，禁止用于任何违法活动。因违规使用导致的一切责任，由您自行承担，与原作者无关。</li>
            <li><strong>免责声明</strong>：本项目按"现状"（AS IS）提供，作者不对其适用性、安全性或稳定性做任何明示或暗示的担保。对于因使用本项目导致的任何直接或间接损失（包括但不限于数据丢失、网络故障、安全漏洞等），作者不承担任何责任。</li>
            <li><strong>网络安全提示</strong>：为确保内网穿透功能正常运行，您可能需要调整防火墙设置或网络配置。请在操作前充分了解相关风险，采取必要的安全防护措施，由此产生的安全问题由您自行负责。</li>
            <li><strong>账户安全责任</strong>：如使用过程中需要创建账户及权限管理，您应妥善保管账户信息，不得擅自与他人共享，由此产生的账户安全风险由您自行承担。</li>
            <li><strong>技术支持范围</strong>：作者提供有限的技术支持，包括项目文档及社区答疑。您应优先参考官方文档及社区资源，作者不对支持的及时性和有效性做任何保证。</li>
            <li><strong>反馈与改进</strong>：欢迎您提出宝贵建议和反馈以帮助项目改进。如需联系我们，请通过项目地址进行沟通。</li>
            <li><strong>衍生作品授权</strong>：基于本项目修改或衍生的作品，需采用与MIT协议兼容的开源许可进行发布，并保留原始版权信息及许可条款。</li>
          </ol>
          <p>使用本项目即表示您已阅读、理解并同意本许可声明的全部条款。</p>
          <p>感谢您的支持与理解。</p>
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
    title: "云端集中管控",
    content: "支持在 Android、Windows、Linux、MacOS 等操作系统及 NAS、Docker 环境中部署，兼容 X86、ARM 等主流 CPU 架构；提供云端配置文件实时下发能力，实现跨地域环境的高效统一管理"
  },
  {
    title: "多设备多租户管理",
    content: "内置精细化用户权限系统，支持超级管理员对子用户进行分级授权与操作审计；子用户可独立完成穿透设备配置及参数管理，满足多租户场景下的隔离化运营需求"
  },
  {
    title: "域名与证书自动化管理",
    content: "支持用户自定义域名绑定，集成 ACME 协议实现 SSL 证书的自动申请、部署及到期续期，保障域名访问的安全性与连续性"
  },
  {
    title: "全协议流量可视化分析",
    content: "提供穿透配置的精细化流量统计功能，支持 UDP、TCP、HTTP 等协议的流量监测，包含 PV、UV 及传输量等关键指标的实时分析与历史数据追溯"
  },
  {
    title: "穿透安全防护体系",
    content: "支持基于 IP 段的黑白名单访问控制，可配置并发连接数限制及流量读写阈值管控，构建多层次的穿透安全防护机制"
  },
  {
    title: "正反向代理一体化支持",
    content: "内置可扩展代理模块，支持反向代理的域名绑定与规则配置；同时集成 HTTP/HTTPS/SOCKS5 协议的正向代理能力，实现代理功能的一站式部署"
  },
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
    background: linear-gradient(135deg, #4b6ff6 0%, #1890ff 100%);
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
  background: linear-gradient(135deg, #4b6ff6 0%, #1890ff 100%);
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
  padding: 50px 0;

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
  padding: 50px 0;
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
  padding: 50px 0;

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