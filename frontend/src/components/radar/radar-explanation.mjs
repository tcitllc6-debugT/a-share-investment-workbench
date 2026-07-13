const eventDetails = {
  momentum: { trigger: '动量异动：价格或涨幅在短时间内明显加速。', nextAction: '打开 K 线，确认量价配合与关键压力位。' },
  strength: { trigger: '强势异动：相对强度进入当日活跃区间。', nextAction: '查看所属板块是否同步走强。' },
  limit: { trigger: '涨跌停异动：接近或触及当日涨跌停状态。', nextAction: '查看封单、换手与次日风险。' },
  risk: { trigger: '风险异动：价格行为出现需要复核的走弱信号。', nextAction: '打开 K 线，检查止损与持仓暴露。' },
  breakdown: { trigger: '破位异动：价格可能跌破近期关键位置。', nextAction: '检查支撑位、成交量和持仓纪律。' },
  capital: { trigger: '资金异动：大单或资金流向出现明显变化。', nextAction: '结合价格走势确认资金方向。' },
}

const changeTypeDetails = {
  8201: { trigger: '火箭发射：盘中价格快速上冲，需确认是否有量能跟随。', nextAction: '打开 K线，核对成交量、分时回落和压力位。' },
  8202: { trigger: '快速反弹：下跌后出现快速回补，需确认不是短线反抽。', nextAction: '查看反弹前低点、量能与所在板块强弱。' },
  8207: { trigger: '竞价上涨：集合竞价阶段价格明显走强。', nextAction: '查看开盘后承接与竞价量，不把竞价单独当作买点。' },
  8209: { trigger: '高开 5 日线：价格高开并站上短期均线。', nextAction: '核对缺口位置、换手和是否有板块共振。' },
  8211: { trigger: '向上缺口：价格跳空上行，关注缺口是否获得承接。', nextAction: '查看成交量、前高压力和缺口回补风险。' },
  8213: { trigger: '60 日新高：价格刷新近 60 个交易日高点。', nextAction: '查看突破量能与历史压力区，避免仅凭新高追涨。' },
  8215: { trigger: '60 日大幅上涨：近 60 日累计涨幅明显。', nextAction: '检查涨幅、估值与回撤风险，先复核而非追随。' },
  8193: { trigger: '大笔买入：盘中出现显著主动买入成交。', nextAction: '结合价格方向和持续性判断资金是否真实流入。' },
  64: { trigger: '资金流入：资金流向出现正向变化。', nextAction: '查看净流入持续性及同板块个股表现。' },
  8194: { trigger: '大笔卖出：盘中出现显著主动卖出成交。', nextAction: '检查承接、持仓风险和是否出现连续卖压。' },
  128: { trigger: '资金流出：资金流向出现负向变化。', nextAction: '查看流出是否持续，并核对价格是否同步走弱。' },
  4: { trigger: '封涨停：价格触及涨停，仍需观察封单与换手。', nextAction: '查看封单稳定性、开板次数及次日风险。' },
  8: { trigger: '打开涨停：涨停状态被打开，强度可能发生变化。', nextAction: '检查开板后的承接和换手，避免把异动等同买卖信号。' },
  16: { trigger: '封跌停：价格触及跌停，流动性与风险需优先复核。', nextAction: '查看封单、成交和持仓风险，不做自动交易。' },
  32: { trigger: '打开跌停：跌停状态被打开，波动仍可能很大。', nextAction: '查看是否有持续承接及后续波动风险。' },
  8203: { trigger: '高台跳水：高位快速回落，属于风险复核信号。', nextAction: '打开 K 线检查跌幅、成交量和持仓暴露。' },
  8204: { trigger: '加速下跌：盘中下跌速度加快。', nextAction: '检查关键支撑和风险控制，不把提示当作交易指令。' },
  8208: { trigger: '竞价下跌：集合竞价阶段价格明显走弱。', nextAction: '观察开盘后承接，不以竞价单独判断趋势。' },
  8210: { trigger: '低开 5 日线：价格低开并跌破短期均线。', nextAction: '查看缺口、量能和是否有板块共振走弱。' },
  8212: { trigger: '向下缺口：价格跳空下行，需复核缺口与支撑。', nextAction: '查看成交量、支撑位和持仓风险。' },
  8214: { trigger: '60 日新低：价格刷新近 60 个交易日低点。', nextAction: '检查下跌趋势、流动性和止损纪律。' },
  8216: { trigger: '60 日大幅下跌：近 60 日累计跌幅明显。', nextAction: '复核基本面事件与持仓风险，不做自动操作。' },
}

export const radarHowItWorks = '雷达每 15 秒扫描一次可用的盘中异动数据。P1 表示优先复核，P2 表示一般关注；异动只提示需要分析的线索，不构成买卖指令。'

export function describeRadarEvent(event = {}) {
  return changeTypeDetails[event.changeType] || eventDetails[event.category] || {
    trigger: `市场异动：${event.typeName || '发现需要复核的价格或资金变化'}。`,
    nextAction: '打开 K 线和 AI 分析核对原因。',
  }
}
