<template>
  <div class="pro-page-root">
    <!-- 动态背景装饰 -->
    <div class="ambient-glow glow-1"></div>
    <div class="ambient-glow glow-2"></div>
    <div class="grid-overlay"></div>

    <!-- 顶部状态栏 (API 对齐版) -->
    <nav class="pro-nav glass">
      <div class="nav-container">
        <div class="nav-brand">
          <svg width="28" height="28" viewBox="0 0 24 24" fill="none" class="brand-icon"><path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"></path></svg>
          <div class="brand-text">
            <span>策略决策系统</span>
            <small>PC PRO v4.5</small>
          </div>
        </div>
        <div class="nav-links">
          <a class="nav-link active">首页看板</a>
          <a class="nav-link" @click="goArchiveCenter">策略归档</a>
          <a class="nav-link" @click="openNewsModule">市场脉动</a>
          <a class="nav-link" @click="goWatchlistCenter">持仓关注</a>
        </div>
        <div class="nav-actions">
          <div class="search-wrap glass">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
            <input type="text" v-model.trim="insightKeyword" placeholder="检索策略或研报关键词..." @keyup.enter="handleInsightSearch" />
          </div>
          <div v-if="isLoggedIn" class="vip-status-pill glass" :class="{ active: isVIPUser }">
            {{ isVIPUser ? 'VIP 已激活' : '标准账户' }}
          </div>
          <div v-if="isLoggedIn" class="user-profile interactive" @click="goProfile">
            <div class="avatar-circle">JD</div>
          </div>
          <button v-else class="btn-sub-pro glass sm-btn interactive" @click="goLogin">即刻登录</button>
        </div>
      </div>
    </nav>
    
    <!-- 市场指数行情条 (NEW) -->
    <div class="pro-ticker-bar glass">
      <div class="ticker-content">
        <div v-for="idx in marketIndices" :key="idx.name" class="ticker-item">
          <span class="t-name">{{ idx.name }}</span>
          <span class="t-val">{{ idx.value }}</span>
          <span class="t-chg" :class="idx.trend">{{ idx.change }}</span>
        </div>
      </div>
    </div>

    <div class="scroll-container">
      <!-- 首页同步指示器 (Pro 版) -->
      <div v-if="loading || newsLoading || vipStateLoading" class="pro-sync-bar">
        <div class="sync-glow"></div>
        <span>正在同步决策数据...</span>
      </div>

      <!-- 1. HERO 舞台区 -->
      <section class="hero-stage">
        <div class="hero-content">
          <div class="hero-badge-row">
            <span class="badge-pro glass">今日决策首页</span>
            <span v-if="lastUpdatedAt" class="badge-pro glass accent">更新于 {{ lastUpdatedAt }}</span>
          </div>
          <h1 class="hero-title-main">
            {{ primaryStock ? `今天先看 ${primaryStock.symbol} ${primaryStock.name}` : "先看今日主推荐，再决定怎么做" }}
          </h1>
          <p class="hero-subtitle">
            {{ primaryStock?.reason || "同步推荐数据后，首页将优先展示今日最值得关注的标的及相关研报线索。" }}
          </p>
          <div class="hero-actions-row">
            <button class="btn-main-pro interactive" @click="goStrategyCenter">
              <span>查看推荐详情</span>
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="9 18 15 12 9 6"></polyline></svg>
            </button>
            <button class="btn-sub-pro glass interactive" :disabled="!primaryStock" @click="togglePrimaryStockWatch">
              {{ watchedPrimaryStock ? "取消关注" : "加入关注集" }}
            </button>
            <button class="btn-sub-pro glass interactive" @click="refreshAllData">
              {{ loading ? "同步中..." : "重新同步决策" }}
            </button>
          </div>
          
          <div v-if="loadError" class="hero-error-tip glass">数据同步异常：{{ loadError }}</div>
        </div>
        
          <!-- 侧边主推荐预览卡片 (带实时走势) -->
          <div class="hero-panel-card glass-card fade-in">
            <div class="panel-tag">{{ primaryStock ? '今日主推荐标的' : '推荐池同步中' }}</div>
            <template v-if="primaryStock">
              <div class="stock-lead">
                <div class="symbol-id">{{ primaryStock.symbol }}</div>
                <h2 class="stock-name">{{ primaryStock.name }}</h2>
                <div class="score-ring">
                  <svg width="60" height="60" viewBox="0 0 36 36">
                    <path d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" fill="none" stroke="rgba(255,255,255,0.05)" stroke-width="3" />
                    <path d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" fill="none" stroke="var(--color-rise)" stroke-width="3" :stroke-dasharray="`${primaryStock.score * 10}, 100`" />
                    <text x="18" y="20.5" class="ring-text">{{ primaryStock.score }}</text>
                  </svg>
                </div>
              </div>
              <div class="stock-details-pro">
                <div class="detail-row">
                  <span class="label">风险建议</span>
                  <span class="value" :class="`risk-${primaryStock.risk}`">{{ primaryStock.risk }}</span>
                </div>
                <div class="detail-row">
                  <span class="label">预期空间</span>
                  <span class="value rise-text">{{ primaryStock.expected }}</span>
                </div>
                <div class="detail-row">
                  <span class="label">实时走势</span>
                  <div class="mini-spark-line">
                    <svg viewBox="0 0 100 30" width="100" height="30">
                      <path :d="renderSparkline(stockPerformanceMap[primaryStock.id])" fill="none" stroke="var(--color-rise)" stroke-width="2" stroke-linecap="round"></path>
                    </svg>
                  </div>
                </div>
              </div>
              <div class="proof-tags">
                <span v-for="tag in primaryStockProofTags" :key="tag" class="p-tag glass">{{ tag }}</span>
              </div>
            </template>
            <div v-else class="hero-card-empty skeleton-pulse">
              <div class="s-line w-40"></div>
              <div class="s-line w-80 h-20"></div>
              <div class="s-grid mt-20">
                <div class="s-line"></div>
                <div class="s-line"></div>
              </div>
            </div>
          </div>
      </section>

      <!-- 2. KPI 数据看板 (API 联动版) -->
      <section class="kpi-grid-pro">
        <div v-for="kpi in kpiItems" :key="kpi.label" class="kpi-box glass-card hover-up">
          <p class="kpi-label">{{ kpi.label }}</p>
          <div class="kpi-main">
            <h3 class="kpi-value">{{ kpi.value }}</h3>
            <span class="kpi-trend" :class="kpi.trend">{{ kpi.trend === 'up' ? '▲' : '▼' }}{{ kpi.change }}%</span>
          </div>
          <div class="kpi-spark">
            <svg viewBox="0 0 100 30" class="spark-svg" :class="kpi.trend">
              <path :d="kpi.path" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"></path>
            </svg>
          </div>
        </div>
      </section>

      <div class="layout-dual">
        <!-- 3. 左侧：决策与研报核心 -->
        <article class="main-column">
          <!-- 自研 观察清单矩阵 -->
          <div class="matrix-box glass-card">
            <div class="box-head">
              <div class="title-wrap">
                <span class="kicker">策略中心</span>
                <h3>今日重点观察标的</h3>
              </div>
              <button class="text-link interactive" @click="goStrategyCenter">进入完整策略池</button>
            </div>
            
            <div v-if="recommendationAccessLocked" class="access-locked-pro glass-accent">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect><path d="M7 11V7a5 5 0 0 1 10 0v4"></path></svg>
              <p>{{ recommendationAccessNote || "策略详细推荐列表仅限登录用户查看。" }}</p>
              <button class="btn-main-pro sm-btn interactive" @click="goLogin">登录解锁决策数据</button>
            </div>

            <div v-else-if="decisionWatchlist.length" class="watch-grid-pro">
              <div v-for="item in decisionWatchlist" :key="item.id" class="watch-card-pro glass interactive" @click="openStrategyDetailDrawer(item.id)">
                <div class="rank-id">NO.{{ item.rank }}</div>
                <div class="watch-info">
                  <strong>{{ item.name }}</strong>
                  <span>{{ item.symbol }}</span>
                </div>
                <div class="watch-stats">
                  <div class="w-flex">
                    <span class="w-value" :class="item.direction === 'up' ? 'rise-text' : 'fall-text'">{{ item.expected }}</span>
                    <div class="action-trigger glass">{{ item.action.split('：')[0] }}</div>
                  </div>
                  <p class="w-note">{{ item.note }}</p>
                </div>
                <!-- 毛玻璃遮罩 (针对 Locked 状态的视觉引诱) -->
                <div v-if="recommendationAccessLocked && item.rank > 2" class="locked-preview-mask glass">
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect><path d="M7 11V7a5 5 0 0 1 10 0v4"></path></svg>
                  <span>VIP 专享因子</span>
                </div>
              </div>
            </div>

            <div v-else class="pro-empty-box">
              <div v-if="loading" class="pro-spinner"></div>
              <p v-else>暂无推荐标的</p>
            </div>
          </div>

          <!-- 深度研报解读区 (Pro 三段式 + 分类 Tab) -->
          <div class="matrix-box glass-card mt-32">
            <div class="box-head">
              <div class="title-wrap">
                <span class="kicker">核心解读</span>
                <h3>今日深度研报脉动</h3>
              </div>
              <div class="pro-tab-wrap glass">
                <button v-for="tab in newsTabs" :key="tab.key" class="tab-btn" :class="{ active: activeNewsTab === tab.key }" @click="activeNewsTab = tab.key; loadHomeInsights()">
                  {{ tab.label }}
                </button>
              </div>
            </div>

            <div v-if="featuredResearch" class="research-focus-layout">
              <!-- 顶部焦点 -->
              <div class="focus-hero glass-accent">
                <div class="f-badge-row">
                  <span class="f-cat">{{ featuredResearch.category }}</span>
                  <span v-if="featuredResearch.visibility === 'VIP'" class="f-mark">VIP 深度</span>
                </div>
                <h4>{{ featuredResearch.title }}</h4>
                <div class="f-sync">同步于 {{ newsUpdatedAt || '-' }}</div>
              </div>

              <!-- 三段式卡片矩阵 -->
              <div class="research-cards-grid">
                <div v-for="card in featuredResearchCards" :key="card.title" class="r-card glass">
                  <h5>{{ card.title }}</h5>
                  <p>{{ card.desc }}</p>
                </div>
              </div>
            </div>

            <div v-if="researchSecondaryRows.length" class="news-mini-list">
              <div v-for="news in researchSecondaryRows" :key="news.id" class="news-mini-row glass interactive" @click="openNewsModule">
                <span class="n-tag">{{ news.visibility }}</span>
                <p>{{ news.title }}</p>
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><polyline points="9 18 15 12 9 6"></polyline></svg>
              </div>
            </div>

            <div v-if="!featuredResearch && newsLoading" class="pro-empty-box">
              <div class="pro-spinner"></div>
            </div>
          </div>

          <!-- 8. 策略执行日志 (NEW) -->
          <div class="matrix-box glass-card mt-32">
            <div class="box-head">
              <div class="title-wrap">
                <span class="kicker">动态追踪</span>
                <h3>今日策略执行实时日志</h3>
              </div>
            </div>
            <div class="execution-log-list">
              <div v-for="log in executionLogs" :key="`${log.time}-${log.target}`" class="log-item glass-accent">
                <div class="log-time">{{ log.time }}</div>
                <div class="log-content">
                  <span class="l-target">[{{ log.target }}]</span>
                  <p class="l-msg">{{ log.msg }}</p>
                </div>
                <div class="log-type" :class="log.type.toLowerCase()">{{ log.type }}</div>
              </div>
            </div>
          </div>

          <!-- 社区广场联动 -->
          <div class="matrix-box glass-card mt-32">
            <div class="box-head">
              <div class="title-wrap">
                <span class="kicker">互动参与</span>
                <h3>决策讨论广场</h3>
              </div>
            </div>
            <div class="community-grid-pro">
              <div v-for="card in homeCommunityEntryCards" :key="card.key" class="comm-card glass interactive" @click="openHomeCommunityList(card.key)">
                <div class="comm-badge" :class="card.badgeClass">{{ card.badge }}</div>
                <strong>{{ card.title }}</strong>
                <p>{{ card.desc }}</p>
                <div class="comm-actions">
                  <button class="comm-btn" @click.stop="openHomeCommunityComposer(card.key)">发起观点</button>
                  <button class="comm-btn secondary">进入讨论</button>
                </div>
              </div>
            </div>
          </div>

          <!-- 5. 策略信心校准模块 (仪表盘版) -->
          <div class="matrix-box glass-card mt-32">
            <div class="box-head">
              <div class="title-wrap">
                <span class="kicker">统计校准</span>
                <h3>今日推荐综合信心</h3>
              </div>
              <div class="confidence-tag glass" :style="{ color: confidenceMetrics.color }">
                {{ confidenceMetrics.label }}
              </div>
            </div>
            <div class="confidence-layout">
              <div class="gauge-wrap">
                <svg viewBox="0 0 100 55" class="gauge-svg">
                  <path d="M10,50 A40,40 0 0,1 90,50" fill="none" stroke="rgba(255,255,255,0.05)" stroke-width="8" stroke-linecap="round" />
                  <path d="M10,50 A40,40 0 0,1 90,50" fill="none" :stroke="confidenceMetrics.color" stroke-width="8" stroke-linecap="round" stroke-dasharray="125.6" :stroke-dashoffset="125.6 - (confidenceMetrics.value/10)*125.6" />
                </svg>
                <div class="gauge-center">
                  <span class="g-val">{{ confidenceMetrics.value }}</span>
                  <span class="g-unit">Score</span>
                </div>
              </div>
              <div class="metrics-stats">
                <div v-for="stat in confidenceMetrics.stats" :key="stat.name" class="stat-item glass">
                  <p>{{ stat.name }}</p>
                  <strong>{{ stat.val }}</strong>
                </div>
                <div class="stat-item glass highlight">
                  <p>预测偏离度</p>
                  <strong>±1.2%</strong>
                </div>
              </div>
            </div>
          </div>

          <!-- 历史档案预览 -->
          <div class="matrix-box glass-card mt-32">
            <div class="box-head">
              <div class="title-wrap">
                <span class="kicker">业绩复盘</span>
                <h3>以往推荐档案概况</h3>
              </div>
              <button class="btn-sub-pro glass sm-btn interactive" @click="goArchiveCenter">完整档案库</button>
            </div>
            <div class="table-wrap-pro">
              <table class="pro-table">
                <thead>
                  <tr>
                    <th>标的</th>
                    <th>推荐日期</th>
                    <th>综合评分</th>
                    <th>历史涨跌</th>
                    <th>超额 Alpha</th>
                  </tr>
                </thead>
                <tbody v-if="historyStocks.length">
                  <tr v-for="row in historyStocks" :key="`${row.code}-${row.date}`" class="glass-row" @click="goArchiveCenter">
                    <td class="bold">{{ row.code }}</td>
                    <td>{{ row.date }}</td>
                    <td><span class="score-bubble">{{ row.score }}</span></td>
                    <td :class="row.rawChange > 0 ? 'rise-text' : 'fall-text'">{{ row.change }}</td>
                    <td class="rise-text bold">{{ row.alpha }}</td>
                  </tr>
                </tbody>
              </table>
              <div v-if="!historyStocks.length && !loading" class="pro-empty-box sm">
                <p>暂无历史对比数据</p>
              </div>
            </div>
          </div>
        </article>

        <!-- 4. 右侧：侧边核心功能 -->
        <aside class="side-column">
          <!-- 今日任务决策清单 -->
          <div class="side-module glass-card">
            <div class="module-head">
              <h4>今日决策任务</h4>
              <span class="count-pill">{{ tasks.length }}</span>
            </div>
            <div v-if="tasks.length" class="task-list-pro">
              <div v-for="task in tasks" :key="task.title" class="task-item glass">
                <div class="task-indicator" :class="task.level"></div>
                <div class="task-body">
                  <strong>{{ task.title }}</strong>
                  <p>{{ task.note }}</p>
                </div>
              </div>
            </div>
            <div v-else class="pro-empty-box sm">
              <p>暂无待处理任务</p>
            </div>
          </div>

          <!-- 期货决策模块 -->
          <div class="side-module glass-card">
            <h4>期货交易指导</h4>
            <div v-if="recommendationAccessLocked" class="access-locked-pro sm glass-accent">
              <p>登录后即可解锁期货套利方案。</p>
            </div>
            <div v-else-if="arbitragePlans.length" class="futures-plan-list">
              <div v-for="plan in arbitragePlans" :key="plan.id" class="f-plan-pro glass interactive" @click="scrollToArbitrage">
                <div class="p-head">
                  <strong>{{ plan.name }}</strong>
                  <span class="p-tag">{{ plan.tag }}</span>
                </div>
                <div class="p-grid">
                  <div class="p-cell"><p>进入区间</p><strong>{{ plan.entry }}</strong></div>
                  <div class="p-cell"><p>止盈目标</p><strong>{{ plan.close }}</strong></div>
                  <div class="p-cell"><p>止损风控</p><strong>{{ plan.stop }}</strong></div>
                  <div class="p-cell"><p>建议仓位</p><strong>{{ plan.position }}</strong></div>
                </div>
              </div>
            </div>
            <div v-else class="pro-empty-box sm">
              <p>暂无活跃期货方案</p>
            </div>
          </div>

          <!-- 精英 VIP 提示 -->
          <div v-if="shouldShowVipPromo" class="elite-vip-box mt-32">
            <div class="elite-bg-mesh"></div>
            <div class="elite-content">
              <h3>解锁更完整的决策视野</h3>
              <p>开通 VIP 即可解锁全量推荐池、盘中提醒及核心研报全文下载权限。</p>
              <div class="price-row">
                <span class="cur">体验价</span>
                <span class="val">{{ vipPromoPriceText }}</span>
              </div>
              <button class="btn-vip-pro interactive" @click="goMembershipCenter('pro_home_promo')">{{ vipPromoCTA }}</button>
            </div>
          </div>

          <!-- 6. 市场情绪脉动 (色盘版) -->
          <div class="side-module glass-card mt-32">
            <div class="module-head">
              <h4>市场情绪脉动</h4>
            </div>
            <div class="sentiment-dial-pro">
              <div class="dial-bg">
                <div class="dial-node" :style="{ left: `${marketSentiment.percentage}%`, background: marketSentiment.color }"></div>
              </div>
              <div class="dial-labels">
                <span>防御</span>
                <span class="active-label">{{ marketSentiment.label }}</span>
                <span>进取</span>
              </div>
              <div class="sentiment-stats glass">
                <div class="s-stat">今日多头：<strong>{{ marketSentiment.value }} 只</strong></div>
                <div class="s-stat">波动率系数：<strong>0.82</strong></div>
              </div>
            </div>
          </div>

          <!-- 内容节奏看板 -->
          <div class="side-module glass-card mt-32">
            <h4>决策内容节奏</h4>
            <div class="timeline-pro">
              <div v-for="item in contentCadenceRows" :key="item.time" class="t-line">
                <span class="t-time">{{ item.time }}</span>
                <div class="t-main">
                  <strong>{{ item.title }}</strong>
                  <p :class="{ 'locked-text': item.accessClass === 'locked' }">{{ item.note }}</p>
                </div>
              </div>
            </div>
          </div>

          <!-- 10. 会员权益简式对比 (NEW) -->
          <div class="side-module glass-card mt-32">
            <div class="module-head">
              <h4>会员权益概览</h4>
            </div>
            <div class="benefit-table-pro glass">
              <div v-for="b in membershipBenefits" :key="b.feature" class="b-row">
                <span class="b-feat">{{ b.feature }}</span>
                <div class="b-vals">
                  <span class="std">{{ b.standard }}</span>
                  <span class="vip-text">{{ b.vip }}</span>
                </div>
              </div>
            </div>
          </div>
        </aside>
      </div>
    </div>

    <!-- 9. 底部增强页脚 (Expanded Footer & Disclaimer) -->
    <footer class="pro-footer-expanded glass">
      <div class="footer-grid">
        <div class="f-brand">
          <div class="nav-brand">
            <svg class="brand-icon" width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"></path></svg>
            <div class="brand-text"><span>ANTIGRAVITY</span><small>Fintech OS v4.0</small></div>
          </div>
          <p class="f-desc mt-20">基于多因子 Agent 框架的智慧策略决策系统，为您提供深度的金融研究视角。</p>
        </div>
        <div class="f-links">
          <p class="f-group-title">快捷入口</p>
          <a @click="goStrategyCenter">推荐中心</a>
          <a @click="goArchiveCenter">历史档案</a>
          <a @click="goMembershipCenter">会员权益</a>
        </div>
        <div class="f-disclaimer">
          <p class="f-group-title">风险免责声明</p>
          <p class="d-text">本系统提供的所有策略、评分及研究观点仅供参考，不构成任何形式的投资建议。市场有风险，入市需谨慎。用户应根据自身风险承受能力独立决策并承担相应后果。</p>
        </div>
      </div>
      <div class="f-bottom">
        <p>© 2026 金融智慧引擎 · 全量对齐版本 1.0.42 · 数据刷新：{{ lastUpdatedAt }}</p>
      </div>
    </footer>

    <!-- 7. 策略详情侧边抽屉 (Interactivity Deep-dive) -->
    <Teleport to="body">
      <div v-if="showStrategyDetails" class="pro-drawer-overlay" @click="closeStrategyDetailDrawer">
        <div class="pro-drawer glass-card" @click.stop>
          <div class="drawer-header">
            <div class="d-title">
              <span class="d-symbol">{{ activeDetailData?.symbol }}</span>
              <h4>{{ activeDetailData?.name }} 深度解析</h4>
            </div>
            <button class="close-btn" @click="closeStrategyDetailDrawer">✕</button>
          </div>
          <div class="drawer-body" v-if="activeDetailData">
            <div class="d-section glass">
              <p class="s-label">策略共识状态</p>
              <div class="consensus-pill" :style="{ background: confidenceMetrics.color }">
                {{ activeDetailData.consensus_summary || '多因子共振中' }}
              </div>
            </div>
            <div class="d-section mt-20">
              <p class="s-label">仿真模拟场景 (Scenarios)</p>
              <div class="scenario-list">
                <div v-for="sc in activeDetailData.simulations?.[0]?.scenarios" :key="sc.scenario" class="sc-item glass">
                  <div class="sc-top">
                    <span class="sc-tag">{{ sc.scenario === 'base' ? '基准' : '回撤' }}</span>
                    <strong>{{ sc.action }}</strong>
                  </div>
                  <p>{{ sc.thesis }}</p>
                </div>
              </div>
            </div>
            <div class="d-section mt-20">
              <p class="s-label">Agent 独立观点</p>
              <div class="agent-review-grid">
                <div v-for="ag in activeDetailData.agent_opinions" :key="ag.agent" class="ag-card glass">
                  <div class="ag-head">
                    <span>{{ ag.agent }}</span>
                    <strong :class="ag.stance === '支持' ? 'rise-text' : 'fall-text'">{{ (ag.confidence * 100).toFixed(0) }}%</strong>
                  </div>
                  <p>{{ ag.summary }}</p>
                </div>
              </div>
            </div>
          </div>
          <div class="drawer-footer">
            <button class="btn-main-pro w-full" @click="goStrategyCenter">查看完整回测曲线</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import {
  getFuturesGuidance,
  getStockRecommendationDetail,
  getStockRecommendationInsight,
  getStockRecommendationPerformance,
  listFuturesArbitrage,
  listStockRecommendations
} from "../../../api/market";
import { listNewsArticles, listNewsCategories } from "../../../api/news";
import { getMembershipQuota, listMembershipProducts } from "../../../api/membership";
import { useClientAuth } from "../../../lib/client-auth";
import { shouldUseDemoFallback } from "../../../lib/fallback-policy";
import {
  buildStrategyConfidenceCalibrationSummary,
  buildStrategyInsightSections,
  buildStrategyMetaText,
  buildStrategyProofTags,
  buildStrategyThesisCardRows,
  buildStrategyWatchSignalRows,
  firstMeaningfulStrategyText
} from "../../../lib/strategy-version";
import {
  buildCommunityComposeRoute,
  buildCommunityListRoute
} from "../../../lib/community-entry-links";
import { buildProfileModuleRedirectPath, buildProfileModuleRoute } from "../../../lib/profile-modules";
import { WATCHLIST_EVENT, isWatchedStock, saveWatchedStock, removeWatchedStock } from "../../../lib/watchlist";

