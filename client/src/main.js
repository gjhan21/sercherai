import { createApp } from "vue";
import App from "./App.vue";
import router from "./router";
import "./styles/tokens.css";
import "./styles/base.css";
import "./styles/finance-shell.css";
import "./styles/finance-pages.css";

createApp(App).use(router).mount("#app");
