<script setup>
import { computed, onMounted, reactive, ref } from "vue";
import { ElMessage } from "element-plus";
import {
  assignAdminAccountRoles,
  createAccessRole,
  createAdminAccount,
  listAccessPermissions,
  listAccessRoles,
  listAdminAccounts,
  resetAdminAccountPassword,
  updateAccessRole,
  updateAccessRoleStatus,
  updateAdminAccountStatus
} from "../api/admin";

const activeTab = ref("admins");
const permissionsLoading = ref(false);
const permissionRows = ref([]);

const roleLoading = ref(false);
const roleFilters = reactive({
  keyword: "",
  status: ""
});
const rolePage = reactive({
  page: 1,
  page_size: 20,
  total: 0
});
const roleRows = ref([]);

const adminLoading = ref(false);
const adminFilters = reactive({
  keyword: "",
  status: ""
});
const adminPage = reactive({
  page: 1,
  page_size: 20,
  total: 0
});
const adminRows = ref([]);

const roleDialogVisible = ref(false);
const roleSubmitting = ref(false);
const roleForm = reactive({
  id: "",
  role_key: "",
  role_name: "",
  description: "",
  status: "ACTIVE",
  permission_codes: []
});

const createAdminDialogVisible = ref(false);
const createAdminSubmitting = ref(false);
const createAdminForm = reactive({
  phone: "",
  password: "",
  email: "",
  status: "ACTIVE",
  role_ids: []
});

const assignRoleDialogVisible = ref(false);
const assignRoleSubmitting = ref(false);
const assignRoleForm = reactive({
  user_id: "",
  role_ids: []
});

const resetPasswordDialogVisible = ref(false);
const resetPasswordSubmitting = ref(false);
const resetPasswordForm = reactive({
  user_id: "",
  password: ""
});

const moduleLabels = {
  DASHBOARD: "仪表盘",
  USERS: "用户",
  MEMBERSHIP: "会员",
  NEWS: "新闻",
  MARKET: "策略",
  REVIEW: "审核",
  RISK: "风控",
  AUDIT: "审计",
  AUTH_SECURITY: "安全",
  SYSTEM_CONFIG: "系统配置",
  SYSTEM_JOB: "任务",
  DATA_SOURCE: "数据源",
  WORKFLOW: "流程消息",
  GROWTH: "增长",
  PAYMENT: "支付",
  REWARD_WALLET: "奖励钱包",
  ACCESS: "管理员权限"
};

const permissionLabels = {
  "dashboard.view": "查看仪表盘",
  "users.view": "查看用户",
  "users.edit": "编辑用户",
  "membership.view": "查看会员模块",
  "membership.edit": "编辑会员模块",
  "news.view": "查看新闻",
  "news.edit": "编辑新闻",
  "market.view": "查看策略",
  "market.edit": "编辑策略",
  "review.view": "查看审核中心",
  "review.edit": "处理审核任务",
  "risk.view": "查看风控",
  "risk.edit": "编辑风控规则",
  "audit.view": "查看操作日志",
  "auth_security.view": "查看安全中心",
  "auth_security.edit": "编辑安全策略",
  "system_config.view": "查看系统配置",
  "system_config.edit": "编辑系统配置",
  "system_job.view": "查看任务中心",
  "system_job.edit": "管理任务中心",
  "data_source.view": "查看数据源",
  "data_source.edit": "管理数据源",
  "workflow.view": "查看流程消息",
  "workflow.edit": "处理流程消息",
  "growth.view": "查看增长模块",
  "growth.edit": "管理增长模块",
  "payment.view": "查看对账",
  "payment.edit": "执行对账操作",
  "reward_wallet.view": "查看奖励钱包",
  "reward_wallet.edit": "审核提现",
  "access.view": "查看管理员与权限",
  "access.edit": "管理管理员与权限"
};

const actionLabels = {
  VIEW: "查看",
  EDIT: "编辑"
};

