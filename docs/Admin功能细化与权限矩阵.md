# Admin 功能细化与权限矩阵（v0.1）

本文件描述管理后台功能模块与角色权限，用于权限设计与审核流程落地。Admin 技术栈为 Vue 3 + Vite + Element Plus。

**1. Admin 角色定义**
1. 超级管理员 `SUPER_ADMIN`
2. 策略管理员 `STRATEGY_ADMIN`
3. 内容审核员 `CONTENT_REVIEWER`
4. 实名审核员 `KYC_REVIEWER`
5. 运营人员 `OPS`

**2. 功能模块**
1. 策略管理
2. 推荐审核
3. 实名审核
4. 数据源管理
5. 用户管理
6. 运营报表
7. 新闻资讯管理
8. 增长激励管理
9. 会员配额管理
10. 支付与对账管理
11. 系统设置

**3. 权限矩阵**

**3.1 策略管理**
- `SUPER_ADMIN` 读写
- `STRATEGY_ADMIN` 读写
- `CONTENT_REVIEWER` 只读
- `OPS` 只读

**3.2 推荐审核**
- `SUPER_ADMIN` 读写
- `CONTENT_REVIEWER` 读写
- `STRATEGY_ADMIN` 只读
- `OPS` 只读

**3.3 实名审核**
- `SUPER_ADMIN` 读写
- `KYC_REVIEWER` 读写
- `OPS` 只读

**3.4 数据源管理**
- `SUPER_ADMIN` 读写
- `STRATEGY_ADMIN` 只读
- `OPS` 只读

**3.5 用户管理**
- `SUPER_ADMIN` 读写
- `OPS` 读写

**3.6 运营报表**
- `SUPER_ADMIN` 读写
- `OPS` 读写
- `STRATEGY_ADMIN` 只读

**3.7 新闻资讯管理**
- `SUPER_ADMIN` 读写
- `CONTENT_REVIEWER` 读写
- `OPS` 读写
- `STRATEGY_ADMIN` 只读

**3.8 增长激励管理**
- `SUPER_ADMIN` 读写
- `OPS` 读写
- `CONTENT_REVIEWER` 只读

**3.9 会员配额管理**
- `SUPER_ADMIN` 读写
- `OPS` 读写
- `CONTENT_REVIEWER` 只读

**3.10 支付与对账管理**
- `SUPER_ADMIN` 读写
- `OPS` 读写
- `CONTENT_REVIEWER` 只读

**3.11 系统设置**
- `SUPER_ADMIN` 读写

**4. 审计与日志**
- 所有角色操作写入审计日志
- 策略发布与推荐发布必须记录审批人
