export const USER_PROFILE = {
  nickname: '股票投资者',
  avatar: '',
  level: 'Lv.5',
  credits: { total: 100, used: 28, remaining: 72, nextReset: '2025-06-01' },
  vip: { tier: 'silver', tierName: '白银会员', expiresAt: '2025-08-15', benefits: ['每日50次AI分析', '进阶技术指标', 'Level-2行情', '无广告'] },
  settings: { notifications: { priceAlert: true, newsAlert: true, systemNotice: true }, display: { theme: 'dark', language: 'zh' } },
  portfolio: {
    totalAssets: 1286542,
    todayPnl: 12845,
    todayPnlPct: 2.59,
    totalReturn: 28.65,
    holdings: [
      { symbol: '300750.SZ', name: '宁德时代', shares: 1500, avgCost: 176.80, currentPrice: 198.62, marketValue: 297930, profitPct: 12.45 },
      { symbol: '600941.SH', name: '中国移动', shares: 3000, avgCost: 101.50, currentPrice: 106.80, marketValue: 320400, profitPct: 5.82 },
      { symbol: '002415.SZ', name: '海康威视', shares: 5000, avgCost: 36.05, currentPrice: 35.27, marketValue: 176350, profitPct: -2.15 },
      { symbol: '600036.SH', name: '招商银行', shares: 4000, avgCost: 35.62, currentPrice: 36.78, marketValue: 147120, profitPct: 3.28 }
    ]
  },
  watchlist: [
    { symbol: '688981.SH', name: '中芯国际', addedAt: '2025-05-01' },
    { symbol: '002594.SZ', name: '比亚迪', addedAt: '2025-04-28' },
    { symbol: '000858.SZ', name: '五粮液', addedAt: '2025-04-25' },
    { symbol: '300059.SZ', name: '东方财富', addedAt: '2025-04-20' },
    { symbol: '002230.SZ', name: '科大讯飞', addedAt: '2025-04-15' }
  ],
  history: [
    { id: 'h1', type: 'analysis', title: '宁德时代深度分析报告', stock: '300750.SZ', date: '2025-05-14T10:30', creditsUsed: 1 },
    { id: 'h2', type: 'analysis', title: '中国移动投资价值分析', stock: '600941.SH', date: '2025-05-13T15:20', creditsUsed: 1 },
    { id: 'h3', type: 'chat', title: 'AI对话：半导体板块投资机会', stock: null, date: '2025-05-12T14:00', creditsUsed: 2 },
    { id: 'h4', type: 'analysis', title: '比亚迪技术面分析', stock: '002594.SZ', date: '2025-05-11T09:30', creditsUsed: 1 },
    { id: 'h5', type: 'analysis', title: '海康威视AI概念评估', stock: '002415.SZ', date: '2025-05-10T16:45', creditsUsed: 1 },
    { id: 'h6', type: 'chat', title: 'AI对话：2025下半年投资策略', stock: null, date: '2025-05-09T20:00', creditsUsed: 2 },
    { id: 'h7', type: 'analysis', title: '招商银行股息率分析', stock: '600036.SH', date: '2025-05-08T11:20', creditsUsed: 1 },
    { id: 'h8', type: 'analysis', title: '东方财富券商板块分析', stock: '300059.SZ', date: '2025-05-07T14:30', creditsUsed: 1 }
  ]
};

export const VIP_TIERS = [
  { tier: 'free', name: '免费版', price: 0, period: '', popular: false, icon: 'free', benefits: ['每日 5 次 AI 分析', '基础市场数据', '社区访问权限', '基础行情浏览'], features: { analysisCount: '5次/日', advancedMetrics: false, level2: false, realtime: false, advisor: false, ads: true } },
  { tier: 'silver', name: '白银会员', price: 99, period: '月', popular: true, icon: 'silver', benefits: ['每日 50 次 AI 分析', '进阶技术指标', 'Level-2 行情', '无广告', 'VIP 交流圈'], features: { analysisCount: '50次/日', advancedMetrics: true, level2: true, realtime: false, advisor: false, ads: false } },
  { tier: 'gold', name: '黄金会员', price: 199, period: '月', popular: false, icon: 'gold', benefits: ['无限次 AI 分析', 'VIP 专属策略', 'Level-2 行情', '实时盯盘提醒', 'AI 专属客服'], features: { analysisCount: '无限', advancedMetrics: true, level2: true, realtime: true, advisor: false, ads: false } },
  { tier: 'platinum', name: '铂金会员', price: 499, period: '月', popular: false, icon: 'platinum', benefits: ['所有黄金权益', '私募级策略', '一对一投顾服务', '优先体验新功能', '线下活动邀请'], features: { analysisCount: '无限', advancedMetrics: true, level2: true, realtime: true, advisor: true, ads: false } }
];

export const FEATURE_COMPARE = [
  { feature: 'AI 分析次数', free: '5次/日', silver: '50次/日', gold: '无限', platinum: '无限' },
  { feature: '技术指标', free: '基础', silver: '进阶', gold: '专业', platinum: '全功能' },
  { feature: '行情数据', free: '延时', silver: 'Level-2', gold: 'Level-2', platinum: 'Level-2' },
  { feature: '实时提醒', free: '×', silver: '×', gold: '✓', platinum: '✓' },
  { feature: '专属策略', free: '×', silver: '×', gold: '✓', platinum: '✓' },
  { feature: '投顾服务', free: '×', silver: '×', gold: 'AI客服', platinum: '真人投顾' },
  { feature: '社区特权', free: '基础', silver: 'VIP圈', gold: 'VIP圈', platinum: '全特权' }
];

export const CREDIT_USAGE_BY_DAY = [
  { day: '05/08', used: 3 },
  { day: '05/09', used: 5 },
  { day: '05/10', used: 2 },
  { day: '05/11', used: 6 },
  { day: '05/12', used: 4 },
  { day: '05/13', used: 3 },
  { day: '05/14', used: 5 }
];
