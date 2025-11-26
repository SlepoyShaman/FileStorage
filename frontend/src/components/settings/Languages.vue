<template>
  <select @change="change" :value="locale">
    <option v-for="(value, label) in locales" :key="label" :value="label">
      {{ $t("languages." + label) }}
    </option>
  </select>
</template>

<script>
import { defineComponent } from "vue";

export default defineComponent({
  name: "Languages",
  props: {
    locale: {
      type: String,
      required: true,
    },
  },
  data() {
    return {
      locales: {
        en: "en",
        ru: "ru",
      },
    };
  },
  methods: {
    change(event) {
      const target = event.target;
      this.$emit("update:locale", target.value);
    },
  },
});
</script>
<style>

</style>
