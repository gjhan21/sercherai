function normalizeText(value) {
  return String(value || "").replace(/\s+/g, " ").trim();
}

function stripLeadingTag(text) {
  const match = text.match(/^[\[【]([^】\]]+)[\]】]\s*/);
  if (!match) {
    return { tag: "", body: text };
  }
  return {
    tag: normalizeText(match[1]),
    body: normalizeText(text.slice(match[0].length))
  };
}

function isMostlyLatin(text) {
  const source = normalizeText(text);
  if (!source) {
    return false;
  }
  const latinChars = (source.match(/[A-Za-z]/g) || []).length;
  return latinChars / source.length >= 0.45;
}

function trimEnglishHeadline(text) {
  let source = normalizeText(text);
  if (!source) {
    return "";
  }
  source = source.replace(/\s+By\s+[A-Z][A-Za-z.'-]+(?:\s+[A-Z][A-Za-z.'-]+){0,3}$/i, "");
  if (isMostlyLatin(source) && source.includes(":")) {
    source = normalizeText(source.split(":")[0]);
  }
  if (isMostlyLatin(source) && source.length > 56) {
    source = `${source.slice(0, 53).trim()}...`;
  }
  return source;
}

export function shapeNewsDisplayTitle(title, categoryLabel = "资讯") {
  const source = normalizeText(title);
  if (!source) {
    return "未命名资讯";
  }

  const { tag, body } = stripLeadingTag(source);
  if (!tag) {
    return trimEnglishHeadline(source) || source;
  }

  const trimmedBody = trimEnglishHeadline(body) || body;
  if (tag === "畅销书") {
    return `畅销书导读：${trimmedBody}`;
  }
  if (tag.includes("畅销书")) {
    return `${categoryLabel || "图书"}导读：${trimmedBody}`;
  }
  return `${tag}：${trimmedBody}`;
}

export function shapeStrategyDisplayTitle({ title = "", symbol = "", name = "", contract = "" } = {}) {
  const preferred = normalizeText(name);
  if (preferred) {
    return preferred;
  }

  const source = normalizeText(title);
  if (!source) {
    return "未命名策略";
  }

  const code = normalizeText(symbol || contract);
  if (code && source.startsWith(`${code} `)) {
    return normalizeText(source.slice(code.length));
  }
  return source;
}
