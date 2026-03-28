# PC 社区替代关注入口、关注收口到我的 - 设计说明

## 背景

当前 PC 客户端的一���导航仍然保留 `关注`，而 `社区` 虽然已有完整页面能力，但还没有进入 PC 主导航主链。这导致：

1. 产品主入口仍以“个人关注清单”为中心，而不是“社区讨论与观点流”为中心。
2. `关注` 作为一级页面暴露过强，不符合“关注只是我的里的个人模块”的目标。
3. 社区已具备主题、评论、我的主题、我的评论等能力，但入口权重不足。
4. 会员页、首页、部分 CTA 和说明文案仍把 `/watchlist` 当作一级主入口宣传，造成口径不统一。

本次调整目标是：

- PC 一级导航由 `社区` 替代 `关注`
- `关注` 从一级入口退到 `我的` 的二级模块
- 产品感知上不再把 `/watchlist` 当作一级主页面
- 保留现有关注页能力，不做大拆页，只做入口与定位收口

## 目标

### 一级目标

1. 一级导航调整为：`首页 / 策略 / 档案 / 社区 / 资讯 / 会员 / 我的`
2. `社区` 成为 PC 主入口之一，直接承接社区发现、讨论与发布行为
3. `关注` 收口到 `我的` 页面内部，以“我的关注”二级入口形式承接
4. 原 `/watchlist` 从独立一级页面语义，改为 `我的 > 我的关注` 语义

### 非目标

1. 不重写社区页面整体结构
2. 不把关注列表直接嵌进 `ProfileView` 主页面正文区域
3. 不修改 H5 端导航结构
4. 不处理后台 admin 路由权限问题
5. 不重构 watchlist 数据模型和本地存储结构

## 方案选型

### 方案 A：导航替换 + 我的页卡片入口 + 关注页收口为 profile 二级页（推荐）

- 一级导航将 `关注` 替换为 `社区`
- 新增 `/community` 到 PC 路由主链
- 新增 `profile/watchlist` 二级路由，作为“我的关注”正式入口
- 原 `/watchlist` 保留兼容重定向到 `/profile/watchlist`
- `ProfileView` 增加“我的关注”入口卡片
- `MyWatchlistView` 页面本体基本复用，仅更新定位文案与返回链路

#### 优点

- 改动面最小，复用现有 `MyWatchlistView.vue`
- 能最快达成“社区替代关注”产品感知
- 对现有会员、首页、搜索链路影响可控
- 可通过重定向保留旧书签和旧 CTA 兼容

#### 缺点

- 关注页仍然是一个独立页面文件，只是被重新收口
- 需要清理多处 `/watchlist` 的主入口文案和按钮

### 方案 B：把关注直接嵌入“我的”首页

- `ProfileView` 直接渲染关注列表和详情摘要
- 不再保留关注独立页面入口

#### 不采用原因

- `MyWatchlistView.vue` 体量过大，直接并入 `ProfileView` 风险高
- `ProfileView` 当前已经承担账户、会员、待办和查询中心，多塞关注列表会明显变重
- 回归成本和拆分成本都高，不适合这一轮入口调整

## 最终设计

### 1. 一级导航调整

#### 现状

PC 顶部导航当前为：

- 首页
- 策略
- 档案
- 关注
- 资讯
- 会员
- 我的

#### 目标

调整后为：

- 首页
- 策略
- 档案
- 社区
- 资讯
- 会员
- 我的

#### 影响文件

- `/Users/gjhan21/cursor/sercherai/client/src/apps/pc/layouts/PcLayout.vue`
- `/Users/gjhan21/cursor/sercherai/client/src/apps/pc/router/index.js`

### 2. 路由结构调整

#### 新路由结构

- `/community`：社区一级主入口
- `/profile`：我的主页
- `/profile/watchlist`：我的关注
- `/watchlist`：兼容重定向到 `/profile/watchlist`

#### 设计说明

1. `CommunityView` 接入 PC 路由主链，成为真正可见的一级入口。
2. `MyWatchlistView` 不删除，改由 `profile/watchlist` 承接。
3. `/watchlist` 继续存在，但只做兼容跳转，不再承担一级页面语义。
4. 所有新的入口、CTA、推荐链路，统一跳 `profile/watchlist`，不再跳 `/watchlist`。

#### 影响文件

- `/Users/gjhan21/cursor/sercherai/client/src/apps/pc/router/index.js`

### 3. 我的页承接“我的关注”

#### 入口形态

`ProfileView` 新增一块明显的功能入口卡片：

- 标题：`我的关注`
- 描述：`查看你保存的标的、变化记录和解释更新`
- 辅助说明：强调“从个人中心进入自己的跟踪清单”
- CTA：`进入我的关注`

#### 入口位置

建议放在 `ProfileView` 的账户总览 / 今日行动 / 查询中心之前的功能工作台区域，保证它是明显的个人模块，而不是埋在低优先级区域。

#### 设计原则

- 入口可见，但不盖过账户身份、会员状态、待办中心
- 不直接把完整 watchlist 列表塞进 `ProfileView`
- 保持“个人中心 -> 二级功能页”的清晰层次

#### 影响文件

- `/Users/gjhan21/cursor/sercherai/client/src/views/ProfileView.vue`

### 4. 关注页重新定位

#### 新定位

`MyWatchlistView` 从“一级主导航页面”调整为“我的关注”二级模块。

#### 必改内容

