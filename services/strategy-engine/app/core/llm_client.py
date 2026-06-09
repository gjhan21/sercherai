import logging
import json
import re
from typing import Any
from concurrent.futures import ThreadPoolExecutor, as_completed
import httpx
from app.settings import get_settings

logger = logging.getLogger(__name__)


class LLMClient:
    def __init__(self) -> None:
        self.settings = get_settings()

    def review_stock(
        self,
        symbol: str,
        name: str,
        metrics: dict[str, Any],
        local_opinions: list[str],
        market_regime: str,
    ) -> dict[str, Any]:
        """
        对单只个股调用大模型进行终审研判。
        返回格式:
        {
            "rating": "强烈推荐" | "观望" | "回避",
            "summary": "一句话推荐理由",
            "analysis": "大模型深度研判文本..."
        }
        """
        # 降级防御：如果未配置 API Key 或关闭了 LLM 评审，直接返回默认的本地规则结果
        if not self.settings.llm_api_key or not self.settings.enable_llm_review:
            logger.debug("LLM review is disabled or API key not configured for %s", symbol)
            return self._build_fallback(symbol, local_opinions)

        url = f"{self.settings.llm_base_url.rstrip('/')}/chat/completions"
        headers = {
            "Authorization": f"Bearer {self.settings.llm_api_key}",
            "Content-Type": "application/json",
        }

        # 整理输入数据提供给大模型
        context_str = json.dumps(
            {
                "code": symbol,
                "name": name,
                "market_regime": market_regime,
                "metrics": metrics,
                "local_agents_opinions": local_opinions,
            },
            ensure_ascii=False,
        )

        system_prompt = (
            "你是一个资深的智能量化投资审查专家。你的任务是对量化 Pipeline 粗筛出来的黄金候选个股进行“临门一脚”的最终审议。\n"
            "你会收到个股的基本面与技术面数据、当前大盘状态（Market Regime）以及本地五个量化智能体（趋势、风险、流动性、基本面、事件）对该股票的初审意见。\n"
            "请结合大局观（大盘状态）与个股的局部表现（技术形态和智能体意见），输出一份逻辑清晰、字数约100字的深度分析研判，并给出最终评级与一句话理由。\n"
            "你必须严格以 JSON 格式输出，不要包含任何 markdown 的 ```json ``` 标记，也不要有任何前后引导废话。输出格式必须如下：\n"
            "{\n"
            '  "rating": "强烈推荐" / "观望" / "回避",\n'
            '  "summary": "一句话核心理由",\n'
            '  "analysis": "100字左右的深度剖析与交易警示"\n'
            "}"
        )

        payload = {
            "model": self.settings.llm_model,
            "messages": [
                {"role": "system", "content": system_prompt},
                {"role": "user", "content": f"输入数据如下：\n{context_str}"},
            ],
            "temperature": 0.2,
            "max_tokens": 500,
        }
        # 如果是 openai gpt 模型，可以强制开启 json_object 格式返回
        if "gpt" in self.settings.llm_model.lower():
            payload["response_format"] = {"type": "json_object"}

        try:
            with httpx.Client(timeout=self.settings.llm_timeout) as client:
                response = client.post(url, headers=headers, json=payload)
                response.raise_for_status()
                data = response.json()
                content = data["choices"][0]["message"]["content"].strip()

                # 对非严格 JSON 格式或带 markdown 的返回进行清洗提取
                if content.startswith("```"):
                    content = re.sub(r"^```(?:json)?\n?", "", content)
                    content = re.sub(r"\n?```$", "", content)

                result = json.loads(content)
                rating = result.get("rating", "观望").strip()
                if rating not in ["强烈推荐", "观望", "回避"]:
                    rating = "观望"

                return {
                    "rating": rating,
                    "summary": result.get("summary", "基本面和技术面偏中性。").strip(),
                    "analysis": result.get("analysis", "根据多智能体共鸣判定为持有或跟踪。").strip(),
                }
        except Exception as exc:
            logger.warning(
                "LLM API review failed for %s, falling back to local opinion. Error: %s",
                symbol,
                str(exc),
            )
            return self._build_fallback(symbol, local_opinions)

    def review_stock_list_concurrent(
        self, stocks: list[dict[str, Any]], market_regime: str
    ) -> dict[str, dict[str, Any]]:
        """
        利用多线程线程池并发审查股票列表，以缩短总体 I/O 等待时间。
        stocks 格式: [{"symbol": "sh600000", "name": "浦发银行", "metrics": {...}, "local_opinions": [...]}, ...]
        返回格式: {"sh600000": {"rating": "...", "summary": "...", "analysis": "..."}, ...}
        """
        results = {}
        if not stocks:
            return results

        max_workers = min(len(stocks), 5)
        with ThreadPoolExecutor(max_workers=max_workers) as executor:
            future_to_symbol = {
                executor.submit(
                    self.review_stock,
                    stock["symbol"],
                    stock["name"],
                    stock["metrics"],
                    stock["local_opinions"],
                    market_regime,
                ): stock["symbol"]
                for stock in stocks
            }

            for future in as_completed(future_to_symbol):
                symbol = future_to_symbol[future]
                try:
                    results[symbol] = future.result()
                except Exception as exc:
                    logger.warning("Future raised exception for %s: %s", symbol, str(exc))
                    results[symbol] = self._build_fallback(symbol, [])
        return results

    def _build_fallback(self, symbol: str, local_opinions: list[str]) -> dict[str, Any]:
        opinions_str = "；".join(local_opinions) if local_opinions else "多智能体维持本地综合评估。"
        return {
            "rating": "观望",
            "summary": "量化共振正常判定。",
            "analysis": f"本地智能体审议结论：{opinions_str}",
        }
