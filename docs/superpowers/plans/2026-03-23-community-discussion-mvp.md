# Community Discussion MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a first-phase financial discussion plaza that lets users publish structured viewpoints tied to stocks, futures, news, and strategy items, while giving admins enough moderation tools to keep the module safe.

**Architecture:** Extend the existing growth monolith instead of creating a standalone service. Add community schema, repo/service/handler methods, client discussion routes and views, and one admin moderation console. Keep phase 1 read-heavy, thread-based, and non-real-time.

**Tech Stack:** Go 1.20 + Gin + MySQL/in-memory repo fallback, Vue 3 + Vue Router + Vite for client, Vue 3 + Element Plus + Vite for admin.

---

## File Structure

### Backend

- Create: `backend/migrations/20260323_01_community_discussion_mvp.sql`
  Responsibility: create discussion tables and RBAC permission rows.
- Create: `backend/internal/growth/model/community_discussion.go`
  Responsibility: community topic, comment, reaction, report models and list-row payloads.
- Create: `backend/internal/growth/dto/community_discussion.go`
  Responsibility: request/query DTOs for public, user, and admin community APIs.
- Create: `backend/internal/growth/repo/community_discussion_mysql.go`
  Responsibility: MySQL queries for topics, comments, reactions, reports, and moderation.
- Create: `backend/internal/growth/repo/community_discussion_inmemory.go`
  Responsibility: in-memory fallback behavior for local/dev usage.
- Create: `backend/internal/growth/repo/community_discussion_repo_test.go`
  Responsibility: repository coverage for list/detail/write/moderation behavior.
- Create: `backend/internal/growth/service/community_discussion.go`
  Responsibility: growthService methods for community flows and reply notifications.
- Create: `backend/internal/growth/handler/user_growth_handler_community.go`
  Responsibility: public and authenticated client-facing community handlers.
- Create: `backend/internal/growth/handler/admin_growth_handler_community.go`
  Responsibility: admin moderation handlers.
- Create: `backend/internal/growth/handler/community_handler_test.go`
  Responsibility: handler-level auth/validation/visibility tests.
- Modify: `backend/internal/growth/repo/interfaces.go`
  Responsibility: add repo interface methods for community workflows.
- Modify: `backend/internal/growth/service/service.go`
  Responsibility: add service interface methods used by handlers.
- Modify: `backend/internal/growth/repo/mysql_repo.go`
  Responsibility: wire any shared helper access needed by the new focused repo file.
- Modify: `backend/internal/growth/repo/inmemory_repo.go`
  Responsibility: wire any shared in-memory structures needed by the new focused repo file.
- Modify: `backend/internal/growth/repo/mysql_repo_data_source_test.go` only if shared test helpers must move.
- Modify: `backend/router/router.go`
  Responsibility: register public, user, and admin community routes.
- Modify: `backend/scripts/seed_demo.sql`
  Responsibility: seed minimal sample topics/comments/reports for local demos.
- Modify: `backend/scripts/seed_admin_extra.sql`
  Responsibility: attach new community permissions to appropriate roles.

### Client

- Create: `client/src/api/community.js`
  Responsibility: client API wrappers for plaza, topic detail, comments, reactions, and reports.
- Create: `client/src/views/CommunityView.vue`
  Responsibility: discussion plaza list page with filters and my-participation states.
- Create: `client/src/views/CommunityTopicView.vue`
  Responsibility: topic detail, comments, linked-object card, and interaction controls.
- Create: `client/src/views/CommunityComposeView.vue`
  Responsibility: create-topic flow with structured financial fields and disclaimer.
- Modify: `client/src/router/index.js`
  Responsibility: add `/community`, `/community/topics/:id`, and `/community/new`.
- Modify: `client/src/components/ClientLayout.vue`
  Responsibility: add new nav entry and footer references.
- Modify: `client/src/views/NewsView.vue`
  Responsibility: add related-discussion and start-discussion entry points.
- Modify: `client/src/views/StrategyView.vue`
  Responsibility: add related-discussion and publish-opinion entry points.
- Modify: `client/src/views/ProfileView.vue`
  Responsibility: add shortcuts into "my topics" / "my comments" views if time allows in phase 1.

### Admin

- Create: `admin/src/views/CommunityView.vue`
  Responsibility: moderation console with tabs for topics, comments, and reports.
- Modify: `admin/src/api/admin.js`
  Responsibility: admin community list/update/review API helpers.
- Modify: `admin/src/router/index.js`
  Responsibility: register admin community route.
- Modify: `admin/src/lib/admin-navigation.js`
  Responsibility: add moderation nav item behind the new permission.
