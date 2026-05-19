export function buildAnalysisReport(symbol) {
  const reports = {
    '300750.SZ': {
      coreLogic: '宁德时代作为全球动力电池龙头企业，当前受益于新能源汽车渗透率持续提升、储能业务高速增长两大核心驱动力。AI 模型综合评估认为，公司在技术路线（麒麟电池、钠离子电池）、产能布局（全球十大基地）和成本控制方面保持显著领先优势。',
      techAnalysis: '日线级别看，股价已突破前期整理平台（185-192 元区间），MACD 金叉向上，成交量温和放大，RSI 处于 62 中性偏强区域，尚未超买。周线级别呈现底部抬升态势，中期趋势向好。AI 技术模型给出 84 分（满分 100）。',
      fundAnalysis: '近 5 个交易日主力资金净流入 12.6 亿元，其中今日超大单净流入 3.8 亿元。北向资金持股比例由 8.2% 提升至 8.7%。融资余额持续增加，市场参与度显著提高。AI 资金模型给出 91 分。',
      riskWarning: '1) 行业竞争加剧，二三线电池厂价格战可能压缩利润率；2) 原材料价格波动风险，碳酸锂价格反弹可能影响成本；3) 海外政策不确定性，欧美电动车关税调整可能影响出口预期。AI 风控模型建议仓位控制在总资金 15% 以内。'
    },
    '600941.SH': {
      coreLogic: '中国移动作为国内最大通信运营商，5G/6G 网络建设持续领先，云计算和数字化业务快速增长。AI 模型认为其高股息属性在当前市场环境下具有显著防御价值。',
      techAnalysis: '股价沿 20 日均线稳步上行，波动率较低，呈现典型的价值股走势特征。RSI 处于 55 中性区间，MACD 轻微金叉，短期有望延续温和上涨。',
      fundAnalysis: '南向资金持续增持，持股比例创新高。社保基金和险资等长线资金持仓稳定。现金流充裕，分红率保持在 70% 以上。',
      riskWarning: '1) 市场竞争加剧，广电和民营运营商可能分流用户；2) 5G 投资回报周期较长；3) 云计算业务面临阿里云等竞争。'
    },
    '002415.SZ': {
      coreLogic: '海康威视作为全球安防龙头，AI 赋能转型成效显著。创新业务（机器人、汽车电子、热成像）快速增长，正在打开第二增长曲线。',
      techAnalysis: '股价放量突破 60 日均线，MACD 零轴上方金叉，短期动能强劲。布林带开口扩大，预示趋势可能加速。',
      fundAnalysis: '北向资金近 10 日净买入 8.5 亿元，两融余额持续增加。机构持仓比例提升至 68%。',
      riskWarning: '1) 海外制裁风险仍是最大不确定性；2) 国内安防市场增速放缓；3) 创新业务尚需时间验证盈利能力。'
    }
  };
  return reports[symbol] || {
    coreLogic: '该标的基本面稳健，行业地位突出。AI 模型正在进行多维度数据分析，建议持续关注。',
    techAnalysis: '技术面呈现震荡整理格局，短期方向尚不明确。建议等待明确信号后再做判断。',
    fundAnalysis: '资金面表现中性，无明显大额流入或流出迹象。',
    riskWarning: '1) 市场整体波动风险；2) 行业政策变化风险；3) 公司经营不确定性。以上分析由 AI 生成，仅供参考。'
  };
}

