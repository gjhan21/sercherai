# strategy-graph

内部图谱服务，负责承接股票/期货研究链输出的结构化实体、关系和 run 级图快照。

默认能力：
- 写入 run 图快照
- 查询图快照
- 按实体查询一跳/两跳子图

默认情况下：
- 若配置了 `STRATEGY_GRAPH_NEO4J_URI`，服务使用 Neo4j 存储
- 否则自动回退到内存仓库，便于本地开发和测试

启动示例：

```bash
cd services/strategy-graph
uvicorn app.main:app --reload --port 18082
```