- Modify: `admin/src/lib/admin-navigation.test.js`
  Responsibility: cover the new permission-gated nav item.

## Cross-Cutting Constraints

- Keep the community module inside the current growth service/repo architecture.
- Do not add WebSocket, SSE, or background worker requirements for phase 1.
- Reuse `messages` for reply notifications instead of creating a new inbox subsystem.
- Keep client UI within the current blue-gold financial design system.
- Preserve one shared PC + H5 structure; do not design a separate app-style chat UI.

## Task 1: Define Backend Contracts And Failing Tests

**Files:**
- Create: `backend/internal/growth/model/community_discussion.go`
- Create: `backend/internal/growth/dto/community_discussion.go`
- Create: `backend/internal/growth/repo/community_discussion_repo_test.go`
- Create: `backend/internal/growth/handler/community_handler_test.go`
- Modify: `backend/internal/growth/repo/interfaces.go`
- Modify: `backend/internal/growth/service/service.go`

- [ ] **Step 1: Write the failing repository test skeleton**

```go
func TestCommunityTopicListFiltersPublishedTopics(t *testing.T) {
    repo := newCommunityRepoForTest(t)
    rows, total, err := repo.ListCommunityTopics("", model.CommunityTopicListQuery{
        TopicType: "STOCK",
        Page:      1,
        PageSize:  20,
    })
    if err != nil {
        t.Fatalf("ListCommunityTopics() error = %v", err)
    }
    if total == 0 || len(rows) == 0 {
        t.Fatalf("expected published stock topics")
    }
}
```

- [ ] **Step 2: Write the failing handler test skeleton**

```go
func TestCreateCommunityTopicRequiresAuth(t *testing.T) {
    growthHandler := newUserGrowthHandlerForTest(t)
    router := gin.New()
    router.POST("/api/v1/community/topics", growthHandler.CreateCommunityTopic)
    req := httptest.NewRequest(http.MethodPost, "/api/v1/community/topics", strings.NewReader(`{"title":"x"}`))
    rec := httptest.NewRecorder()
    router.ServeHTTP(rec, req)
    if rec.Code != http.StatusUnauthorized {
        t.Fatalf("expected 401, got %d", rec.Code)
    }
}
```

- [ ] **Step 3: Run the focused tests to confirm they fail**

Run: `cd backend && go test ./internal/growth/repo ./internal/growth/handler -run 'Community|community'`

Expected: FAIL because the new contracts and handlers do not exist yet.

- [ ] **Step 4: Add the new model, DTO, repo-interface, and service-interface contracts**

Define explicit enums and payloads:

```go
type CommunityTopicType string

const (
    CommunityTopicTypeStock    CommunityTopicType = "STOCK"
    CommunityTopicTypeFutures  CommunityTopicType = "FUTURES"
    CommunityTopicTypeNews     CommunityTopicType = "NEWS"
    CommunityTopicTypeStrategy CommunityTopicType = "STRATEGY"
)
```

- [ ] **Step 5: Re-run the same tests to confirm failures are now implementation-level, not compile-level**

Run: `cd backend && go test ./internal/growth/repo ./internal/growth/handler -run 'Community|community'`

Expected: FAIL on missing repo/handler behavior instead of undefined types.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/growth/model/community_discussion.go \
  backend/internal/growth/dto/community_discussion.go \
  backend/internal/growth/repo/interfaces.go \
  backend/internal/growth/service/service.go \
  backend/internal/growth/repo/community_discussion_repo_test.go \
  backend/internal/growth/handler/community_handler_test.go
git commit -m "test: define community discussion contracts"
```

## Task 2: Add Schema, RBAC, And Repo Implementations

**Files:**
- Create: `backend/migrations/20260323_01_community_discussion_mvp.sql`
- Create: `backend/internal/growth/repo/community_discussion_mysql.go`
- Create: `backend/internal/growth/repo/community_discussion_inmemory.go`
- Modify: `backend/internal/growth/repo/mysql_repo.go`
- Modify: `backend/internal/growth/repo/inmemory_repo.go`
- Modify: `backend/scripts/seed_demo.sql`
- Modify: `backend/scripts/seed_admin_extra.sql`
- Test: `backend/internal/growth/repo/community_discussion_repo_test.go`

- [ ] **Step 1: Implement the migration and seed data**

Add tables and permission rows:

```sql
CREATE TABLE discussion_topics (...);
CREATE TABLE discussion_topic_links (...);
CREATE TABLE discussion_comments (...);
CREATE TABLE discussion_reactions (...);
CREATE TABLE discussion_reports (...);

