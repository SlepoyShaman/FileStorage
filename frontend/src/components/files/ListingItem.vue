<template>
  <a
    :href="getUrl()"
    class="item listing-item clickable no-select"
    :class="{
      activebutton: isSelected,
      hiddenFile: isHiddenNotSelected && this && !this.isDraggedOver,
      'half-selected': isDraggedOver,
      'drag-hover': isDraggedOver,
    }"
    :id="getID"
    role="button"
    tabindex="0"
    :draggable="isDraggable"
    @dragstart="dragStart"
    @dragleave="dragLeave"
    @dragover="dragOver"
    @drop="drop"
    :data-dir="isDir"
    :data-type="type"
    :data-name="name"
    :data-index="index"
    :aria-label="name"
    :aria-selected="isSelected"
    @contextmenu="onRightClick"
    @click="click"
    @touchstart="addSelected"
    @touchmove="handleTouchMove"
    @touchend="cancelContext"
    @mouseup="cancelContext"
  >
    <div :class="{ 'gallery-div': galleryView }">
      <Icon
        :mimetype="type"
        :active="isSelected"
        :thumbnailUrl="isThumbnailInView ? thumbnailUrl : ''"
        :filename="name"
        :hasPreview="hasPreview"
      />
    </div>

    <div class="text">
      <p :class="{ adjustment: quickDownloadEnabled }" class="name">{{ name }}</p>
      <p
        class="size"
        :class="{ adjustment: quickDownloadEnabled }"
        :data-order="humanSize()"
      >
        {{ humanSize() }}
      </p>
      <p class="modified">
        <time :datetime="modified">{{ getTime() }}</time>
      </p>
    </div>
    <Icon
      @click.stop="downloadFile"
      v-if="quickDownloadEnabled"
      :filename="name"
      :hasPreview="hasPreview"
      mimetype="file_download"
      style="padding-right: 0.5em"
      class="download-icon"
      role="button"
      aria-label="Download"
      tabindex="0"
      :clickable=true
    />
  </a>
</template>

<script>
import { globalVars, serverHasMultipleSources, shareInfo } from "@/utils/constants";
import downloadFiles from "@/utils/download";

import { getHumanReadableFilesize } from "@/utils/filesizes";
import { filesApi,publicApi } from "@/api";
import * as upload from "@/utils/upload";
import { state, getters, mutations } from "@/store";
import { url } from "@/utils";
import Icon from "@/components/files/Icon.vue";