const groupedPermissions = computed(() => {
  const groups = {};
  permissionRows.value.forEach((item) => {
    const key = item.module || "OTHER";
    if (!groups[key]) {
      groups[key] = [];
    }
    groups[key].push(item);
  });
  return Object.entries(groups)
    .map(([module, items]) => ({
      module,
      label: moduleLabels[module] || module,
      items: [...items].sort((a, b) => permissionDisplayName(a).localeCompare(permissionDisplayName(b), "zh-Hans-CN"))
    }))
    .sort((a, b) => a.module.localeCompare(b.module));
});

const roleOptions = computed(() =>
  roleRows.value
    .filter((item) => item.status === "ACTIVE")
    .map((item) => ({
      label: `${item.role_name} (${item.role_key})`,
      value: item.id
    }))
);

const allPermissionCodes = computed(() => permissionRows.value.map((item) => item.code));
const allPermissionsChecked = computed(
  () => allPermissionCodes.value.length > 0 && allPermissionCodes.value.every((code) => roleForm.permission_codes.includes(code))
);
const allPermissionsIndeterminate = computed(
  () => roleForm.permission_codes.length > 0 && !allPermissionsChecked.value
);

function permissionDisplayName(item) {
  const mapped = permissionLabels[item.code];
  if (mapped) {
    return mapped;
  }
  if (item.name) {
    return item.name;
  }
  const moduleLabel = moduleLabels[item.module] || item.module || "其他";
  const actionLabel = actionLabels[item.action] || item.action || "操作";
  return `${actionLabel}${moduleLabel}`;
}

function normalizePermissionCodes(codes) {
  return [...new Set((codes || []).filter(Boolean))].sort();
}

function hasRolePermission(code) {
  return roleForm.permission_codes.includes(code);
}

function toggleRolePermission(code, checked) {
  const next = new Set(roleForm.permission_codes);
  if (checked) {
    next.add(code);
  } else {
    next.delete(code);
  }
  roleForm.permission_codes = normalizePermissionCodes([...next]);
}

function getGroupCodes(group) {
  return (group?.items || []).map((item) => item.code);
}

function selectedCountInGroup(group) {
  const codes = getGroupCodes(group);
  return codes.filter((code) => roleForm.permission_codes.includes(code)).length;
}

function isGroupChecked(group) {
  const codes = getGroupCodes(group);
  return codes.length > 0 && codes.every((code) => roleForm.permission_codes.includes(code));
}

function isGroupIndeterminate(group) {
  const codes = getGroupCodes(group);
  const selected = selectedCountInGroup(group);
  return selected > 0 && selected < codes.length;
}

function toggleGroupPermissions(group, checked) {
  const codes = getGroupCodes(group);
  const next = new Set(roleForm.permission_codes);
  codes.forEach((code) => {
    if (checked) {
      next.add(code);
    } else {
      next.delete(code);
    }
  });
  roleForm.permission_codes = normalizePermissionCodes([...next]);
}

function toggleAllPermissions(checked) {
  roleForm.permission_codes = checked ? normalizePermissionCodes(allPermissionCodes.value) : [];
}

async function loadPermissions() {
  permissionsLoading.value = true;
  try {
    const data = await listAccessPermissions({
      page: 1,
      page_size: 200
    });
    permissionRows.value = data.items || [];
  } finally {
    permissionsLoading.value = false;
  }
}

async function loadRoles(resetPage = false) {
  if (resetPage) {
    rolePage.page = 1;
  }
  roleLoading.value = true;
  try {
    const data = await listAccessRoles({
      page: rolePage.page,
      page_size: rolePage.page_size,
      keyword: roleFilters.keyword.trim(),
      status: roleFilters.status
    });
    roleRows.value = data.items || [];
    rolePage.total = data.total || 0;
  } finally {
    roleLoading.value = false;
  }
}

async function loadAdminAccounts(resetPage = false) {
  if (resetPage) {
    adminPage.page = 1;
  }
  adminLoading.value = true;
  try {
    const data = await listAdminAccounts({
      page: adminPage.page,
      page_size: adminPage.page_size,
      keyword: adminFilters.keyword.trim(),
      status: adminFilters.status
    });
    adminRows.value = data.items || [];
    adminPage.total = data.total || 0;
  } finally {
    adminLoading.value = false;
  }
}