INSERT INTO rbac_permissions (code, name, module, action, description, status, created_at, updated_at)
VALUES
  ('community.view', 'Community View', 'COMMUNITY', 'VIEW', 'view community moderation', 'ACTIVE', NOW(), NOW()),
  ('community.edit', 'Community Edit', 'COMMUNITY', 'EDIT', 'edit topics and comments', 'ACTIVE', NOW(), NOW()),
  ('community.review', 'Community Review', 'COMMUNITY', 'REVIEW', 'review community reports', 'ACTIVE', NOW(), NOW());
```

- [ ] **Step 2: Implement MySQL repo methods for list/detail/create/comment/reaction/report/moderation**

Cover:

- public list only returns `PUBLISHED`
- authenticated detail can include `liked_by_me` / `favorited_by_me`
- reaction writes are idempotent per `(user_id, target_type, target_id, reaction_type)`
- topic `last_active_at` updates when new comments land

- [ ] **Step 3: Implement the in-memory repo equivalents**

Keep the in-memory path simple but behaviorally aligned for local development and handler tests.

- [ ] **Step 4: Run the repository tests**

Run: `cd backend && go test ./internal/growth/repo -run 'Community|community'`

Expected: PASS

- [ ] **Step 5: Run a wider backend safety pass**

Run: `cd backend && go test ./internal/growth/repo ./internal/growth/handler`

Expected: PASS without regressions in existing repo/handler tests.

- [ ] **Step 6: Commit**

```bash
git add backend/migrations/20260323_01_community_discussion_mvp.sql \
  backend/internal/growth/repo/community_discussion_mysql.go \
  backend/internal/growth/repo/community_discussion_inmemory.go \
  backend/internal/growth/repo/mysql_repo.go \
  backend/internal/growth/repo/inmemory_repo.go \
  backend/scripts/seed_demo.sql \
  backend/scripts/seed_admin_extra.sql \
  backend/internal/growth/repo/community_discussion_repo_test.go
git commit -m "feat: add community discussion persistence"
```

## Task 3: Implement Public And User Community APIs

**Files:**
- Create: `backend/internal/growth/service/community_discussion.go`
- Create: `backend/internal/growth/handler/user_growth_handler_community.go`
- Modify: `backend/router/router.go`
- Modify: `backend/internal/growth/handler/community_handler_test.go`
- Test: `backend/internal/growth/handler/community_handler_test.go`

- [ ] **Step 1: Extend the failing handler tests for public visibility and auth writes**

Add coverage for:

- visitor can list plaza
- visitor cannot publish
- logged-in user can publish
- logged-in user can comment
- invalid stance/topic_type returns 400

- [ ] **Step 2: Implement service methods and reply-notification hook**

When creating a comment or reply:

- increment counters
- update `last_active_at`
- create a `messages` record for the topic owner or replied-to user, skipping self-notifications

- [ ] **Step 3: Implement handlers and register routes**

Add:

- `GET /api/v1/public/community/topics`
- `GET /api/v1/public/community/topics/:id`
- `GET /api/v1/public/community/topics/:id/comments`
- `GET /api/v1/community/topics`
- `POST /api/v1/community/topics`
- `GET /api/v1/community/topics/:id`
- `GET /api/v1/community/topics/:id/comments`
- `POST /api/v1/community/topics/:id/comments`
- `POST /api/v1/community/reactions`
- `DELETE /api/v1/community/reactions`
- `POST /api/v1/community/reports`
- `GET /api/v1/community/me/topics`
- `GET /api/v1/community/me/comments`

- [ ] **Step 4: Run handler tests**

Run: `cd backend && go test ./internal/growth/handler -run 'Community|community'`

Expected: PASS

- [ ] **Step 5: Run the broader backend package tests again**

Run: `cd backend && go test ./internal/growth/repo ./internal/growth/handler`

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add backend/internal/growth/service/community_discussion.go \
  backend/internal/growth/handler/user_growth_handler_community.go \
  backend/router/router.go \
  backend/internal/growth/handler/community_handler_test.go
git commit -m "feat: add community discussion user APIs"
```

## Task 4: Implement Admin Moderation APIs And Permissions

**Files:**
- Create: `backend/internal/growth/handler/admin_growth_handler_community.go`
- Modify: `backend/router/router.go`
- Modify: `backend/internal/growth/service/community_discussion.go`
- Modify: `backend/internal/growth/repo/community_discussion_mysql.go`
- Modify: `backend/internal/growth/repo/community_discussion_inmemory.go`
- Modify: `backend/internal/growth/handler/community_handler_test.go`

- [ ] **Step 1: Add failing admin handler tests**

