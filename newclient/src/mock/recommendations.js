export const DAILY_RECS = [
  { id: 1, symbol: '300750.SZ', name: '宁德时代', rank: 1, price: '198.62', change: 3.45, score: 92, aiReason: 'AI 检测到主力资金持续流入，技术形态突破前期平台，量价配合良好', strategy: { entry: '195-200', stopLoss: '185', takeProfit: '228', positionSize: '15%' }, dimensions: { tech: 84, fund: 91, fundamental: 78, sentiment: 88, valuation: 65 }, tags: ['VIP', '资金流入'] },
  { id: 2, symbol: '600941.SH', name: '中国移动', rank: 2, price: '106.80', change: 1.82, score: 88, aiReason: '高股息防御属性突出，AI 模型评分持续走高，资金避险情绪下有望受益', strategy: { entry: '105-107', stopLoss: '102', takeProfit: '115', positionSize: '20%' }, dimensions: { tech: 72, fund: 85, fundamental: 92, sentiment: 80, valuation: 90 }, tags: ['高股息', '稳健'] },
  { id: 3, symbol: '002415.SZ', name: '海康威视', rank: 3, price: '35.27', change: 2.65, score: 85, aiReason: 'AI 概念催化，量价齐升突破平台，资金面明显改善', strategy: { entry: '34.5-35.5', stopLoss: '33', takeProfit: '40', positionSize: '10%' }, dimensions: { tech: 82, fund: 78, fundamental: 72, sentiment: 90, valuation: 75 }, tags: ['AI概念', '放量'] },
  { id: 4, symbol: '000858.SZ', name: '五粮液', rank: 4, price: '152.30', change: -0.58, score: 82, aiReason: '估值处于历史低位，AI 判断反弹窗口临近，消费复苏预期下值得跟踪', strategy: { entry: '148-153', stopLoss: '142', takeProfit: '168', positionSize: '12%' }, dimensions: { tech: 65, fund: 72, fundamental: 85, sentiment: 68, valuation: 88 }, tags: ['消费', '估值修复'] },
  { id: 5, symbol: '002594.SZ', name: '比亚迪', rank: 5, price: '268.45', change: 2.18, score: 80, aiReason: '新能源政策利好频出，AI 趋势模型看多，但需注意短期涨幅过大', strategy: { entry: '260-270', stopLoss: '248', takeProfit: '300', positionSize: '10%' }, dimensions: { tech: 78, fund: 82, fundamental: 76, sentiment: 85, valuation: 60 }, tags: ['新能源', '趋势'] },
  { id: 6, symbol: '688981.SH', name: '中芯国际', rank: 6, price: '56.78', change: 4.12, score: 79, aiReason: '半导体周期回暖信号明确，AI 情绪模型积极，但估值偏高需谨慎', strategy: { entry: '55-57', stopLoss: '51', takeProfit: '65', positionSize: '8%' }, dimensions: { tech: 86, fund: 88, fundamental: 62, sentiment: 82, valuation: 45 }, tags: ['半导体', '热点'] }
];

