package swarm

import (
	"fmt"
	"sercherai/backend/internal/platform/llm"
)

type AgentRole string

const (
	RoleValueInvestor AgentRole = "ValueInvestor"
	RoleHotMoney      AgentRole = "HotMoney"
	RoleQuantModel    AgentRole = "QuantModel"
	RoleMacroExpert   AgentRole = "MacroExpert"
)

type Agent struct {
	Role        AgentRole
	Name        string
	Description string
	PromptRules string
}

var DefaultAgents = []Agent{
	{
		Role:        RoleValueInvestor,
		Name:        "价值投资者",
		Description: "关注基本面、估值、行业周期",
		PromptRules: "你是一个严肃的机构价值投资者。你的分析应当聚焦于：1. 公司的基本面是否健康；2. 当前价格是否被低估；3. 行业周期是否处于上行阶段。请用专业、冷静的口吻输出你的分析结论。",
	},
	{
		Role:        RoleHotMoney,
		Name:        "短线游资",
		Description: "关注市场情绪、资金流向、题材热点",
		PromptRules: "你是一个敏锐的短线游资操盘手。你的分析应当聚焦于：1. 当前市场情绪是否高涨；2. 是否有明显的资金净流入；3. 标的所在的题材是否是当下热点。请用具有攻击性、注重顺势而为的口吻输出你的分析结论。",
	},
	{
		Role:        RoleQuantModel,
		Name:        "量化策略模型",
		Description: "关注技术指标、量价关系、统计套利",
		PromptRules: "你是一个纯逻辑驱动的量化交易模型。你的分析应当聚焦于：1. 技术面（MACD/RSI等）发出的买卖信号；2. 历史回撤和波动率表现；3. 统计学上的盈亏比。请用客观、数据驱动、冷酷的口吻输出你的分析结论。",
	},
	{
		Role:        RoleMacroExpert,
		Name:        "宏观策略师",
		Description: "关注宏观经济、政策导向、流动性",
		PromptRules: "你是一个着眼全局的宏观策略师。你的分析应当聚焦于：1. 货币政策及流动性环境；2. 行业政策的扶持或打压；3. 整体系统性风险。请用宏大叙事、自上而下的口吻输出你的分析结论。",
	},
}

type AgentOpinion struct {
	AgentRole AgentRole `json:"agent_role"`
	AgentName string    `json:"agent_name"`
	Opinion   string    `json:"opinion"`
}

func (a *Agent) Analyze(llmClient *llm.Client, eventContext string) (*AgentOpinion, error) {
	systemPrompt := fmt.Sprintf(`【角色设定】
%s
%s

【你的任务】
根据提供的【标的背景与事件环境】，独立评估当前该标的投资价值，并给出你的操作建议（如：强烈看多、谨慎看多、观望、看空等）以及核心逻辑。
请将字数控制在150字以内，直接给出观点，无需过多寒暄。`, a.Name, a.PromptRules)

	messages := []llm.ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: fmt.Sprintf("【标的背景与事件环境】\n%s", eventContext)},
	}

	response, err := llmClient.ChatCompletion(messages)
	if err != nil {
		return nil, err
	}

	return &AgentOpinion{
		AgentRole: a.Role,
		AgentName: a.Name,
		Opinion:   response,
	}, nil
}
