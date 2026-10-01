<template>
  <div class="landing">
    <!-- 导航栏 -->
    <header class="nav" :class="{ 'is-scrolled': isScrolled }">
      <div class="nav__inner">
        <a class="brand" href="javascript:void(0)" @click="scrollToTop">
          <!-- 两张图叠放交叉淡入：直接换 src 会重新加载导致闪烁 -->
          <span class="brand__logo">
            <img class="brand__logo-img brand__logo-img--light" src="/logo-back.png" alt=""/>
            <img class="brand__logo-img brand__logo-img--color" src="/logo.png" alt=""/>
          </span>
          <span class="brand__text">HP-Lite</span>
        </a>

        <nav class="nav__links">
          <a v-for="item in navItems" :key="item.id" href="javascript:void(0)"
             :class="{ 'is-active': activeNav === item.id }"
             @click="scrollToSection(item.id)">{{ item.label }}</a>
        </nav>

        <div class="nav__actions">
          <t-button variant="outline" theme="primary" @click="goLogin">登录</t-button>
        </div>
      </div>
    </header>

    <!-- 英雄区 -->
    <section class="hero" id="hero">
      <div class="hero__glow hero__glow--1"></div>
      <div class="hero__glow hero__glow--2"></div>
      <div class="hero__grid"></div>

      <div class="container hero__inner">
        <div class="hero__text">
          <span class="hero__badge">
            <span class="hero__badge-dot"></span>
            轻量 · 高性能 · 内网穿透
          </span>
          <h1 class="hero__title">HP-Lite 内网穿透</h1>
          <p class="hero__subtitle">
            无需公网 IP、无需路由器端口映射，让内网应用随时可以通过域名进行外网访问
          </p>
          <div class="hero__actions">
            <t-button theme="primary" size="large" @click="openSource">
              开源地址
              <template #suffix><arrow-right-icon/></template>
            </t-button>
            <t-button size="large" variant="outline" class="hero__ghost" @click="goLogin">
              进入控制台
            </t-button>
          </div>
        </div>

        <div class="hero__qr">
          <div class="qr-card">
            <img src="/wx.png" alt="微信打赏二维码"/>
            <p class="qr-card__text">扫码支持我们</p>
          </div>
        </div>
      </div>

      <div class="hero__wave"></div>
    </section>

    <!-- 特色区 -->
    <section class="section" id="features">
      <div class="container">
        <!-- 站点数据统计：无数据时组件整块不渲染（含描述） -->
        <LaStats class="stats-band" title="全网访问数据汇总"/>

        <div class="section__head">
          <span class="section__eyebrow">CORE FEATURES</span>
          <h2 class="section__title">核心特色</h2>
          <p class="section__desc">一套覆盖设备、域名、安全与监控的完整内网穿透方案</p>
        </div>

        <div class="feature-grid">
          <article v-for="item in serverList" :key="item.title" class="feature-card">
            <div class="feature-card__icon">
              <component :is="item.icon"/>
            </div>
            <h3 class="feature-card__title">{{ item.title }}</h3>
            <p class="feature-card__desc">{{ item.content }}</p>
          </article>
        </div>
      </div>
    </section>

    <!-- 声明区 -->
    <section class="section section--muted" id="policy">
      <div class="container container--narrow">
        <div class="section__head">
          <span class="section__eyebrow">LICENSE</span>
          <h2 class="section__title">项目许可声明（MIT License）</h2>
        </div>

        <div class="policy">
          <p>版权所有 (c) [2025-{{ new Date().getFullYear() }}] [黑小马]</p>
          <p>
            本内网穿透项目采用
            <a class="link" href="https://opensource.org/licenses/MIT" target="_blank" rel="noopener noreferrer">MIT 开源许可证</a>
            授权。在使用本项目前，请您仔细阅读以下条款：
          </p>
          <ol>
            <li><strong>授权范围</strong>：您有权对本项目进行复制、使用、修改、合并、发布、分发、再授权及销售副本，可用于个人搭建、学习研究及符合法律法规的商业用途。但在所有副本或重要衍生部分中，必须保留本版权声明及许可条款。</li>
            <li><strong>合法使用义务</strong>：本项目旨在帮助用户实现内网穿透以访问自身本地网络资源。您必须在合法范围内使用，严格遵守国家法律法规及相关规定，禁止用于任何违法活动。因违规使用导致的一切责任，由您自行承担，与原作者无关。</li>
            <li><strong>免责声明</strong>：本项目按"现状"（AS IS）提供，作者不对其适用性、安全性或稳定性做任何明示或暗示的担保。对于因使用本项目导致的任何直接或间接损失（包括但不限于数据丢失、网络故障、安全漏洞等），作者不承担任何责任。</li>
            <li><strong>网络安全提示</strong>：为确保内网穿透功能正常运行，您可能需要调整防火墙设置或网络配置。请在操作前充分了解相关风险，采取必要的安全防护措施，由此产生的安全问题由您自行负责。</li>
            <li><strong>账户安全责任</strong>：如使用过程中需要创建账户及权限管理，您应妥善保管账户信息，不得擅自与他人共享，由此产生的账户安全风险由您自行承担。</li>
            <li><strong>技术支持范围</strong>：作者提供有限的技术支持，包括项目文档及社区答疑。您应优先参考官方文档及社区资源，作者不对支持的及时性和有效性做任何保证。</li>
            <li><strong>反馈与改进</strong>：欢迎您提出宝贵建议和反馈以帮助项目改进。如需联系我们，请通过项目地址进行沟通。</li>
            <li><strong>衍生作品授权</strong>：基于本项目修改或衍生的作品，需采用与 MIT 协议兼容的开源许可进行发布，并保留原始版权信息及许可条款。</li>
          </ol>
          <p>使用本项目即表示您已阅读、理解并同意本许可声明的全部条款。</p>
          <p>感谢您的支持与理解。</p>
        </div>
      </div>
    </section>

    <!-- 联系区 -->
    <section class="section" id="contact">
      <div class="container container--narrow">
        <div class="section__head">
          <span class="section__eyebrow">CONTACT</span>
          <h2 class="section__title">联系我们</h2>
        </div>

        <div class="contact">
          <div class="contact__qq">
            <chat-icon size="20px"/>
            <span>QQ 群：1065301527</span>
          </div>
          <p class="contact__note">
            本站默认用户都具有互联网基础知识，和阅读文档能力，如果您阅读文档后仍对本产品有使用上的疑问，
            我们虽有用户交流群，但不代表一定会有人工客服提供解答，因为本站规模较小，没有资金和能力提供专业的解答服务。
          </p>
        </div>
      </div>
    </section>

    <!-- 页脚 -->
    <footer class="footer">
      <div class="container footer__inner">
        <p>&copy; 2025-{{ new Date().getFullYear() }} HP-Lite 内网穿透项目</p>
      </div>
    </footer>
  </div>