function openCreateRoleDialog() {
  roleForm.id = "";
  roleForm.role_key = "";
  roleForm.role_name = "";
  roleForm.description = "";
  roleForm.status = "ACTIVE";
  roleForm.permission_codes = [];
  roleDialogVisible.value = true;
}

function openEditRoleDialog(role) {
  roleForm.id = role.id;
  roleForm.role_key = role.role_key;
  roleForm.role_name = role.role_name;
  roleForm.description = role.description || "";
  roleForm.status = role.status;
  roleForm.permission_codes = normalizePermissionCodes(role.permission_codes || []);
  roleDialogVisible.value = true;
}

async function submitRoleDialog() {
  if (!roleForm.role_name.trim()) {
    ElMessage.error("角色名称不能为空");
    return;
  }
  if (!roleForm.id && !roleForm.role_key.trim()) {
    ElMessage.error("角色标识不能为空");
    return;
  }
  if (!roleForm.permission_codes.length) {
    ElMessage.error("至少选择一个权限");
    return;
  }
  roleSubmitting.value = true;
  try {
    const payload = {
      role_key: roleForm.role_key.trim().toUpperCase(),
      role_name: roleForm.role_name.trim(),
      description: roleForm.description.trim(),
      status: roleForm.status,
      permission_codes: roleForm.permission_codes
    };
    if (roleForm.id) {
      await updateAccessRole(roleForm.id, payload);
      ElMessage.success("角色已更新");
    } else {
      await createAccessRole(payload);
      ElMessage.success("角色已创建");
    }
    roleDialogVisible.value = false;
    await loadRoles();
  } finally {
    roleSubmitting.value = false;
  }
}

async function toggleRoleStatus(role) {
  const nextStatus = role.status === "ACTIVE" ? "DISABLED" : "ACTIVE";
  await updateAccessRoleStatus(role.id, nextStatus);
  ElMessage.success(nextStatus === "ACTIVE" ? "角色已启用" : "角色已禁用");
  await loadRoles();
}

function openCreateAdminDialog() {
  createAdminForm.phone = "";
  createAdminForm.password = "";
  createAdminForm.email = "";
  createAdminForm.status = "ACTIVE";
  createAdminForm.role_ids = [];
  createAdminDialogVisible.value = true;
}

async function submitCreateAdminDialog() {
  if (!createAdminForm.phone.trim()) {
    ElMessage.error("手机号不能为空");
    return;
  }
  if (!createAdminForm.password) {
    ElMessage.error("密码不能为空");
    return;
  }
  if (createAdminForm.password.length < 8) {
    ElMessage.error("密码至少 8 位");
    return;
  }
  if (!createAdminForm.role_ids.length) {
    ElMessage.error("至少分配一个角色");
    return;
  }
  createAdminSubmitting.value = true;
  try {
    await createAdminAccount({
      phone: createAdminForm.phone.trim(),
      password: createAdminForm.password,
      email: createAdminForm.email.trim(),
      status: createAdminForm.status,
      role_ids: createAdminForm.role_ids
    });
    ElMessage.success("管理员已创建");
    createAdminDialogVisible.value = false;
    await loadAdminAccounts();
  } finally {
    createAdminSubmitting.value = false;
  }
}

function openAssignRoleDialog(row) {
  assignRoleForm.user_id = row.id;
  assignRoleForm.role_ids = [...(row.role_ids || [])];
  assignRoleDialogVisible.value = true;
}

async function submitAssignRoleDialog() {
  if (!assignRoleForm.role_ids.length) {
    ElMessage.error("至少选择一个角色");
    return;
  }
  assignRoleSubmitting.value = true;
  try {
    await assignAdminAccountRoles(assignRoleForm.user_id, assignRoleForm.role_ids);
    ElMessage.success("角色分配已更新");
    assignRoleDialogVisible.value = false;
    await loadAdminAccounts();
  } finally {
    assignRoleSubmitting.value = false;
  }
}

function openResetPasswordDialog(row) {
  resetPasswordForm.user_id = row.id;
  resetPasswordForm.password = "";
  resetPasswordDialogVisible.value = true;
}

