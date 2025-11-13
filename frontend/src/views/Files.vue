<template>
  <errors v-if="error && !(isShare && error.status === 401)" :errorCode="error.status" />
  <component v-else-if="currentViewLoaded" :is="currentView"></component>
  <div v-else>
    <h2 class="message delayed">
      <div class="spinner">
        <div class="bounce1"></div>
        <div class="bounce2"></div>
        <div class="bounce3"></div>
      </div>
      <span>{{ $t("files.loading") }}</span>
    </h2>
  </div>
  <PopupPreview v-if="popupEnabled" />
</template>

<script>
import { filesApi } from "@/api";
import ListingView from "@/views/files/ListingView.vue";
import Editor from "@/views/files/Editor.vue";
import { state, mutations, getters } from "@/store";
import { url } from "@/utils";
import router from "@/router";
import { globalVars } from "@/utils/constants";
import { extractSourceFromPath } from "@/utils/url";

export default {
  name: "files",
  components: {
    Breadcrumbs,
    Errors,
    Preview,
    ListingView,
    Editor,
    EpubViewer,
    DocViewer,
    OnlyOfficeEditor,
    MarkdownViewer,
    PopupPreview,
    ShareInfoCard,
  },
  data() {
    return {
      error: null,
      width: window.innerWidth,
      lastPath: "",
      lastHash: "",
      popupSource: "",
      sharePassword: "",
      attemptedPasswordLogin: false,
      shareHash: null,
      shareSubPath: "",
      shareToken: "",
    };
  },
  computed: {
    share() {
      return state.share;
    },
    currentView() {
      return getters.currentView();
    },
    currentViewLoaded() {
      return getters.currentView() != "";
    },
    reload() {
      return state.reload;
    },
  },
  created() {
    this.fetchData();
  },
  watch: {
    $route: "fetchData",
    reload(value) {
      if (value) {
        this.fetchData();
      }
    },
  },
  mounted() {
    window.addEventListener("hashchange", this.scrollToHash);
    window.addEventListener("keydown", this.keyEvent);
    if (getters.isInvalidShare()) {
      this.error = {
        status: "share404",
        message: "errors.shareNotFound",
      };
    }
  },
  beforeUnmount() {
    window.removeEventListener("keydown", this.keyEvent);
  },
  unmounted() {
    mutations.replaceRequest({}); // Use mutation
  },
  methods: {
    scrollToHash() {
      let scrollToId = "";
      const noHashChange = window.location.hash === this.lastHash
      if (noHashChange && state.previousHistoryItem.name === "") return;
      this.lastHash = window.location.hash;
      if (window.location.hash) {
        const rawHash = window.location.hash.slice(1);
        let decodedName = rawHash;
        try {
          decodedName = decodeURIComponent(rawHash);
        } catch (e) {
          decodedName = rawHash;
        }
        scrollToId = url.base64Encode(encodeURIComponent(decodedName));

      } else if (state.previousHistoryItem.name && state.previousHistoryItem.path === state.req.path && state.previousHistoryItem.source === state.req.source) {
        scrollToId = url.base64Encode(encodeURIComponent(state.previousHistoryItem.name));
      }
      const element = document.getElementById(scrollToId);
        if (element) {
          element.scrollIntoView({
            behavior: "instant",
            block: "center",
          });
    
          element.classList.add('scroll-glow');
          
          setTimeout(() => {
            element.classList.remove('scroll-glow');
          }, 1000);
        }
    },
    async fetchData() {
      if (state.deletedItem || getters.isInvalidShare() || shareInfo.shareType == "upload") {
        return
      }

      if (!state.user.sorting) {
        mutations.updateListingSortConfig({
          field: "name",
          asc: true,
        });
      }
   
      mutations.setLoading(getters.isShare() ? "share" : "files", true);
      this.error = null;
      mutations.setReload(false);

      try {
        await this.fetchFilesData();
      } catch (e) {
        if (e.message) {
          notify.showError(e.message);
        } else {
          notify.showError(e);
        }
        this.error = e;
        mutations.replaceRequest({});
        if (e.status === 404) {
          router.push({ name: "notFound" });
        } else if (e.status === 403) {
          router.push({ name: "forbidden" });
        } else {
          router.push({ name: "error" });
        }
      } 

      setTimeout(() => {
        this.scrollToHash();
      }, 25);
      this.lastPath = state.route.path;
    },

    async fetchFilesData() {
      if (!getters.isLoggedIn()) {
        return;
      }

      mutations.clearShareData();

      const routePath = url.removeTrailingSlash(getters.routePath(`${globalVars.baseURL}files`));
      const rootRoute =
        routePath == "/files" ||
        routePath == "/files/" ||
        routePath == "" ||
        routePath == "/";

      if (state.serverHasMultipleSources && rootRoute) {
        const targetPath = `/files/${state.sources.current}`;
        if (routePath !== targetPath) {
          router.push(targetPath);
          return;
        }
      }

      const result = extractSourceFromPath(getters.routePath());

      if (result.source === "") {
        this.error = { message: $t("index.noSources") };
        mutations.replaceRequest({});
        return;
      }

      this.lastHash = "";
      mutations.resetSelected();
      let data = {};
      try {
        const fetchSource = decodeURIComponent(result.source);
        const fetchPath = decodeURIComponent(result.path);
        let res = await filesApi.fetchFiles(fetchSource, fetchPath );

        if (res.type != "directory" && !res.type.startsWith("image")) {
          const content = !getters.fileViewingDisabled(res.name);
          let directoryPath = url.removeLastDir(res.path);

          if (!directoryPath || directoryPath === '') {
            directoryPath = '/';
          }

          const shouldFetchParent = directoryPath !== res.path;

          const promises = [
            filesApi.fetchFiles(res.source, res.path, content)
          ];

          if (shouldFetchParent) {
            promises.push(
              filesApi.fetchFiles(res.source, directoryPath).catch(() => null)
            );
          }

          const results = await Promise.all(promises);
          res = results[0];

          if (shouldFetchParent && results[1] && results[1].items) {
            res.parentDirItems = results[1].items;
          }
        }
        data = res;

        if (state.sources.count > 1) {
          mutations.setCurrentSource(data.source);
        }
        document.title = `${document.title} - ${res.name}`;
      } catch (e) {
        notify.showError(e);
        this.error = e;
        mutations.replaceRequest({});
      } finally {
        mutations.replaceRequest(data);
        mutations.setLoading("files", false);
      }
    },
  },
};
</script>

<style>
.scroll-glow {
  animation: scrollGlowAnimation 1s ease-out;
}

@keyframes scrollGlowAnimation {
  0% {
    color: inherit;
  }
  50% {
    color: var(--primaryColor);
  }
  100% {
    color: inherit;
  }
}

.share-info-component {
  margin-top: 0.5em;
}
</style>
