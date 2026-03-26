import { mountH5App } from "./apps/h5/bootstrap";
import { mountPcApp } from "./apps/pc/bootstrap";
import { resolveClientAppMode } from "./app-entry";

const pathname = typeof window !== "undefined" ? window.location.pathname : "/";

if (resolveClientAppMode(pathname) === "h5") {
  mountH5App();
} else {
  mountPcApp();
}