export function buildLogicChain(symbol) {
  return [
    { step: 1, title: '数据采集', icon: 'data', sources: ['日线行情数据 (90天)', 'Level-2 资金流向', '财报数据 (最近4季)', '新闻情绪分析', '社交媒体热度'], detail: '从多个数据源采集了与该标的相关的量价数据、资金流向、财务指标和舆情信息，共计 128 个原始特征。', score: 95 },
    { step: 2, title: '因子计算', icon: 'calc', factors: ['趋势因子: +8.2', '资金因子: +12.5', '估值因子: -3.1', '情绪因子: +6.8', '波动因子: -2.4'], detail: '基于12个技术指标、5个资金指标、8个基本面指标和3个情绪指标，计算出多维度因子得分。', score: 88 },
    { step: 3, title: '模型推理', icon: 'model', models: ['XGBoost 评分模型', 'LSTM 趋势预测', 'Transformer 情绪分析', '随机森林分类器'], detail: '集成 4 个深度学习模型的推理结果，经过加权融合后产出综合评分。', score: 91 },
    { step: 4, title: '结果生成', icon: 'output', output: { score: 92, judgment: '强烈看多', confidence: '87.3%', targetPrice: '215-228' }, detail: '综合各模型输出，生成最终评分、判断方向、目标价区间和风险等级。', score: 90 },
    { step: 5, title: '置信度评估', icon: 'confidence', checks: ['数据完整性检查 ✓', '模型一致性验证 ✓', '异常值检测 ✓', '历史回测校验 ✓'], detail: '对最终结果进行多维度质量校验，确保输出结果的可靠性和一致性。', score: 85 }
  ];
}

export const SCORE_DIMENSIONS = [
  { name: '技术面', score: 84, detail: 'MACD金叉、突破平台、量能配合' },
  { name: '资金面', score: 91, detail: '主力连续3日净流入、北向资金加仓' },
  { name: '基本面', score: 78, detail: '营收增长25%、利润率稳定' },
  { name: '情绪面', score: 88, detail: '新闻偏积极、社交媒体热度上升' },
  { name: '估值面', score: 65, detail: 'PE处于行业中上水平、PEG合理' }
];

export const INSIGHTS = [
  { title: '趋势判断', desc: '日线、周线同步多头排列，中期上涨趋势确立', color: 'var(--positive)' },
  { title: '资金信号', desc: '主力连续三日净流入，北向资金加速配置', color: 'var(--accent-gold)' },
  { title: '风险提示', desc: '短期涨幅较大，注意回调风险，建议分批建仓', color: 'var(--accent-cyan)' },
  { title: '目标价位', desc: 'AI 模型预测 1 个月内目标价 215-228 元区间', color: 'var(--accent-purple)' }
];

export const SENTIMENT_DATA = { positive: 62, neutral: 24, negative: 14 };

export const PRICE_LEVELS = [
  { label: '强阻力位', price: '228.00', note: '前高 + 筹码密集区', color: 'var(--negative)' },
  { label: '弱阻力位', price: '210.00', note: '整数关口', color: 'var(--accent-gold)' },
  { label: '当前价位', price: '198.62', note: 'AI 推荐建仓区间', color: 'var(--accent-cyan)' },
  { label: '弱支撑位', price: '185.00', note: '平台下沿', color: 'var(--accent-blue)' },
  { label: '强支撑位', price: '175.00', note: '60 日均线', color: 'var(--positive)' }
];

// Simulated LLM analysis response
export function generateLLMAnalysis(stockName, symbol) {
  const templates = [
    `【AI 智能分析报告 — ${stockName}（${symbol}）】\n\n` +
    `一、公司概览\n${stockName}是${symbol.includes('300') ? '创业板' : 'A股'}优质标的，AI 模型从多维度进行了综合评估。\n\n` +
    `二、核心亮点\n✅ 行业地位突出，市场份额持续提升\n✅ 财务健康，现金流充裕\n✅ 技术面呈现多头排列\n\n` +
    `三、风险关注\n⚠️ 短期估值偏高，需注意回调风险\n⚠️ 行业竞争加剧可能影响利润率\n\n` +
    `四、AI 建议\n🎯 中长期看好，建议在回调时分批建仓，仓位控制在总资金 15% 以内。`,
    `【AI 综合研判结果】\n\n标的：${stockName}（${symbol}）\n\n` +
    `📊 综合评分：${Math.floor(75 + Math.random() * 20)}/100\n\n` +
    `📈 趋势判断：中期趋势向好，短期注意节奏\n💰 资金信号：近5日主力资金净流入\n🎯 目标区间：AI 预测 1-3 个月内存在 8-15% 上行空间\n\n` +
    `⚠️ 风险提示：以上分析由 AI 自动生成，不构成投资建议。`
  ];
  return templates.join('\n\n---\n\n');
}
