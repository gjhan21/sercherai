<script setup>
import { computed, onMounted, reactive, ref } from "vue";
import { listSystemConfigs, testOSSQiniuConfig, testPaymentYolkPayConfig, upsertSystemConfig } from "../api/admin";
import { hasPermission } from "../lib/session";

const activeTab = ref("oss");

const errorMessage = ref("");
const message = ref("");

const refreshingAll = ref(false);

const ossLoading = ref(false);
const ossSaving = ref(false);
const ossTesting = ref(false);
const ossTestResult = ref(null);
const ossForm = reactive({
  provider: "QINIU",
  enabled: false,
  access_key: "",
  secret_key: "",
  bucket: "",
  domain: "",
  region: "z0",
  path_prefix: "uploads/",
  use_https: true,
  upload_max_size_mb: 20,
  upload_allowed_exts: "jpg,jpeg,png,gif,webp,pdf,doc,docx,xls,xlsx,ppt,pptx,zip"
});

const paymentLoading = ref(false);
const paymentSaving = ref(false);
const paymentTesting = ref(false);
const paymentTestResult = ref(null);
const paymentForm = reactive({
  enabled: false,
  default_channel: "ALIPAY",
  signing_secret: "",
  order_timeout_minutes: 30,
  alipay_enabled: false,
  alipay_app_id: "",
  alipay_merchant_id: "",
  alipay_private_key: "",
  alipay_public_key: "",
  alipay_notify_url: "",
  wechat_enabled: false,
  wechat_mch_id: "",
  wechat_app_id: "",
  wechat_api_v3_key: "",
  wechat_cert_serial_no: "",
  wechat_private_key: "",
  wechat_notify_url: "",
  yolkpay_enabled: false,
  yolkpay_pid: "",
  yolkpay_key: "",
  yolkpay_gateway: "https://www.yolkpay.net",
  yolkpay_mapi_path: "/mapi.php",
  yolkpay_notify_url: "",
  yolkpay_return_url: "",
  yolkpay_pay_type: "airpay",
  yolkpay_device: "pc"
});

const listLoading = ref(false);
const listSubmitting = ref(false);
const listPage = ref(1);
const listPageSize = ref(20);
const listTotal = ref(0);
const listItems = ref([]);
const listFilters = reactive({
  keyword: ""
});

const dialogVisible = ref(false);
const dialogForm = reactive({
  config_key: "",
  config_value: "",
  description: ""
});

const ossRegionOptions = [
  { label: "华东(z0)", value: "z0" },
  { label: "华北(z1)", value: "z1" },
  { label: "华南(z2)", value: "z2" },
  { label: "北美(na0)", value: "na0" },
  { label: "新加坡(as0)", value: "as0" }
];
const paymentChannelOptions = ["ALIPAY", "WECHAT", "YOLKPAY"];

const futuresScoreLoading = ref(false);
const futuresScoreSaving = ref(false);
const futuresScoreConfigKey = "futures.strategy.score_weights";
const futuresScoreForm = reactive({
  trend: 25,
  structure: 20,
  flow: 15,
  risk: 20,
  news: 10,
  performance: 10
});

const futuresFactorItems = [
  { key: "trend", label: "趋势因子", desc: "方向、趋势强度与突破结构" },
  { key: "structure", label: "结构因子", desc: "入场/止盈/止损结构完整度" },
  { key: "flow", label: "资金与事件因子", desc: "事件驱动、资金流与活跃度" },
  { key: "risk", label: "风险控制因子", desc: "风险等级、失效条件与风控纪律" },
  { key: "news", label: "资讯因子", desc: "相关资讯情绪与相关度" },
  { key: "performance", label: "绩效因子", desc: "历史胜率、超额收益与回撤" }
];
const canEditSystemConfigs = hasPermission("system_config.edit");

function normalizeErrorMessage(error, fallback) {
  return error?.message || fallback || "操作失败";
}

function clearMessages() {
  errorMessage.value = "";
  message.value = "";
}

function ensureCanEditSystemConfigs() {
  if (canEditSystemConfigs) {
    return true;
  }
  errorMessage.value = "当前账号只有查看权限，无法测试或修改系统配置";
  return false;
}

function parseConfigBool(raw, fallback = false) {
  const text = String(raw ?? "").trim().toLowerCase();
  if (!text) {
    return fallback;
  }
  if (["1", "true", "yes", "on", "y"].includes(text)) {
    return true;
  }
  if (["0", "false", "no", "off", "n"].includes(text)) {
    return false;
  }
  return fallback;
}

function parseConfigInt(raw, fallback, min, max) {
  const parsed = Number.parseInt(String(raw ?? "").trim(), 10);
  if (!Number.isFinite(parsed)) {
    return fallback;
  }
  return Math.max(min, Math.min(max, parsed));
}

function toConfigMap(items) {
  const map = {};
  (items || []).forEach((item) => {
    const key = String(item?.config_key || "").trim().toLowerCase();
    if (!key) {
      return;
    }
    map[key] = String(item?.config_value ?? "").trim();
  });
  return map;
}

