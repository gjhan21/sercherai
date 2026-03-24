import { createApp } from "vue";
import App from "./App.vue";
import router from "./router";
import "../../styles/tokens.css";
import "../../styles/base.css";
import "../../styles/finance-pages.css";
import "./styles/pc-shell.css";

export function mountPcApp(selector = "#app") {
  return createApp(App).use(router).mount(selector);
}
