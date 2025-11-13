import { state } from "./state.js";
import { getters } from "./getters.js";
import { emitStateChanged } from './eventBus'; 
import { notify } from "@/notify";
import { filesApi, publicApi } from "@/api";

export const mutations = {
  setPreviewSource: (value) => {
    if (value === state.popupPreviewSource) {
      return;
    }
    state.popupPreviewSource = value;
    emitStateChanged();
  },
  updateListing: (value) => {
    if (value == state.listing) {
      return;
    }
    state.listing = value;
    emitStateChanged();
  },
  closeHovers: () => {
    const hasActiveUploads = state.upload.isUploading;
    const hasUploadPrompt = state.prompts.some(p => p.name === "upload");
    const hasWarningPrompt = state.prompts.some(p => p.name === "CloseWithActiveUploads");

    if (hasActiveUploads && hasUploadPrompt && !hasWarningPrompt) {
      mutations.showHover({
        name: "CloseWithActiveUploads",
        confirm: () => {
          state.prompts = [];
          if (!state.stickySidebar) {
            state.showSidebar = false;
          }
          emitStateChanged();
        },
        cancel: () => {
          mutations.closeTopHover();
        },
      });
      return;
    }
    state.prompts = [];
    if (!state.stickySidebar) {
      state.showSidebar = false;
    }
    emitStateChanged();
  },
  closeTopHover: () => {
    state.prompts.pop();
    if (state.prompts.length === 0) {
      if (!state.stickySidebar) {
        state.showSidebar = false;
      }
    }
    emitStateChanged();
  },
  showHover: (value) => {
    if (typeof value === "object") {
      state.prompts.push({
        name: value?.name,
        confirm: value?.confirm,
        action: value?.action,
        props: value?.props,
        discard: value?.discard,
        cancel: value?.cancel,
      });
    } else {
      state.prompts.push({
        name: value,
        confirm: value?.confirm,
        action: value?.action,
        props: value?.props,
        discard: value?.discard,
        cancel: value?.cancel,
      });
    }
    emitStateChanged();
  },
  setLoading: (loadType, status) => {
    if (status === false) {
      delete state.loading[loadType];
    } else {
      state.loading = { ...state.loading, [loadType]: true };
    }
    emitStateChanged();
  },
  setMultiple: (value) => {
    if (value == state.multiple) {
      return;
    }
    state.multiple = value;
    if (value == true) {
      notify.showMultipleSelection()
    }
    emitStateChanged();
  },
  addSelected: (value) => {
    state.selected.push(value);
    emitStateChanged();
  },
  removeSelected: (value) => {
    let i = state.selected.indexOf(value);
    if (i === -1) return;
    state.selected.splice(i, 1);
    emitStateChanged();
  },
  resetSelected: () => {
    state.selected = [];
    mutations.setMultiple(false);
    emitStateChanged();
  },
  setLastSelectedIndex: (index) => {
    if (index === state.lastSelectedIndex) {
      return;
    }
    state.lastSelectedIndex = index;
    emitStateChanged();
  },
  updateListingSortConfig: ({ field, asc }) => {
    if (!state.user.sorting) {
      state.user.sorting = {};
    }
    mutations.updateDisplayPreferences({ sorting: { by: field, asc: asc } });
    emitStateChanged();
  },
  updateDisplayPreferences: (payload) => {
    let source = state.sources.current;
    if (getters.isShare()) {
      source = getters.currentHash();
    }
    const path = state.route.path;

    if (!source || !path) return;
    if (!state.displayPreferences) {
      state.displayPreferences = {};
    }
    if (!state.displayPreferences[source]) {
      state.displayPreferences[source] = {};
    }
    if (!state.displayPreferences[source][path]) {
      state.displayPreferences[source][path] = {};
    }

    state.displayPreferences[source][path] = {
      ...state.displayPreferences[source][path],
      ...payload,
    };

    const allPreferences = JSON.parse(localStorage.getItem("displayPreferences") || "{}");
    if (!allPreferences[state.user.username]) {
      allPreferences[state.user.username] = {};
    }
    allPreferences[state.user.username] = state.displayPreferences;
    localStorage.setItem("displayPreferences", JSON.stringify(allPreferences));

    emitStateChanged();
  },
  getPrefetchUrl: (item) => {
    if (getters.isShare()) {
      return publicApi.getDownloadURL(
        {
          path: item.path,
          hash: state.share.hash,
          token: state.share.token,
        },
        [item.path],
        true,
      );
    }
    return filesApi.getDownloadURL(item.source, item.path, true);
  },
};