</template>

<script setup>
import {onMounted, onUnmounted, ref, shallowRef} from 'vue';
import {useRouter} from 'vue-router';
import {
  ArrowRightIcon,
  ChartBubbleIcon,
  ChatIcon,
  CloudIcon,
  LockOnIcon,
  SecuredIcon,
  SwapIcon,
  UserCircleIcon
} from 'tdesign-icons-vue-next';
import LaStats from '../../components/LaStats.vue';

const router = useRouter();

const navItems = [
  {id: 'features', label: '特色'},
  {id: 'policy', label: '声明'},
  {id: 'contact', label: '联系'},
];

// 特色数据
const serverList = shallowRef([
  {
    icon: CloudIcon,
    title: "云端集中管控",
    content: "支持在 Android、Windows、Linux、MacOS 等操作系统及 NAS、Docker 环境中部署，兼容 X86、ARM 等主流 CPU 架构；提供云端配置文件实时下发能力，实现跨地域环境的高效统一管理"
  },
  {
    icon: UserCircleIcon,
    title: "多设备多租户管理",
    content: "内置精细化用户权限系统，支持超级管理员对子用户进行分级授权与操作审计；子用户可独立完成穿透设备配置及参数管理，满足多租户场景下的隔离化运营需求"
  },
  {
    icon: SecuredIcon,
    title: "域名与证书自动化管理",
    content: "支持用户自定义域名绑定，集成 ACME 协议实现 SSL 证书的自动申请、部署及到期续期，保障域名访问的安全性与连续性"
  },
  {
    icon: ChartBubbleIcon,
    title: "全协议流量可视化分析",
    content: "提供穿透配置的精细化流量统计功能，支持 UDP、TCP、HTTP 等协议的流量监测，包含 PV、UV 及传输量等关键指标的实时分析与历史数据追溯"
  },
  {
    icon: LockOnIcon,
    title: "穿透安全防护体系",
    content: "支持基于 IP 段的黑白名单访问控制，可配置并发连接数限制及流量读写阈值管控，构建多层次的穿透安全防护机制"
  },
  {
    icon: SwapIcon,
    title: "正反向代理一体化支持",
    content: "内置可扩展代理模块，支持反向代理的域名绑定与规则配置；同时集成 HTTP/HTTPS/SOCKS5 协议的正向代理能力，实现代理功能的一站式部署"
  },
]);

