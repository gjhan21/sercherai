package swarm

import (
	"encoding/json"
	"fmt"
	"sync"

	"sercherai/backend/internal/platform/llm"
)

type SimulationEngine struct {
	llmClient *llm.Client
	agents    []Agent
}

func NewSimulationEngine(llmClient *llm.Client) *SimulationEngine {
	return &SimulationEngine{
		llmClient: llmClient,
		agents:    DefaultAgents,
	}
}

type SimulationResult struct {
	EventContext string         `json:"event_context"`
	Opinions     []AgentOpinion `json:"opinions"`
	Summary      string         `json:"summary"`
}

func (e *SimulationEngine) Run(eventContext string) (*SimulationResult, error) {
	result := &SimulationResult{
		EventContext: eventContext,
		Opinions:     make([]AgentOpinion, 0, len(e.agents)),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	errs := make([]error, 0)

	for _, agent := range e.agents {
		wg.Add(1)
		go func(a Agent) {
			defer wg.Done()
			opinion, err := a.Analyze(e.llmClient, eventContext)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, fmt.Errorf("agent %s error: %w", a.Name, err))
			} else if opinion != nil {
				result.Opinions = append(result.Opinions, *opinion)
			}
		}(agent)
	}

	wg.Wait()

	if len(result.Opinions) == 0 && len(errs) > 0 {
		return nil, fmt.Errorf("all agents failed to analyze: %v", errs)
	}

	summary, err := e.generateReport(eventContext, result.Opinions)
	if err != nil {
		return nil, fmt.Errorf("report agent error: %w", err)
	}

	result.Summary = summary
	return result, nil
}

func (e *SimulationEngine) generateReport(eventContext string, opinions []AgentOpinion) (string, error) {
	opinionsJSON, _ := json.MarshalIndent(opinions, "", "  ")

	systemPrompt := `你是一个【ReportAgent】，拥有“上帝视角”。
你的任务是观察多个独立金融智能体（Agent）对于特定标的/事件的推演结果，并从中提取【共识】与【分歧】，最终给出一份给用户的“AI复盘报告”。

【要求】
1. 报告需直击核心，字数控制在 300 字以内。
2. 明确指出不同资金流派/角色的共识点和主要分歧点。
3. 给出最终的操作倾斜方向（偏向多头、空头还是震荡）。
4. 不要返回 Markdown 标题格式，直接输出纯文本的正文段落即可。
5. 必须在开头标明：【基于MiroFish多Agent博弈引擎生成】`

	userPrompt := fmt.Sprintf(`【标的背景与事件环境】
%s

【各Agent推演意见】
%s

请根据以上信息，生成综合AI复盘报告。`, eventContext, string(opinionsJSON))

	messages := []llm.ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	return e.llmClient.ChatCompletion(messages)
}