function applyOSSConfigMap(map) {
  ossForm.provider = (map["oss.provider"] || "QINIU").toUpperCase();
  ossForm.enabled = parseConfigBool(map["oss.enabled"], false);
  ossForm.access_key = map["oss.qiniu.access_key"] || "";
  ossForm.secret_key = map["oss.qiniu.secret_key"] || "";
  ossForm.bucket = map["oss.qiniu.bucket"] || "";
  ossForm.domain = map["oss.qiniu.domain"] || "";
  ossForm.region = (map["oss.qiniu.region"] || "z0").toLowerCase();
  ossForm.path_prefix = map["oss.qiniu.path_prefix"] || "uploads/";
  ossForm.use_https = parseConfigBool(map["oss.qiniu.use_https"], true);
  ossForm.upload_max_size_mb = parseConfigInt(map["oss.upload.max_size_mb"], 20, 1, 500);
  ossForm.upload_allowed_exts =
    map["oss.upload.allowed_exts"] || "jpg,jpeg,png,gif,webp,pdf,doc,docx,xls,xlsx,ppt,pptx,zip";
}

function applyPaymentConfigMap(map) {
  paymentForm.enabled = parseConfigBool(map["payment.enabled"], false);
  paymentForm.default_channel = (map["payment.default_channel"] || "ALIPAY").toUpperCase();
  paymentForm.signing_secret = map["payment.signing_secret"] || "";
  paymentForm.order_timeout_minutes = parseConfigInt(map["payment.order_timeout_minutes"], 30, 1, 1440);
  paymentForm.alipay_enabled = parseConfigBool(map["payment.channel.alipay.enabled"], false);
  paymentForm.alipay_app_id = map["payment.channel.alipay.app_id"] || "";
  paymentForm.alipay_merchant_id = map["payment.channel.alipay.merchant_id"] || "";
  paymentForm.alipay_private_key = map["payment.channel.alipay.private_key"] || "";
  paymentForm.alipay_public_key = map["payment.channel.alipay.public_key"] || "";
  paymentForm.alipay_notify_url = map["payment.channel.alipay.notify_url"] || "";
  paymentForm.wechat_enabled = parseConfigBool(map["payment.channel.wechat.enabled"], false);
  paymentForm.wechat_mch_id = map["payment.channel.wechat.mch_id"] || "";
  paymentForm.wechat_app_id = map["payment.channel.wechat.app_id"] || "";
  paymentForm.wechat_api_v3_key = map["payment.channel.wechat.api_v3_key"] || "";
  paymentForm.wechat_cert_serial_no = map["payment.channel.wechat.cert_serial_no"] || "";
  paymentForm.wechat_private_key = map["payment.channel.wechat.private_key"] || "";
  paymentForm.wechat_notify_url = map["payment.channel.wechat.notify_url"] || "";
  paymentForm.yolkpay_enabled = parseConfigBool(map["payment.channel.yolkpay.enabled"], false);
  paymentForm.yolkpay_pid = map["payment.channel.yolkpay.pid"] || "";
  paymentForm.yolkpay_key = map["payment.channel.yolkpay.key"] || "";
  paymentForm.yolkpay_gateway = map["payment.channel.yolkpay.gateway"] || "https://www.yolkpay.net";
  paymentForm.yolkpay_mapi_path = map["payment.channel.yolkpay.mapi_path"] || "/mapi.php";
  paymentForm.yolkpay_notify_url = map["payment.channel.yolkpay.notify_url"] || "";
  paymentForm.yolkpay_return_url = map["payment.channel.yolkpay.return_url"] || "";
  paymentForm.yolkpay_pay_type = map["payment.channel.yolkpay.pay_type"] || "airpay";
  paymentForm.yolkpay_device = map["payment.channel.yolkpay.device"] || "pc";
}

function defaultFuturesScoreWeightsPercent() {
  return {
    trend: 25,
    structure: 20,
    flow: 15,
    risk: 20,
    news: 10,
    performance: 10
  };
}

function normalizeFuturesFactorKey(rawKey) {
  const key = String(rawKey || "").trim().toLowerCase();
  if (!key) {
    return "";
  }
  if (["trend", "trend_factor"].includes(key)) return "trend";
  if (["structure", "structure_factor"].includes(key)) return "structure";
  if (["flow", "flow_factor", "event", "events"].includes(key)) return "flow";
  if (["risk", "risk_factor", "risk_control"].includes(key)) return "risk";
  if (["news", "news_factor"].includes(key)) return "news";
  if (["performance", "performance_factor", "perf"].includes(key)) return "performance";
  return "";
}

function clampWeightPercent(value) {
  const numeric = Number(value);
  if (!Number.isFinite(numeric)) {
    return 0;
  }
  return Math.max(0, Math.min(100, numeric));
}

function normalizeFuturesScoreWeightsPercent(raw) {
  const defaults = defaultFuturesScoreWeightsPercent();
  const result = { ...defaults };
  Object.entries(raw || {}).forEach(([key, value]) => {
    const normalizedKey = normalizeFuturesFactorKey(key);
    if (!normalizedKey) {
      return;
    }
    const numeric = Number(value);
    if (!Number.isFinite(numeric)) {
      return;
    }
    const asPercent = numeric <= 1 ? numeric * 100 : numeric;
    result[normalizedKey] = clampWeightPercent(asPercent);
  });
  const total = Object.values(result).reduce((sum, value) => sum + Number(value || 0), 0);
  if (total <= 0) {
    return defaults;
  }
  const normalized = {};
  Object.entries(result).forEach(([key, value]) => {
    normalized[key] = Number(((Number(value || 0) * 100) / total).toFixed(1));
  });
  return normalized;
}

