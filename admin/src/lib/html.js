export function sanitizeHTML(rawHTML) {
  if (typeof window === "undefined" || typeof DOMParser === "undefined") {
    return String(rawHTML || "")
      .replace(/<script[\s\S]*?>[\s\S]*?<\/script>/gi, "")
      .replace(/<style[\s\S]*?>[\s\S]*?<\/style>/gi, "");
  }

  const parser = new DOMParser();
  const doc = parser.parseFromString(String(rawHTML || ""), "text/html");
  const disallowedSelector =
    "script,style,iframe,object,embed,link,meta,form,input,button,textarea,select";

  doc.querySelectorAll(disallowedSelector).forEach((node) => node.remove());
  doc.querySelectorAll("*").forEach((element) => {
    Array.from(element.attributes).forEach((attr) => {
      const attrName = String(attr.name || "").toLowerCase();
      const attrValue = String(attr.value || "").trim();

      if (attrName.startsWith("on")) {
        element.removeAttribute(attr.name);
        return;
      }

      if (["href", "src", "xlink:href"].includes(attrName)) {
        const loweredValue = attrValue.toLowerCase();
        const safeValue =
          loweredValue.startsWith("http://") ||
          loweredValue.startsWith("https://") ||
          loweredValue.startsWith("/") ||
          loweredValue.startsWith("#") ||
          loweredValue.startsWith("mailto:") ||
          loweredValue.startsWith("tel:") ||
          loweredValue.startsWith("data:image/");

        if (!safeValue) {
          element.removeAttribute(attr.name);
        }
      }
    });
  });

  return doc.body.innerHTML || "<p>-</p>";
}
