const DEFAULT_REDIRECT = "/home";
const ALLOWED_REDIRECTS = new Set([
  "/home",
  "/news",
  "/strategies",
  "/watchlist",
  "/archive",
  "/membership",
  "/profile"
]);
const AUTH_ONLY_PATHS = new Set([
  "/membership",
  "/profile"
]);

function toURL(rawPath) {
  const value = String(rawPath || "").trim();
  if (!value.startsWith("/")) {
    return null;
  }
  try {
    return new URL(value, "https://sercherai.local");
  } catch {
    return null;
  }
}

export function normalizeInviteCode(query = {}) {
  return String(query.invite_code || query.code || query.invite || "")
    .trim()
    .toUpperCase();
}

export function resolveAuthInitialMode(query = {}) {
  return normalizeInviteCode(query) ? "register" : "login";
}

export function normalizeAuthRedirect(rawPath) {
  const targetURL = toURL(rawPath);
  if (!targetURL) {
    return DEFAULT_REDIRECT;
  }

  const normalizedPath = targetURL.pathname;
  if (!ALLOWED_REDIRECTS.has(normalizedPath)) {
    return DEFAULT_REDIRECT;
  }

  return `${normalizedPath}${targetURL.search}${targetURL.hash}`;
}

export function describeAuthScene(rawPath) {
  const redirectPath = normalizeAuthRedirect(rawPath);
  const pathname = toURL(redirectPath)?.pathname || DEFAULT_REDIRECT;

  if (pathname === "/membership") {
    return {
      label: "会员中心",
      title: "登录后继续会员服务",
      tip: "完成登录后，将返回会员中心，继续升级、续费或确认激活状态。",
      highlights: ["权益升级", "续费开通", "激活确认"]
    };
  }

  if (pathname === "/news") {
    return {
      label: "资讯正文",
      title: "登录后继续深度阅读",
      tip: "完成登录后，将返回资讯页，继续查看正文、附件与关联内容。",
      highlights: ["正文阅读", "附件下载", "关联策略"]
    };
  }

  if (pathname === "/strategies") {
    return {
      label: "策略详情",
      title: "登录后继续查看策略",
      tip: "完成登录后，将返回策略页，继续阅读推荐理由、风险边界与后续动作。",
      highlights: ["主推荐", "风险边界", "后续动作"]
    };
  }

  if (pathname === "/watchlist") {
    return {
      label: "我的关注",
      title: "登录后继续跟踪已关注对象",
      tip: "完成登录后，将返回关注页，继续查看变化、风险边界和下一步动作。",
      highlights: ["关注变化", "风险边界", "持续跟踪"]
    };
  }

  if (pathname === "/archive") {
    return {
      label: "历史档案",
      title: "登录后继续查看历史复盘",
      tip: "完成登录后，将返回历史档案页，继续查看样本结果、版本变化与来源说明。",
      highlights: ["历史样本", "结果复盘", "版本变化"]
    };
  }

  if (pathname === "/profile") {
    return {
      label: "我的账户",
      title: "登录后继续处理账户事项",
      tip: "完成登录后，将返回我的页面，继续处理实名、消息、邀请与会员事项。",
      highlights: ["账户状态", "消息提醒", "邀请关系"]
    };
  }

  return {
    label: "首页",
    title: "登录后进入今日主线",
    tip: "完成登录后，将返回首页，继续查看今日主推荐、资讯节奏与会员权益。",
    highlights: ["今日主推荐", "资讯节奏", "会员权益"]
  };
}

function resolvePrimaryActionLabel(scene, mode) {
  if (mode === "register") {
    if (scene.label === "会员中心") {
      return "注册并继续会员服务";
    }
    if (scene.label === "资讯正文") {
      return "注册并继续阅读";
    }
    if (scene.label === "策略详情") {
      return "注册并继续看策略";
    }
    if (scene.label === "我的关注") {
      return "注册并继续看关注";
    }
    if (scene.label === "历史档案") {
      return "注册并继续看档案";
    }
    if (scene.label === "我的账户") {
      return "注册并继续处理账户";
    }
    return "注册并进入首页";
  }

  if (scene.label === "会员中心") {
    return "登录并继续会员服务";
  }
  if (scene.label === "资讯正文") {
    return "登录并继续阅读";
  }
  if (scene.label === "策略详情") {
    return "登录并继续看策略";
  }
  if (scene.label === "我的关注") {
    return "登录并继续看关注";
  }
  if (scene.label === "历史档案") {
    return "登录并继续看档案";
  }
  if (scene.label === "我的账户") {
    return "登录并继续处理账户";
  }
  return "登录并进入首页";
}

function resolveSubmitHint(scene, mode, inviteCode) {
  if (mode === "register" && inviteCode) {
    const sceneName = scene.label === "资讯正文" ? "资讯" : scene.label;
    return `完成注册后将自动绑定邀请码，并返回当前${sceneName}场景。`;
  }
  if (mode === "register") {
    return `完成注册后将自动建立会话，并返回${scene.label}继续当前操作。`;
  }
  return `完成登录后将自动保存当前会话，并返回${scene.label}继续当前操作。`;
}

export function buildAuthSurfaceModel({
  redirectPath = DEFAULT_REDIRECT,
  inviteCode = "",
  mode = "login"
} = {}) {
  const scene = describeAuthScene(redirectPath);
  const normalizedMode = mode === "register" ? "register" : "login";
  const hasInviteCode = Boolean(String(inviteCode || "").trim());

  return {
    heroKicker: "研究决策会员入口",
    sceneLabel: scene.label,
    heroTitle: scene.title,
    heroTip: scene.tip,
    highlights: scene.highlights,
    primaryActionLabel: resolvePrimaryActionLabel(scene, normalizedMode),
    agreementLabel: `${normalizedMode === "register" ? "注册" : "登录"}即代表你已阅读并同意《用户服务协议》与《隐私说明》。`,
    submitHint: resolveSubmitHint(scene, normalizedMode, inviteCode),
    cards: [
      {
        title: "认证后返回",
        desc: `当前认证完成后将返回${scene.label}，继续当前主线。`
      },
      {
        title: normalizedMode === "register" ? "注册后可继续" : "登录后可继续",
        desc: scene.tip
      },
      {
        title: hasInviteCode ? "邀请码已识别" : "会话保持",
        desc: hasInviteCode
          ? `邀请码 ${inviteCode} 已识别，注册完成后将自动绑定邀请关系。`
          : "登录成功后会自动保存当前会话，后续打开移动端时可继续使用。"
      }
    ]
  };
}

export function resolveAuthBackTarget(rawPath) {
  const redirectPath = normalizeAuthRedirect(rawPath);
  const pathname = toURL(redirectPath)?.pathname || DEFAULT_REDIRECT;
  if (AUTH_ONLY_PATHS.has(pathname)) {
    return DEFAULT_REDIRECT;
  }
  return redirectPath;
}