function applyFuturesScoreForm(weights) {
  const normalized = normalizeFuturesScoreWeightsPercent(weights);
  futuresScoreForm.trend = normalized.trend;
  futuresScoreForm.structure = normalized.structure;
  futuresScoreForm.flow = normalized.flow;
  futuresScoreForm.risk = normalized.risk;
  futuresScoreForm.news = normalized.news;
  futuresScoreForm.performance = normalized.performance;
}

function buildFuturesScoreMethodText(weightsPercent) {
  const compact = (value) => {
    const numeric = Number(value);
    if (!Number.isFinite(numeric)) {
      return "0";
    }
    return Number.isInteger(numeric) ? String(numeric) : numeric.toFixed(1);
  };
  return `futures-v1 (trend${compact(weightsPercent.trend)} + structure${compact(
    weightsPercent.structure
  )} + flow${compact(weightsPercent.flow)} + risk${compact(weightsPercent.risk)} + news${compact(
    weightsPercent.news
  )} + performance${compact(weightsPercent.performance)})`;
}

const futuresScoreMethod = computed(() =>
  buildFuturesScoreMethodText(
    normalizeFuturesScoreWeightsPercent({
      trend: futuresScoreForm.trend,
      structure: futuresScoreForm.structure,
      flow: futuresScoreForm.flow,
      risk: futuresScoreForm.risk,
      news: futuresScoreForm.news,
      performance: futuresScoreForm.performance
    })
  )
);

const futuresWeightTotal = computed(() =>
  Number(
    (
      Number(futuresScoreForm.trend || 0) +
      Number(futuresScoreForm.structure || 0) +
      Number(futuresScoreForm.flow || 0) +
      Number(futuresScoreForm.risk || 0) +
      Number(futuresScoreForm.news || 0) +
      Number(futuresScoreForm.performance || 0)
    ).toFixed(1)
  )
);

async function fetchOSSConfig() {
  ossLoading.value = true;
  try {
    const data = await listSystemConfigs({ keyword: "oss.", page: 1, page_size: 200 });
    applyOSSConfigMap(toConfigMap(data?.items || []));
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "加载OSS配置失败");
  } finally {
    ossLoading.value = false;
  }
}

async function fetchPaymentConfig() {
  paymentLoading.value = true;
  try {
    const data = await listSystemConfigs({ keyword: "payment.", page: 1, page_size: 200 });
    applyPaymentConfigMap(toConfigMap(data?.items || []));
    paymentTestResult.value = null;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "加载支付配置失败");
  } finally {
    paymentLoading.value = false;
  }
}

async function fetchFuturesScoreConfig() {
  futuresScoreLoading.value = true;
  try {
    const data = await listSystemConfigs({ keyword: futuresScoreConfigKey, page: 1, page_size: 20 });
    const map = toConfigMap(data?.items || []);
    const raw = String(map[futuresScoreConfigKey] || "").trim();
    if (!raw) {
      applyFuturesScoreForm(defaultFuturesScoreWeightsPercent());
      return;
    }
    let parsed = null;
    try {
      parsed = JSON.parse(raw);
    } catch {
      parsed = null;
    }
    if (!parsed || typeof parsed !== "object") {
      applyFuturesScoreForm(defaultFuturesScoreWeightsPercent());
      return;
    }
    applyFuturesScoreForm(parsed);
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "加载期货评分权重失败");
  } finally {
    futuresScoreLoading.value = false;
  }
}

async function fetchConfigList(options = {}) {
  const { keepMessage = false } = options;
  listLoading.value = true;
  errorMessage.value = "";
  if (!keepMessage) {
    message.value = "";
  }
  try {
    const data = await listSystemConfigs({
      keyword: listFilters.keyword.trim(),
      page: listPage.value,
      page_size: listPageSize.value
    });
    listItems.value = data.items || [];
    listTotal.value = data.total || 0;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "加载系统配置列表失败");
  } finally {
    listLoading.value = false;
  }
}

async function refreshAll() {
  refreshingAll.value = true;
  clearMessages();
  try {
    await Promise.all([fetchOSSConfig(), fetchPaymentConfig(), fetchFuturesScoreConfig(), fetchConfigList()]);
    message.value = "配置中心数据已刷新";
  } finally {
    refreshingAll.value = false;
  }
}

function boolToConfigValue(value) {
  return value ? "true" : "false";
}

async function saveOSSConfig() {
  if (!ensureCanEditSystemConfigs()) {
    return;
  }
  if (ossForm.enabled) {
    if (!ossForm.access_key.trim() || !ossForm.secret_key.trim() || !ossForm.bucket.trim() || !ossForm.domain.trim()) {
      errorMessage.value = "启用七牛云时，access_key/secret_key/bucket/domain 不能为空";
      return;
    }
  }
  ossSaving.value = true;
  clearMessages();
  try {
    const payloads = [
      { config_key: "oss.provider", config_value: "QINIU", description: "对象存储服务商" },
      { config_key: "oss.enabled", config_value: boolToConfigValue(ossForm.enabled), description: "对象存储总开关" },
      { config_key: "oss.qiniu.access_key", config_value: ossForm.access_key.trim(), description: "七牛云AccessKey" },
      { config_key: "oss.qiniu.secret_key", config_value: ossForm.secret_key.trim(), description: "七牛云SecretKey" },
      { config_key: "oss.qiniu.bucket", config_value: ossForm.bucket.trim(), description: "七牛云存储空间Bucket" },
      { config_key: "oss.qiniu.domain", config_value: ossForm.domain.trim(), description: "七牛云访问域名" },
      { config_key: "oss.qiniu.region", config_value: ossForm.region.trim().toLowerCase(), description: "七牛云存储区域" },
      { config_key: "oss.qiniu.path_prefix", config_value: ossForm.path_prefix.trim(), description: "七牛云对象前缀" },
      { config_key: "oss.qiniu.use_https", config_value: boolToConfigValue(ossForm.use_https), description: "七牛云访问是否启用HTTPS" },
      { config_key: "oss.upload.max_size_mb", config_value: String(parseConfigInt(ossForm.upload_max_size_mb, 20, 1, 500)), description: "上传文件最大大小(MB)" },
      { config_key: "oss.upload.allowed_exts", config_value: ossForm.upload_allowed_exts.trim(), description: "允许上传扩展名，逗号分隔" }
    ];
    await Promise.all(payloads.map((payload) => upsertSystemConfig(payload)));
    await Promise.all([fetchOSSConfig(), fetchConfigList({ keepMessage: true })]);
    message.value = "OSS配置已保存";
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "保存OSS配置失败");
  } finally {
    ossSaving.value = false;
  }
}