async function submitResetPasswordDialog() {
  if (!resetPasswordForm.password || resetPasswordForm.password.length < 8) {
    ElMessage.error("新密码至少 8 位");
    return;
  }
  resetPasswordSubmitting.value = true;
  try {
    await resetAdminAccountPassword(resetPasswordForm.user_id, resetPasswordForm.password);
    ElMessage.success("密码已重置");
    resetPasswordDialogVisible.value = false;
  } finally {
    resetPasswordSubmitting.value = false;
  }
}

async function toggleAdminStatus(row) {
  const nextStatus = row.status === "ACTIVE" ? "DISABLED" : "ACTIVE";
  await updateAdminAccountStatus(row.id, nextStatus);
  ElMessage.success(nextStatus === "ACTIVE" ? "账号已启用" : "账号已禁用");
  await loadAdminAccounts();
}

onMounted(async () => {
  await Promise.all([loadPermissions(), loadRoles(), loadAdminAccounts()]);
});
</script>

<template>
  <div class="page">
    <div class="head">
      <div>
        <h2>管理员与权限</h2>
        <p>管理员账号、角色模板、权限矩阵统一在这里维护。</p>
      </div>
    </div>

    <el-tabs v-model="activeTab" class="tabs">
      <el-tab-pane label="管理员账号" name="admins">
        <el-card shadow="never" class="block">
          <div class="toolbar">
            <div class="filters">
              <el-input
                v-model="adminFilters.keyword"
                clearable
                placeholder="搜索 ID / 手机号 / 邮箱"
                style="width: 260px"
                @keyup.enter="loadAdminAccounts(true)"
              />
              <el-select v-model="adminFilters.status" clearable placeholder="状态" style="width: 140px">
                <el-option label="ACTIVE" value="ACTIVE" />
                <el-option label="DISABLED" value="DISABLED" />
              </el-select>
              <el-button type="primary" @click="loadAdminAccounts(true)">查询</el-button>
            </div>
            <el-button type="primary" @click="openCreateAdminDialog">新增管理员</el-button>
          </div>

          <el-table :data="adminRows" border stripe v-loading="adminLoading">
            <el-table-column prop="id" label="管理员ID" min-width="210" />
            <el-table-column prop="phone" label="手机号" min-width="140" />
            <el-table-column prop="email" label="邮箱" min-width="200" />
            <el-table-column label="角色" min-width="260">
              <template #default="{ row }">
                <el-tag v-for="roleName in row.role_names || []" :key="`${row.id}-${roleName}`" size="small" class="tag">
                  {{ roleName }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="110" />
            <el-table-column prop="last_login" label="最近登录" min-width="180" />
            <el-table-column prop="created_at" label="创建时间" min-width="180" />
            <el-table-column label="操作" width="260" fixed="right">
              <template #default="{ row }">
                <el-space>
                  <el-button link type="primary" @click="openAssignRoleDialog(row)">分配角色</el-button>
                  <el-button link type="warning" @click="openResetPasswordDialog(row)">重置密码</el-button>
                  <el-button link type="danger" @click="toggleAdminStatus(row)">
                    {{ row.status === "ACTIVE" ? "禁用" : "启用" }}
                  </el-button>
                </el-space>
              </template>
            </el-table-column>
          </el-table>

          <el-pagination
            class="pagination"
            background
            layout="total, prev, pager, next, sizes"
            v-model:current-page="adminPage.page"
            v-model:page-size="adminPage.page_size"
            :total="adminPage.total"
            @current-change="loadAdminAccounts()"
            @size-change="loadAdminAccounts(true)"
          />
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="角色管理" name="roles">
        <el-card shadow="never" class="block">
          <div class="toolbar">
            <div class="filters">
              <el-input
                v-model="roleFilters.keyword"
                clearable
                placeholder="搜索角色标识 / 名称"
                style="width: 260px"
                @keyup.enter="loadRoles(true)"
              />
              <el-select v-model="roleFilters.status" clearable placeholder="状态" style="width: 140px">
                <el-option label="ACTIVE" value="ACTIVE" />
                <el-option label="DISABLED" value="DISABLED" />
              </el-select>
              <el-button type="primary" @click="loadRoles(true)">查询</el-button>
            </div>
            <el-button type="primary" @click="openCreateRoleDialog">新增角色</el-button>
          </div>

          <el-table :data="roleRows" border stripe v-loading="roleLoading">
            <el-table-column prop="role_key" label="角色标识" min-width="140" />
            <el-table-column prop="role_name" label="角色名称" min-width="160" />
            <el-table-column prop="description" label="说明" min-width="220" />
            <el-table-column label="权限数" width="90">
              <template #default="{ row }">{{ (row.permission_codes || []).length }}</template>
            </el-table-column>
            <el-table-column prop="user_count" label="账号数" width="90" />
            <el-table-column prop="status" label="状态" width="100" />
            <el-table-column label="内置" width="80">
              <template #default="{ row }">{{ row.built_in ? "是" : "否" }}</template>
            </el-table-column>
            <el-table-column prop="updated_at" label="更新时间" min-width="180" />
            <el-table-column label="操作" width="210" fixed="right">
              <template #default="{ row }">
                <el-space>
                  <el-button link type="primary" @click="openEditRoleDialog(row)">编辑</el-button>
                  <el-button link type="danger" @click="toggleRoleStatus(row)">
                    {{ row.status === "ACTIVE" ? "禁用" : "启用" }}
                  </el-button>
                </el-space>
              </template>
            </el-table-column>
          </el-table>

          <el-pagination
            class="pagination"
            background
            layout="total, prev, pager, next, sizes"
            v-model:current-page="rolePage.page"
            v-model:page-size="rolePage.page_size"
            :total="rolePage.total"
            @current-change="loadRoles()"
            @size-change="loadRoles(true)"
          />
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="权限字典" name="permissions">
        <el-card shadow="never" class="block">
          <div class="permission-list" v-loading="permissionsLoading">
            <el-card v-for="group in groupedPermissions" :key="group.module" shadow="never" class="permission-group">
              <template #header>
                <div class="permission-title">{{ group.label }}</div>
              </template>
              <div class="permission-dict-items">
                <div v-for="item in group.items" :key="item.code" class="permission-dict-row">
                  <el-tag class="permission-name-tag" effect="plain">{{ permissionDisplayName(item) }}</el-tag>
                  <span class="permission-code-text">{{ item.code }}</span>
                </div>
              </div>
            </el-card>
          </div>
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="roleDialogVisible" :title="roleForm.id ? '编辑角色' : '新建角色'" width="720px">
      <el-form label-position="top">
        <el-form-item label="角色标识">
          <el-input v-model="roleForm.role_key" :disabled="!!roleForm.id" placeholder="例如 CONTENT_EDITOR" />
        </el-form-item>
        <el-form-item label="角色名称">
          <el-input v-model="roleForm.role_name" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="roleForm.status" style="width: 160px">
            <el-option label="ACTIVE" value="ACTIVE" />
            <el-option label="DISABLED" value="DISABLED" />
          </el-select>
        </el-form-item>
        <el-form-item label="说明">
          <el-input v-model="roleForm.description" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="权限">
          <div class="permission-picker">
            <div class="permission-picker-head">
              <el-checkbox
                :model-value="allPermissionsChecked"
                :indeterminate="allPermissionsIndeterminate"
                @change="toggleAllPermissions"
              >
                全选权限
              </el-checkbox>
              <span class="permission-counter">
                已选择 {{ roleForm.permission_codes.length }} / {{ allPermissionCodes.length }}
              </span>
            </div>
            <el-scrollbar class="permission-picker-scroll" max-height="340px">
              <div v-for="group in groupedPermissions" :key="group.module" class="permission-module-block">
                <div class="permission-module-head">
                  <el-checkbox
                    :model-value="isGroupChecked(group)"
                    :indeterminate="isGroupIndeterminate(group)"
                    @change="(checked) => toggleGroupPermissions(group, checked)"
                  >
                    {{ group.label }}
                  </el-checkbox>
                  <span class="permission-counter small">
                    {{ selectedCountInGroup(group) }} / {{ group.items.length }}
                  </span>
                </div>
                <div class="permission-module-items">
                  <el-checkbox
                    v-for="item in group.items"
                    :key="item.code"
                    :model-value="hasRolePermission(item.code)"
                    class="permission-checkbox"
                    @change="(checked) => toggleRolePermission(item.code, checked)"
                  >
                    <span class="permission-checkbox-label">{{ permissionDisplayName(item) }}</span>
                    <span class="permission-checkbox-code">{{ item.code }}</span>
                  </el-checkbox>
                </div>
              </div>
            </el-scrollbar>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="roleDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="roleSubmitting" @click="submitRoleDialog">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="createAdminDialogVisible" title="新增管理员" width="560px">
      <el-form label-position="top">
        <el-form-item label="手机号">
          <el-input v-model="createAdminForm.phone" />
        </el-form-item>
        <el-form-item label="初始密码">
          <el-input v-model="createAdminForm.password" type="password" show-password />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="createAdminForm.email" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="createAdminForm.status" style="width: 160px">
            <el-option label="ACTIVE" value="ACTIVE" />
            <el-option label="DISABLED" value="DISABLED" />
          </el-select>
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="createAdminForm.role_ids" multiple filterable style="width: 100%">
            <el-option v-for="item in roleOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createAdminDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="createAdminSubmitting" @click="submitCreateAdminDialog">创建</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="assignRoleDialogVisible" title="分配角色" width="560px">
      <el-form label-position="top">
        <el-form-item label="管理员 ID">
          <el-input :model-value="assignRoleForm.user_id" disabled />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="assignRoleForm.role_ids" multiple filterable style="width: 100%">
            <el-option v-for="item in roleOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="assignRoleDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="assignRoleSubmitting" @click="submitAssignRoleDialog">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="resetPasswordDialogVisible" title="重置密码" width="420px">
      <el-form label-position="top">
        <el-form-item label="管理员 ID">
          <el-input :model-value="resetPasswordForm.user_id" disabled />
        </el-form-item>
        <el-form-item label="新密码">
          <el-input v-model="resetPasswordForm.password" type="password" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="resetPasswordDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="resetPasswordSubmitting" @click="submitResetPasswordDialog">重置</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page {
  padding: 20px;
}

