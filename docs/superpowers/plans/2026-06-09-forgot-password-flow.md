# Phone+Email Registration and Forgot/Reset Password Flow Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Allow users to register with both phone and email, and recover/reset their password via a public email verification code flow.

**Architecture:** We will create a database table `password_reset_codes` to store generated codes. The backend will expose two public endpoints (`/auth/forgot-password` and `/auth/reset-password`). For development convenience, `/auth/forgot-password` returns the generated code in the JSON payload. Frontend registration and login views will be updated to support the new flows.

**Tech Stack:** Go (Gin, SQL), Vue 3 (Composition API), Vite

---

### Task 1: Database Migration

**Files:**
- Create: `backend/migrations/20260609_05_password_reset_codes.sql`

- [ ] **Step 1: Write database migration SQL**
  Create the migration file `backend/migrations/20260609_05_password_reset_codes.sql` containing:
  ```sql
  CREATE TABLE IF NOT EXISTS `password_reset_codes` (
      `id` VARCHAR(64) NOT NULL,
      `email` VARCHAR(128) NOT NULL,
      `code` VARCHAR(10) NOT NULL,
      `status` VARCHAR(20) NOT NULL DEFAULT 'UNUSED', -- UNUSED, USED
      `expired_at` TIMESTAMP NOT NULL,
      `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
      PRIMARY KEY (`id`),
      INDEX `idx_email_code` (`email`, `code`, `status`)
  ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
  ```

- [ ] **Step 2: Commit DDL migration**
  ```bash
  git add backend/migrations/20260609_05_password_reset_codes.sql
  git commit -m "migration: create password_reset_codes table"
  ```

---

### Task 2: Backend API Router and Handlers

**Files:**
- Modify: `backend/router/auth.go`
- Modify: `backend/internal/growth/handler/auth_handler.go`

- [ ] **Step 1: Add endpoints to router**
  Modify `backend/router/auth.go` to expose the new public endpoints:
  ```go
  // Target Content around line 16:
  authGroup.POST("/register", authHandler.Register)
  authGroup.POST("/login", authHandler.Login)
  
  // Replacement Content:
  authGroup.POST("/register", authHandler.Register)
  authGroup.POST("/login", authHandler.Login)
  authGroup.POST("/forgot-password", authHandler.ForgotPassword)
  authGroup.POST("/reset-password", authHandler.ResetPassword)
  ```

- [ ] **Step 2: Implement handlers and helpers in auth_handler.go**
  Add imports `"crypto/rand"`, `"math/big"`, `"time"`, `"net/http"`, and `"strings"` to `backend/internal/growth/handler/auth_handler.go` if not present.
  At the end of `backend/internal/growth/handler/auth_handler.go`, append:
  ```go
  func generateRandomCode() string {
  	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
  	return fmt.Sprintf("%06d", n.Int64())
  }
  
  func (h *AuthHandler) ForgotPassword(c *gin.Context) {
  	var req struct {
  		Email string `json:"email" binding:"required,email"`
  	}
  	if err := c.ShouldBindJSON(&req); err != nil {
  		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
  		return
  	}
  	email := strings.ToLower(strings.TrimSpace(req.Email))
  
  	exists, err := h.emailExists(email)
  	if err != nil {
  		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
  		return
  	}
  	if !exists {
  		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40002, Message: "email not found", Data: struct{}{}})
  		return
  	}
  
  	code := generateRandomCode()
  	expiredAt := time.Now().Add(10 * time.Minute)
  	id := newID("prc")
  
  	_, err = h.db.Exec(
  		"INSERT INTO password_reset_codes (id, email, code, status, expired_at, created_at) VALUES (?, ?, ?, 'UNUSED', ?, ?)",
  		id, email, code, expiredAt, time.Now(),
  	)
  	if err != nil {
  		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
  		return
  	}
  
  	c.JSON(http.StatusOK, dto.APIResponse{
  		Code:    0,
  		Message: "success",
  		Data: gin.H{
  			"email": email,
  			"code":  code,
  		},
  	})
  }
  
  func (h *AuthHandler) ResetPassword(c *gin.Context) {
  	var req struct {
  		Email       string `json:"email" binding:"required,email"`
  		Code        string `json:"code" binding:"required"`
  		NewPassword string `json:"new_password" binding:"required,min=8"`
  	}
  	if err := c.ShouldBindJSON(&req); err != nil {
  		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
  		return
  	}
  	email := strings.ToLower(strings.TrimSpace(req.Email))
  	code := strings.TrimSpace(req.Code)
  
  	// Validate Code
  	var codeID string
  	var expiredAt time.Time
  	err := h.db.QueryRow(
  		"SELECT id, expired_at FROM password_reset_codes WHERE email = ? AND code = ? AND status = 'UNUSED' LIMIT 1",
  		email, code,
  	).Scan(&codeID, &expiredAt)
  	if err != nil {
  		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40002, Message: "invalid verification code", Data: struct{}{}})
  		return
  	}
  
  	if time.Now().After(expiredAt) {
  		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40003, Message: "verification code expired", Data: struct{}{}})
  		return
  	}
  
  	// Mark code as USED
  	_, err = h.db.Exec("UPDATE password_reset_codes SET status = 'USED' WHERE id = ?", codeID)
  	if err != nil {
  		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
  		return
  	}
  
  	// Hash password
  	passwordHash, err := bcryptHash(req.NewPassword)
  	if err != nil {
  		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
  		return
  	}
  
  	// Update password
  	_, err = h.db.Exec("UPDATE users SET password_hash = ?, updated_at = ? WHERE email = ?", passwordHash, time.Now(), email)
  	if err != nil {
  		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
  		return
  	}
  
  	c.JSON(http.StatusOK, dto.APIResponse{
  		Code:    0,
  		Message: "密码重置成功",
  		Data:    struct{}{},
  	})
  }
  ```

- [ ] **Step 3: Compile and verify backend**
  Run compilation check:
  ```bash
  go build -o /dev/null ./...
  ```
  Expected: Success without errors.

- [ ] **Step 4: Commit backend changes**
  ```bash
  git add backend/router/auth.go backend/internal/growth/handler/auth_handler.go
  git commit -m "feat(backend): implement forgot-password and reset-password API endpoints"
  ```

---

### Task 3: Client API Helpers Update

**Files:**
- Modify: `newclient/src/api/auth.js`

- [ ] **Step 1: Export forgotPassword and resetPassword functions**
  Modify `newclient/src/api/auth.js` to add helpers:
  ```javascript
  // Append to end of file:
  export function forgotPassword(email) {
    return http.post("/auth/forgot-password", { email });
  }
  
  export function resetPassword(payload) {
    return http.post("/auth/reset-password", payload);
  }
  ```

- [ ] **Step 2: Commit helper exports**
  ```bash
  git add newclient/src/api/auth.js
  git commit -m "feat(newclient): export forgotPassword and resetPassword API helpers"
  ```

---

### Task 4: PC Client Views & Router Update

**Files:**
- Modify: `newclient/src/apps/pc/views/user/Register.vue`
- Modify: `newclient/src/apps/pc/views/user/Login.vue`
- Modify: `newclient/src/apps/pc/router/index.js`
- Create: `newclient/src/apps/pc/views/user/ForgotPassword.vue`

- [ ] **Step 1: Require email and phone in PC Register.vue**
  Modify [Register.vue](file:///Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/user/Register.vue) to include the Email field, note, and form validations:
  ```html
  <!-- Target Content around line 12: -->
        <div class="auth-form">
          <div class="auth-field">
            <label>手机号</label>
            <input type="text" placeholder="请输入手机号" v-model="phone" />
          </div>
  
  <!-- Replacement Content: -->
        <div class="auth-form">
          <div class="auth-field">
            <label>手机号</label>
            <input type="text" placeholder="请输入手机号" v-model="phone" />
          </div>
          <div class="auth-field">
            <label>邮箱</label>
            <input type="email" placeholder="请输入电子邮箱" v-model="email" />
            <p class="field-hint">💡 邮箱为帐号找回，密码重置唯一凭证</p>
          </div>
  ```
  ```javascript
  // Target Content around line 38:
  const router = useRouter();
  const phone = ref("");
  const password = ref("");
  const confirmPwd = ref("");
  const loading = ref(false);
  const errorMsg = ref("");
  
  async function handleRegister() {
    if (!phone.value || !password.value) return;
    if (password.value !== confirmPwd.value) { errorMsg.value = "两次密码不一致"; return; }
    if (password.value.length < 6) { errorMsg.value = "密码至少6位"; return; }
    loading.value = true;
    errorMsg.value = "";
    try {
      const payload = { phone: phone.value, password: password.value };
      const result = await register(payload);
  
  // Replacement Content:
  const router = useRouter();
  const phone = ref("");
  const email = ref("");
  const password = ref("");
  const confirmPwd = ref("");
  const loading = ref(false);
  const errorMsg = ref("");
  
  async function handleRegister() {
    if (!phone.value || !email.value || !password.value) { errorMsg.value = "请填写完整"; return; }
    if (password.value !== confirmPwd.value) { errorMsg.value = "两次密码不一致"; return; }
    if (password.value.length < 6) { errorMsg.value = "密码至少6位"; return; }
    loading.value = true;
    errorMsg.value = "";
    try {
      const payload = { phone: phone.value, email: email.value, password: password.value };
      const result = await register(payload);
  
  // Also add .field-hint style in style section:
  // Add at the bottom of the style section:
  .field-hint { font-size: 11px; color: var(--accent-gold); margin-top: 4px; opacity: 0.9; }
  ```

- [ ] **Step 2: Add Forgot Password link in PC Login.vue**
  Modify [Login.vue](file:///Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/user/Login.vue) to add forget password link:
  ```html
  <!-- Target Content around line 16: -->
          <div class="auth-field">
            <label>密码</label>
            <input type="password" placeholder="请输入密码" v-model="password" @keyup.enter="handleLogin" />
          </div>
  
  <!-- Replacement Content: -->
          <div class="auth-field">
            <div class="field-header">
              <label>密码</label>
              <button class="link-btn-sm" @click="$router.push('/forgot-password')">忘记密码？</button>
            </div>
            <input type="password" placeholder="请输入密码" v-model="password" @keyup.enter="handleLogin" />
          </div>
  ```
  And add CSS to style section of `Login.vue`:
  ```css
  .field-header { display: flex; justify-content: space-between; align-items: center; }
  .link-btn-sm { background: none; border: none; color: var(--accent-gold); font-size: 12px; cursor: pointer; padding: 0; }
  .link-btn-sm:hover { text-decoration: underline; }
  ```

- [ ] **Step 3: Define /forgot-password route in PC router**
  Modify `newclient/src/apps/pc/router/index.js` to add the route:
  ```javascript
  // Target Content around line 42:
    { path: "/login", name: "login", component: () => import("../views/user/Login.vue"), meta: { title: "登录", layout: "blank" } },
    { path: "/register", name: "register", component: () => import("../views/user/Register.vue"), meta: { title: "注册", layout: "blank" } },
  
  // Replacement Content:
    { path: "/login", name: "login", component: () => import("../views/user/Login.vue"), meta: { title: "登录", layout: "blank" } },
    { path: "/register", name: "register", component: () => import("../views/user/Register.vue"), meta: { title: "注册", layout: "blank" } },
    { path: "/forgot-password", name: "forgot-password", component: () => import("../views/user/ForgotPassword.vue"), meta: { title: "找回密码", layout: "blank" } },
  ```

- [ ] **Step 4: Create ForgotPassword.vue for PC**
  Create `newclient/src/apps/pc/views/user/ForgotPassword.vue` containing:
  ```vue
  <template>
    <div class="auth-layout">
      <div class="auth-card glass">
        <div class="auth-brand">
          <div class="auth-brand-icon">
            <svg viewBox="0 0 32 32" fill="none" stroke="currentColor" stroke-width="2"><circle cx="16" cy="16" r="14"/><path d="M10 20l4-6 4 3 4-7" stroke-width="2.5"/></svg>
          </div>
          <h1>找回密码</h1>
          <p>请输入注册时绑定的邮箱以重置密码</p>
        </div>
        <div class="auth-form">
          <!-- Step 1: Input Email -->
          <div v-if="step === 1">
            <div class="auth-field">
              <label>注册邮箱</label>
              <input type="email" placeholder="请输入电子邮箱" v-model="email" />
            </div>
            <p v-if="errorMsg" class="auth-error">{{ errorMsg }}</p>
            <button class="auth-btn" :disabled="!email || loading" @click="handleSendCode">{{ loading ? '发送中...' : '发送验证码' }}</button>
          </div>

          <!-- Step 2: Verification and Reset -->
          <div v-else>
            <div class="auth-field">
              <label>验证码</label>
              <input type="text" placeholder="输入6位验证码" v-model="code" />
            </div>
            <div class="auth-field">
              <label>设置新密码</label>
              <input type="password" placeholder="至少8位密码" v-model="newPassword" />
            </div>
            <div class="auth-field">
              <label>确认新密码</label>
              <input type="password" placeholder="再次输入新密码" v-model="confirmPassword" />
            </div>
            <p v-if="errorMsg" class="auth-error">{{ errorMsg }}</p>
            <p v-if="successMsg" class="auth-success">{{ successMsg }}</p>
            <button class="auth-btn" :disabled="!code || !newPassword || !confirmPassword || loading" @click="handleResetPassword">{{ loading ? '提交中...' : '重置密码' }}</button>
          </div>

          <p class="auth-switch"><button class="link" @click="$router.push('/login')">返回登录</button></p>
        </div>
      </div>
    </div>
  </template>

  <script setup>
  import { ref } from "vue";
  import { useRouter } from "vue-router";
  import { forgotPassword, resetPassword } from "@/api/auth";

  const router = useRouter();
  const step = ref(1);
  const email = ref("");
  const code = ref("");
  const newPassword = ref("");
  const confirmPassword = ref("");
  const loading = ref(false);
  const errorMsg = ref("");
  const successMsg = ref("");

  async function handleSendCode() {
    if (!email.value) return;
    loading.value = true;
    errorMsg.value = "";
    try {
      const res = await forgotPassword(email.value);
      step.value = 2;
      // Dev helper: auto-fill code if present in response
      if (res?.code) {
        code.value = res.code;
      }
    } catch (e) {
      errorMsg.value = e?.message || "发送验证码失败";
    } finally {
      loading.value = false;
    }
  }

  async function handleResetPassword() {
    if (!code.value || !newPassword.value) return;
    if (newPassword.value !== confirmPassword.value) { errorMsg.value = "两次密码不一致"; return; }
    if (newPassword.value.length < 8) { errorMsg.value = "新密码至少8位"; return; }
    loading.value = true;
    errorMsg.value = "";
    successMsg.value = "";
    try {
      await resetPassword({
        email: email.value,
        code: code.value,
        new_password: newPassword.value
      });
      successMsg.value = "密码重置成功，正在返回登录页...";
      setTimeout(() => {
        router.push("/login");
      }, 2000);
    } catch (e) {
      errorMsg.value = e?.message || "密码重置失败";
    } finally {
      loading.value = false;
    }
  }
  </script>

  <style scoped>
  .auth-card { width: 400px; max-width: 90vw; padding: 40px 32px; border-radius: var(--radius-xl); text-align: center; }
  .auth-brand { margin-bottom: 32px; }
  .auth-brand-icon { width: 48px; height: 48px; color: var(--accent-gold); margin: 0 auto 12px; filter: drop-shadow(0 0 8px rgba(240,185,11,.3)); }
  .auth-brand h1 { font-size: 24px; font-weight: 800; background: linear-gradient(135deg,var(--accent-gold),#fff); -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text; }
  .auth-brand p { font-size: 13px; color: var(--text-secondary); }
  .auth-form { text-align: left; }
  .auth-field { margin-bottom: 16px; }
  .auth-field label { display: block; font-size: 13px; color: var(--text-secondary); margin-bottom: 6px; }
  .auth-field input { width: 100%; padding: 12px 14px; border-radius: var(--radius-sm); background: rgba(255,255,255,.04); border: 1px solid var(--border); color: var(--text-primary); font-size: 14px; }
  .auth-field input:focus { border-color: var(--accent-gold); }
  .auth-error { color: var(--negative); font-size: 12px; margin-bottom: 8px; }
  .auth-success { color: var(--positive); font-size: 12px; margin-bottom: 8px; }
  .auth-btn { width: 100%; padding: 14px; border-radius: var(--radius-full); background: linear-gradient(135deg,var(--accent-gold),var(--accent-gold-dim)); color: #000; font-size: 15px; font-weight: 600; margin: 8px 0 16px; }
  .auth-btn:disabled { opacity: .5; }
  .auth-switch { text-align: center; font-size: 13px; color: var(--text-secondary); }
  .link { color: var(--accent-gold); font-weight: 600; background: none; border: none; cursor: pointer; }
  </style>
  ```

- [ ] **Step 5: Commit PC client changes**
  ```bash
  git add newclient/src/apps/pc/views/user/Register.vue newclient/src/apps/pc/views/user/Login.vue newclient/src/apps/pc/router/index.js newclient/src/apps/pc/views/user/ForgotPassword.vue
  git commit -m "feat(newclient): implement forgot password entry, page, and updated registration for PC client"
  ```

---

### Task 5: H5 Client Views & Router Update

**Files:**
- Modify: `newclient/src/apps/h5/views/user/Register.vue`
- Modify: `newclient/src/apps/h5/views/user/Login.vue`
- Modify: `newclient/src/apps/h5/router/index.js`
- Create: `newclient/src/apps/h5/views/user/ForgotPassword.vue`

- [ ] **Step 1: Require email and phone in H5 Register.vue**
  Modify [Register.vue](file:///Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/user/Register.vue) to add Email input and hints:
  ```html
  <!-- Target Content around line 8: -->
        <div class="h5-auth-field"><input type="text" placeholder="手机号" v-model="phone" /></div>
        <div class="h5-auth-field"><input type="password" placeholder="密码" v-model="password" /></div>
  
  <!-- Replacement Content: -->
        <div class="h5-auth-field"><input type="text" placeholder="手机号" v-model="phone" /></div>
        <div class="h5-auth-field">
          <input type="email" placeholder="电子邮箱" v-model="email" />
          <p class="field-hint">💡 邮箱为帐号找回，密码重置唯一凭证</p>
        </div>
        <div class="h5-auth-field"><input type="password" placeholder="密码" v-model="password" /></div>
  ```
  ```javascript
  // Target Content around line 28:
  async function handleRegister() {
    if (!phone.value || !password.value) return;
    if (password.value !== confirmPwd.value) { errorMsg.value = "两次密码不一致"; return; }
    if (password.value.length < 6) { errorMsg.value = "密码至少6位"; return; }
    loading.value = true;
    errorMsg.value = "";
    try {
      const result = await register({ phone: phone.value, password: password.value });
  
  // Replacement Content:
  const email = ref("");
  async function handleRegister() {
    if (!phone.value || !email.value || !password.value) { errorMsg.value = "请填写完整"; return; }
    if (password.value !== confirmPwd.value) { errorMsg.value = "两次密码不一致"; return; }
    if (password.value.length < 6) { errorMsg.value = "密码至少6位"; return; }
    loading.value = true;
    errorMsg.value = "";
    try {
      const result = await register({ phone: phone.value, email: email.value, password: password.value });
  ```
  And add CSS to style section of H5 `Register.vue`:
  ```css
  .field-hint { font-size: 11px; color: var(--accent-gold); margin-top: 4px; opacity: 0.9; text-align: left; padding-left: 4px; }
  ```

- [ ] **Step 2: Add Forgot Password link in H5 Login.vue**
  Modify [Login.vue](file:///Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/user/Login.vue) to add forget password link:
  ```html
  <!-- Target Content around line 8: -->
        <div class="h5-auth-field"><input type="text" placeholder="手机号/邮箱" v-model="account" /></div>
        <div class="h5-auth-field"><input type="password" placeholder="密码" v-model="password" @keyup.enter="handleLogin" /></div>
  
  <!-- Replacement Content: -->
        <div class="h5-auth-field"><input type="text" placeholder="手机号/邮箱" v-model="account" /></div>
        <div class="h5-auth-field">
          <input type="password" placeholder="密码" v-model="password" @keyup.enter="handleLogin" />
          <div class="forgot-link-container">
            <button class="link-btn-sm" @click="$router.push('/profile/forgot-password')">忘记密码？</button>
          </div>
        </div>
  ```
  And add CSS to style section of H5 `Login.vue`:
  ```css
  .forgot-link-container { display: flex; justify-content: flex-end; margin-top: 6px; }
  .link-btn-sm { background: none; border: none; color: var(--accent-gold); font-size: 12px; cursor: pointer; padding: 0; }
  .link-btn-sm:active { opacity: 0.7; }
  ```

- [ ] **Step 3: Define /profile/forgot-password route in H5 router**
  Modify `newclient/src/apps/h5/router/index.js` to add the route:
  ```javascript
  // Target Content around line 18:
    { path: "/login", name: "h5-login", component: () => import("../views/user/Login.vue"), meta: { layout: "blank" } },
    { path: "/register", name: "h5-register", component: () => import("../views/user/Register.vue"), meta: { layout: "blank" } },
  
  // Replacement Content:
    { path: "/login", name: "h5-login", component: () => import("../views/user/Login.vue"), meta: { layout: "blank" } },
    { path: "/register", name: "h5-register", component: () => import("../views/user/Register.vue"), meta: { layout: "blank" } },
    { path: "/profile/forgot-password", name: "h5-forgot-password", component: () => import("../views/user/ForgotPassword.vue"), meta: { layout: "blank" } },
  ```

- [ ] **Step 4: Create ForgotPassword.vue for H5**
  Create `newclient/src/apps/h5/views/user/ForgotPassword.vue` containing:
  ```vue
  <template>
    <div class="auth-layout">
      <div class="h5-auth-card">
        <div class="h5-auth-brand">
          <svg viewBox="0 0 32 32" fill="none" stroke="currentColor" stroke-width="2" width="36" height="36" style="color:var(--accent-gold)"><circle cx="16" cy="16" r="14"/><path d="M10 20l4-6 4 3 4-7" stroke-width="2.5"/></svg>
          <h1>找回密码</h1>
          <p style="font-size:12px;color:var(--text-secondary);margin-top:6px;">输入注册时绑定的邮箱以重置密码</p>
        </div>

        <!-- Step 1: Input Email -->
        <div v-if="step === 1" class="form-container">
          <div class="h5-auth-field"><input type="email" placeholder="注册邮箱" v-model="email" /></div>
          <p v-if="errorMsg" class="auth-error">{{ errorMsg }}</p>
          <button class="h5-auth-btn" :disabled="!email || loading" @click="handleSendCode">{{ loading ? '发送中...' : '发送验证码' }}</button>
        </div>

        <!-- Step 2: Verification and Reset -->
        <div v-else class="form-container">
          <div class="h5-auth-field"><input type="text" placeholder="输入6位验证码" v-model="code" /></div>
          <div class="h5-auth-field"><input type="password" placeholder="设置新密码" v-model="newPassword" /></div>
          <div class="h5-auth-field"><input type="password" placeholder="确认新密码" v-model="confirmPassword" /></div>
          <p v-if="errorMsg" class="auth-error">{{ errorMsg }}</p>
          <p v-if="successMsg" class="auth-success">{{ successMsg }}</p>
          <button class="h5-auth-btn" :disabled="!code || !newPassword || !confirmPassword || loading" @click="handleResetPassword">{{ loading ? '提交中...' : '重置密码' }}</button>
        </div>

        <p class="h5-auth-link"><button class="link" @click="$router.push('/login')">返回登录</button></p>
      </div>
    </div>
  </template>

  <script setup>
  import { ref } from "vue";
  import { useRouter } from "vue-router";
  import { forgotPassword, resetPassword } from "@/api/auth";

  const router = useRouter();
  const step = ref(1);
  const email = ref("");
  const code = ref("");
  const newPassword = ref("");
  const confirmPassword = ref("");
  const loading = ref(false);
  const errorMsg = ref("");
  const successMsg = ref("");

  async function handleSendCode() {
    if (!email.value) return;
    loading.value = true;
    errorMsg.value = "";
    try {
      const res = await forgotPassword(email.value);
      step.value = 2;
      if (res?.code) {
        code.value = res.code;
      }
    } catch (e) {
      errorMsg.value = e?.message || "发送失败";
    } finally {
      loading.value = false;
    }
  }

  async function handleResetPassword() {
    if (!code.value || !newPassword.value) return;
    if (newPassword.value !== confirmPassword.value) { errorMsg.value = "两次密码不一致"; return; }
    if (newPassword.value.length < 8) { errorMsg.value = "新密码至少8位"; return; }
    loading.value = true;
    errorMsg.value = "";
    successMsg.value = "";
    try {
      await resetPassword({
        email: email.value,
        code: code.value,
        new_password: newPassword.value
      });
      successMsg.value = "重置成功，正在返回登录页...";
      setTimeout(() => {
        router.push("/login");
      }, 2000);
    } catch (e) {
      errorMsg.value = e?.message || "重置失败";
    } finally {
      loading.value = false;
    }
  }
  </script>

  <style scoped>
  .h5-auth-card { width: 100%; max-width: 360px; margin: 0 auto; display: grid; gap: 14px; }
  .h5-auth-brand { text-align: center; margin-bottom: 20px; }
  .h5-auth-brand h1 { font-size: 24px; font-weight: 800; background: linear-gradient(135deg,var(--accent-gold),#fff); -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text; }
  .form-container { display: grid; gap: 14px; }
  .h5-auth-field input { width: 100%; padding: 14px; border-radius: var(--radius-sm); background: rgba(255,255,255,.04); border: 1px solid var(--border); color: var(--text-primary); font-size: 14px; }
  .h5-auth-field input:focus { border-color: var(--accent-gold); }
  .auth-error { color: var(--negative); font-size: 12px; text-align: center; }
  .auth-success { color: var(--positive); font-size: 12px; text-align: center; }
  .h5-auth-btn { width: 100%; padding: 14px; border-radius: var(--radius-full); background: linear-gradient(135deg,var(--accent-gold),var(--accent-gold-dim)); color: #000; font-size: 15px; font-weight: 600; }
  .h5-auth-btn:disabled { opacity: .5; }
  .h5-auth-link { text-align: center; font-size: 13px; color: var(--text-secondary); }
  .link { color: var(--accent-gold); font-weight: 600; background: none; border: none; cursor: pointer; }
  </style>
  ```

- [ ] **Step 5: Commit H5 client changes**
  ```bash
  git add newclient/src/apps/h5/views/user/Register.vue newclient/src/apps/h5/views/user/Login.vue newclient/src/apps/h5/router/index.js newclient/src/apps/h5/views/user/ForgotPassword.vue
  git commit -m "feat(newclient): implement forgot password entry, page, and updated registration for H5 client"
  ```

---

### Task 6: Full Client Build Validation

**Files:**
- Test: `newclient/package.json`

- [ ] **Step 1: Run production build for PC and H5**
  Verify that all Vue code bundles correctly:
  ```bash
  npm run build
  ```
  Expected: Command finishes successfully with output showing build results for both PC and H5, and exit code `0`.