async function testOSSConfig() {
  if (!ensureCanEditSystemConfigs()) {
    return;
  }
  ossTesting.value = true;
  clearMessages();
  try {
    const data = await testOSSQiniuConfig();
    ossTestResult.value = data || null;
    message.value = `七牛云测试成功：${data?.object_key || ""}`;
  } catch (error) {
    ossTestResult.value = null;
    errorMessage.value = normalizeErrorMessage(error, "七牛云测试失败");
  } finally {
    ossTesting.value = false;
  }
}

async function testPaymentConfig() {
  if (!ensureCanEditSystemConfigs()) {
    return;
  }
  paymentTesting.value = true;
  clearMessages();
  try {
    const data = await testPaymentYolkPayConfig();
    paymentTestResult.value = data || null;
    message.value = `蛋黄支付测试成功：商户 ${data?.pid || "-"}`;
  } catch (error) {
    paymentTestResult.value = null;
    errorMessage.value = normalizeErrorMessage(error, "蛋黄支付测试失败");
  } finally {
    paymentTesting.value = false;
  }
}

async function savePaymentConfig() {
  if (!ensureCanEditSystemConfigs()) {
    return;
  }
  if (paymentForm.enabled) {
    if (!paymentForm.signing_secret.trim()) {
      errorMessage.value = "启用支付时，签名密钥不能为空";
      return;
    }
    if (!paymentForm.alipay_enabled && !paymentForm.wechat_enabled && !paymentForm.yolkpay_enabled) {
      errorMessage.value = "启用支付时，至少开启一个支付通道";
      return;
    }
  }
  if (paymentForm.yolkpay_enabled) {
    if (!paymentForm.yolkpay_pid.trim() || !paymentForm.yolkpay_key.trim()) {
      errorMessage.value = "启用蛋黄支付时，商户ID和商户密钥不能为空";
      return;
    }
  }
  paymentSaving.value = true;
  clearMessages();
  try {
    const payloads = [
      { config_key: "payment.enabled", config_value: boolToConfigValue(paymentForm.enabled), description: "平台支付总开关" },
      { config_key: "payment.default_channel", config_value: paymentForm.default_channel, description: "默认支付通道" },
      { config_key: "payment.signing_secret", config_value: paymentForm.signing_secret.trim(), description: "支付回调签名密钥" },
      { config_key: "payment.order_timeout_minutes", config_value: String(parseConfigInt(paymentForm.order_timeout_minutes, 30, 1, 1440)), description: "支付订单超时时间(分钟)" },
      { config_key: "payment.channel.alipay.enabled", config_value: boolToConfigValue(paymentForm.alipay_enabled), description: "支付宝通道开关" },
      { config_key: "payment.channel.alipay.app_id", config_value: paymentForm.alipay_app_id.trim(), description: "支付宝应用ID" },
      { config_key: "payment.channel.alipay.merchant_id", config_value: paymentForm.alipay_merchant_id.trim(), description: "支付宝商户号" },
      { config_key: "payment.channel.alipay.private_key", config_value: paymentForm.alipay_private_key.trim(), description: "支付宝应用私钥" },
      { config_key: "payment.channel.alipay.public_key", config_value: paymentForm.alipay_public_key.trim(), description: "支付宝公钥" },
      { config_key: "payment.channel.alipay.notify_url", config_value: paymentForm.alipay_notify_url.trim(), description: "支付宝异步通知地址" },
      { config_key: "payment.channel.wechat.enabled", config_value: boolToConfigValue(paymentForm.wechat_enabled), description: "微信支付通道开关" },
      { config_key: "payment.channel.wechat.mch_id", config_value: paymentForm.wechat_mch_id.trim(), description: "微信支付商户号" },
      { config_key: "payment.channel.wechat.app_id", config_value: paymentForm.wechat_app_id.trim(), description: "微信支付应用ID" },
      { config_key: "payment.channel.wechat.api_v3_key", config_value: paymentForm.wechat_api_v3_key.trim(), description: "微信支付APIv3密钥" },
      { config_key: "payment.channel.wechat.cert_serial_no", config_value: paymentForm.wechat_cert_serial_no.trim(), description: "微信支付证书序列号" },
      { config_key: "payment.channel.wechat.private_key", config_value: paymentForm.wechat_private_key.trim(), description: "微信支付私钥" },
      { config_key: "payment.channel.wechat.notify_url", config_value: paymentForm.wechat_notify_url.trim(), description: "微信支付异步通知地址" },
      { config_key: "payment.channel.yolkpay.enabled", config_value: boolToConfigValue(paymentForm.yolkpay_enabled), description: "蛋黄支付通道开关" },
      { config_key: "payment.channel.yolkpay.pid", config_value: paymentForm.yolkpay_pid.trim(), description: "蛋黄支付商户ID" },
      { config_key: "payment.channel.yolkpay.key", config_value: paymentForm.yolkpay_key.trim(), description: "蛋黄支付商户密钥" },
      { config_key: "payment.channel.yolkpay.gateway", config_value: paymentForm.yolkpay_gateway.trim(), description: "蛋黄支付网关域名" },
      { config_key: "payment.channel.yolkpay.mapi_path", config_value: paymentForm.yolkpay_mapi_path.trim(), description: "蛋黄支付下单路径" },
      { config_key: "payment.channel.yolkpay.notify_url", config_value: paymentForm.yolkpay_notify_url.trim(), description: "蛋黄支付异步通知地址" },
      { config_key: "payment.channel.yolkpay.return_url", config_value: paymentForm.yolkpay_return_url.trim(), description: "蛋黄支付同步跳转地址" },
      { config_key: "payment.channel.yolkpay.pay_type", config_value: paymentForm.yolkpay_pay_type.trim().toLowerCase(), description: "蛋黄支付支付类型" },
      { config_key: "payment.channel.yolkpay.device", config_value: paymentForm.yolkpay_device.trim().toLowerCase(), description: "蛋黄支付设备类型" }
    ];
    await Promise.all(payloads.map((payload) => upsertSystemConfig(payload)));
    await Promise.all([fetchPaymentConfig(), fetchConfigList({ keepMessage: true })]);
    message.value = "支付配置已保存";
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "保存支付配置失败");
  } finally {
    paymentSaving.value = false;
  }
}

