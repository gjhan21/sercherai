function toArray(value) {
  return Array.isArray(value) ? value : [];
}

export function buildExplanationView(explanation = {}, fallbackReason = "") {
  const evidenceTags = toArray(explanation.evidence_cards)
    .map((item) => {
      const title = String(item?.title || "").trim();
      const value = String(item?.value || "").trim();
      if (!title || !value) {
        return "";
      }
      return `${title}: ${value}`;
    })
    .filter(Boolean);

  const proofTags = [
    ...toArray(explanation.theme_tags).filter(Boolean),
    ...evidenceTags
  ];

  const metaParts = [
    String(explanation.market_regime || "").trim(),
    String(explanation.graph_summary || "").trim()
  ].filter(Boolean);

  return {
    whyNow: String(explanation.confidence_reason || fallbackReason || "").trim(),
    proofTags,
    meta: metaParts.join(" · "),
    riskBoundary: String(explanation.risk_boundary || "").trim(),
    versionSummary: String(explanation.version_diff?.summary || "").trim()
  };
}