// 导航栏状态
const isScrolled = ref(false);
const activeNav = ref('hero');

// 滚动处理
const handleScroll = () => {
  isScrolled.value = window.scrollY > 50;

  const sections = document.querySelectorAll('section');
  let current = 'hero';

  sections.forEach(section => {
    const sectionTop = section.offsetTop - 80;
    const sectionHeight = section.clientHeight;

    if (window.scrollY >= sectionTop && window.scrollY < sectionTop + sectionHeight) {
      current = section.getAttribute('id');
    }
  });

  activeNav.value = current;
};

// 平滑滚动
const scrollToSection = (sectionId) => {
  const section = document.getElementById(sectionId);
  if (section) {
    section.scrollIntoView({behavior: 'smooth'});
  }
};

const scrollToTop = () => {
  window.scrollTo({top: 0, behavior: 'smooth'});
};

const goLogin = () => {
  router.push('/home/login');
};

const openSource = () => {
  window.open('https://gitee.com/HServer/hp-lite', '_blank');
};

// 生命周期钩子
onMounted(() => {
  window.addEventListener('scroll', handleScroll);
});

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll);
});
</script>

<style scoped>
.landing {
  overflow-x: hidden;
}

.container {
  width: 100%;
  max-width: 1180px;
  margin: 0 auto;
  padding: 0 24px;
}

.container--narrow {
  max-width: 900px;
}

/* ============ 导航 ============ */
.nav {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 100;
  transition: all .3s ease;
  padding: 18px 0;
}

.nav.is-scrolled {
  padding: 10px 0;
  background: rgba(255, 255, 255, .78);
  backdrop-filter: saturate(180%) blur(14px);
  -webkit-backdrop-filter: saturate(180%) blur(14px);
  border-bottom: 1px solid var(--hp-border);
  box-shadow: 0 2px 14px rgba(16, 24, 40, .05);
}

.nav__inner {
  width: 100%;
  max-width: 1180px;
  margin: 0 auto;
  padding: 0 24px;
  display: flex;
  align-items: center;
  gap: 24px;
}

.brand {
  display: flex;
  align-items: center;
  gap: 10px;
  color: #fff;
  transition: color .3s ease;
}

.nav.is-scrolled .brand {
  color: var(--hp-text);
}

.brand img {
  width: 38px;
  height: 38px;
  transition: transform .3s ease, opacity .3s ease;
}

.brand:hover img {
  transform: scale(1.06);
}

/* logo 两张图叠放：未滚动用白色版（配蓝色 hero），滚动后用彩色版（配白色导航） */
.brand .brand__logo {
  position: relative;
  flex: none;
  width: 38px;
  height: 38px;
}

.brand .brand__logo-img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
}

.brand .brand__logo-img--light {
  opacity: 1;
}

.brand .brand__logo-img--color {
  opacity: 0;
}

.nav.is-scrolled .brand__logo-img--light {
  opacity: 0;
}

.nav.is-scrolled .brand__logo-img--color {
  opacity: 1;
}

.brand__text {
  font-size: 19px;
  font-weight: 700;
  letter-spacing: .3px;
}

.nav__links {
  display: none;
  gap: 30px;
  margin-left: auto;
}

@media (min-width: 768px) {
  .nav__links {
    display: flex;
  }
}

.nav__links a {
  position: relative;
  font-size: 15px;
  color: rgba(255, 255, 255, .88);
  transition: color .25s ease;
}

.nav.is-scrolled .nav__links a {
  color: var(--hp-text-2);
}

