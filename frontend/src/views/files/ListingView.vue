<template>
  <div v-if="shareInfo.shareType != 'upload'" class="no-select">
    <div v-if="loading">
      <h2 class="message delayed">
        <div class="spinner">
          <div class="bounce1"></div>
          <div class="bounce2"></div>
          <div class="bounce3"></div>
        </div>
        <span>{{ $t("files.loading") }}</span>
      </h2>
    </div>
    <div v-else>
      <div
        id="listingView"
        ref="listingView"
        :class="{
          'add-padding': isStickySidebar,
          [listingViewMode]: true,
          dropping: isDragging,
          'rectangle-selecting': isRectangleSelecting
        }"
        class="file-icons"
      >
        <!-- Rectangle selection overlay -->
        <div class="selection-rectangle"
          :style="rectangleStyle"
        ></div>
        <div>
          <div class="header card" :class="{ 'dark-mode-item-header': isDarkMode }">
            <p
              :class="{ active: nameSorted }"
              class="name"
              role="button"
              tabindex="0"
              @click="sort('name')"
              :title="$t('files.sortByName')"
              :aria-label="$t('files.sortByName')"
            >
              <span>{{ $t("files.name") }}</span>
              <i class="material-icons">{{ nameIcon }}</i>
            </p>

            <p
              :class="{ active: sizeSorted }"
              class="size"
              role="button"
              tabindex="0"
              @click="sort('size')"
              :title="$t('files.sortBySize')"
              :aria-label="$t('files.sortBySize')"
            >
              <span>{{ $t("files.size") }}</span>
              <i class="material-icons">{{ sizeIcon }}</i>
            </p>
            <p
              :class="{ active: modifiedSorted }"
              class="modified"
              role="button"
              tabindex="0"
              @click="sort('modified')"
              :title="$t('files.sortByLastModified')"
              :aria-label="$t('files.sortByLastModified')"
            >
              <span>{{ $t("files.lastModified") }}</span>
              <i class="material-icons">{{ modifiedIcon }}</i>
            </p>
          </div>
        </div>
        <div v-if="numDirs > 0">
          <div class="header-items">
            <h2>{{ $t("general.folders") }}</h2>
          </div>
        </div>
        <div
          v-if="numDirs > 0"
          class="folder-items"
          aria-label="Folder Items"
          :class="{ lastGroup: numFiles === 0 }"
        >
          <item
            v-for="item in dirs"
            :key="base64(item.name)"
            v-bind:index="item.index"
            v-bind:name="item.name"
            v-bind:isDir="item.type == 'directory'"
            v-bind:source="req.source"
            v-bind:modified="item.modified"
            v-bind:type="item.type"
            v-bind:size="item.size"
            v-bind:path="item.path"
            v-bind:reducedOpacity="item.hidden || isDragging"
            v-bind:hash="shareInfo.hash"
            v-bind:hasPreview="item.hasPreview"
          />
        </div>
        <div v-if="numFiles > 0">
          <div class="header-items">
            <h2>{{ $t("general.files") }}</h2>
          </div>
        </div>
        <div v-if="numFiles > 0" class="file-items" :class="{ lastGroup: numFiles > 0 }" aria-label="File Items">
          <item
            v-for="item in files"
            :key="base64(item.name)"
            v-bind:index="item.index"
            v-bind:name="item.name"
            v-bind:isDir="item.type == 'directory'"
            v-bind:modified="item.modified"
            v-bind:source="req.source"
            v-bind:type="item.type"
            v-bind:size="item.size"
            v-bind:path="item.path"
            v-bind:reducedOpacity="item.hidden || isDragging"
            v-bind:hash="shareInfo.hash"
            v-bind:hasPreview="item.hasPreview"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import throttle from "@/utils/throttle";
import { state, mutations, getters } from "@/store";
import { url } from "@/utils";
import { shareInfo } from "@/utils/constants";

import Item from "@/components/files/ListingItem.vue";

