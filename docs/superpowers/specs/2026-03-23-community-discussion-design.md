# Community Discussion Plaza Design

## Summary

Build a first-phase community module for the client as a discussion plaza focused on financial viewpoints, not a real-time chatroom. Users can publish structured opinions around stocks, futures, news reports, and strategy items; other users can comment, like, favorite, and report content. Admin must be able to review, hide, and resolve reported content before the module is exposed broadly across the site.

The design intentionally avoids real-time chat, private messaging, image upload, deep social graphs, and complex recommendation feeds. The goal is to validate whether users will produce and engage with high-signal discussion content that extends the current decision, news, and watchlist workflows.

## Product Positioning

The module is a "personal viewpoint and discussion" layer for the existing Chinese financial site. It is not:

- a signal room
- a paid lead-generation group
- a guaranteed stock recommendation product
- a general-purpose social feed

The platform hosts user discussions, but does not endorse, guarantee, or operationally follow user recommendations.

## Goals

- Let logged-in users publish structured opinions tied to real financial content already present in the product.
- Let readers discuss a stock, futures contract, news article, or strategy item in a durable thread model.
- Create a reusable interaction layer that can later be linked from Home, News, Strategy, Watchlist, and Profile.
- Reuse the existing account, membership, message, admin permission, and content-management infrastructure where possible.
- Keep phase 1 small enough to ship without introducing WebSocket infrastructure or a new recommendation engine.

## Non-Goals

- Real-time chat, group chat, or presence indicators
- Private messaging between users
- Images, file attachments, or voice messages
- Multi-level nested comments
- Follower graphs, badges, or gamification
- Automatic profit tracking or "verified picks" scoring
- AI moderation pipelines or semantic ranking in phase 1

## Approved Phase 1 Scope

Phase 1 includes:

- discussion plaza list page
- topic detail page
- dedicated create-topic flow
- topic comments with one reply level
- likes and favorites
- report flow
- admin moderation pages for topics, comments, and reports
- basic site-message notifications for replies
- link-in entry points from News and Strategy first

Phase 1 excludes:

- Home and Watchlist deep surfacing beyond simple future placeholders
- automatic hot-ranking algorithms beyond simple latest/most active ordering
- real-time polling or live socket updates

## User Roles And Permissions

### Visitor

- can browse public discussion plaza lists
- can view public topic details and comments
- cannot publish, comment, like, favorite, or report

### Logged-In User

- can publish topics
- can comment and reply
- can like and favorite
- can report abusive content
- can see "my topics" and "my comments"

### VIP User

VIP is not required to participate in phase 1. The module should not be gated behind membership. If later needed, VIP can receive higher posting quotas or stronger profile labels, but that is out of scope for the MVP.

### Admin Moderator

- can list topics, comments, and reports
- can update topic status
- can update comment status
- can resolve reports
- can review user-generated discussion content through new RBAC permissions

## Information Architecture

## Client Pages

### 1. Discussion Plaza

Primary destination for browsing topics.

Sections:

- page intro and risk note
- filters: all / stocks / futures / reports / strategies
- sort: latest / most active
- topic feed
- right rail on desktop: posting rules, risk statement, related-page entry guidance

### 2. Topic Detail

Primary destination for reading a discussion in full.

Sections:

- topic body
- linked object card
- metrics and actions
- comment list
- reply composer
- related object links back to News or Strategy

### 3. Create Topic

Dedicated page, not a modal, on desktop and mobile for phase 1.

Fields:

- title
- topic type
- linked object
- stance
- time horizon
- reason text
- risk text
- body
- agreement checkbox for community risk disclaimer

### 4. My Discussions

May live inside Profile in phase 1 to avoid creating yet another standalone route tree.

Sections:

- my topics
- my comments
- saved topics

## Entry Points

Phase 1 link-in points:

- News detail: "related discussions" and "start discussion"
- Strategy detail: "view discussion" and "publish opinion"

Phase 2 link-in points:

- Home spotlight block
- Watchlist recent discussion changes

## Topic Model Design

Each topic is anchored to one financial object family:

- STOCK
- FUTURES
- NEWS
- STRATEGY

Each topic also carries a stance:

- BULLISH
- BEARISH
- WATCH

Recommendation-style topics must be structured, not free-form only. Phase 1 requires explicit reason and risk text so that the module does not degrade into one-line "buy this now" posts.

## Data Model

## New Tables

### discussion_topics

- id
- user_id
- title
- summary
- content
- topic_type
- stance
- time_horizon
- reason_text
- risk_text
- status
- comment_count
- like_count
- favorite_count
- report_count
- last_active_at
- created_at
- updated_at

Status values:

- PUBLISHED
- PENDING_REVIEW
- HIDDEN
- DELETED

### discussion_topic_links

- id
- topic_id
- target_type
- target_id
- target_snapshot
- created_at

This stores a display snapshot so the topic card remains readable if the linked object later changes title or becomes unavailable.

### discussion_comments

- id
- topic_id
- user_id
- parent_comment_id
- reply_to_user_id
- content
- status
- like_count
- created_at
- updated_at