Cover:

- `community.view` required for list endpoints
- `community.edit` required for topic/comment status changes
- `community.review` required for report resolution

- [ ] **Step 2: Implement admin list/update/review methods**

Required admin actions:

- list topics with filters by status, topic type, user, report count
- update topic status
- list comments with filters by status/topic
- update comment status
- list reports
- resolve or reject reports with review note

- [ ] **Step 3: Register admin routes under `/api/v1/admin/community`**

Example:

```go
adminCommunity := v1.Group("/admin/community")
adminCommunity.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
```

- [ ] **Step 4: Run handler tests**

Run: `cd backend && go test ./internal/growth/handler -run 'Community|community'`

Expected: PASS

- [ ] **Step 5: Run a full backend test sweep used by this feature**

Run: `cd backend && go test ./internal/growth/repo ./internal/growth/handler`

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add backend/internal/growth/handler/admin_growth_handler_community.go \
  backend/router/router.go \
  backend/internal/growth/service/community_discussion.go \
  backend/internal/growth/repo/community_discussion_mysql.go \
  backend/internal/growth/repo/community_discussion_inmemory.go \
  backend/internal/growth/handler/community_handler_test.go
git commit -m "feat: add community moderation APIs"
```

## Task 5: Build Client Routing, Navigation, And Read-Only Plaza

**Files:**
- Create: `client/src/api/community.js`
- Create: `client/src/views/CommunityView.vue`
- Modify: `client/src/router/index.js`
- Modify: `client/src/components/ClientLayout.vue`

- [ ] **Step 1: Add the client API wrapper**

Expose wrappers such as:

```js
export function listCommunityTopics(params) {
  return http.get(resolveCommunityPath("/community/topics"), { params: buildParams(params) });
}
```

Use the current auth token presence to choose `/public/community/...` vs `/community/...` when the operation is read-only.

- [ ] **Step 2: Register client routes**

Add:

- `/community`
- `/community/topics/:id`
- `/community/new` with `meta: { requiresAuth: true }`

- [ ] **Step 3: Add the nav entry and shell copy**

Label recommendation:

- nav label: `讨论`
- signal/focus copy should reinforce discussion and risk awareness, not "chat"

- [ ] **Step 4: Build the plaza page in read-only mode first**

The first pass should support:

- page intro and risk note
- topic type filter
- latest / most active sorting
- desktop right rail
- mobile single-column compression

- [ ] **Step 5: Run the client build**

Run: `cd client && npm run build`

Expected: PASS

- [ ] **Step 6: Manual verification**

Check:

- `/community` loads as visitor
- desktop right rail stays readable
- H5 keeps one-column reading order

- [ ] **Step 7: Commit**

```bash
git add client/src/api/community.js \
  client/src/views/CommunityView.vue \
  client/src/router/index.js \
  client/src/components/ClientLayout.vue
git commit -m "feat: add community plaza shell"
```

## Task 6: Add Topic Detail, Composer, Comments, And Interactions

**Files:**
- Create: `client/src/views/CommunityTopicView.vue`
- Create: `client/src/views/CommunityComposeView.vue`
- Modify: `client/src/api/community.js`
- Modify: `client/src/router/index.js`

- [ ] **Step 1: Build topic detail page**

Include:

- topic header
- linked-object card
- risk disclaimer strip
- metrics and actions
- comments and single-level replies

- [ ] **Step 2: Build create-topic flow**

Fields:

- title
- topic_type
- linked object context
- stance
- time horizon
- reason text
- risk text
- body
- agreement checkbox

- [ ] **Step 3: Add write interactions**

Support:

- create topic
- create comment
- reply to comment
- like
- favorite
- report

- [ ] **Step 4: Ensure visitor gating is explicit**

Visitors should see:

- read access
- disabled write buttons
- auth redirect preserving return path

- [ ] **Step 5: Run the client build**

Run: `cd client && npm run build`

Expected: PASS

- [ ] **Step 6: Manual verification**

Check:

- visitor vs logged-in action differences
- topic create validation
- reply notification flow after backend is wired
- H5 composer layout remains usable

- [ ] **Step 7: Commit**

```bash
git add client/src/views/CommunityTopicView.vue \
  client/src/views/CommunityComposeView.vue \
  client/src/api/community.js \
  client/src/router/index.js