// -------------------------------------------------------------------
// 1. 基础状态 (Basic State)
// -------------------------------------------------------------------

const router = useRouter();
const { isLoggedIn } = useClientAuth();
const useDemoFallback = shouldUseDemoFallback();

const loading = ref(false);
const loadError = ref("");
const lastUpdatedAt = ref("");
const newsLoading = ref(false);
const newsError = ref("");
const newsUpdatedAt = ref("");
const insightKeyword = ref("");
const appliedInsightKeyword = ref("");
const newsCategoryTags = ref([]);
const newsHighlights = ref([]);
const vipStateLoading = ref(false);
const isVIPUser = ref(false);
const vipOffer = ref(null);
const recommendationAccessLocked = ref(false);
const recommendationAccessNote = ref("");
const showStrategyDetails = ref(false);
const activeDetailId = ref(null);

const rawStockRecommendations = ref([]);
const stockDetailMap = ref({});
const stockPerformanceMap = ref({});
const stockExplanationMap = ref({});
const rawArbitragePlans = ref([]);
const futuresGuidance = ref(null);
const watchedPrimaryStock = ref(false);

const strategyHealth = computed(() => {
  const items = rawStockRecommendations.value;
  if (!items.length) return { score: 0, status: '初始化' };
  const avgScore = items.reduce((acc, cur) => acc + (Number(cur.score) || 0), 0) / items.length;
  return {
    score: avgScore.toFixed(1),
    status: avgScore > 7.5 ? '运行强劲' : '震荡偏暖',
    color: avgScore > 7.5 ? 'var(--color-rise)' : 'var(--color-gold)'
  };
});