Comment status values:

- PUBLISHED
- HIDDEN
- DELETED
- PENDING_REVIEW

### discussion_reactions

- id
- user_id
- target_type
- target_id
- reaction_type
- created_at

Reaction types:

- LIKE
- FAVORITE

### discussion_reports

- id
- reporter_user_id
- target_type
- target_id
- reason
- status
- review_note
- created_at
- updated_at

Report status values:

- PENDING
- RESOLVED
- REJECTED

## Reused Existing Tables

- users and auth/session data
- messages for notification delivery
- existing news and strategy tables for object linking
- rbac_permissions and rbac_role_permissions for admin moderation access

## API Design

## Public Read Endpoints

- GET /api/v1/public/community/topics
- GET /api/v1/public/community/topics/:id
- GET /api/v1/public/community/topics/:id/comments

These must expose only published topics and published comments.

## Authenticated User Endpoints

- GET /api/v1/community/topics
- POST /api/v1/community/topics
- GET /api/v1/community/topics/:id
- GET /api/v1/community/topics/:id/comments
- POST /api/v1/community/topics/:id/comments
- POST /api/v1/community/reactions
- DELETE /api/v1/community/reactions
- POST /api/v1/community/reports
- GET /api/v1/community/me/topics
- GET /api/v1/community/me/comments

## Admin Endpoints

- GET /api/v1/admin/community/topics
- PUT /api/v1/admin/community/topics/:id/status
- GET /api/v1/admin/community/comments
- PUT /api/v1/admin/community/comments/:id/status
- GET /api/v1/admin/community/reports
- PUT /api/v1/admin/community/reports/:id/review

## Request And Response Shape Principles

- reuse the existing page/page_size pattern
- reuse dto validation style from current handlers
- keep reaction and report payloads minimal
- include current user interaction flags in topic detail and list rows when authenticated:
  - liked_by_me
  - favorited_by_me
  - can_comment

## Moderation And Risk Controls

This feature is financially sensitive and must ship with explicit controls.

### Product Rules

- user posts represent personal opinion only
- no guarantee-of-profit language
- no lead generation or private contact solicitation
- no paid signal-group promotion
- no fake insider news or manipulative claims

### Posting Constraints

For recommendation-style topics, require:

- stance
- reason_text
- risk_text
- time_horizon

### Release Strategy

Phase 1 moderation model:

- default publish for normal content
- configurable keyword blocking or pending-review routing for risky text
- user reporting
- admin moderation after publish

### UI Risk Surfaces

- pre-submit disclaimer on create topic page
- topic detail disclaimer below header
- stronger warning strip for recommendation-style topics

## Notification Strategy

Do not create a new notification subsystem in phase 1.

Instead, reuse the existing user messages infrastructure for:

- someone commented on my topic
- someone replied to my comment

Notification payload should include:

- topic title
- short reply excerpt
- deep link path to topic detail

## Admin And RBAC Plan

Add new permissions:

- community.view
- community.edit
- community.review

Admin navigation should expose one new module entry, labeled in Chinese for the admin UI, that consolidates:

- topic moderation
- comment moderation
- report resolution

## Client UX Principles

- keep the same blue-gold financial visual system already used by the rest of client
- maintain shared PC + H5 structure; no separate app-like chat UI
- keep the plaza in content-reading rhythm, not message-bubble rhythm
- avoid giant empty social widgets if there is not enough data

## Technical Constraints

- no new global state management system
- no backend contract changes to existing news, strategy, or membership APIs
- no independent community service; extend current growth service/repo pattern
- keep view-level client logic consistent with the current app structure

## Rollout Plan

### Phase 1

- schema
- public browse
- auth publish/comment
- reactions and reports
- admin moderation
- News/Strategy entry points

### Phase 2

- Home spotlight
- Watchlist discussion linkage
- richer sorting
- stronger moderator tools
- quota tuning by membership or trust level

### Phase 3

- evaluate whether stronger social features are justified
- only then revisit tags, ranking, or near-real-time refresh

## Testing And Acceptance

### Backend

- repository tests for topic CRUD, list filtering, reaction toggles, and moderation status
- handler tests for auth, validation, and visibility boundaries

### Client

- build passes with Vite
- plaza and topic detail render correctly on desktop and H5
- visitor vs logged-in write permissions are respected
- News and Strategy entry points navigate with the correct linked object context

### Admin

- build passes with Vite
- permission-gated navigation works
- moderators can list, hide, and resolve without breaking existing modules

## Assumptions

- linked objects will use existing IDs from stock, futures, news article, and strategy sources
- phase 1 can accept simple linked-object selection via existing page entry context rather than building a brand-new global search picker
- if direct stock or futures lookup is needed during manual post creation, a lightweight server-backed lookup can be added later, not in the MVP
- reply notifications can be delivered through the existing messages table without introducing a new inbox model

## Final Recommendation

Ship the first version as a discussion plaza anchored to real financial content. Do not attempt a real-time chatroom in the same phase. Validate discussion behavior first, then expand only if engagement quality and moderation load remain healthy.
