function normalizeJobName(value) {
  return String(value || "").trim();
}

function normalizeModule(value) {
  return String(value || "").trim().toUpperCase();
}

function normalizeDisplayName(value, fallback) {
  return String(value || "").trim() || String(fallback || "").trim();
}

function compareDefinitionOptions(left, right) {
  const moduleCompare = String(left?.module || "").localeCompare(String(right?.module || ""), "zh-Hans-CN");
  if (moduleCompare !== 0) {
    return moduleCompare;
  }
  return String(left?.job_name || "").localeCompare(String(right?.job_name || ""), "zh-Hans-CN");
}

function collectDefinitionJobNameMap(definitions = []) {
  const map = new Map();
  (Array.isArray(definitions) ? definitions : []).forEach((item) => {
    const jobName = normalizeJobName(item?.job_name);
    if (!jobName) {
      return;
    }
    map.set(jobName, String(item?.id || "").trim());
  });
  return map;
}

export function buildSchedulerDefinitionOptions(supportedJobs = [], definitions = []) {
  const existingJobMap = collectDefinitionJobNameMap(definitions);
  const optionMap = new Map();

  (Array.isArray(supportedJobs) ? supportedJobs : []).forEach((item) => {
    const jobName = normalizeJobName(item?.job_name);
    if (!jobName) {
      return;
    }
    optionMap.set(jobName, {
      job_name: jobName,
      display_name: normalizeDisplayName(item?.display_name, jobName),
      module: normalizeModule(item?.module) || "SYSTEM",
      alias_of: normalizeJobName(item?.alias_of),
      used: existingJobMap.has(jobName)
    });
  });

  (Array.isArray(definitions) ? definitions : []).forEach((item) => {
    const jobName = normalizeJobName(item?.job_name);
    if (!jobName || optionMap.has(jobName)) {
      return;
    }
    optionMap.set(jobName, {
      job_name: jobName,
      display_name: normalizeDisplayName(item?.display_name, jobName),
      module: normalizeModule(item?.module) || "SYSTEM",
      alias_of: "",
      used: true
    });
  });

  return Array.from(optionMap.values()).sort(compareDefinitionOptions);
}

export function buildSchedulerDefinitionCreateOptions(supportedJobs = [], definitions = []) {
  return buildSchedulerDefinitionOptions(supportedJobs, definitions).filter((item) => !item.used);
}

export function validateSchedulerDefinitionJobName(jobName, definitions = [], currentDefinitionID = "") {
  const normalizedJobName = normalizeJobName(jobName);
  if (!normalizedJobName) {
    return "任务编码不能为空";
  }
  const normalizedCurrentID = String(currentDefinitionID || "").trim();
  const duplicated = (Array.isArray(definitions) ? definitions : []).find((item) => {
    const itemJobName = normalizeJobName(item?.job_name);
    if (itemJobName !== normalizedJobName) {
      return false;
    }
    if (!normalizedCurrentID) {
      return true;
    }
    return String(item?.id || "").trim() !== normalizedCurrentID;
  });
  if (duplicated) {
    return `任务编码 ${normalizedJobName} 已存在，请直接编辑原定义或改用其他编码`;
  }
  return "";
}