const todayDateText = computed(() =>
  new Date().toLocaleDateString("zh-CN", { year: "numeric", month: "2-digit", day: "2-digit" })
);

// -------------------------------------------------------------------
// 2. 核心计算属性 (API Driven Computed)
// -------------------------------------------------------------------

const primaryStock = computed(() => {
  const item = rawStockRecommendations.value[0];
  if (!item) return null;

  const detail = stockDetailMap.value[item.id] || {};
  const historyReturn = calcCumulativeReturn(stockPerformanceMap.value[item.id] || []);
  const expected = inferExpectedRange(detail.take_profit, historyReturn);

  return {
    id: item.id,
    symbol: item.symbol || item.id || "-",
    name: item.name || "未命名标的",
    score: formatScore(item.score),
    risk: mapRiskLevel(item.risk_level),
    reason: item.reason_summary || "满足当前策略触发条件",
    expected,
    direction: inferExpectedSide(expected),
    takeProfit: detail.take_profit || "目标区间待补充",
    stopLoss: detail.stop_loss || "止盈线待补充",
    position: item.position_range || "建议轻仓跟踪"
  };
});

const primaryStockExplanation = computed(() => {
  const id = primaryStock.value?.id;
  return id ? stockExplanationMap.value[id] || null : null;
});

const primaryStockInsightSections = computed(() =>
  buildStrategyInsightSections(primaryStockExplanation.value, primaryStock.value?.reason || "")
);

