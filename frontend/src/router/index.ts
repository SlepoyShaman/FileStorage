import { RouteLocation, createRouter, createWebHistory, RouteRecordRaw } from "vue-router";
import Files from "@/views/Files.vue";
import { globalVars } from "@/utils/constants";


const titles = {
  Files: "general.files",
};

const routes: RouteRecordRaw[] = [
  {
    path: "/files",
    children: [
      {
        path: ":path*",
        name: "Files",
        component: Files,
      },
    ],
  },
  {
    path: "/:catchAll(.*)*",
    redirect: (to: RouteLocation) => {
      const path = Array.isArray(to.params.catchAll)
        ? to.params.catchAll.join("/")
        : to.params.catchAll || "";
      return `/files/${path}`;
    },
  },
];

const router = createRouter({
  history: createWebHistory(globalVars.baseURL),
  routes,
});


router.afterEach((to) => {
  if (window.self !== window.top) {
    window.parent.postMessage(
      {
        type: "filebrowser:navigation",
        url: to.fullPath,
      },
      "*"
    );
  }
});

export { router, router as default };