git commit -m "feat: add community topic detail and composer"
```

## Task 7: Link Community Into News, Strategy, And Profile

**Files:**
- Modify: `client/src/views/NewsView.vue`
- Modify: `client/src/views/StrategyView.vue`
- Modify: `client/src/views/ProfileView.vue`
- Modify: `client/src/views/CommunityView.vue`
- Modify: `client/src/views/CommunityComposeView.vue`

- [ ] **Step 1: Add News entry points**

From article detail:

- "related discussions"
- "start discussion about this article"

The compose page should receive query/context values such as:

```js
{ topic_type: "NEWS", target_type: "NEWS_ARTICLE", target_id: article.id }
```

- [ ] **Step 2: Add Strategy entry points**

From strategy detail:

- "view related discussion"
- "publish opinion"

Use stock/futures context where available so users do not need to manually reselect the object.

- [ ] **Step 3: Add Profile shortcuts or mine filters**

Minimum acceptable phase 1 behavior:

- profile shortcut into `/community?mine=topics`
- profile shortcut into `/community?mine=comments`

- [ ] **Step 4: Run the client build**

Run: `cd client && npm run build`

Expected: PASS

- [ ] **Step 5: Manual verification**

Check:

- News -> Community context carry-over
- Strategy -> Community context carry-over
- "my topics" and "my comments" paths remain reachable

- [ ] **Step 6: Commit**

```bash
git add client/src/views/NewsView.vue \
  client/src/views/StrategyView.vue \
  client/src/views/ProfileView.vue \
  client/src/views/CommunityView.vue \
  client/src/views/CommunityComposeView.vue
git commit -m "feat: connect community with news and strategy flows"
```

## Task 8: Build Admin Moderation Console

**Files:**
- Create: `admin/src/views/CommunityView.vue`
- Modify: `admin/src/api/admin.js`
- Modify: `admin/src/router/index.js`
- Modify: `admin/src/lib/admin-navigation.js`
- Modify: `admin/src/lib/admin-navigation.test.js`

- [ ] **Step 1: Add admin API wrappers**

Support:

- list community topics
- update topic status
- list community comments
- update comment status
- list reports
- review reports

- [ ] **Step 2: Register the admin route and nav item**

Permission gate:

- route permission: `community.view`
- update buttons: `community.edit`
- report review actions: `community.review`

- [ ] **Step 3: Build one consolidated moderation page**

Recommended desktop layout:

- tabs: Topics / Comments / Reports
- filters per tab
- status actions inline or in a side drawer

- [ ] **Step 4: Update navigation tests**

Add expectation that the new nav item is only visible when `community.view` is present.

- [ ] **Step 5: Run the admin build and tests**

Run:

- `cd admin && npm run build`
- `cd admin && node --test src/lib/admin-navigation.test.js`

Expected: PASS

- [ ] **Step 6: Manual verification**

Check:

- no nav leak without permission
- moderators can hide topics and comments
- report review updates status correctly

- [ ] **Step 7: Commit**

```bash
git add admin/src/views/CommunityView.vue \
  admin/src/api/admin.js \
  admin/src/router/index.js \
  admin/src/lib/admin-navigation.js \
  admin/src/lib/admin-navigation.test.js
git commit -m "feat: add community moderation console"
```

## Task 9: Final Integration Verification

**Files:**
- No new files required.

- [ ] **Step 1: Run backend verification**

Run: `cd backend && go test ./internal/growth/repo ./internal/growth/handler`

Expected: PASS

- [ ] **Step 2: Run frontend verification**

Run:

- `cd client && npm run build`
- `cd admin && npm run build`

Expected: PASS

- [ ] **Step 3: Run manual end-to-end checks**

Desktop checks:

- visitor can browse plaza and detail
- logged-in user can create topic and comment
- topic reply creates a site message
- News and Strategy entry points deep-link correctly
- admin can hide a topic and resolve a report

H5 checks:

- plaza is one-column
- topic detail reads top-to-bottom without side-rail breakage
- composer is usable without modal layering issues

- [ ] **Step 4: Capture rollout notes**

Document:

- migration order
- seed expectations
- required RBAC role assignment
- phase-2 backlog items intentionally deferred

- [ ] **Step 5: Commit**

```bash
git add .
git commit -m "chore: finish community discussion mvp integration"
```

## Deferred Items

Do not add during this plan unless the human explicitly expands scope:

- real-time transport
- image upload
- direct message flows
- follower graph
- ranking algorithm work beyond latest/most active
- automated investment-performance scoring of user posts

## Execution Notes

- Because `client` currently has no dedicated unit-test harness, treat `npm run build` plus desktop/H5 manual checks as the required verification bar there.
- Because `admin` currently has minimal test coverage, keep the new admin surface consolidated in one page and protect it with RBAC plus navigation test coverage.
- Keep route and API naming consistent with the existing monolith conventions so future agents can discover the feature quickly.