function resetFuturesScoreDefaults() {
  applyFuturesScoreForm(defaultFuturesScoreWeightsPercent());
  clearMessages();
  message.value = "已恢复默认权重";
}

function normalizeFuturesScoreForm() {
  applyFuturesScoreForm({
    trend: futuresScoreForm.trend,
    structure: futuresScoreForm.structure,
    flow: futuresScoreForm.flow,
    risk: futuresScoreForm.risk,
    news: futuresScoreForm.news,
    performance: futuresScoreForm.performance
  });
}

async function saveFuturesScoreConfig() {
  if (!ensureCanEditSystemConfigs()) {
    return;
  }
  futuresScoreSaving.value = true;
  clearMessages();
  try {
    const normalized = normalizeFuturesScoreWeightsPercent({
      trend: futuresScoreForm.trend,
      structure: futuresScoreForm.structure,
      flow: futuresScoreForm.flow,
      risk: futuresScoreForm.risk,
      news: futuresScoreForm.news,
      performance: futuresScoreForm.performance
    });
    applyFuturesScoreForm(normalized);
    const configValue = JSON.stringify({
      trend: Number((normalized.trend / 100).toFixed(4)),
      structure: Number((normalized.structure / 100).toFixed(4)),
      flow: Number((normalized.flow / 100).toFixed(4)),
      risk: Number((normalized.risk / 100).toFixed(4)),
      news: Number((normalized.news / 100).toFixed(4)),
      performance: Number((normalized.performance / 100).toFixed(4))
    });
    await upsertSystemConfig({
      config_key: futuresScoreConfigKey,
      config_value: configValue,
      description: "期货策略评分因子权重(JSON，自动归一化)"
    });
    await Promise.all([fetchFuturesScoreConfig(), fetchConfigList({ keepMessage: true })]);
    message.value = "期货评分权重已保存";
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "保存期货评分权重失败");
  } finally {
    futuresScoreSaving.value = false;
  }
}

function openCreateDialog() {
  if (!ensureCanEditSystemConfigs()) {
    return;
  }
  Object.assign(dialogForm, {
    config_key: "",
    config_value: "",
    description: ""
  });
  dialogVisible.value = true;
}

function openEditDialog(item) {
  if (!ensureCanEditSystemConfigs()) {
    return;
  }
  Object.assign(dialogForm, {
    config_key: item.config_key || "",
    config_value: item.config_value || "",
    description: item.description || ""
  });
  dialogVisible.value = true;
}

async function submitDialog() {
  if (!ensureCanEditSystemConfigs()) {
    return;
  }
  const payload = {
    config_key: dialogForm.config_key.trim(),
    config_value: dialogForm.config_value,
    description: dialogForm.description.trim()
  };
  if (!payload.config_key || payload.config_value === "") {
    errorMessage.value = "config_key 和 config_value 不能为空";
    return;
  }
  listSubmitting.value = true;
  clearMessages();
  try {
    await upsertSystemConfig(payload);
    dialogVisible.value = false;
    await Promise.all([fetchConfigList({ keepMessage: true }), fetchOSSConfig(), fetchPaymentConfig(), fetchFuturesScoreConfig()]);
    message.value = `系统配置 ${payload.config_key} 已保存`;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "保存系统配置失败");
  } finally {
    listSubmitting.value = false;
  }
}

function applyListFilters() {
  listPage.value = 1;
  fetchConfigList();
}

function resetListFilters() {
  listFilters.keyword = "";
  listPage.value = 1;
  fetchConfigList();
}

