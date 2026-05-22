function buildLineChart(points, width = 420, height = 176, padding = 18) {
  const max = Math.max(...points);
  const min = Math.min(...points);
  const range = Math.max(max - min, 1);
  const innerWidth = width - padding * 2;
  const innerHeight = height - padding * 2;
  const dots = points.map((point, index) => {
    const x = padding + (innerWidth * index) / Math.max(points.length - 1, 1);
    const y = padding + innerHeight - ((point - min) / range) * innerHeight;
    return { key: `${index}-${point}`, x, y };
  });
  const line = dots.map((point, index) => `${index === 0 ? "M" : "L"} ${point.x} ${point.y}`).join(" ");
  const last = dots[dots.length - 1];
  const first = dots[0];
  const area = `${line} L ${last.x} ${height - padding} L ${first.x} ${height - padding} Z`;
  return { width, height, dots, line, area };
}

export const sparkline = buildLineChart([18, 24, 23, 31, 30, 39, 37, 46]);

export const homeDemoData = {
  trustMarks: ["机构研究流程", "多接口证据链", "历史样本回看", "会员分层承接", "期货策略扩展", "统一检索入口"],
  heroTags: ["主推荐 中风险", "历史命中 72%", "观察清单 4 只", "重点研报 1 条"],
  trustStats: [
    { label: "主推荐评分", value: "92", note: "先看强度，再决定是否进入策略详情。" },
    { label: "观察清单", value: "4 只", note: "首页优先聚焦最值得继续跟踪的对象。" },
    { label: "资讯线索", value: "3 条", note: "焦点研报解释为什么今天先看这些标的。" },
    { label: "历史样本", value: "6 条", note: "历史命中率重新接回今天的阅读链。" }
  ],
  primaryStock: {
    symbol: "300750",
    name: "宁德时代",
    risk: "中风险",
    score: "92 / 100",
    expected: "目标区间 212 - 226",
    takeProfit: "止盈 226",
    stopLoss: "风控线 198",
    reason: "储能链订单确认继续强化，今天更适合先验证强势延续，而不是追逐弱修复。"
  },
  featuredResearch: {
    title: "储能链与高股息同时占优，市场在寻找更确定的胜率资产",
    summary: "首页重点内容不再只是摆一篇文章，而是明确承接今日主推荐、资讯线索和后续社区讨论。",
    category: "焦点研报",
    visibility: "公开",
    time: "10:20 更新"
  },
  watchlist: [
    {
      rank: 1,
      symbol: "300750",
      name: "宁德时代",
      risk: "中风险",
      expected: "预期区间 212 - 226",
      note: "优先确认强势延续是否继续获得订单与情绪双重支持。"
    },
    {
      rank: 2,
      symbol: "600036",
      name: "招商银行",
      risk: "低风险",
      expected: "预期区间 37.8 - 39.6",
      note: "作为低波动防守线，适合做资金风险对冲参考。"
    },
    {
      rank: 3,
      symbol: "601899",
      name: "紫金矿业",
      risk: "中风险",
      expected: "预期区间 18.4 - 19.8",
      note: "金属价格韧性仍在，但需要继续盯住盘中量能确认。"
    },
    {
      rank: 4,
      symbol: "002371",
      name: "北方华创",
      risk: "中高风险",
      expected: "预期区间 334 - 356",
      note: "弹性更强，但必须把风险边界和止损纪律前置。"
    }
  ],
  insightCards: [
    {
      title: "这篇内容在说什么",
      desc: "把今天的主推荐放到更大的行业与情绪框架里解释，而不是单点推荐。"
    },
    {
      title: "它对今天推荐的影响",
      desc: "资讯继续强化储能与防守双主线，因此首页的推荐优先级也相应收敛。"
    },
    {
      title: "今天应该怎么用",
      desc: "先读摘要，再决定去资讯中心看全文，最后回到策略页确认动作边界。"
    }
  ],
  newsPulse: [
    { time: "08:30", label: "盘前导读", value: 46 },
    { time: "10:20", label: "热点确认", value: 72 },
    { time: "13:15", label: "午后分歧", value: 58 },
    { time: "15:30", label: "收盘复盘", value: 88 }
  ],
  historyRows: [
    { code: "300750", date: "06-18", risk: "中风险", alpha: "+12.4%", width: 84 },
    { code: "601899", date: "06-12", risk: "中风险", alpha: "+9.1%", width: 68 },
    { code: "600036", date: "06-05", risk: "低风险", alpha: "+5.8%", width: 49 },
    { code: "002371", date: "05-29", risk: "中高风险", alpha: "+14.2%", width: 92 },
    { code: "688041", date: "05-23", risk: "高风险", alpha: "+7.6%", width: 57 },
    { code: "603986", date: "05-15", risk: "中风险", alpha: "+10.3%", width: 74 }
  ],
  futuresPlan: {
    name: "RB2510 - HC2510 套利方案",
    summary: "首页只展示最关键的进场、平仓、止损区间，详细执行仍然留给深层策略页面。"
  },
  futuresLadder: [
    { label: "止损", value: "-84", offset: "18%" },
    { label: "触发", value: "-46", offset: "42%" },
    { label: "目标一", value: "-12", offset: "68%" },
    { label: "目标二", value: "+8", offset: "82%" }
  ],
  futuresSteps: [
    { title: "步骤 1：确认触发区间", desc: "只有在价差进入触发区间且成交量同步配合时才执行首笔建仓。" },
    { title: "步骤 2：写入平仓目标", desc: "目标区间不靠主观推测，先写好，再决定是否分批止盈。" },
    { title: "步骤 3：严格执行止损", desc: "当价差进入失效区间时立即止损，不把首页展示误认为交易建议。" }
  ],
  vipOffer: {
    name: "研究会员月度版",
    price: "¥199 / 月",
    desc: "解锁完整推荐池、午盘变化提醒、收盘复盘和更高配额。",
    cta: "查看会员权益"
  },
  conversionStages: [
    { title: "游客", badge: "当前入口", desc: "先看公开首页、主推荐和最新资讯。", state: "neutral" },
    { title: "注册用户", badge: "可保存关注", desc: "开始建立个人观察清单与回看路径。", state: "active" },
    { title: "会员", badge: "完整内容", desc: "进入完整推荐、复盘与更深层解释链。", state: "vip" }
  ],
  roleCards: [
    {
      kicker: "For Research Teams",
      title: "投研团队",
      desc: "更关心主推荐、解释链和历史样本，适合直接进入策略与档案页继续判断。",
      path: "/strategies",
      cta: "进入策略页"
    },
    {
      kicker: "For Advisors",
      title: "财富顾问",
      desc: "更关心资讯解释和会员转化路径，适合把首页作为展示和讲解入口。",
      path: "/news",
      cta: "进入资讯中心"
    },
    {
      kicker: "For Individual Users",
      title: "个人研究者",
      desc: "更关心检索、关注和社区承接，适合从首页分流到统一检索与讨论广场。",
      path: "/search",
      cta: "进入统一检索"
    }
  ],
  evidenceCards: [
    {
      kicker: "Why this homepage",
      title: "主推荐不是孤立信号",
      desc: "首页直接承接推荐理由、资讯线索和风险边界，让用户先理解再动作。"
    },
    {
      kicker: "Why this research",
      title: "重点研报负责解释逻辑是否延续",
      desc: "不是重复推荐，而是帮助判断今天的主线是否仍在强化。"
    },
    {
      kicker: "Why keep archives",
      title: "命中与失效都要留下复盘痕迹",
      desc: "这样首页才不是一次性推荐页，而是可持续的研究入口。"
    },
    {
      kicker: "Why conversion later",
      title: "转化放在阅读链之后",
      desc: "先让用户理解价值，再承接权益升级，不在首屏强推。"
    }
  ],
  nextActions: [
    {
      title: "进入讨论广场",
      desc: "看完主推荐和研报后，继续去社区补充观点分歧和市场反馈。",
      path: "/community"
    },
    {
      title: "进入统一检索",
      desc: "当今天的焦点不够明确，就让搜索先接手关键词、资讯和标的查询。",
      path: "/search"
    },
    {
      title: "查看会员权益",
      desc: "当用户已经理解价值，再承接权益升级和深度内容开放。",
      path: "/membership"
    }
  ]
};