- 页面主标题从“关注页 / 我的关注股票”统一收口为“我的关注”
- CTA 文案从“返回关注页 / 去关注页”收口为“返回我的关注 / 去我的关注”
- 页面说明文案减少“一级固定回访入口”表述，改成“个人跟踪清单”表述

#### 第一轮不改内容

- 不拆 watchlist explanation、history、version compare、proof grid 等现有能力
- 不改 watchlist 本地存储、快照同步和版本对比逻辑

#### 影响文件

- `/Users/gjhan21/cursor/sercherai/client/src/views/MyWatchlistView.vue`

### 5. 社区功能完善范围

本轮“完善社区功能”不做大规模重构，重点是补足它作为一级主入口时最缺的部分。

#### 第一轮增强目标

1. 社区页文案更明确地表达“发现观点 / 跟进讨论 / 发布观点”
2. 一级入口下的默认视觉焦点更清晰
3. 登录用户能够更快到达：
   - 我的主题
   - 我的评论
   - 发布我的观点
4. 与策略页、资讯页的跳转关系保持连续

#### 不做的事情

- 不新建社区设计系统
- 不改动社区数据接口契约
- 不重写 topic detail 或 composer

#### 影响文件

- `/Users/gjhan21/cursor/sercherai/client/src/views/CommunityView.vue`

### 6. 全站 CTA 与文案回收

这一步是本次调整是否真正“收口”的关键。

#### 必须回收的口径

凡是把 `/watchlist` 当成一级主入口宣传的地方，都要改成：

- “去我的关注”
- “进入我的关注”
- “在我的里查看关注”

#### 优先级最高的页面

1. `/Users/gjhan21/cursor/sercherai/client/src/views/MembershipView.vue`
2. `/Users/gjhan21/cursor/sercherai/client/src/views/HomeView.vue`
3. `/Users/gjhan21/cursor/sercherai/client/src/views/ProfileView.vue`
4. `/Users/gjhan21/cursor/sercherai/client/src/views/MyWatchlistView.vue`
5. `/Users/gjhan21/cursor/sercherai/client/src/views/CommunityView.vue` 中可能引用 watchlist 的链接文案

#### 第一轮最低要求

- 改关键 CTA 的目标路径
- 改最显眼的页面文案
- 保留少量深层说明中的旧措辞可以接受，但不能再让用户通过主链路感知到“关注是一级主入口”

## 数据流和兼容策略

### 路由兼容

- 旧链接 `/watchlist` -> 重定向到 `/profile/watchlist`
- 旧按钮若暂时没来得及全部替换，也不至于 404

### 页面复用

- `MyWatchlistView.vue` 继续复用现有数据流：
  - `/Users/gjhan21/cursor/sercherai/client/src/lib/watchlist.js`
  - 本地 watchlist 存储
  - explanation / history / version compare 聚合

### 社区复用

- `CommunityView.vue` 保持原接口和原页面主结构
- 只增强入口感知和首屏行为，不改 API

## 测试与验收标准

### 路由

1. PC 路由存在 `/community`
2. PC 路由存在 `/profile/watchlist`
3. `/watchlist` 会重定向到 `/profile/watchlist`

### 顶部导航

1. 顶部导航不再出现 `关注`
2. 顶部导航出现 `社区`
3. 点击 `社区` 能进入正确页面

### 我的页

1. `ProfileView` 出现“我的关注”入口卡片
2. 点击后进入 `profile/watchlist`

### 关注页

1. 页面可正常加载原有关注数据
2. 标题和关键 CTA 文案已收口为“我的关注”
3. 旧地址访问仍然可达

### 社区页

1. 能作为一级入口正常访问
2. 首屏 CTA 明确可见
3. 登录态下“我的主题 / 我的评论 / 发布观点”路径清晰

### 文案回归

1. 一级主链不再把 watchlist 当成主入口宣传
2. 会员页/首页关键 CTA 已改向“我的关注”

## 实施顺序

### 包 1：路由与导航替换

- 接入 `/community`
- 增加 `/profile/watchlist`
- `/watchlist` 做兼容重定向
- 顶部导航 `关注 -> 社区`

### 包 2：我的页承接关注入口

- `ProfileView` 增加“我的关注”入口卡片
- 入口跳转到 `/profile/watchlist`

### 包 3：关注页定位收口

- `MyWatchlistView` 标题、说明、关键按钮文案改为“我的关注”口径

### 包 4：社区一级入口增强

- 强化 `CommunityView` 首屏入口感
- 明确“发布观点 / 我的主题 / 我的评论”动作优先级

### 包 5：全站 CTA 收尾

- 首页、会员页和相关页面把 `/watchlist` CTA 改成 `/profile/watchlist`
- 清理关键一级主入口文案

## 风险

1. `MembershipView.vue` 中 watchlist 文案和 CTA 很多，容易遗漏
2. `MyWatchlistView.vue` 页面体量较大，改文案时要避免误改逻辑
3. 如果只改导航不改 CTA，用户仍会通过旧按钮感知到 watchlist 是一级入口
4. 社区页如果只加路由不补首屏强化，会出现“入口换了，但心智没换”

## 决策结论

本次采用方案 A：

- 社区升级为 PC 一级主入口
- 关注收口为“我的”中的二级模块
- 保留现有关注页面能力，但重构入口、路由语义和产品文案
- 以最小风险达成产品信息架构调整，并为后续继续加强社区能力保留空间