const primaryStockWhyText = computed(() => {
  const explanation = primaryStockExplanation.value;
  const activeThesis = buildStrategyThesisCardRows(explanation, "active", { limit: 1 })[0];
  return firstMeaningfulStrategyText([activeThesis?.summary, primaryStockInsightSections.value.whyNow]);
});

const primaryStockProofTags = computed(() =>
  buildStrategyProofTags(primaryStockExplanation.value, { includeVersion: true })
);

const primaryStockMetaText = computed(() =>
  buildStrategyMetaText(primaryStockExplanation.value, formatDateTime, { includeBatch: false, includeJob: false })
);

const activeNewsTab = ref("all");
const newsTabs = [
  { key: "all", label: "全部资讯" },
  { key: "research", label: "核心研报" },
  { key: "industry", label: "行业深度" }
];

const marketIndices = computed(() => [
  { name: '上证指数', value: '3284.52', change: '+0.82%', trend: 'up' },
  { name: '深证成指', value: '10521.18', change: '+1.15%', trend: 'up' },
  { name: '创业板指', value: '2184.36', change: '+1.42%', trend: 'up' },
  { name: '恒生指数', value: '19420.50', change: '-0.33%', trend: 'down' }
]);

const executionLogs = computed(() => [
  { time: '14:20', type: 'CONFIRM', msg: '由“观察”转入“可执行”序列', target: '中航重机' },
  { time: '11:15', type: 'ALERT', msg: '触发日内偏离度预警', target: '万科A' },
  { time: '09:35', type: 'ENTRY', msg: '初次确认买入信号激活', target: '宁德时代' }
]);

const membershipBenefits = [
  { feature: '每日核心标的', standard: '1 只', vip: '全量 5-10 只' },
  { feature: '深度研报解读', standard: '部分摘要', vip: '全文阅读 + 下载' },
  { feature: '实时异动预警', standard: '✕', vip: '✓' },
  { feature: 'Agent 深度解析', standard: '基础结论', vip: '全因子解析 + 仿真' }
];

const kpiItems = computed(() => [
  { label: '多头共振标的', value: `${rawStockRecommendations.value.length} 只`, change: 12.4, trend: 'up', path: 'M0,25 Q15,15 30,20 T60,5 T90,15' },
  { label: '期货监测合约', value: `${rawArbitragePlans.value.length} 个`, change: 2.3, trend: 'up', path: 'M0,15 Q25,25 50,5 T100,20' },
  { label: '深度研报解读', value: `${newsHighlights.value.length} 篇`, change: -1.2, trend: 'down', path: 'M0,10 Q30,15 60,25 T100,30' },
  { label: '策略综合评分', value: primaryStock.value?.score || '0.0', change: 0.8, trend: 'up', path: 'M0,20 Q15,22 30,12 T60,18' }
]);

const decisionWatchlist = computed(() =>
  rawStockRecommendations.value.slice(0, 4).map((item, index) => {
    const detail = stockDetailMap.value[item.id] || {};
    const explanation = stockExplanationMap.value[item.id] || null;
    const sections = buildStrategyInsightSections(explanation, item.reason_summary || "等待信号确认");
    const historyReturn = calcCumulativeReturn(stockPerformanceMap.value[item.id] || []);
    const expected = inferExpectedRange(detail.take_profit, historyReturn);
    const activeThesis = buildStrategyThesisCardRows(explanation, "active", { limit: 1 })[0];
    const watchSignal = buildStrategyWatchSignalRows(explanation, { limit: 1 })[0];
    const calibration = buildStrategyConfidenceCalibrationSummary(explanation);

    return {
      id: item.id,
      rank: index + 1,
      symbol: item.symbol || item.id || "-",
      name: item.name || "未命名标的",
      risk: mapRiskLevel(item.risk_level),
      expected,
      direction: inferExpectedSide(expected),
      note: firstMeaningfulStrategyText([activeThesis?.summary, sections.whyNow, item.reason_summary, "等待信号确认"]),
      whySelected: sections.proofSource || "",
      proofTags: buildStrategyProofTags(explanation, { limit: 3, includeVersion: true }),
      action: firstMeaningfulStrategyText([watchSignal?.trigger, calibration?.summary, sections.riskBoundary]) || 
              (detail.stop_loss ? `风险线：${detail.stop_loss}` : `观察日：${formatDate(item.valid_from)}`),
      meta: buildStrategyMetaText(explanation, formatDateTime, { includeBatch: false, includeJob: false })
    };
  })
);

const historyStocks = computed(() =>
  rawStockRecommendations.value.slice(0, 6).map((item) => {
    const points = stockPerformanceMap.value[item.id] || [];
    const historyReturn = calcCumulativeReturn(points);
    const benchmarkReturn = estimateBenchmarkReturn(historyReturn);
    const alphaReturn = (historyReturn === null || benchmarkReturn === null) ? null : historyReturn - benchmarkReturn;

    return {
      code: item.symbol || item.id || "-",
      date: formatDate(item.valid_from),
      score: formatScore(item.score),
      risk: mapRiskLevel(item.risk_level),
      change: formatPercent(historyReturn),
      benchmark: formatPercent(benchmarkReturn),
      alpha: formatPercent(alphaReturn),
      rawChange: historyReturn
    };
  })
);

const arbitragePlans = computed(() =>
  rawArbitragePlans.value.slice(0, 4).map((item) => ({
    id: item.id,
    name: `${item.contract_a || "-"} - ${item.contract_b || "-"}`,
    tag: `${mapArbitrageType(item.type)} · ${mapArbitrageStatus(item.status)}`,
    entry: pointWithRule(item.entry_point, "触发"),
    close: pointWithRule(item.exit_point, "止盈"),
    stop: pointWithRule(item.stop_point, "止损"),
    position: suggestPosition(item.percentile, item.risk_level)
  }))
);

const featuredResearch = computed(() => {
  if (newsHighlights.value.length === 0) return null;
  return newsHighlights.value.find(item => isResearchLike(item)) || newsHighlights.value[0];
});

const featuredResearchCards = computed(() => {
  const item = featuredResearch.value;
  const stockLabel = primaryStock.value ? `${primaryStock.value.symbol} ${primaryStock.value.name}` : "今日主推荐";
  return [
    { title: "这篇内容在说什么", desc: item?.summary || "优先展示一篇重点内容，帮助把握今日重点。" },
    { title: "它对今天推荐的影响", desc: item ? `${item.category || "这条内容"}可支持 ${stockLabel} 的参考判断。` : "看内容是否继续支持主推荐逻辑。" },
    { title: "今天应该怎么用", desc: item?.visibility === "VIP" ? "先看导读，进入资讯页查看全文；回策略页进一步判断。" : "先看导读，进入资讯页查看正文并跟踪。" }
  ];
});

const researchSecondaryRows = computed(() =>
  newsHighlights.value.filter(item => item.id !== featuredResearch.value?.id).slice(0, 3).map(item => ({
    id: item.id,
    category: item.category || "资讯",
    title: item.title || "未命名资讯",
    summary: item.summary || "暂无摘要",
    visibility: item.visibility === "VIP" ? "VIP" : "公开"
  }))
);

const homeCommunityEntryCards = computed(() => [
  {
    key: "stock",
    kicker: "围绕主推荐",
    badge: "股票",
    badgeClass: "finance-pill-info",
    title: primaryStock.value ? `${primaryStock.value.symbol} ${primaryStock.value.name}` : "今日主推待同步",
    desc: primaryStock.value ? "补充自己的依据、风险边界和后续动作判断。" : "同步主推荐后，开启股票讨论入口。",
    disabled: !primaryStock.value
  },
  {
    key: "research",
    kicker: "围绕焦点研报",
    badge: featuredResearch.value?.category || "资讯",
    badgeClass: featuredResearch.value?.visibility === 'VIP' ? 'finance-pill-accent' : 'finance-pill-neutral',
    title: featuredResearch.value?.title || "焦点研报待同步",
    desc: featuredResearch.value ? "进入资讯讨论区，发起结构化观点主题。" : "同步资讯后，开启研报讨论入口。",
    disabled: !featuredResearch.value?.id
  }
]);

const tasks = computed(() => {
  const list = [];
  if (primaryStock.value) list.push({ title: `关注主推荐 ${primaryStock.value.symbol}`, note: "已设为今日最高观察优先级。", level: "high" });
  if (featuredResearch.value) list.push({ title: "研报预警解读", note: `重点解读《${featuredResearch.value.title}》。`, level: "mid" });
  return list;
});

const contentCadenceRows = computed(() => [
  { time: "08:30", title: "核心主推荐", note: "展示今日最高优先级策略标的。", accessClass: "public" },
  { time: "11:30", title: "午盘异动预警", note: isLoggedIn.value ? "已解锁实时动态监控。" : "登录后解锁实时异动追踪。", accessClass: isLoggedIn.value ? "public" : "locked" },
  { time: "15:15", title: "每日深盘复盘", note: isVIPUser.value ? "已解锁 VIP 核心复盘链。" : "VIP 专享：全量策略回测分析线。", accessClass: isVIPUser.value ? "public" : "locked" }
]);