function handleListPageChange(nextPage) {
  if (nextPage === listPage.value) {
    return;
  }
  listPage.value = nextPage;
  fetchConfigList();
}

onMounted(refreshAll);
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">配置中心</h1>
        <p class="muted">统一管理 OSS、支付通道与期货评分权重配置，支持七牛云与策略模型参数维护</p>
      </div>
      <el-button :loading="refreshingAll" @click="refreshAll">刷新全部</el-button>
    </div>

    <el-alert
      v-if="errorMessage"
      :title="errorMessage"
      type="error"
      show-icon
      style="margin-bottom: 12px"
    />
    <el-alert
      v-if="message"
      :title="message"
      type="success"
      show-icon
      style="margin-bottom: 12px"
    />

    <el-tabs v-model="activeTab" type="border-card">
      <el-tab-pane label="OSS配置（七牛云）" name="oss">
        <div class="card" v-loading="ossLoading">
          <div class="section-head">
            <div class="section-title">对象存储</div>
            <div class="toolbar" style="margin-bottom: 0">
              <el-button :loading="ossLoading" @click="fetchOSSConfig">刷新</el-button>
              <el-button
                v-if="canEditSystemConfigs"
                :loading="ossTesting"
                :disabled="ossSaving"
                @click="testOSSConfig"
              >
                测试七牛配置
              </el-button>
              <el-button
                v-if="canEditSystemConfigs"
                type="primary"
                :loading="ossSaving"
                @click="saveOSSConfig"
              >
                保存OSS配置
              </el-button>
            </div>
          </div>

          <el-form label-width="140px">
            <div class="form-grid">
              <el-form-item label="存储服务商">
                <el-select v-model="ossForm.provider" disabled>
                  <el-option label="七牛云 QINIU" value="QINIU" />
                </el-select>
              </el-form-item>
              <el-form-item label="OSS启用状态">
                <el-switch v-model="ossForm.enabled" />
              </el-form-item>
              <el-form-item label="AccessKey">
                <el-input v-model="ossForm.access_key" placeholder="七牛云 AccessKey" />
              </el-form-item>
              <el-form-item label="SecretKey">
                <el-input v-model="ossForm.secret_key" show-password placeholder="七牛云 SecretKey" />
              </el-form-item>
              <el-form-item label="Bucket">
                <el-input v-model="ossForm.bucket" placeholder="存储空间名称" />
              </el-form-item>
              <el-form-item label="访问域名">
                <el-input v-model="ossForm.domain" placeholder="例如 cdn.example.com" />
              </el-form-item>
              <el-form-item label="存储区域">
                <el-select v-model="ossForm.region">
                  <el-option v-for="item in ossRegionOptions" :key="item.value" :label="item.label" :value="item.value" />
                </el-select>
              </el-form-item>
              <el-form-item label="对象前缀">
                <el-input v-model="ossForm.path_prefix" placeholder="例如 uploads/news/" />
              </el-form-item>
              <el-form-item label="启用HTTPS">
                <el-switch v-model="ossForm.use_https" />
              </el-form-item>
              <el-form-item label="最大上传大小(MB)">
                <el-input-number v-model="ossForm.upload_max_size_mb" :min="1" :max="500" :step="1" controls-position="right" />
              </el-form-item>
            </div>
            <el-form-item label="允许上传扩展名">
              <el-input
                v-model="ossForm.upload_allowed_exts"
                type="textarea"
                :rows="3"
                placeholder="例如 jpg,png,pdf,docx,zip"
              />
            </el-form-item>
          </el-form>

          <el-alert
            title="说明：当前 OSS 先对接七牛云，后续可扩展腾讯云COS、阿里云OSS等存储服务商。"
            type="info"
            :closable="false"
            show-icon
          />
          <el-descriptions
            v-if="ossTestResult"
            title="最近一次七牛测试结果"
            :column="2"
            border
            size="small"
            style="margin-top: 10px"
          >
            <el-descriptions-item label="Provider">{{ ossTestResult.provider || "-" }}</el-descriptions-item>
            <el-descriptions-item label="Region">{{ ossTestResult.region || "-" }}</el-descriptions-item>
            <el-descriptions-item label="Bucket">{{ ossTestResult.bucket || "-" }}</el-descriptions-item>
            <el-descriptions-item label="对象Key">{{ ossTestResult.object_key || "-" }}</el-descriptions-item>
            <el-descriptions-item label="文件地址" :span="2">
              <el-link v-if="ossTestResult.file_url" :href="ossTestResult.file_url" target="_blank" type="primary">
                {{ ossTestResult.file_url }}
              </el-link>
              <span v-else>-</span>
            </el-descriptions-item>
          </el-descriptions>
        </div>
      </el-tab-pane>

      <el-tab-pane label="支付配置" name="payment">
        <div class="card" v-loading="paymentLoading">
          <div class="section-head">
            <div class="section-title">平台支付通道</div>
            <div class="toolbar" style="margin-bottom: 0">
              <el-button :loading="paymentLoading" @click="fetchPaymentConfig">刷新</el-button>
              <el-button
                v-if="canEditSystemConfigs"
                :loading="paymentTesting"
                :disabled="paymentSaving"
                @click="testPaymentConfig"
              >
                测试蛋黄支付
              </el-button>
              <el-button
                v-if="canEditSystemConfigs"
                type="primary"
                :loading="paymentSaving"
                @click="savePaymentConfig"
              >
                保存支付配置
              </el-button>
            </div>
          </div>

          <el-form label-width="150px">
            <div class="form-grid">
              <el-form-item label="支付总开关">
                <el-switch v-model="paymentForm.enabled" />
              </el-form-item>
              <el-form-item label="默认支付通道">
                <el-select v-model="paymentForm.default_channel">
                  <el-option v-for="item in paymentChannelOptions" :key="item" :label="item" :value="item" />
                </el-select>
              </el-form-item>
              <el-form-item label="签名密钥">
                <el-input v-model="paymentForm.signing_secret" show-password placeholder="用于回调验签" />
              </el-form-item>
              <el-form-item label="订单超时(分钟)">
                <el-input-number v-model="paymentForm.order_timeout_minutes" :min="1" :max="1440" :step="1" controls-position="right" />
              </el-form-item>
            </div>

            <div class="subsection">支付宝通道</div>
            <div class="form-grid">
              <el-form-item label="支付宝开关">
                <el-switch v-model="paymentForm.alipay_enabled" />
              </el-form-item>
              <el-form-item label="AppID">
                <el-input v-model="paymentForm.alipay_app_id" placeholder="支付宝应用ID" />
              </el-form-item>
              <el-form-item label="商户号">
                <el-input v-model="paymentForm.alipay_merchant_id" placeholder="支付宝商户号" />
              </el-form-item>
              <el-form-item label="异步通知地址">
                <el-input v-model="paymentForm.alipay_notify_url" placeholder="https://api.example.com/api/v1/payment/callbacks/alipay" />
              </el-form-item>
            </div>
            <el-form-item label="应用私钥">
              <el-input v-model="paymentForm.alipay_private_key" type="textarea" :rows="4" placeholder="-----BEGIN PRIVATE KEY-----" />
            </el-form-item>
            <el-form-item label="支付宝公钥">
              <el-input v-model="paymentForm.alipay_public_key" type="textarea" :rows="4" placeholder="-----BEGIN PUBLIC KEY-----" />
            </el-form-item>

            <div class="subsection">微信支付通道</div>
            <div class="form-grid">
              <el-form-item label="微信支付开关">
                <el-switch v-model="paymentForm.wechat_enabled" />
              </el-form-item>
              <el-form-item label="商户号">
                <el-input v-model="paymentForm.wechat_mch_id" placeholder="微信支付商户号" />
              </el-form-item>
              <el-form-item label="AppID">
                <el-input v-model="paymentForm.wechat_app_id" placeholder="微信应用AppID" />
              </el-form-item>
              <el-form-item label="异步通知地址">
                <el-input v-model="paymentForm.wechat_notify_url" placeholder="https://api.example.com/api/v1/payment/callbacks/wechat" />
              </el-form-item>
              <el-form-item label="APIv3密钥">
                <el-input v-model="paymentForm.wechat_api_v3_key" show-password placeholder="32位APIv3密钥" />
              </el-form-item>
              <el-form-item label="证书序列号">
                <el-input v-model="paymentForm.wechat_cert_serial_no" placeholder="微信支付证书序列号" />
              </el-form-item>
            </div>
            <el-form-item label="商户私钥">
              <el-input v-model="paymentForm.wechat_private_key" type="textarea" :rows="4" placeholder="-----BEGIN PRIVATE KEY-----" />
            </el-form-item>

            <div class="subsection">蛋黄支付通道</div>
            <div class="form-grid">
              <el-form-item label="蛋黄支付开关">
                <el-switch v-model="paymentForm.yolkpay_enabled" />
              </el-form-item>
              <el-form-item label="商户ID(pid)">
                <el-input v-model="paymentForm.yolkpay_pid" placeholder="例如 1005" />
              </el-form-item>
              <el-form-item label="商户密钥(key)">
                <el-input v-model="paymentForm.yolkpay_key" show-password placeholder="蛋黄支付商户密钥" />
              </el-form-item>
              <el-form-item label="网关地址">
                <el-input v-model="paymentForm.yolkpay_gateway" placeholder="https://www.yolkpay.net" />
              </el-form-item>
              <el-form-item label="下单路径">
                <el-input v-model="paymentForm.yolkpay_mapi_path" placeholder="/mapi.php" />
              </el-form-item>
              <el-form-item label="支付类型(type)">
                <el-input v-model="paymentForm.yolkpay_pay_type" placeholder="airpay / unionpay" />
              </el-form-item>
              <el-form-item label="设备类型(device)">
                <el-select v-model="paymentForm.yolkpay_device">
                  <el-option label="pc" value="pc" />
                  <el-option label="mobile" value="mobile" />
                </el-select>
              </el-form-item>
              <el-form-item label="异步通知地址">
                <el-input v-model="paymentForm.yolkpay_notify_url" placeholder="https://api.example.com/api/v1/payment/callbacks/yolkpay/notify" />
              </el-form-item>
              <el-form-item label="同步跳转地址">
                <el-input v-model="paymentForm.yolkpay_return_url" placeholder="https://client.example.com/membership" />
              </el-form-item>
            </div>
          </el-form>

          <el-alert
            title="说明：支付参数保存后会进入系统配置表；上线前请先联调回调地址与签名校验。"
            type="info"
            :closable="false"
            show-icon
          />
          <el-descriptions
            v-if="paymentTestResult"
            title="最近一次蛋黄支付测试结果"
            :column="2"
            border
            size="small"
            style="margin-top: 10px"
          >
            <el-descriptions-item label="商户ID">{{ paymentTestResult.pid || "-" }}</el-descriptions-item>
            <el-descriptions-item label="激活状态">{{ paymentTestResult.active || "-" }}</el-descriptions-item>
            <el-descriptions-item label="余额">{{ paymentTestResult.money || "-" }}</el-descriptions-item>
            <el-descriptions-item label="订单总数">{{ paymentTestResult.orders || "-" }}</el-descriptions-item>
          </el-descriptions>
        </div>
      </el-tab-pane>

      <el-tab-pane label="期货评分配置" name="futures-score">
        <div class="card" v-loading="futuresScoreLoading">
          <div class="section-head">
            <div class="section-title">期货策略评分权重</div>
            <div class="toolbar" style="margin-bottom: 0">
              <el-button :loading="futuresScoreLoading" @click="fetchFuturesScoreConfig">刷新</el-button>
              <el-button
                v-if="canEditSystemConfigs"
                :disabled="futuresScoreSaving"
                @click="normalizeFuturesScoreForm"
              >
                归一化到100%
              </el-button>
              <el-button
                v-if="canEditSystemConfigs"
                :disabled="futuresScoreSaving"
                @click="resetFuturesScoreDefaults"
              >
                恢复默认
              </el-button>
              <el-button
                v-if="canEditSystemConfigs"
                type="primary"
                :loading="futuresScoreSaving"
                @click="saveFuturesScoreConfig"
              >
                保存权重
              </el-button>
            </div>
          </div>

          <el-alert
            :title="`当前总权重：${futuresWeightTotal}%`"
            :type="Math.abs(futuresWeightTotal - 100) < 0.1 ? 'success' : 'warning'"
            :closable="false"
            show-icon
            style="margin-bottom: 10px"
          />

          <el-form label-width="120px">
            <el-form-item
              v-for="item in futuresFactorItems"
              :key="item.key"
              :label="item.label"
            >
              <div class="weight-row">
                <el-slider
                  v-model="futuresScoreForm[item.key]"
                  :min="0"
                  :max="100"
                  :step="0.5"
                  style="flex: 1"
                />
                <el-input-number
                  v-model="futuresScoreForm[item.key]"
                  :min="0"
                  :max="100"
                  :step="0.5"
                  controls-position="right"
                  style="width: 120px"
                />
                <span class="weight-unit">%</span>
              </div>
              <div class="muted">{{ item.desc }}</div>
            </el-form-item>
          </el-form>

          <el-alert
            :title="`当前模型标识：${futuresScoreMethod}`"
            type="info"
            :closable="false"
            show-icon
            style="margin-top: 8px"
          />
          <el-alert
            title="保存后将写入 system_configs：futures.strategy.score_weights，调度与客户端详情页会实时使用新权重。"
            type="info"
            :closable="false"
            show-icon
            style="margin-top: 8px"
          />
        </div>
      </el-tab-pane>

      <el-tab-pane label="系统配置列表" name="list">
        <div class="card" style="margin-bottom: 12px">
          <div class="toolbar" style="margin-bottom: 0">
            <el-input v-model="listFilters.keyword" clearable placeholder="按 key/description 关键词检索" style="width: 280px" />
            <el-button type="primary" plain @click="applyListFilters">查询</el-button>
            <el-button @click="resetListFilters">重置</el-button>
            <el-button :loading="listLoading" @click="fetchConfigList">刷新</el-button>
            <el-button v-if="canEditSystemConfigs" type="primary" @click="openCreateDialog">新增配置</el-button>
          </div>
        </div>

        <div class="card">
          <el-table :data="listItems" border stripe v-loading="listLoading" empty-text="暂无系统配置">
            <el-table-column prop="config_key" label="配置键" min-width="260" />
            <el-table-column prop="config_value" label="配置值" min-width="320" />
            <el-table-column prop="description" label="描述" min-width="220" />
            <el-table-column prop="updated_by" label="更新人" min-width="120" />
            <el-table-column prop="updated_at" label="更新时间" min-width="180" />
            <el-table-column label="操作" align="right" min-width="110">
              <template #default="{ row }">
                <el-button v-if="canEditSystemConfigs" size="small" @click="openEditDialog(row)">编辑</el-button>
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination">
            <el-text type="info">第 {{ listPage }} 页，共 {{ listTotal }} 条</el-text>
            <el-pagination
              background
              layout="prev, pager, next"
              :current-page="listPage"
              :page-size="listPageSize"
              :total="listTotal"
              @current-change="handleListPageChange"
            />
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="dialogVisible" title="系统配置" width="640px" destroy-on-close>
      <el-form label-width="110px">
        <el-form-item label="配置键" required>
          <el-input v-model="dialogForm.config_key" placeholder="如 oss.provider" />
        </el-form-item>
        <el-form-item label="配置值" required>
          <el-input v-model="dialogForm.config_value" type="textarea" :rows="4" placeholder="配置内容" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="dialogForm.description" placeholder="用途说明" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button v-if="canEditSystemConfigs" type="primary" :loading="listSubmitting" @click="submitDialog">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  margin-bottom: 12px;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  color: #1f2937;
}

.subsection {
  margin: 10px 0 8px;
  font-size: 14px;
  font-weight: 600;
  color: #374151;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 0 12px;
}

.weight-row {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 10px;
}

.weight-unit {
  color: #6b7280;
  font-size: 12px;
}
</style>
