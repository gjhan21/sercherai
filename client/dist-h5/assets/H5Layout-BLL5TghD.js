import{c as d,a as t,b as m,w as f,t as a,u as p,d as T,e as k,F as L,r as V,f as b,g,h as A,i as j,j as I,k as R,l as N,o as r,m as W,n as H}from"./index-CP1V6P-6.js";import{l as P}from"./auth-Cl7-qgmD.js";import"./http-yVn_aQk7.js";const z=[{path:"/home",label:"首页",icon:"home"},{path:"/news",label:"资讯",icon:"news"},{path:"/strategies",label:"策略",icon:"insight"},{path:"/membership",label:"会员",icon:"vip"},{path:"/profile",label:"我的",icon:"user"}];function $(u=""){const s=String(u||""),[n]=s.split("?");return n||"/home"}function D(){return z}function E(u){const s=$(u);return s.startsWith("/news")?{section:"资讯",title:"市场资讯",subtitle:"栏目切换后直接进入正文与内容流",pulse:"阅读中"}:s.startsWith("/strategies")?{section:"策略",title:"精选观点",subtitle:"像内容 App 一样查看结论、理由和风险边界",pulse:"观点流"}:s.startsWith("/membership")?{section:"会员",title:"会员中心",subtitle:"套餐、支付与激活状态集中在一页完成",pulse:"收银台"}:s.startsWith("/profile")?{section:"我的",title:"账户中心",subtitle:"消息、订单、实名和会员状态统一管理",pulse:"账户"}:{section:"首页",title:"今日观点",subtitle:"先看核心判断，再顺着内容流继续阅读",pulse:"内容优先"}}const F={class:"h5-shell"},O={class:"h5-header"},q={class:"h5-header-inner"},G={class:"h5-header-main"},J={class:"h5-brand-copy"},K={class:"h5-brand-title"},Q={class:"h5-header-actions"},U={class:"h5-shell-pulse"},X={key:0,class:"h5-user"},Y=["disabled"],Z={class:"h5-main"},ee={class:"h5-tabbar","aria-label":"H5 底部导航"},te=["innerHTML"],ae={__name:"H5Layout",setup(u){const s={home:`
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
      <path d="M4.5 10.5 12 4l7.5 6.5" />
      <path d="M6.5 10.5V20h11v-9.5" />
    </svg>
  `,news:`
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
      <path d="M5 6.5h14" />
      <path d="M5 11.5h14" />
      <path d="M5 16.5h9" />
      <path d="M17 15.5h2v3h-2z" />
    </svg>
  `,insight:`
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
      <path d="M5 18h14" />
      <path d="M7 16V9" />
      <path d="M12 16V6" />
      <path d="M17 16v-4" />
    </svg>
  `,vip:`
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
      <path d="m5 8 3 3 4-5 4 5 3-3" />
      <path d="M6 10.5 8 18h8l2-7.5" />
    </svg>
  `,user:`
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
      <circle cx="12" cy="8" r="3.2" />
      <path d="M5 19c1.8-3 4.2-4.5 7-4.5S17.2 16 19 19" />
    </svg>
  `},n=j(),w=N(),l=R(!1),{session:c,isLoggedIn:_}=A(),M=D(),h=g(()=>E(n.fullPath)),y=g(()=>{const e=c.value?.phone||"";if(/^\d{11}$/.test(e))return`${e.slice(0,3)}****${e.slice(-2)}`;const o=String(c.value?.email||"").trim();return o.includes("@")?o.split("@")[0]||"当前用户":c.value?.userID||"当前用户"});function C(e){return e==="/home"?n.path==="/home"||n.path==="/":n.path.startsWith(e)}function B(e){return s[e]||s.home}async function x(){if(!l.value){l.value=!0;try{const e=c.value?.refreshToken||"";e&&await P(e)}catch(e){console.warn("h5 logout failed:",e?.message||e)}finally{I(),l.value=!1,await w.replace("/auth")}}}return(e,o)=>{const v=b("RouterLink"),S=b("RouterView");return r(),d("div",F,[t("header",O,[t("div",q,[t("div",G,[m(v,{class:"h5-brand",to:"/home","aria-label":"返回首页"},{default:f(()=>[o[0]||(o[0]=t("span",{class:"h5-brand-mark"},"S",-1)),t("div",J,[t("span",K,a(h.value.section),1),t("strong",null,a(h.value.title),1),t("small",null,a(h.value.subtitle),1)])]),_:1})]),t("div",Q,[t("span",U,a(h.value.pulse),1),p(_)?(r(),d("span",X,a(y.value),1)):T("",!0),p(_)?(r(),d("button",{key:2,type:"button",class:"h5-header-link",disabled:l.value,onClick:x},a(l.value?"退出中":"退出"),9,Y)):(r(),k(v,{key:1,class:"h5-header-link",to:{path:"/auth",query:{redirect:p(n).fullPath}}},{default:f(()=>[...o[1]||(o[1]=[W(" 登录 ",-1)])]),_:1},8,["to"]))])])]),t("main",Z,[m(S)]),t("nav",ee,[(r(!0),d(L,null,V(p(M),i=>(r(),k(v,{key:i.path,to:i.path,class:H(["h5-tabbar-link",{active:C(i.path)}])},{default:f(()=>[t("span",{class:"h5-tabbar-icon","aria-hidden":"true",innerHTML:B(i.icon)},null,8,te),t("span",null,a(i.label),1)]),_:2},1032,["to","class"]))),128))])])}}};export{ae as default};