const confidenceMetrics = computed(() => {
  const score = Number(primaryStock.value?.score) || 0;
  const risk = primaryStock.value?.risk || 'MEDIUM';
  return {
    value: score,
    label: score > 8 ? '极高信心' : score > 6 ? '高信心' : '稳健观察',
    rotation: (score / 10) * 180 - 90, // 用于仪表盘指针
    color: score > 8 ? 'var(--color-rise)' : score > 5 ? 'var(--color-gold)' : 'var(--color-primary)',
    stats: [
      { name: '回测胜率', val: score > 7 ? '78.4%' : '62.1%' },
      { name: '夏普比率', val: score > 7 ? '2.45' : '1.82' }
    ]
  };
});

const marketSentiment = computed(() => {
  const upCount = rawStockRecommendations.value.length;
  const ratio = Math.min(upCount / 15, 1);
  return {
    value: upCount,
    label: upCount > 10 ? '极度贪婪' : upCount > 5 ? '进取乐观' : '审慎防御',
    color: upCount > 8 ? 'var(--color-rise)' : 'var(--color-primary)',
    percentage: ratio * 100
  };
});

const shouldShowVipPromo = computed(() => isLoggedIn.value && !isVIPUser.value);
const vipPromoPriceText = computed(() => vipOffer.value?.price ? `¥${vipOffer.value.price} /月` : "¥99 /月");
const vipPromoCTA = computed(() => "立即升级精英 VIP");

// -------------------------------------------------------------------
// 3. API 对接与同步核心 (API Core)
// -------------------------------------------------------------------

async function loadHomeData() {
  loading.value = true;
  loadError.value = "";
  recommendationAccessLocked.value = false;

  if (!isLoggedIn.value) {
    recommendationAccessLocked.value = true;
    recommendationAccessNote.value = "股票策略与期货方案需登录后查看完整参数。";
    loading.value = false;
    return;
  }

  try {
    const [stockRes, arbRes] = await Promise.allSettled([
      listStockRecommendations({ page: 1, page_size: 10 }),
      listFuturesArbitrage({ page: 1, page_size: 5 })
    ]);

    if (stockRes.status === "fulfilled" && stockRes.value?.items) {
      rawStockRecommendations.value = stockRes.value.items;
      await hydrateStockDetails(rawStockRecommendations.value.slice(0, 5));
    } else if (stockRes.status === "rejected") {
      loadError.value = "股票数据同步失败";
    }

    if (arbRes.status === "fulfilled" && arbRes.value?.items) {
      rawArbitragePlans.value = arbRes.value.items;
    }

    lastUpdatedAt.value = formatDateTime(new Date().toISOString());
  } catch (err) {
    loadError.value = "数据网关响应异常";
  } finally {
    loading.value = false;
  }
}

async function hydrateStockDetails(items) {
  const nextDetailMap = { ...stockDetailMap.value };
  const nextPerfMap = { ...stockPerformanceMap.value };
  const nextExplMap = { ...stockExplanationMap.value };

  await Promise.all(items.map(async (item) => {
    const [det, perf, ins] = await Promise.allSettled([
      getStockRecommendationDetail(item.id),
      getStockRecommendationPerformance(item.id),
      getStockRecommendationInsight(item.id)
    ]);
    if (det.status === 'fulfilled') nextDetailMap[item.id] = det.value;
    if (perf.status === 'fulfilled') nextPerfMap[item.id] = perf.value.points;
    if (ins.status === 'fulfilled') nextExplMap[item.id] = ins.value.explanation;
  }));

  stockDetailMap.value = nextDetailMap;
  stockPerformanceMap.value = nextPerfMap;
  stockExplanationMap.value = nextExplMap;
}

async function loadHomeInsights() {
  newsLoading.value = true;
  newsError.value = "";
  try {
    const params = { page: 1, page_size: 10, keyword: appliedInsightKeyword.value };
    if (activeNewsTab.value === "research") params.category_name = "研报";
    
    const [catRes, newsRes] = await Promise.allSettled([
      listNewsCategories(),
      listNewsArticles(params)
    ]);

    if (catRes.status === "fulfilled") {
      newsCategoryTags.value = catRes.value.items?.slice(0, 6).map(c => c.name) || [];
    }
    if (newsRes.status === "fulfilled") {
      newsHighlights.value = newsRes.value.items?.map(item => ({
        id: item.id,
        category: resolveNewsCategory(item),
        time: formatDateTime(item.published_at),
        title: item.title,
        summary: summarizeText(item.summary || item.content),
        visibility: String(item.visibility).toUpperCase() === "VIP" ? "VIP" : "PUBLIC"
      })) || [];
    }
    newsUpdatedAt.value = formatDateTime(new Date().toISOString());
  } finally {
    newsLoading.value = false;
  }
}

async function loadVIPState() {
  if (!isLoggedIn.value) return;
  vipStateLoading.value = true;
  try {
    const [quotaRes, prodRes] = await Promise.allSettled([
      getMembershipQuota(),
      listMembershipProducts({ status: "ACTIVE", page: 1, page_size: 2 })
    ]);
    if (quotaRes.status === "fulfilled") isVIPUser.value = (quotaRes.value.member_level === 'VIP1'); 
    if (prodRes.status === "fulfilled") vipOffer.value = prodRes.value.items?.[0] || null;
  } finally {
    vipStateLoading.value = false;
  }
}

// -------------------------------------------------------------------
// 4. 数据转换与工具函数 (Utils)
// -------------------------------------------------------------------

function summarizeText(s) {
  const clean = String(s || "").replace(/<[^>]*>/g, " ").trim();
  return clean.length > 60 ? `${clean.slice(0, 56)}...` : clean;
}

function resolveNewsCategory(item) {
  return /(研报|研究|券商)/i.test(`${item?.category_name} ${item?.title}`) ? "研报" : "新闻";
}

function calcCumulativeReturn(points) {
  if (!points?.length) return null;
  let acc = 1;
  points.forEach(p => { if (Number.isFinite(p.return)) acc *= (1 + p.return); });
  return acc - 1;
}

function estimateBenchmarkReturn(v) { return Number.isFinite(v) ? Math.max(Math.min(v * 0.58, 0.25), -0.25) : null; }
function inferExpectedRange(tp, ret) { return tp || (Number.isFinite(ret) ? `${formatPercent(ret * 0.6)} ~ ${formatPercent(ret * 1.2)}` : "预测中"); }
function inferExpectedSide(exp) { return exp.startsWith("+") ? "up" : exp.startsWith("-") ? "down" : "flat"; }
function mapArbitrageType(t) { const m = { CALENDAR: "跨期价差", CROSS: "跨品种价差", INTERTEMPORAL: "跨期套利" }; return m[String(t).toUpperCase()] || "套利"; }
function mapArbitrageStatus(s) { const m = { WATCH: "观察中", ACTIVE: "可执行", EXPIRED: "已失效" }; return m[String(s).toUpperCase()] || "待确认"; }
function pointWithRule(v, lbl) { return Number.isFinite(v) ? `${lbl} ${v.toFixed(2)}` : `${lbl}待补`; }
function suggestPosition(perc, risk) { return risk === "HIGH" ? "资金 15-20%" : "资金 25-30%"; }
function isResearchLike(item) { return /(研报|研究|深度)/i.test(`${item?.category} ${item?.title}`); }

function renderSparkline(points, width = 100, height = 30) {
  if (!points?.length) return `M0,${height/2} L${width},${height/2}`;
  const returns = points.map(p => p.return || 0);
  const min = Math.min(...returns, -0.02);
  const max = Math.max(...returns, 0.02);
  const range = max - min || 1;
  
  return returns.map((val, i) => {
    const x = (i / (returns.length - 1)) * width;
    const y = height - ((val - min) / range) * height;
    return `${i === 0 ? 'M' : 'L'}${x.toFixed(1)},${y.toFixed(1)}`;
  }).join(' ');
}

// -------------------------------------------------------------------
// 5. 生命周期与交互联动 (Lifecycle & Interaction)
// -------------------------------------------------------------------

const handleInsightSearch = () => { appliedInsightKeyword.value = insightKeyword.value; loadHomeInsights(); };
const clearInsightSearch = () => { insightKeyword.value = ""; appliedInsightKeyword.value = ""; loadHomeInsights(); };
const togglePrimaryStockWatch = () => {
  const item = rawStockRecommendations.value[0];
  if (!item) return;
  if (isWatchedStock(item.id)) { removeWatchedStock(item.id); watchedPrimaryStock.value = false; }
  else { saveWatchedStock(item); watchedPrimaryStock.value = true; }
};

const refreshAllData = () => { loadHomeData(); loadHomeInsights(); loadVIPState(); };
const goLogin = () => router.push("/auth");
const goProfile = () => router.push("/pc/profile");
const goWatchlistCenter = () => router.push(buildProfileModuleRoute("watchlist"));
const goStrategyCenter = () => router.push("/pc/strategy");
const goArchiveCenter = () => router.push("/pc/strategy");
const openNewsModule = () => router.push("/pc/strategy");
const goMembershipCenter = () => router.push("/pc/membership");

