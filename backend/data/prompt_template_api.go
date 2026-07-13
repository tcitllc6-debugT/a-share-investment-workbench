package data

import (
	"go-stock/backend/db"
	"go-stock/backend/logger"
	"go-stock/backend/models"
)

type PromptTemplateApi struct {
}

func DefaultPromptTemplates() []models.PromptTemplate {
	templates := []models.PromptTemplate{
		{
			Name: "A股投研系统角色",
			Type: "模型系统Prompt",
			Content: `你是一名专注A股市场的本地投研助手。回答必须基于可验证的数据、工具查询结果和用户给出的上下文，不编造实时价格、成交额、公告、研报或政策信息。

分析时优先覆盖：市场环境、行业/概念位置、资金流、技术结构、基本面、事件催化、风险因素、仓位与止损纪律。所有结论都要区分事实、推断和操作假设。

输出不是投资建议，只用于研究和复盘。`,
		},
		{
			Name: "A股盘前盘后复盘",
			Type: "模型用户Prompt",
			Content: `请基于当前市场指数、行业/概念涨跌、资金流向、重要资讯和外媒简讯，生成一份A股市场复盘。

请按以下结构输出：
1. 市场总览：上证、深成指、创业板、科创相关指数的状态与情绪。
2. 主线题材：列出3-5条强势方向，说明驱动逻辑和持续性。
3. 风险信号：列出可能影响次日交易的风险。
4. 观察名单：给出可继续跟踪的板块/个股方向，并说明跟踪条件。
5. 明日策略：区分激进、稳健、防守三种仓位思路。`,
		},
		{
			Name: "个股多因子分析",
			Type: "模型用户Prompt",
			Content: `请对目标A股进行多因子分析。必须先确认股票代码、名称、最新行情、K线趋势、所属行业/概念、资金流和近期资讯。

请按以下结构输出：
1. 核心结论：一句话说明偏多、震荡或偏空。
2. 技术面：趋势、均线、量能、支撑压力、异常波动。
3. 基本面：主营、估值、财务质量、业绩变化。
4. 资金与情绪：主力资金、龙虎榜、热度、新闻催化。
5. 交易计划：观察价位、买入条件、止盈止损、无效条件。
6. 风险提示：至少列出3条。`,
		},
		{
			Name: "热点题材与资金流追踪",
			Type: "模型用户Prompt",
			Content: `请围绕A股热点题材做资金流追踪，结合行业/概念排名、涨跌幅、成交额、资金净流入、热门个股和资讯催化。

输出要求：
- 先判断题材处于启动、发酵、高潮、分歧还是退潮阶段。
- 给出龙头、补涨、中军和低位观察方向。
- 标注哪些信号代表题材延续，哪些信号代表风险升高。
- 最后给出适合短线、中线、观望三类用户的跟踪动作。`,
		},
		{
			Name: "AI推荐记录结构化输出",
			Type: "模型用户Prompt",
			Content: `请基于工具查询结果筛选A股候选标的，并在最后用结构化方式给出推荐记录。

每个标的必须包含：股票代码、股票名称、行业/概念、评级、推荐理由、当前价格、建议关注/买入区间、止盈区间、止损价、风险备注。

如果证据不足，请明确写出“不推荐”，不要为了凑数量而输出标的。`,
		},
		{
			Name: "AI自动推荐A股候选股",
			Type: "模型用户Prompt",
			Content: `请作为A股投研助手，基于当前市场行情、热点题材、资金流、涨跌幅排行、个股K线、公告/新闻，筛选3-5只候选股票，并调用股票推荐记录工具写入推荐记录。

要求：
1. 先说明筛选范围、使用的数据来源和时间范围，不要编造实时数据。
2. 每只股票必须包含：代码、名称、所属板块/题材、推荐理由、风险等级、观察价、止损价、目标价/目标区间、推荐周期。
3. 对涨停、ST、退市风险、成交量异常、消息面风险和高位追涨风险要单独提示。
4. 如果证据不足，明确写出“不推荐”或“仅观察”，不要为了凑数量而输出标的。
5. 最后输出“已写入股票推荐记录”的摘要；如果工具不可用，输出可手工录入的结构化表格。`,
		},
		{
			Name: "风险控制与仓位纪律",
			Type: "模型用户Prompt",
			Content: `请只从风险控制角度审查当前A股交易计划。

请检查：
1. 买入理由是否可验证。
2. 仓位是否匹配波动率和亏损承受能力。
3. 止损是否具体、可执行。
4. 是否存在追高、消息滞后、流动性不足、财报/监管/退市风险。
5. 如果计划不合格，请直接给出否决原因和等待条件。`,
		},
	}
	return append(templates, models.PromptTemplate{
		Name: "自选持仓成本复盘与投资建议",
		Type: "模型用户Prompt",
		Content: `请基于当前自选股票列表，优先统计已经填写成本的持仓股票：只纳入 costPrice > 0 且 costVolume/volume > 0 的记录。

请先输出一张持仓汇总表，字段至少包括：股票代码、股票名称、costPrice（买入成本价）、costVolume/volume（持股数量）、投入本金、当前价、当前市值、浮动盈亏、盈亏比例、仓位建议。

然后结合可用工具数据做投资建议：K线趋势、多空线位置与拐点、成交量、资金流、行业/概念、公告新闻、市场情绪、交易日志。若交易日志没有对应买入记录，请把自选成本视为初始买入记录，不要编造不存在的交易。

输出结构：
1. 总体仓位与风险暴露
2. 单股持仓复盘
3. 加仓/减仓/持有/止损/观察建议
4. 未来 1-3 个交易日的跟踪条件
5. 需要补充的数据或无法确认的信息

所有结论必须区分事实、推断和操作假设；这不是投资承诺，仅用于研究复盘。`,
	})
}