export default {
  name: "item",
  components: {
    Icon,
  },
  data() {
    return {
      isThumbnailInView: false,
      touches: 0,
      touchStartX: 0,
      touchStartY: 0,
      isLongPress: false,
      isSwipe: false,
      isDraggedOver: false,
      contextTimeout: null,
    };
  },
  props: [
    "name",
    "isDir",
    "source",
    "type",
    "size",
    "modified",
    "index",
    "readOnly",
    "path",
    "reducedOpacity",
    "hash",
    "hasPreview",
  ],
  computed: {
    galleryView() {
      return getters.viewMode() === "gallery";
    },
    quickDownloadEnabled() {
      // @ts-ignore
      if (getters.isShare()) {
        // @ts-ignore
        return shareInfo.quickDownload && !this.isDir;
      }
      // @ts-ignore
      return state.user?.quickDownload && !this.galleryView && !this.isDir;
    },
    isHiddenNotSelected() {
      return !this.isSelected && this.reducedOpacity;
    },
    getID() {
      return url.base64Encode(encodeURIComponent(this.name));
    },
    quickNav() {
      return state.user.singleClick && !state.multiple;
    },
    user() {
      return state.user;
    },
    selected() {
      return state.selected;
    },
    isClicked() {
      // @ts-ignore
      if (state.user.singleClick || !this.allowedView) {
        return false;
      }
      return this.isSelected;
    },
    isSelected() {
      return state.selected.indexOf(this.index) !== -1;
    },
    isDraggable() {
      // @ts-ignore
      return this.readOnly == undefined && state.user.permissions?.modify || shareInfo.allowCreate;
    },
    canDrop() {
      if (!this.isDir || this.readOnly !== undefined) return false;

      for (const i of this.selected) {
        if (
          // @ts-ignore
          state.req.items[i].path === this.path &&
          // @ts-ignore
          state.req.source === this.source
        ) {
          return false;
        }

        // @ts-ignore
        if (state.req.items[i].index === this.index) {
          return false;
        }
      }
      return true;
    },
    thumbnailUrl() {
      if (!globalVars.enableThumbs || !state.req.path || !this.name) {
        return "";
      }
      const previewPath = url.joinPath(state.req.path, this.name);
      if (getters.isShare()) {
        return publicApi.getPreviewURL(previewPath);
      }
      // @ts-ignore
      return filesApi.getPreviewURL(state.req.source, previewPath, this.modified);
    },
    isThumbsEnabled() {
      return globalVars.enableThumbs;
    },
  },
  mounted() {
    const observer = new IntersectionObserver(this.handleIntersect, {
      root: null,
      rootMargin: "0px",
      threshold: 0.1,
    });

    observer.observe(this.$el);
  },
  methods: {
    /** @param {MouseEvent} event */
    downloadFile(event) {
      event.preventDefault();
      event.stopPropagation();
      mutations.resetSelected();
      // @ts-ignore
      mutations.addSelected(this.index);
      downloadFiles(state.selected);
    },
    /** @param {TouchEvent} event */
    handleTouchMove(event) {
      if (!state.isSafari) return;
      const touch = event.touches[0];
      const deltaX = Math.abs(touch.clientX - this.touchStartX);
      const deltaY = Math.abs(touch.clientY - this.touchStartY);
      const movementThreshold = 10;
      if (deltaX > movementThreshold || deltaY > movementThreshold) {
        this.isSwipe = true;
        // @ts-ignore
        this.cancelContext();
      }
    },
    handleTouchEnd() {
      if (!state.isSafari) return;
      // @ts-ignore
      this.cancelContext(); // Clear timeout
      this.isSwipe = false; // Reset swipe state
    },
    cancelContext() {
      if (this.contextTimeout) {
        clearTimeout(this.contextTimeout);
        this.contextTimeout = null;
      }
      this.isLongPress = false;
    },
    /** @param {string} path */
    updateHashAndNavigate(path) {
      window.location.hash = path;
    },
    getUrl() {
      if (this.hash) {
        return globalVars.baseURL + "public/share/" + this.hash + "/" + url.encodedPath(this.path);
      }
      if (serverHasMultipleSources) {
        return globalVars.baseURL + "files/" + encodeURIComponent(this.source) + url.encodedPath(this.path);
      }
      return globalVars.baseURL + "files" + url.encodedPath(this.path);
    },
    /** @param {MouseEvent} event */
    onRightClick(event) {
      event.preventDefault();
      if (!state.multiple && getters.selectedCount() < 2) {
        mutations.resetSelected();
        // @ts-ignore
        mutations.addSelected(this.index);
      }
      mutations.showHover({
        name: "ContextMenu",
        props: {
          posX: event.clientX,
          posY: event.clientY,
        },
      });
    },
    /**
     * @param {IntersectionObserverEntry[]} entries
     * @param {IntersectionObserver} observer
     */
    handleIntersect(entries, observer) {
      entries.forEach((entry) => {
        if (entry.isIntersecting) {
          this.isThumbnailInView = true;
          // Stop observing once thumbnail is in view
          observer.unobserve(entry.target);
        }
      });
    },
    humanSize() {
      return this.type == "invalid_link"
        ? "invalid link"
        : getHumanReadableFilesize(this.size);
    },
    getTime() {
      // @ts-ignore
      return getters.getTime(this.modified);
    },
    /** @param {DragEvent} event */
    dragLeave(event) {
      // Only reset visual state for internal drags
      const isInternal = Array.from(event.dataTransfer.types).includes(
        "application/x-filebrowser-internal-drag"
      );
      if (isInternal) {
        this.isDraggedOver = false;
      }
    },
    /** @param {DragEvent} event */
    dragStart(event) {
      if (state.selected.indexOf(this.index) === -1) {
        mutations.resetSelected();
        // @ts-ignore
        mutations.addSelected(this.index);
      }
      if (event.dataTransfer) {
        event.dataTransfer.setData(
          "application/x-filebrowser-internal-drag",
          "true"
        );
      }
    },
    /** @param {DragEvent} event */
    dragOver(event) {
      if (!this.canDrop) return;

      const isInternal = Array.from(event.dataTransfer.types).includes(
        "application/x-filebrowser-internal-drag"
      );

      if (!isInternal) return;

      event.preventDefault();
      this.isDraggedOver = true;
    },
    /** @param {DragEvent} event */
    async drop(event) {
      this.isDraggedOver = false;

      const isInternal = Array.from(event.dataTransfer.types).includes(
        "application/x-filebrowser-internal-drag"
      );

      if (!isInternal) {
        return;
      }

      event.preventDefault();
      event.stopPropagation();

      let items = [];
      for (let i of state.selected) {
        items.push({
          // @ts-ignore
          from: state.req.items[i].path,
          // @ts-ignore
          fromSource: state.req.items[i].source,
          // @ts-ignore
          to: url.joinPath(this.path, state.req.items[i].name),
          toSource: this.source,
        });
      }

      items = items.filter(item => {
        if (item.from === item.to) {
          return false;
        }

        const fromPath = item.from;
        const destinationDir = this.path;

        if (fromPath === destinationDir || fromPath.startsWith(destinationDir + '/')) {
          return false;
        }

        return true;
      });

      if (items.length === 0) {
        return;
      }

      let checkAction = async () => {
        if (getters.isShare()) {
          return await publicApi.fetchPub(this.path, shareInfo.hash);
        } else {
          return await filesApi.fetchFiles(this.source, this.path);
        }
      }
      const response = await checkAction();
      const conflict = upload.checkConflict(items, response?.items || [] );

      /**
       * @param {boolean} overwrite
       * @param {boolean} rename
       */
      let action = async (overwrite, rename) => {
        mutations.showHover({
          name: "move",
          props: {
            operationInProgress: true,
          },
        });

        try {
          if (getters.isShare()) {
            await publicApi.moveCopy(items, "move", overwrite, rename);
          } else {
            await filesApi.moveCopy(items, "move", overwrite, rename);
          }
          mutations.closeHovers();
        } catch (error) {
          mutations.closeHovers();
          throw error;
        }
      };

      if (conflict) {
        mutations.showHover({
          name: "replace-rename",
          /**
           * @param {Event} event
           * @param {string} option
           */
          confirm: async (event, option) => {
            const overwrite = option === "overwrite";
            const rename = option === "rename";

            event.preventDefault();
            mutations.closeHovers();
            await action(overwrite, rename);
          },
        });
        return;
      }

      await action(false, false);
    },
    /** @param {TouchEvent} event */
    addSelected(event) {
      if (!state.isSafari) {
        return;
      }
      if (event.type !== "touchstart") {
        return;
      }
      const touch = event.touches[0];
      this.touchStartX = touch.clientX;
      this.touchStartY = touch.clientY;
      this.isLongPress = false; 
      this.isSwipe = false; 
      if (state.multiple) {
        return;
      }
      // @ts-ignore
      this.contextTimeout = setTimeout(() => {
        if (!this.isSwipe) {
          if (!this.isSelected) {
            mutations.resetSelected();
            // @ts-ignore
            mutations.addSelected(this.index);
          }
        }
      }, 500);
    },
    /** @param {MouseEvent} event */
    click(event) {
      if (event.button === 0) {
        // Left-click
        event.preventDefault();
        if (this.quickNav) {
          this.open();
        }
      }

      if (
        !state.user.singleClick &&
        getters.selectedCount() !== 0 &&
        event.button === 0
      ) {
        event.preventDefault();
      }
      setTimeout(() => {
        this.touches = 0;
      }, 500);
      this.touches++;
      if (this.touches > 1) {
        this.open();
      }

      if (event.shiftKey && this.selected.length > 0) {
        let fi = 0;
        let la = 0;

        if (state.lastSelectedIndex !== null) {
          if (this.index > state.lastSelectedIndex) {
            fi = state.lastSelectedIndex;
            la = this.index;
          } else {
            fi = this.index;
            la = state.lastSelectedIndex;
          }
        }

        mutations.resetSelected();

        for (; fi <= la; fi++) {
          if (this.selected.indexOf(fi) === -1) {
            // @ts-ignore
            mutations.addSelected(fi);
          }
        }
        return;
      }

      if (this.selected.indexOf(this.index) !== -1) {
        if (event.ctrlKey || event.metaKey) {
          mutations.removeSelected(this.index);
          mutations.setLastSelectedIndex(this.index);
          return;
        }

        if (state.multiple) {
          mutations.removeSelected(this.index);
          mutations.setLastSelectedIndex(this.index);
          return;
        }

        if (this.selected.length > 1) {
          mutations.resetSelected();
          // @ts-ignore
          mutations.addSelected(this.index);
          mutations.setLastSelectedIndex(this.index);
        }
        return;
      }

      if (
        !state.user.singleClick &&
        !event.ctrlKey &&
        !event.metaKey &&
        !state.multiple
      ) {
        mutations.resetSelected();
      }
      mutations.addSelected(this.index);
      mutations.setLastSelectedIndex(this.index);
    },
    open() {
      const previousHistoryItem = {
        name: state.req.items[this.index].name,
        source: state.req.source,
        path: state.req.path,
      };
      url.goToItem(this.source, this.path, previousHistoryItem);
    },
  },
};
</script>

<style>
.download-icon {
  font-size: 1.5em;
  cursor: pointer;
  color: var(--secondaryColor);
}

.icon-download {
  font-size: 0.5em;
}

.item {
  -webkit-touch-callout: none; 
}

.hiddenFile {
  opacity: 0.5;
}

.activebutton {
  height: 10em;
}

.iconActive {
  font-size: 6em !important;
}

.activetitle {
  width: 9em !important;
  margin-right: 1em !important;
}

.half-selected {
  border-color: var(--primaryColor) !important;
  border-style: solid !important;
}
</style>