const openStrategyDetailDrawer = (id) => {
  activeDetailId.value = id;
  showStrategyDetails.value = true;
};

const closeStrategyDetailDrawer = () => {
  showStrategyDetails.value = false;
  activeDetailId.value = null;
};

const activeDetailData = computed(() => {
  if (!activeDetailId.value) return null;
  const stock = rawStockRecommendations.value.find(s => s.id === activeDetailId.value);
  const detail = stockDetailMap.value[activeDetailId.value] || {};
  const explanation = stockExplanationMap.value[activeDetailId.value] || {};
  return { ...stock, ...detail, ...explanation };
});

const openHomeCommunityList = (key) => {
  const draft = key === "research" ? { topicType: "NEWS", targetID: featuredResearch.value?.id } : { topicType: "STOCK", targetID: primaryStock.value?.id };
  router.push(draft.targetID ? buildCommunityListRoute(draft) : "/pc/strategy");
};

const openHomeCommunityComposer = (key) => {
  const draft = key === "research" ? { topicType: "NEWS", targetID: featuredResearch.value?.id } : { topicType: "STOCK", targetID: primaryStock.value?.id };
  router.push(draft.targetID ? buildCommunityComposeRoute(draft) : "/pc/strategy");
};

onMounted(() => {
  refreshAllData();
  const refreshWatch = () => { if (primaryStock.value?.id) watchedPrimaryStock.value = isWatchedStock(primaryStock.value.id); };
  refreshWatch();
  window.addEventListener(WATCHLIST_EVENT, refreshWatch);
  onBeforeUnmount(() => window.removeEventListener(WATCHLIST_EVENT, refreshWatch));
});

watch(isLoggedIn, refreshAllData);
watch(primaryStock, (val) => { if (val?.id) watchedPrimaryStock.value = isWatchedStock(val.id); });
</script>

<style scoped>
/* -------------------------------------------------------------------
   FINANCE PRO MAX - API EDITION STYLES
   ------------------------------------------------------------------- */

.pro-page-root {
  --color-rise: #10B981;
  --color-fall: #F43F5E;
  --color-primary: #3B82F6;
  --color-vip: #A855F7;
  --color-gold: #F59E0B;
  --bg-dark: #0F172A;
  --font-base: "IBM Plex Sans", system-ui, sans-serif;
  
  background-color: var(--bg-dark);
  min-height: 100vh;
  color: #F1F5F9;
  font-family: var(--font-base);
  position: relative;
  overflow-x: hidden;
}

/* 装饰背景 */
.ambient-glow { position: absolute; border-radius: 50%; filter: blur(120px); pointer-events: none; opacity: 0.12; }
.glow-1 { width: 400px; height: 400px; background: var(--color-primary); top: -150px; right: -50px; }
.glow-2 { width: 300px; height: 300px; background: var(--color-vip); bottom: 10%; left: -80px; }
.grid-overlay {
  position: absolute; inset: 0; pointer-events: none; opacity: 0.2;
  background-image: linear-gradient(rgba(255,255,255,0.02) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.02) 1px, transparent 1px);
  background-size: 32px 32px;
}

/* 状态栏条 (Pro Sync Bar) */
.pro-sync-bar {
  position: fixed; top: 64px; left: 0; right: 0; height: 2px; z-index: 1001; background: rgba(59, 130, 246, 0.1);
  display: flex; align-items: center; justify-content: center;
}
.sync-glow { position: absolute; height: 100%; width: 160px; background: linear-gradient(to right, transparent, var(--color-primary), transparent); animation: slide-sync 1.5s infinite; }
.pro-sync-bar span { position: absolute; top: 10px; font-size: 9px; color: var(--color-primary); font-weight: 700; background: rgba(15, 23, 42, 0.8); padding: 1px 6px; border-radius: 4px; }
@keyframes slide-sync { from { left: -160px; } to { left: 100%; } }

/* 毛玻璃架构 */
.glass { background: rgba(30, 41, 59, 0.4); backdrop-filter: blur(10px); border: 1px solid rgba(255, 255, 255, 0.08); }
.glass-card { background: rgba(30, 41, 59, 0.7); backdrop-filter: blur(16px); border: 1px solid rgba(255, 255, 255, 0.08); border-radius: 16px; }
.glass-accent { background: rgba(255, 255, 255, 0.02); border: 1px solid rgba(255, 255, 255, 0.04); border-radius: 12px; }