export const REC_HISTORY = [
  { id: 1, symbol: '300750.SZ', name: '宁德时代', recDate: '2025-03-10', recPrice: 182.50, currentPrice: 198.62, actualReturn: 8.83, recScore: 89, hitTarget: true, outcome: 'success', strategy: '突破买入' },
  { id: 2, symbol: '600941.SH', name: '中国移动', recDate: '2025-03-12', recPrice: 102.30, currentPrice: 106.80, actualReturn: 4.40, recScore: 86, hitTarget: true, outcome: 'success', strategy: '高股息策略' },
  { id: 3, symbol: '002415.SZ', name: '海康威视', recDate: '2025-03-15', recPrice: 32.80, currentPrice: 35.27, actualReturn: 7.53, recScore: 82, hitTarget: true, outcome: 'success', strategy: '趋势跟踪' },
  { id: 4, symbol: '688981.SH', name: '中芯国际', recDate: '2025-03-18', recPrice: 52.40, currentPrice: 56.78, actualReturn: 8.36, recScore: 76, hitTarget: true, outcome: 'success', strategy: '突破买入' },
  { id: 5, symbol: '000858.SZ', name: '五粮液', recDate: '2025-03-20', recPrice: 156.80, currentPrice: 152.30, actualReturn: -2.87, recScore: 80, hitTarget: false, outcome: 'fail', strategy: '价值投资' },
  { id: 6, symbol: '002594.SZ', name: '比亚迪', recDate: '2025-03-22', recPrice: 255.00, currentPrice: 268.45, actualReturn: 5.27, recScore: 78, hitTarget: true, outcome: 'success', strategy: '趋势跟踪' },
  { id: 7, symbol: '300059.SZ', name: '东方财富', recDate: '2025-03-25', recPrice: 15.20, currentPrice: 16.82, actualReturn: 10.66, recScore: 84, hitTarget: true, outcome: 'success', strategy: '突破买入' },
  { id: 8, symbol: '601318.SH', name: '中国平安', recDate: '2025-03-28', recPrice: 47.50, currentPrice: 48.26, actualReturn: 1.60, recScore: 72, hitTarget: false, outcome: 'neutral', strategy: '价值投资' },
  { id: 9, symbol: '600036.SH', name: '招商银行', recDate: '2025-04-01', recPrice: 35.80, currentPrice: 36.78, actualReturn: 2.74, recScore: 74, hitTarget: false, outcome: 'neutral', strategy: '高股息策略' },
  { id: 10, symbol: '002230.SZ', name: '科大讯飞', recDate: '2025-04-03', recPrice: 45.60, currentPrice: 48.90, actualReturn: 7.24, recScore: 81, hitTarget: true, outcome: 'success', strategy: '趋势跟踪' },
  { id: 11, symbol: '300750.SZ', name: '宁德时代', recDate: '2025-04-08', recPrice: 190.00, currentPrice: 198.62, actualReturn: 4.54, recScore: 90, hitTarget: true, outcome: 'success', strategy: '突破买入' },
  { id: 12, symbol: '688012.SH', name: '中微公司', recDate: '2025-04-10', recPrice: 155.00, currentPrice: 168.20, actualReturn: 8.52, recScore: 86, hitTarget: true, outcome: 'success', strategy: '突破买入' },
  { id: 13, symbol: '603259.SH', name: '药明康德', recDate: '2025-04-15', recPrice: 56.20, currentPrice: 58.60, actualReturn: 4.27, recScore: 75, hitTarget: false, outcome: 'neutral', strategy: '价值投资' },
  { id: 14, symbol: '300760.SZ', name: '迈瑞医疗', recDate: '2025-04-18', recPrice: 278.00, currentPrice: 286.50, actualReturn: 3.06, recScore: 79, hitTarget: false, outcome: 'neutral', strategy: '趋势跟踪' },
  { id: 15, symbol: '002475.SZ', name: '立讯精密', recDate: '2025-04-20', recPrice: 30.80, currentPrice: 32.56, actualReturn: 5.71, recScore: 77, hitTarget: true, outcome: 'success', strategy: '趋势跟踪' },
  { id: 16, symbol: '002594.SZ', name: '比亚迪', recDate: '2025-04-22', recPrice: 258.00, currentPrice: 268.45, actualReturn: 4.05, recScore: 80, hitTarget: true, outcome: 'success', strategy: '趋势跟踪' },
  { id: 17, symbol: '600519.SH', name: '贵州茅台', recDate: '2025-04-25', recPrice: 1670.00, currentPrice: 1685.00, actualReturn: 0.90, recScore: 71, hitTarget: false, outcome: 'neutral', strategy: '价值投资' },
  { id: 18, symbol: '300750.SZ', name: '宁德时代', recDate: '2025-05-04', recPrice: 192.00, currentPrice: 198.62, actualReturn: 3.45, recScore: 88, hitTarget: true, outcome: 'success', strategy: '突破买入' },
  { id: 19, symbol: '000858.SZ', name: '五粮液', recDate: '2025-05-06', recPrice: 150.00, currentPrice: 152.30, actualReturn: 1.53, recScore: 78, hitTarget: false, outcome: 'neutral', strategy: '价值投资' },
  { id: 20, symbol: '002371.SZ', name: '北方华创', recDate: '2025-05-08', recPrice: 268.00, currentPrice: 286.80, actualReturn: 7.01, recScore: 85, hitTarget: true, outcome: 'success', strategy: '突破买入' }
];