.head {
  margin-bottom: 16px;
}

.head h2 {
  margin: 0;
  font-size: 22px;
}

.head p {
  margin: 6px 0 0;
  color: #64748b;
}

.tabs :deep(.el-tabs__content) {
  padding-top: 8px;
}

.block {
  border: 1px solid #e5e7eb;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.filters {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.tag {
  margin-right: 6px;
  margin-bottom: 6px;
}

.pagination {
  margin-top: 14px;
  justify-content: flex-end;
}

.permission-list {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 12px;
}

.permission-group {
  min-height: 120px;
}

.permission-title {
  font-weight: 600;
}

.permission-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.permission-dict-items {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.permission-dict-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  padding: 6px 0;
  border-bottom: 1px dashed #e5e7eb;
}

.permission-name-tag {
  max-width: 65%;
}

.permission-code-text {
  color: #64748b;
  font-family: Menlo, Monaco, Consolas, "Courier New", monospace;
  font-size: 12px;
}

.permission-picker {
  width: 100%;
  border: 1px solid #dcdfe6;
  border-radius: 8px;
  background: #fafafa;
}

.permission-picker-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  border-bottom: 1px solid #ebeef5;
  background: #fff;
}

.permission-counter {
  color: #64748b;
  font-size: 12px;
}

.permission-counter.small {
  font-size: 11px;
}

.permission-picker-scroll {
  padding: 10px 12px;
}

.permission-module-block + .permission-module-block {
  margin-top: 10px;
}

.permission-module-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.permission-module-items {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 8px 10px;
  padding: 8px 10px;
  border: 1px solid #ebeef5;
  border-radius: 6px;
  background: #fff;
}

.permission-checkbox {
  margin-right: 0;
}

.permission-checkbox :deep(.el-checkbox__label) {
  display: inline-flex;
  flex-direction: column;
  line-height: 1.25;
}

.permission-checkbox-label {
  color: #1f2937;
  font-size: 13px;
}

.permission-checkbox-code {
  color: #94a3b8;
  font-size: 11px;
  font-family: Menlo, Monaco, Consolas, "Courier New", monospace;
}
</style>
