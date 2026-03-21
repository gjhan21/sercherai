<script setup>
import { useRouter } from "vue-router";
import StatePanel from "../components/StatePanel.vue";
import { clearSession } from "../lib/session";

const router = useRouter();

function backToLogin() {
  clearSession();
  router.replace("/login");
}
</script>

<template>
  <section class="no-access-page">
    <StatePanel
      tone="warning"
      eyebrow="403"
      title="当前账号暂无可访问页面"
      description="你已经完成登录，但当前角色还没有分配任何后台页面权限。请联系超级管理员分配权限后再重试。"
    >
      <template #actions>
        <el-button type="primary" @click="backToLogin">返回登录页</el-button>
        <el-button @click="router.go(0)">刷新权限</el-button>
      </template>
    </StatePanel>
  </section>
</template>

<style scoped>
.no-access-page {
  min-height: calc(100vh - 56px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
}

@media (max-width: 640px) {
  :deep(.state-actions) {
    flex-direction: column;
  }
}
</style>
