import { createApp } from "vue";
import App from "./App.vue";
import router from "./router";
import "../../styles/tokens.css";
import "../../styles/base.css";
import "./styles/h5-shell.css";
import "./styles/h5-ui.css";

export function mountH5App(selector = "#app") {
  return createApp(App).use(router).mount(selector);
}