.nav__links a::after {
  content: '';
  position: absolute;
  left: 0;
  bottom: -6px;
  width: 0;
  height: 2px;
  border-radius: 2px;
  background: currentColor;
  transition: width .28s ease;
}

.nav__links a:hover,
.nav__links a.is-active {
  color: #fff;
}

.nav.is-scrolled .nav__links a:hover,
.nav.is-scrolled .nav__links a.is-active {
  color: var(--td-brand-color);
}

.nav__links a:hover::after,
.nav__links a.is-active::after {
  width: 100%;
}

.nav__actions {
  margin-left: auto;
  display: flex;
  align-items: center;
}

@media (min-width: 768px) {
  .nav__actions {
    margin-left: 0;
  }
}

.nav:not(.is-scrolled) .nav__actions :deep(.t-button) {
  background: rgba(255, 255, 255, .16);
  border-color: rgba(255, 255, 255, .5);
  color: #fff;
}

.nav:not(.is-scrolled) .nav__actions :deep(.t-button:hover) {
  background: rgba(255, 255, 255, .26);
}

/* ============ 英雄区 ============ */
.hero {
  position: relative;
  overflow: hidden;
  background: var(--hp-grad);
  color: #fff;
  padding: 190px 0 130px;
}

.hero__glow {
  position: absolute;
  width: 620px;
  height: 620px;
  border-radius: 50%;
  filter: blur(90px);
  opacity: .38;
  pointer-events: none;
}

.hero__glow--1 {
  background: #7aa2ff;
  top: -280px;
  left: -180px;
}

.hero__glow--2 {
  background: #22d3ee;
  bottom: -320px;
  right: -160px;
  opacity: .26;
}

.hero__grid {
  position: absolute;
  inset: 0;
  background-image: linear-gradient(rgba(255, 255, 255, .07) 1px, transparent 1px),
  linear-gradient(90deg, rgba(255, 255, 255, .07) 1px, transparent 1px);
  background-size: 44px 44px;
  mask-image: radial-gradient(ellipse at 50% 0%, #000 30%, transparent 72%);
  -webkit-mask-image: radial-gradient(ellipse at 50% 0%, #000 30%, transparent 72%);
  pointer-events: none;
}

.hero__inner {
  position: relative;
  display: flex;
  align-items: center;
  gap: 56px;
}

.hero__text {
  flex: 1;
}

.hero__badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 5px 14px;
  border-radius: 999px;
  background: rgba(255, 255, 255, .16);
  border: 1px solid rgba(255, 255, 255, .26);
  font-size: 13px;
  margin-bottom: 22px;
}

.hero__badge-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #4ade80;
  box-shadow: 0 0 0 4px rgba(74, 222, 128, .25);
}

.hero__title {
  color: #fff;
  font-size: 52px;
  line-height: 1.15;
  font-weight: 800;
  margin: 0 0 18px;
  letter-spacing: -.5px;
}

.hero__subtitle {
  font-size: 17px;
  line-height: 1.7;
  margin: 0 0 32px;
  color: rgba(255, 255, 255, .9);
  max-width: 560px;
}

.hero__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.hero__ghost {
  background: rgba(255, 255, 255, .14) !important;
  border-color: rgba(255, 255, 255, .45) !important;
  color: #fff !important;
}

.hero__ghost:hover {
  background: rgba(255, 255, 255, .24) !important;
}

.hero__qr {
  flex: none;
}

.qr-card {
  background: #fff;
  padding: 22px;
  border-radius: 20px;
  box-shadow: var(--hp-shadow-lg);
  text-align: center;
  width: 240px;
  transition: transform .3s ease;
}

.qr-card:hover {
  transform: translateY(-6px);
}

.qr-card img {
  width: 194px;
  height: 194px;
  border-radius: 12px;
  display: block;
  margin-bottom: 14px;
}

.qr-card__text {
  margin: 0;
  color: var(--hp-text);
  font-size: 15px;
  font-weight: 600;
}

.hero__wave {
  position: absolute;
  left: 0;
  right: 0;
  bottom: -1px;
  height: 70px;
  background: var(--hp-bg);
  clip-path: ellipse(75% 100% at 50% 100%);
}

/* ============ 通用区块 ============ */
.section {
  padding: 90px 0;
}

.section--muted {
  background: #eef1f8;
}