export default {
  name: "listingView",
  components: {
    Item,
  },
  data() {
    return {
      columnWidth: 250 + state.user.gallerySize * 50,
      dragCounter: 0,
      width: window.innerWidth,
      lastSelected: {},
      contextTimeout: null,
      ctrKeyPressed: false,
      clipboard: { items: [] },
      isRectangleSelecting: false,
      rectangleStart: { x: 0, y: 0 },
      rectangleEnd: { x: 0, y: 0 },
      rectangleSelection: [],
    };
  },
  watch: {
    gallerySize() {
      this.columnWidth = 250 + state.user.gallerySize * 50;
      this.colunmsResize();
    },
    scrolling() {
      const scrollContainer = this.$refs.listingView;
      if (!scrollContainer) return;

      const itemNodes = scrollContainer.querySelectorAll(".listing-item");

      let topItem = null;
      let minTop = Infinity;
      itemNodes.forEach((el) => {
        const rect = el.getBoundingClientRect();
        if (rect.top >= 0 && rect.top < minTop) {
          minTop = rect.top;
          topItem = el;
        }
      });

      if (!topItem) return;

      const letter = topItem.getAttribute("data-name")?.[0]?.toUpperCase() || "A";
      let category = "folders"; 
      if (this.numFiles > 0) {
        const fileSection = this.$el.querySelector(".file-items");
        const fileTop = fileSection?.getBoundingClientRect().top ?? 0;
        category = fileTop <= 0 ? "files" : "folders";
      }
      if (this.numDirs == 0) {
        category = "files";
      }

      mutations.updateListing({
        ...state.listing,
        category,
        letter,
      });
    },
  },
  computed: {
    shareInfo() {
      return shareInfo;
    },
    state() {
      return state;
    },
    isDragging() {
      return this.dragCounter > 0;
    },
    scrolling() {
      return state.listing.scrollRatio;
    },
    isStickySidebar() {
      return getters.isStickySidebar();
    },
    lastFolderIndex() {
      const allItems = [...this.items.dirs, ...this.items.files];
      for (let i = 0; i < allItems.length; i++) {
        if (allItems[i].type != "directory") {
          return i - 1;
        }
      }
      if (allItems.length > 0) {
        return allItems.length;
      }

      return null;
    },
    numColumns() {
      if (!getters.isCardView()) {
        return 1;
      }
      const elem = document.querySelector("#main");
      if (!elem) {
        return 1;
      }
      let columns = Math.floor(elem.offsetWidth / this.columnWidth);
      if (columns === 0) columns = 1;
      return columns;
    },
    gallerySize() {
      return state.user.gallerySize;
    },
    isDarkMode() {
      return getters.isDarkMode();
    },
    getMultiple() {
      return state.multiple;
    },
    nameSorted() {
      return getters.sorting().by === "name";
    },
    sizeSorted() {
      return getters.sorting().by === "size";
    },
    modifiedSorted() {
      return getters.sorting().by === "modified";
    },
    ascOrdered() {
      return getters.sorting().asc;
    },
    items() {
      return getters.reqItems();
    },
    numDirs() {
      return getters.reqNumDirs();
    },
    numFiles() {
      return getters.reqNumFiles();
    },
    dirs() {
      return this.items.dirs;
    },
    files() {
      return this.items.files;
    },
    nameIcon() {
      if (this.nameSorted && !this.ascOrdered) {
        return "arrow_upward";
      }

      return "arrow_downward";
    },
    sizeIcon() {
      if (this.sizeSorted && this.ascOrdered) {
        return "arrow_downward";
      }

      return "arrow_upward";
    },
    modifiedIcon() {
      if (this.modifiedSorted && this.ascOrdered) {
        return "arrow_downward";
      }

      return "arrow_upward";
    },
    viewIcon() {
      const icons = {
        list: "view_module",
        compact: "view_module",
        normal: "grid_view",
        gallery: "view_list",
      };
      return icons[getters.viewMode()];
    },
    listingViewMode() {
      this.colunmsResize();
      return getters.viewMode();
    },
    selectedCount() {
      return state.selected.length;
    },
    req() {
      return state.req;
    },
    loading() {
      return getters.isLoading();
    },
    rectangleStyle() {
      if (!this.isRectangleSelecting) return { display: 'none' };
      const left = Math.min(this.rectangleStart.x, this.rectangleEnd.x);
      const top = Math.min(this.rectangleStart.y, this.rectangleEnd.y);
      const width = Math.abs(this.rectangleStart.x - this.rectangleEnd.x);
      const height = Math.abs(this.rectangleStart.y - this.rectangleEnd.y);
      return {
        left: left + 'px',
        top: top + 'px',
        width: width + 'px',
        height: height + 'px',
      };
    },
  },
  mounted() {
    mutations.setSearch(false);
    this.lastSelected = state.selected;
    this.colunmsResize();
    window.addEventListener("keydown", this.keyEvent);
    window.addEventListener("resize", this.windowsResize);
    window.addEventListener("click", this.clickClear);
    window.addEventListener("keyup", this.clearCtrKey);
    window.addEventListener("dragover", this.preventDefault);
    this.$el.addEventListener("touchmove", this.handleTouchMove);

    this.$el.addEventListener("contextmenu", this.openContext);
    if (state.isSafari) {
      this.$el.addEventListener("touchstart", this.openContextForSafari, {
        passive: true,
      });
      this.$el.addEventListener("mousedown", this.openContextForSafari);

      this.$el.addEventListener("touchend", this.cancelContext);
      this.$el.addEventListener("mouseup", this.cancelContext);
    }

    if (state.user.permissions?.modify || shareInfo.allowCreate) {
      this.$el.addEventListener("dragenter", this.dragEnter);
      this.$el.addEventListener("dragleave", this.dragLeave);
      this.$el.addEventListener("drop", this.drop);
      this.$el.addEventListener('mousedown', this.startRectangleSelection);
      document.addEventListener('mousemove', this.updateRectangleSelection);
      document.addEventListener('mouseup', this.endRectangleSelection);
    }

  },
  beforeUnmount() {
    window.removeEventListener("keydown", this.keyEvent);
    window.removeEventListener("resize", this.windowsResize);
    window.removeEventListener("click", this.clickClear);
    window.removeEventListener("keyup", this.clearCtrKey);
    window.removeEventListener("dragover", this.preventDefault);

    this.$el.removeEventListener("touchmove", this.handleTouchMove);
    this.$el.removeEventListener("contextmenu", this.openContext);

    if (state.isSafari) {
      this.$el.removeEventListener("touchstart", this.openContextForSafari);
      this.$el.removeEventListener("mousedown", this.openContextForSafari);
      this.$el.removeEventListener("touchend", this.cancelContext);
      this.$el.removeEventListener("mouseup", this.cancelContext);
    }

    if (state.user && state.user?.permissions?.modify || shareInfo.allowCreate) {
      this.$el.removeEventListener("dragenter", this.dragEnter);
      this.$el.removeEventListener("dragleave", this.dragLeave);
      this.$el.removeEventListener("drop", this.drop);
      this.$el.removeEventListener('mousedown', this.startRectangleSelection);
      document.removeEventListener('mousemove', this.updateRectangleSelection);
      document.removeEventListener('mouseup', this.endRectangleSelection);
    }
  },
  methods: {
    cancelContext() {
      if (this.contextTimeout) {
        clearTimeout(this.contextTimeout);
        this.contextTimeout = null;
      }
      this.isLongPress = false;
    },
    openContextForSafari(event) {
      this.cancelContext();
      this.isLongPress = false; 
      this.isSwipe = false;

      const touch = event.touches[0];
      this.touchStartX = touch.clientX;
      this.touchStartY = touch.clientY;

      this.contextTimeout = setTimeout(() => {
        if (!this.isSwipe) {
          this.isLongPress = true;
          event.preventDefault(); 
          this.openContext(event);
        }
      }, 500); 
    },
    handleTouchMove(event) {
      const touch = event.touches[0];
      const deltaX = Math.abs(touch.clientX - this.touchStartX);
      const deltaY = Math.abs(touch.clientY - this.touchStartY);
      const movementThreshold = 10; 
      if (deltaX > movementThreshold || deltaY > movementThreshold) {
        this.isSwipe = true;
        this.cancelContext();
      }
    },
    handleTouchEnd() {
      this.cancelContext(); 
      this.isSwipe = false; 
    },
    base64(name) {
      return url.base64Encode(name);
    },
    selectFirstItem() {
      mutations.resetSelected();
      const allItems = [...this.items.dirs, ...this.items.files];
      if (allItems.length > 0) {
        mutations.addSelected(allItems[0].index);
      }
    },
    selectItem(index) {
      mutations.resetSelected();
      mutations.addSelected(index);
    },
    selectAll() {
      for (let file of this.items.files) {
        if (state.selected.indexOf(file.index) === -1) {
          mutations.addSelected(file.index);
        }
      }
      for (let dir of this.items.dirs) {
        if (state.selected.indexOf(dir.index) === -1) {
          mutations.addSelected(dir.index);
        }
      }
    },
    alphanumericKeyPress(key) {
      const searchLetter = key.toLowerCase();
      const currentSelected = getters.getFirstSelected();
      let currentName = null;
      let findNextWithName = false;

      if (currentSelected != undefined) {
        currentName = currentSelected.name.toLowerCase();
        if (currentName.startsWith(searchLetter)) {
          findNextWithName = true;
        }
      }
      const allItems = [...this.items.dirs, ...this.items.files];
      let foundPrevious = false;
      let firstFound = null;
      for (let i = 0; i < allItems.length; i++) {
        const itemName = allItems[i].name.toLowerCase();
        if (!itemName.startsWith(searchLetter)) {
          continue;
        }
        if (firstFound == null) {
          firstFound = allItems[i].index;
        }
        if (!findNextWithName) {
          this.selectItem(allItems[i].index);
          return;
        }
        if (itemName == currentName) {
          foundPrevious = true;
          continue;
        }
        if (foundPrevious) {
          this.selectItem(allItems[i].index);
          return;
        }
      }
      if (firstFound != null) {
        this.selectItem(firstFound);
      }
    },
    preventDefault(event) {
      event.preventDefault();
    },
    colunmsResize() {
      document.documentElement.style.setProperty(
        "--item-width",
        `calc(${100 / this.numColumns}% - 1em)`
      );

      if (getters.viewMode() == "gallery") {
        document.documentElement.style.setProperty(
          "--item-height",
          `calc(${this.columnWidth / 20}em)`
        );
      } else {
        document.documentElement.style.setProperty("--item-height", `auto`);
      }
    },
    sort(field) {
      let asc = false;
      if (
        (field === "name" && this.nameIcon === "arrow_upward") ||
        (field === "size" && this.sizeIcon === "arrow_upward") ||
        (field === "modified" && this.modifiedIcon === "arrow_upward")
      ) {
        asc = true;
      }

      mutations.updateListingSortConfig({ field, asc });
      mutations.updateListingItems();
      this.lastSelected = state.selected;
    },
    setMultiple(val) {
      mutations.setMultiple(val == true);
      showMultipleSelection();
    },
    windowsResize: throttle(function () {
      this.colunmsResize();
      this.width = window.innerWidth;
      if (this.$refs.listingView == null) return;
    }, 100),
    openContext(event) {
      event.preventDefault();
      event.stopPropagation();
      mutations.showHover({
        name: "ContextMenu",
        props: {
          showCentered: getters.isMobile(),
          posX: event.clientX,
          posY: event.clientY,
        },
      });
    },
    clickClear(event) {
      if (this.ctrKeyPressed || event.shiftKey) return;
      const sameAsBefore = state.selected == this.lastSelected;
      if (sameAsBefore && !state.multiple && getters.currentPromptName() == "") {
        mutations.resetSelected();
      }
      this.lastSelected = state.selected;
    },

    startRectangleSelection(event) {
      if (event.target.closest('.item') || event.target.closest('.header')) {
        return;
      }

      if (event.button !== 0) return;

      this.isRectangleSelecting = true;

      const listingRect = this.$refs.listingView.getBoundingClientRect();
      this.rectangleStart = {
        x: event.clientX - listingRect.left,
        y: event.clientY - listingRect.top
      };
      this.rectangleEnd = {
        x: event.clientX - listingRect.left,
        y: event.clientY - listingRect.top
      };
 
      this.initialSelectionState = [...state.selected];

      const hasModifier = event.ctrlKey || event.metaKey;
      if (!hasModifier) {
        mutations.resetSelected();
      }

      event.preventDefault();
    },

    updateRectangleSelection(event) {
      if (!this.isRectangleSelecting) return;

      const listingRect = this.$refs.listingView.getBoundingClientRect();
      this.rectangleEnd = {
        x: event.clientX - listingRect.left,
        y: event.clientY - listingRect.top
      };

      this.updateSelectedItemsInRectangle(event.ctrlKey || event.metaKey);
    },

    endRectangleSelection(event) {
      if (!this.isRectangleSelecting) return;

      this.isRectangleSelecting = false;
      this.updateSelectedItemsInRectangle(event.ctrlKey || event.metaKey);

      setTimeout(() => {
        this.rectangleStart = { x: 0, y: 0 };
        this.rectangleEnd = { x: 0, y: 0 };
        this.initialSelectionState = [];
      }, 100);
    },

    updateSelectedItemsInRectangle(isAdditive) {
      if (!this.isRectangleSelecting) return;

      const listingRect = this.$refs.listingView.getBoundingClientRect();
      const rect = {
        left: Math.min(this.rectangleStart.x, this.rectangleEnd.x),
        top: Math.min(this.rectangleStart.y, this.rectangleEnd.y),
        right: Math.max(this.rectangleStart.x, this.rectangleEnd.x),
        bottom: Math.max(this.rectangleStart.y, this.rectangleEnd.y)
      };

      const rectangleSelectedIndexes = [];

      const itemElements = this.$el.querySelectorAll('.item');

      itemElements.forEach((element) => {
        const elementRect = element.getBoundingClientRect();

        const elementRelativeRect = {
          left: elementRect.left - listingRect.left,
          top: elementRect.top - listingRect.top,
          right: elementRect.right - listingRect.left,
          bottom: elementRect.bottom - listingRect.top
        };

        if (
          elementRelativeRect.left < rect.right &&
          elementRelativeRect.right > rect.left &&
          elementRelativeRect.top < rect.bottom &&
          elementRelativeRect.bottom > rect.top
        ) {
          const index = parseInt(element.getAttribute('data-index'));
          if (!isNaN(index)) {
            rectangleSelectedIndexes.push(index);
          }
        }
      });

      if (isAdditive) {
        const newSelection = [...state.selected];
        rectangleSelectedIndexes.forEach(index => {
          if (!newSelection.includes(index)) {
            newSelection.push(index);
          }
        });

        mutations.resetSelected();
        newSelection.forEach(index => mutations.addSelected(index));
      } else {
        mutations.resetSelected();
        rectangleSelectedIndexes.forEach(index => mutations.addSelected(index));
      }
    },
  },
};
</script>

<style>
.dark-mode-item-header {
  border-color: var(--divider) !important;
  background: var(--surfacePrimary) !important;
  user-select: none;
}

.header-items {
  width: 100% !important;
  max-width: 100% !important;
  justify-content: center;
}

.add-padding {
  padding-left: 0.5em;
}
.font-size-large h2 {
  font-size: 2em !important;
}

#listingView.dropping {
  transform: scale(0.97);
  box-shadow: var(--primaryColor) 0 0 1em;
}

#listingView {
  min-height: 90vh !important;
  position: relative;
}

.folder-items a {
  border-color: #d1d1d1;
  border-style: solid;
}

/* Upload Share Styles */
.upload-share-embed {
  padding: 2em;
  max-width: 800px;
  margin: 0 auto;
}

.selection-rectangle {
  position: absolute;
  border: 2px solid var(--primaryColor);
  background-color: color-mix(in srgb, var(--primaryColor) 25%, transparent);
  border-radius: 8px;
  pointer-events: none;
  z-index: 10;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

#listingView.rectangle-selecting {
  cursor: crosshair;
  user-select: none;
}

#listingView.rectangle-selecting .item {
  pointer-events: none;
}

</style>