export const STRATEGIES = [
  { id: 1, name: '突破买入策略', type: 'breakout', symbol: '300750.SZ', stockName: '宁德时代', entry: { condition: '股价突破前高192，成交量>5日均量150%', confidence: '高' }, exit: { condition: '跌破10日均线或达到止盈位' }, stopLoss: { price: '185.00', percent: '-3.5%' }, takeProfit: { level1: '210', level2: '228', level3: '250' }, positionSize: '建议仓位15%，分3批建仓' },
  { id: 2, name: '高股息防御策略', type: 'defensive', symbol: '600941.SH', stockName: '中国移动', entry: { condition: '股息率>4.5%且股价处于20日均线上方', confidence: '中' }, exit: { condition: '股息率降至3.5%以下或基本面恶化' }, stopLoss: { price: '95.00', percent: '-8%' }, takeProfit: { level1: '115', level2: '125' }, positionSize: '建议仓位20%，可一次性建仓' },
  { id: 3, name: '趋势跟踪策略', type: 'trend', symbol: '002415.SZ', stockName: '海康威视', entry: { condition: 'MACD金叉+OBV创新高+股价站上60日均线', confidence: '高' }, exit: { condition: 'MACD死叉或跌破20日均线' }, stopLoss: { price: '31.00', percent: '-6%' }, takeProfit: { level1: '40', level2: '45', level3: '50' }, positionSize: '建议仓位10-12%，分2批建仓' }
];

export const BACKTEST_RESULTS = {
  strategy: '突破买入策略',
  period: '2024-06 ~ 2025-05',
  summary: { totalReturn: '32.45%', annualized: '28.6%', winRate: '68.4%', maxDrawdown: '-8.2%', sharpeRatio: '1.86', totalTrades: 38, profitFactor: '2.45' },
  trades: [
    { id: 1, entryDate: '2024-06-10', entryPrice: 168.50, exitDate: '2024-07-15', exitPrice: 195.30, returnPct: 15.9, duration: '35天' },
    { id: 2, entryDate: '2024-06-18', entryPrice: 45.20, exitDate: '2024-07-22', exitPrice: 52.40, returnPct: 15.9, duration: '34天' },
    { id: 3, entryDate: '2024-07-01', entryPrice: 152.00, exitDate: '2024-07-28', exitPrice: 148.50, returnPct: -2.3, duration: '27天' },
    { id: 4, entryDate: '2024-07-15', entryPrice: 185.00, exitDate: '2024-08-20', exitPrice: 210.50, returnPct: 13.8, duration: '36天' },
    { id: 5, entryDate: '2024-08-01', entryPrice: 35.60, exitDate: '2024-08-25', exitPrice: 33.80, returnPct: -5.1, duration: '24天' },
    { id: 6, entryDate: '2024-08-15', entryPrice: 210.00, exitDate: '2024-09-28', exitPrice: 232.00, returnPct: 10.5, duration: '44天' },
    { id: 7, entryDate: '2024-09-05', entryPrice: 48.20, exitDate: '2024-10-10', exitPrice: 55.60, returnPct: 15.4, duration: '35天' },
    { id: 8, entryDate: '2024-09-20', entryPrice: 260.00, exitDate: '2024-10-18', exitPrice: 242.00, returnPct: -6.9, duration: '28天' },
    { id: 9, entryDate: '2024-10-10', entryPrice: 172.00, exitDate: '2024-11-15', exitPrice: 195.00, returnPct: 13.4, duration: '36天' },
    { id: 10, entryDate: '2024-10-25', entryPrice: 38.50, exitDate: '2024-11-28', exitPrice: 42.80, returnPct: 11.2, duration: '34天' },
    { id: 11, entryDate: '2024-11-10', entryPrice: 225.00, exitDate: '2024-12-15', exitPrice: 218.00, returnPct: -3.1, duration: '35天' },
    { id: 12, entryDate: '2024-11-25', entryPrice: 50.60, exitDate: '2025-01-08', exitPrice: 56.20, returnPct: 11.1, duration: '44天' },
    { id: 13, entryDate: '2024-12-10', entryPrice: 280.00, exitDate: '2025-01-20', exitPrice: 305.00, returnPct: 8.9, duration: '41天' },
    { id: 14, entryDate: '2025-01-05', entryPrice: 42.00, exitDate: '2025-02-10', exitPrice: 40.20, returnPct: -4.3, duration: '36天' },
    { id: 15, entryDate: '2025-01-20', entryPrice: 310.00, exitDate: '2025-02-28', exitPrice: 338.00, returnPct: 9.0, duration: '39天' }
  ]
};