.section__head {
  text-align: center;
  margin-bottom: 46px;
}

.section__eyebrow {
  display: inline-block;
  font-size: 12px;
  letter-spacing: 2px;
  font-weight: 600;
  color: var(--td-brand-color);
  margin-bottom: 10px;
}

.section__title {
  font-size: 34px;
  font-weight: 700;
  margin: 0;
  letter-spacing: -.3px;
}

.section__desc {
  margin: 12px 0 0;
  color: var(--hp-text-2);
  font-size: 15px;
}

/* ============ 站点数据统计（核心特色上方） ============ */
/* 组件内部负责居中与留白，这里只管与下方区块的间距 + 数据到达时的入场动画 */
.stats-band {
  margin-bottom: 58px;
  animation: stats-fade-in .45s ease-out both;
}

@keyframes stats-fade-in {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: none;
  }
}

@media (prefers-reduced-motion: reduce) {
  .stats-band {
    animation: none;
  }
}

/* ============ 特色卡片 ============ */
.feature-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 22px;
}

.feature-card {
  position: relative;
  background: #fff;
  border: 1px solid var(--hp-border);
  border-radius: var(--hp-radius);
  padding: 28px;
  box-shadow: var(--hp-shadow-sm);
  transition: transform .28s ease, box-shadow .28s ease, border-color .28s ease;
  overflow: hidden;
}

.feature-card::before {
  content: '';
  position: absolute;
  inset: 0 0 auto 0;
  height: 3px;
  background: var(--hp-grad);
  opacity: 0;
  transition: opacity .28s ease;
}

.feature-card:hover {
  transform: translateY(-5px);
  border-color: #d9e2ff;
  box-shadow: var(--hp-shadow);
}

.feature-card:hover::before {
  opacity: 1;
}

.feature-card__icon {
  width: 46px;
  height: 46px;
  border-radius: 13px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--td-brand-color-1);
  color: var(--td-brand-color);
  font-size: 22px;
  margin-bottom: 16px;
}

.feature-card__title {
  font-size: 17px;
  font-weight: 600;
  margin: 0 0 8px;
}

.feature-card__desc {
  margin: 0;
  color: var(--hp-text-2);
  line-height: 1.75;
  font-size: 14px;
}

/* ============ 声明 ============ */
.policy {
  background: #fff;
  border: 1px solid var(--hp-border);
  border-radius: var(--hp-radius);
  padding: 32px 34px;
  box-shadow: var(--hp-shadow-sm);
  color: var(--hp-text-2);
}

.policy p {
  margin: 0 0 14px;
  line-height: 1.85;
}

.policy ol {
  margin: 0 0 18px;
  padding-left: 22px;
}

.policy li {
  margin-bottom: 12px;
  line-height: 1.85;
}

.policy strong {
  color: var(--hp-text);
}

.link {
  color: var(--td-brand-color);
  font-weight: 500;
}

.link:hover {
  text-decoration: underline;
}

/* ============ 联系 ============ */
.contact {
  text-align: center;
}

.contact__qq {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 22px;
  border-radius: 999px;
  background: #fff;
  border: 1px solid var(--hp-border);
  box-shadow: var(--hp-shadow-sm);
  font-size: 17px;
  font-weight: 600;
  color: var(--td-brand-color);
  margin-bottom: 18px;
}

.contact__note {
  margin: 0;
  font-size: 14px;
  color: var(--hp-text-2);
  line-height: 1.9;
}

/* ============ 页脚 ============ */
.footer {
  background: #101a34;
  color: rgba(255, 255, 255, .72);
  padding: 34px 0;
  text-align: center;
}

.footer__inner p {
  margin: 0;
}

/* ============ 响应式 ============ */
@media (max-width: 900px) {
  .hero {
    padding: 160px 0 110px;
  }

  .hero__inner {
    flex-direction: column;
    text-align: center;
  }

  .hero__subtitle {
    margin-inline: auto;
  }

  .hero__actions {
    justify-content: center;
  }

  .hero__title {
    font-size: 38px;
  }
}

@media (max-width: 768px) {
  .section {
    padding: 64px 0;
  }

  .section__title {
    font-size: 27px;
  }

  .feature-grid {
    grid-template-columns: 1fr;
  }

  .policy {
    padding: 24px 20px;
  }
}
</style>