func (t PromptTemplateApi) SeedDefaultPromptTemplates() error {
	for _, template := range DefaultPromptTemplates() {
		record := models.PromptTemplate{
			Name:    template.Name,
			Type:    template.Type,
			Content: template.Content,
		}
		if err := db.Dao.Where(models.PromptTemplate{
			Name: template.Name,
			Type: template.Type,
		}).FirstOrCreate(&record).Error; err != nil {
			return err
		}
	}
	return nil
}

func (t PromptTemplateApi) GetPromptTemplates(name string, promptType string) *[]models.PromptTemplate {
	var result []models.PromptTemplate
	if name != "" && promptType != "" {
		db.Dao.Model(&models.PromptTemplate{}).Where("name=? and type=?", name, promptType).Find(&result)
	}
	if name != "" && promptType == "" {
		db.Dao.Model(&models.PromptTemplate{}).Where("name=?", name).Find(&result)
	}
	if name == "" && promptType != "" {
		db.Dao.Model(&models.PromptTemplate{}).Where("type=?", promptType).Find(&result)
	}
	if name == "" && promptType == "" {
		db.Dao.Model(&models.PromptTemplate{}).Find(&result)
	}

	return &result
}

// GetPromptTemplateList 分页查询PromptTemplate记录
func (t PromptTemplateApi) GetPromptTemplateList(query *models.PromptTemplateQuery) (*models.PromptTemplatePageData, error) {
	var list []models.PromptTemplate
	var total int64

	q := db.Dao.Model(&models.PromptTemplate{})

	// 构建查询条件
	if query.Name != "" {
		q = q.Where("name LIKE ?", "%"+query.Name+"%")
	}
	if query.Type != "" {
		q = q.Where("type LIKE ?", "%"+query.Type+"%")
	}
	if query.Content != "" {
		q = q.Where("content LIKE ?", "%"+query.Content+"%")
	}

	// 计算总数
	err := q.Count(&total).Error
	if err != nil {
		return nil, err
	}

	// 设置默认分页参数
	page := query.Page
	pageSize := query.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	// 执行分页查询
	offset := (page - 1) * pageSize
	err = q.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&list).Error
	if err != nil {
		return nil, err
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return &models.PromptTemplatePageData{
		List:       list,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (t PromptTemplateApi) AddPrompt(template models.PromptTemplate) string {
	var tmp models.PromptTemplate
	db.Dao.Model(&models.PromptTemplate{}).Where("id=?", template.ID).First(&tmp)
	if tmp.ID == 0 {
		err := db.Dao.Model(&models.PromptTemplate{}).Create(&models.PromptTemplate{
			Content: template.Content,
			Name:    template.Name,
			Type:    template.Type,
		}).Error
		if err != nil {
			return "添加失败"
		} else {
			return "添加成功"
		}
	} else {
		err := db.Dao.Model(&models.PromptTemplate{}).Where("id=?", template.ID).Updates(template).Error
		if err != nil {
			return "更新失败"
		} else {
			return "更新成功"
		}
	}
}

func (t PromptTemplateApi) DelPrompt(Id uint) string {
	template := &models.PromptTemplate{}
	db.Dao.Model(template).Where("id=?", Id).Find(template)
	if template.ID > 0 {
		err := db.Dao.Model(template).Delete(template).Error
		if err != nil {
			return "删除失败"
		} else {
			return "删除成功"
		}
	}
	return "模板信息不存在"
}

func (t PromptTemplateApi) GetPromptTemplateByID(id int) string {
	prompt := &models.PromptTemplate{}
	db.Dao.Model(&models.PromptTemplate{}).Where("id=?", id).First(prompt)
	logger.SugaredLogger.Infof("GetPromptTemplateByID:%d %s", id, prompt.Content)
	return prompt.Content
}
func NewPromptTemplateApi() *PromptTemplateApi {
	return &PromptTemplateApi{}
}