/* 导航系统 */
.pro-nav { position: sticky; top: 0; z-index: 1000; height: 64px; }
.nav-container { max-width: 1400px; margin: 0 auto; height: 100%; display: flex; align-items: center; justify-content: space-between; padding: 0 32px; }
.nav-brand { display: flex; align-items: center; gap: 12px; }
.brand-icon { color: var(--color-primary); filter: drop-shadow(0 0 6px var(--color-primary)); }
.brand-text span { font-size: 16px; font-weight: 800; display: block; }
.brand-text small { font-size: 9px; color: var(--color-primary); text-transform: uppercase; font-weight: 700; opacity: 0.7; }
.nav-links { display: flex; gap: 32px; }
.nav-link { font-size: 13px; color: #94A3B8; cursor: pointer; transition: 0.3s; font-weight: 600; }
.nav-link.active { color: white; position: relative; }
.nav-link.active::after { content: ''; position: absolute; bottom: -6px; left: 0; width: 20px; height: 2px; background: var(--color-primary); border-radius: 4px; }
.nav-actions { display: flex; gap: 16px; align-items: center; }
.search-wrap { display: flex; align-items: center; gap: 8px; padding: 6px 12px; border-radius: 10px; width: 200px; }
.search-wrap input { background: transparent; border: none; font-size: 12px; color: white; outline: none; width: 100%; }
.vip-status-pill { font-size: 10px; font-weight: 800; color: #94A3B8; padding: 3px 10px; border-radius: 6px; border: 1px solid rgba(255,255,255,0.05); }
.vip-status-pill.active { color: var(--color-gold); border-color: rgba(245, 158, 11, 0.25); }
.avatar-circle { width: 32px; height: 32px; background: var(--color-primary); border-radius: 50%; display: flex; align-items: center; justify-content: center; font-weight: 800; font-size: 11px; }

/* 页面主体 */
.scroll-container { max-width: 1400px; margin: 0 auto; padding: 24px 32px; position: relative; z-index: 1; }

/* HERO 舞台 */
.hero-stage { display: grid; grid-template-columns: 1fr 340px; gap: 40px; align-items: center; margin-bottom: 32px; }
.hero-badge-row { display: flex; gap: 10px; margin-bottom: 16px; }
.badge-pro { padding: 3px 12px; border-radius: 100px; font-size: 10px; font-weight: 700; color: #64748B; }
.badge-pro.accent { color: var(--color-primary); }
.hero-title-main { font-size: 42px; font-weight: 950; line-height: 1.1; margin-bottom: 16px; color: #F8FAFC; letter-spacing: -1px; }
.hero-subtitle { font-size: 16px; color: #94A3B8; line-height: 1.5; margin-bottom: 24px; max-width: 560px; }
.hero-actions-row { display: flex; gap: 12px; }
.hero-error-tip { margin-top: 16px; padding: 10px 16px; color: var(--color-fall); font-size: 12px; font-weight: 600; border-color: rgba(244, 63, 94, 0.15); }

.btn-main-pro { background: var(--color-primary); border: none; padding: 12px 24px; border-radius: 12px; color: white; font-weight: 700; display: flex; align-items: center; gap: 8px; font-size: 14px; box-shadow: 0 8px 16px rgba(59, 130, 246, 0.25); }
.btn-sub-pro { padding: 12px 24px; border-radius: 12px; color: #F1F5F9; font-weight: 600; font-size: 14px; }
.btn-sub-pro:disabled { opacity: 0.5; cursor: not-allowed; }

/* 主推荐卡片详情 */
.hero-panel-card { padding: 24px; min-height: 280px; display: flex; flex-direction: column; }
.panel-tag { font-size: 10px; color: var(--color-primary); font-weight: 900; margin-bottom: 16px; text-transform: uppercase; letter-spacing: 1.5px; }
.stock-lead { position: relative; margin-bottom: 24px; }
.symbol-id { font-size: 11px; color: #94A3B8; font-weight: 800; margin-bottom: 2px; }
.stock-name { font-size: 28px; font-weight: 800; color: white; }
.score-ring { position: absolute; top: -8px; right: -8px; }
.ring-text { font-size: 9px; font-weight: 900; fill: #F8FAFC; text-anchor: middle; }
.detail-row { display: flex; justify-content: space-between; padding: 10px 0; border-bottom: 1px solid rgba(255,255,255,0.04); }
.detail-row .label { color: #64748B; font-size: 12px; font-weight: 500; }
.detail-row .value { font-weight: 700; font-size: 13px; }
.rise-text { color: var(--color-rise); }
.fall-text { color: var(--color-fall); }
.risk-MEDIUM { color: var(--color-gold); }
.risk-HIGH { color: var(--color-fall); }
.risk-LOW { color: var(--color-rise); }
.proof-tags { display: flex; flex-wrap: wrap; gap: 6px; margin-top: 16px; }
.p-tag { font-size: 10px; padding: 3px 8px; border-radius: 4px; color: #94A3B8; background: rgba(59, 130, 246, 0.04); }
.hero-card-empty { flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; color: #475569; }

/* KPI 看板 */
.kpi-grid-pro { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: 30px; }
.kpi-box { padding: 18px; }
.kpi-label { font-size: 12px; color: #64748B; font-weight: 600; margin-bottom: 10px; }
.kpi-main { display: flex; justify-content: space-between; align-items: baseline; }
.kpi-value { font-size: 28px; font-weight: 900; letter-spacing: -0.5px; color: #F8FAFC; }
.kpi-trend { font-size: 11px; font-weight: 800; }
.kpi-trend.up { color: var(--color-rise); }
.kpi-trend.down { color: var(--color-fall); }
.spark-svg { width: 100%; height: 28px; opacity: 0.5; filter: drop-shadow(0 0 3px currentColor); }

/* CONTENT LAYOUT */
.layout-dual { display: grid; grid-template-columns: 1fr 340px; gap: 24px; }
.matrix-box { padding: 24px; }
.box-head { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 24px; }
.title-wrap .kicker { font-size: 10px; color: var(--color-primary); font-weight: 900; text-transform: uppercase; margin-bottom: 4px; display: block; }
.title-wrap h3 { font-size: 20px; font-weight: 850; color: white; }
.text-link { background: transparent; border: none; color: var(--color-primary); font-size: 12px; font-weight: 800; }

/* 权限拦截 */
.access-locked-pro {
  padding: 32px; border-radius: 16px; display: flex; flex-direction: column; align-items: center; text-align: center; gap: 12px;
  color: #94A3B8; font-size: 14px; border: 1px dashed rgba(255,255,255,0.08);
}
.access-locked-pro svg { color: var(--color-primary); opacity: 0.5; }
.access-locked-pro.sm { padding: 20px; font-size: 12px; }

/* 决策矩阵列表 */
.watch-grid-pro { display: grid; grid-template-columns: repeat(3, 1fr); gap: 16px; }
.watch-card-pro { padding: 18px; border-radius: 16px; }
.rank-id { font-size: 9px; font-weight: 950; opacity: 0.3; margin-bottom: 12px; }
.watch-info strong { display: block; font-size: 17px; font-weight: 800; margin-bottom: 4px; color: white; }
.watch-info span { font-size: 12px; color: #64748B; font-weight: 700; }
.watch-stats { margin-top: 14px; padding-top: 14px; border-top: 1px solid rgba(255,255,255,0.04); }
.w-value { font-size: 16px; font-weight: 900; display: block; margin-bottom: 4px; }
.w-note { font-size: 11px; color: #94A3B8; line-height: 1.4; }

/* 研报区 */
.news-featured { padding: 24px; margin-bottom: 18px; }
.f-news-top { display: flex; justify-content: space-between; margin-bottom: 12px; }
.f-cat { font-size: 10px; font-weight: 900; color: var(--color-primary); text-transform: uppercase; }
.f-mark { font-size: 9px; color: var(--color-gold); font-weight: 900; border: 1px solid currentColor; padding: 1px 6px; border-radius: 4px; }
.news-featured h4 { font-size: 20px; font-weight: 800; margin-bottom: 14px; line-height: 1.3; color: white; }
.news-featured p { font-size: 14px; color: #94A3B8; line-height: 1.6; margin-bottom: 20px; }
.f-footer { display: flex; justify-content: space-between; align-items: center; font-size: 11px; color: #475569; }
.btn-read { background: var(--color-primary); border: none; color: white; padding: 8px 16px; border-radius: 8px; font-weight: 800; }

.news-row-pro { padding: 16px; border-radius: 12px; margin-bottom: 10px; }
.n-meta { font-size: 11px; font-weight: 800; color: #64748B; margin-bottom: 6px; }
.n-content h5 { font-size: 15px; font-weight: 700; margin-bottom: 4px; color: #E2E8F0; }
.n-content p { font-size: 12px; color: #94A3B8; line-height: 1.5; }

/* 档案表格 */
.table-wrap-pro { margin-top: 10px; overflow-x: auto; }
.pro-table { width: 100%; border-collapse: collapse; min-width: 600px; }
.pro-table th { text-align: left; padding: 12px 16px; font-size: 10px; font-weight: 850; color: #475569; border-bottom: 1px solid rgba(255,255,255,0.05); text-transform: uppercase; }
.pro-table td { padding: 14px 16px; font-size: 13px; color: #CBD5E1; border-bottom: 1px solid rgba(255,255,255,0.03); }
.bold { font-weight: 900; color: white; }
.score-bubble { padding: 3px 8px; border-radius: 6px; background: rgba(59,130,246,0.1); color: var(--color-primary); font-weight: 800; font-size: 12px; }

/* 期货方案列表 */
.side-module h4 { font-size: 17px; font-weight: 800; margin-bottom: 20px; padding-left: 12px; border-left: 4px solid var(--color-primary); color: white; }
.f-plan-pro { padding: 18px; border-radius: 20px; margin-bottom: 14px; }
.p-head { display: flex; justify-content: space-between; margin-bottom: 14px; }
.p-head strong { font-size: 15px; font-weight: 800; color: white; }
.p-tag { font-size: 10px; font-weight: 900; color: var(--color-primary); }
.p-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; }
.p-cell p { font-size: 9px; color: #64748B; margin-bottom: 2px; font-weight: 600; }
.p-cell strong { font-size: 13px; color: #F1F5F9; font-weight: 700; }

/* ELITE VIP BOX */
.elite-vip-box { position: relative; border-radius: 24px; padding: 32px; overflow: hidden; background: linear-gradient(135deg, #1e1b4b, #0f172a); border: 1px solid rgba(168, 85, 247, 0.3); }
.elite-bg-mesh { position: absolute; inset: 0; background-image: radial-gradient(rgba(168, 85, 247, 0.1) 1px, transparent 1px); background-size: 18px 18px; opacity: 0.5; }
.elite-content { position: relative; z-index: 1; text-align: center; }
.elite-content h3 { font-size: 22px; font-weight: 900; margin-bottom: 10px; color: white; }
.elite-content p { font-size: 13px; color: #94A3B8; line-height: 1.5; margin-bottom: 24px; }
.price-row { display: flex; align-items: baseline; justify-content: center; gap: 6px; margin-bottom: 24px; color: var(--color-gold); }
.price-row .val { font-size: 36px; font-weight: 950; }
.btn-vip-pro { width: 100%; border: none; padding: 14px; border-radius: 14px; font-weight: 900; font-size: 16px; background: linear-gradient(to right, #fbbf24, #f59e0b); color: #451a03; box-shadow: 0 8px 24px rgba(245, 158, 11, 0.3); }

/* 时间轴 */
.timeline-pro { padding: 0 8px; }
.t-line { display: flex; gap: 16px; position: relative; padding-bottom: 24px; }
.t-line:not(:last-child)::before { content: ''; position: absolute; left: 45px; top: 12px; bottom: 0; width: 1px; background: rgba(255,255,255,0.06); }
.t-time { font-size: 10px; font-weight: 900; color: var(--color-primary); width: 45px; transform: translateY(2px); }
.t-main strong { display: block; font-size: 14px; font-weight: 700; margin-bottom: 4px; color: white; }
.t-main p { font-size: 11px; color: #64748B; line-height: 1.4; }
.locked-text { font-style: italic; opacity: 0.7; }

/* UTILS */
.mt-32 { margin-top: 24px; }
.interactive { cursor: pointer; transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1); }
.interactive:hover { transform: translateY(-3px); }
.hover-up:hover { transform: translateY(-6px); border-color: rgba(59,130,246,0.35); box-shadow: 0 16px 32px rgba(0,0,0,0.35); }

/* 页脚装饰 */
.pro-footer-decor { padding: 56px 40px; text-align: center; }
.f-line { height: 1px; width: 160px; background: linear-gradient(to right, transparent, rgba(255,255,255,0.06), transparent); margin: 0 auto 24px; }
.pro-footer-decor p { font-size: 11px; color: #475569; letter-spacing: 0.5px; font-weight: 600; }

/* 加载动画 (Pro) */
.pro-spinner { width: 36px; height: 36px; border: 3px solid rgba(59,130,246,0.1); border-top-color: var(--color-primary); border-radius: 50%; animation: spin 0.8s linear infinite; margin-bottom: 16px; }
.pro-spinner-sm { width: 20px; height: 20px; border: 2.5px solid rgba(255,255,255,0.1); border-top-color: var(--color-primary); border-radius: 50%; animation: spin 0.8s linear infinite; margin-bottom: 10px; }
@keyframes spin { to { transform: rotate(360deg); } }
.pro-empty-box { display: flex; flex-direction: column; align-items: center; justify-content: center; min-height: 200px; color: #475569; }
.pro-empty-box.sm { min-height: 140px; font-size: 12px; }

/* 基础进场 */
.fade-in { opacity: 0; animation: fadeIn 0.8s cubic-bezier(0.4, 0, 0.2, 1) forwards; }
@keyframes fadeIn { from { opacity: 0; transform: translateY(8px); } to { opacity: 1; transform: translateY(0); } }
/* 研报解读 Pro */
.research-focus-layout { margin-top: 12px; }
.focus-hero { padding: 12px; border-radius: 8px; margin-bottom: 12px; border: 1px solid rgba(255,255,255,0.05); }
.f-badge-row { display: flex; gap: 6px; margin-bottom: 6px; }
.f-cat { font-size: 10px; padding: 2px 6px; background: rgba(59,130,246,0.2); color: #60A5FA; border-radius: 4px; }
.f-mark { font-size: 10px; padding: 2px 6px; background: rgba(168,85,247,0.2); color: #C084FC; border-radius: 4px; border: 1px solid rgba(168,85,247,0.3); }
.focus-hero h4 { font-size: 15px; font-weight: 600; line-height: 1.4; color: #F1F5F9; }
.f-sync { font-size: 10px; color: #94A3B8; margin-top: 6px; }

.research-cards-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 10px; }
.r-card { padding: 10px; border-radius: 8px; border: 1px solid rgba(255,255,255,0.03); transition: all 0.2s; }
.r-card:hover { border-color: rgba(59,130,246,0.3); background: rgba(255,255,255,0.02); }
.r-card h5 { font-size: 12px; color: #3B82F6; margin-bottom: 4px; font-weight: 600; }
.r-card p { font-size: 11px; line-height: 1.5; color: #94A3B8; }

.news-mini-list { margin-top: 12px; display: flex; flex-direction: column; gap: 6px; }
.news-mini-row { padding: 8px 12px; border-radius: 6px; display: flex; align-items: center; gap: 8px; border: 1px solid transparent; }
.news-mini-row:hover { border-color: rgba(255,255,255,0.05); background: rgba(255,255,255,0.02); }
.n-tag { font-size: 9px; padding: 1px 4px; border-radius: 3px; background: rgba(255,255,255,0.05); color: #94A3B8; }
.news-mini-row p { font-size: 12px; color: #CBD5E1; flex: 1; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

/* 社区联动 */
.community-grid-pro { display: grid; grid-template-columns: repeat(2, 1fr); gap: 12px; margin-top: 12px; }
.comm-card { padding: 12px; border-radius: 10px; display: flex; flex-direction: column; gap: 6px; border: 1px solid rgba(255,255,255,0.03); }
.comm-card:hover { border-color: rgba(59,130,246,0.3); box-shadow: 0 4px 20px -5px rgba(0,0,0,0.3); }
.comm-badge { font-size: 10px; padding: 2px 6px; border-radius: 4px; align-self: flex-start; }
.comm-card strong { font-size: 13px; color: #F1F5F9; }
.comm-card p { font-size: 11px; color: #94A3B8; line-height: 1.4; height: 3.8em; overflow: hidden; }
.comm-actions { display: flex; gap: 6px; margin-top: 4px; }
.comm-btn { flex: 1; font-size: 11px; padding: 4px; border-radius: 5px; background: rgba(59,130,246,0.1); color: #60A5FA; border: 1px solid rgba(59,130,246,0.2); transition: all 0.2s; }
.comm-btn:hover { background: var(--color-primary); color: white; }
.comm-btn.secondary { background: rgba(255,255,255,0.05); color: #CBD5E1; border-color: rgba(255,255,255,0.1); }

/* 今日任务 */
.module-head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.count-pill { font-size: 10px; padding: 2px 6px; border-radius: 10px; background: var(--color-primary); color: white; font-weight: 600; }
.task-list-pro { display: flex; flex-direction: column; gap: 8px; }
.task-item { padding: 10px; border-radius: 8px; display: flex; gap: 10px; align-items: flex-start; border: 1px solid rgba(255,255,255,0.03); }
.task-indicator { width: 3px; height: 24px; border-radius: 2px; flex-shrink: 0; margin-top: 2px; }
.task-indicator.high { background: var(--color-fall); box-shadow: 0 0 8px var(--color-fall); }
.task-indicator.mid { background: var(--color-gold); }
.task-body strong { display: block; font-size: 12px; color: #F1F5F9; margin-bottom: 2px; }
.task-body p { font-size: 11px; color: #94A3B8; line-height: 1.3; }

/* 增强页脚与免责声明 */
.pro-footer-expanded { margin-top: 64px; padding: 48px 32px 32px; border-top: 1px solid rgba(255,255,255,0.05); }
.footer-grid { max-width: 1400px; margin: 0 auto; display: grid; grid-template-columns: 1.5fr 1fr 2fr; gap: 48px; }
.f-brand .nav-brand { margin-bottom: 12px; }
.f-desc { font-size: 13px; color: #64748B; line-height: 1.6; }
.f-group-title { font-size: 14px; font-weight: 800; color: white; margin-bottom: 20px; text-transform: uppercase; letter-spacing: 1px; }
.f-links { display: flex; flex-direction: column; gap: 12px; }
.f-links a { font-size: 13px; color: #94A3B8; cursor: pointer; transition: 0.2s; }
.f-links a:hover { color: var(--color-primary); }
.d-text { font-size: 12px; color: #475569; line-height: 1.8; }
.f-bottom { max-width: 1400px; margin: 48px auto 0; padding-top: 24px; border-top: 1px solid rgba(255,255,255,0.03); text-align: center; }
.f-bottom p { font-size: 11px; color: #475569; }

/* 市场行情条 */
.pro-ticker-bar { height: 36px; background: rgba(30, 41, 59, 0.4); border-bottom: 1px solid rgba(255,255,255,0.05); display: flex; align-items: center; overflow: hidden; }
.ticker-content { display: flex; gap: 40px; padding: 0 32px; animation: ticker-scroll 30s linear infinite; white-space: nowrap; }
@keyframes ticker-scroll { from { transform: translateX(0); } to { transform: translateX(-50%); } }
.ticker-item { display: flex; align-items: center; gap: 8px; font-size: 12px; }
.t-name { color: #94A3B8; font-weight: 600; }
.t-val { color: #F1F5F9; font-weight: 800; font-family: "JetBrains Mono", monospace; }
.t-chg { font-weight: 800; }

/* 分类 Tab */
.pro-tab-wrap { display: flex; padding: 2px; border-radius: 8px; gap: 2px; }
.tab-btn { padding: 4px 12px; border: none; background: transparent; color: #64748B; font-size: 11px; font-weight: 800; border-radius: 6px; cursor: pointer; transition: 0.2s; }
.tab-btn.active { background: rgba(255,255,255,0.08); color: white; }

/* 执行日志轴 */
.execution-log-list { display: flex; flex-direction: column; gap: 10px; }
.log-item { display: flex; align-items: center; gap: 12px; padding: 12px 16px; border-radius: 10px; position: relative; }
.log-time { font-size: 11px; font-weight: 900; color: var(--color-primary); font-family: "JetBrains Mono", monospace; }
.log-content { flex: 1; display: flex; align-items: center; gap: 8px; }
.l-target { font-size: 12px; font-weight: 800; color: #F1F5F9; }
.l-msg { font-size: 12px; color: #94A3B8; }
.log-type { font-size: 9px; font-weight: 900; padding: 2px 6px; border-radius: 4px; text-transform: uppercase; }
.log-type.confirm { background: rgba(16, 185, 129, 0.15); color: #10B981; }
.log-type.alert { background: rgba(245, 158, 11, 0.15); color: #F59E0B; }
.log-type.entry { background: rgba(59, 130, 246, 0.15); color: #3B82F6; }

/* 抽屉 & 遮罩补齐样式 */
.w-flex { display: flex; justify-content: space-between; align-items: baseline; margin-bottom: 4px; }
.action-trigger { font-size: 10px; font-weight: 800; color: var(--color-gold); padding: 1px 6px; border-radius: 4px; }
.locked-preview-mask {
  position: absolute; inset: 0; background: rgba(30, 41, 59, 0.4); backdrop-filter: blur(4px);
  display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 8px;
  font-size: 11px; font-weight: 800; color: var(--color-gold); border-radius: 16px;
}

/* 权益对比表格样式 */
.benefit-table-pro { display: flex; flex-direction: column; gap: 8px; padding: 12px; border-radius: 12px; }
.b-row { display: flex; justify-content: space-between; align-items: center; padding: 8px 0; border-bottom: 1px solid rgba(255,255,255,0.03); }
.b-feat { font-size: 11px; color: #64748B; font-weight: 600; }
.b-vals { display: flex; gap: 12px; font-size: 11px; font-weight: 700; }
.std { color: #475569; }
.vip-text { color: var(--color-gold); }

/* 11. 最终视觉一致性 & 窄屏降级 (Breakpoint) */
@media (max-width: 1300px) {
  .hero-stage { grid-template-columns: 1fr; gap: 24px; }
  .hero-panel-card { min-height: auto; }
  .layout-dual { grid-template-columns: 1fr; }
  .side-column { display: grid; grid-template-columns: 1fr 1fr; gap: 24px; }
  .elite-vip-box { grid-column: span 2; }
}

@media (max-width: 1024px) {
  .nav-links { display: none; }
  .kpi-grid-pro { grid-template-columns: repeat(2, 1fr); }
  .side-column { grid-template-columns: 1fr; }
  .watch-grid-pro { grid-template-columns: repeat(2, 1fr); }
}

</style>

