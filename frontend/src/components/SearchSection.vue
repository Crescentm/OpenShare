<script setup lang="ts">
import { computed, ref } from "vue";
import { Search } from "lucide-vue-next";

import SearchTagChip from "./SearchTagChip.vue";

const props = withDefaults(
  defineProps<{
    tags?: string[];
  }>(),
  {
    tags: () => [],
  },
);

const keyword = ref("");
const selectedTags = ref<string[]>([]);
const canSearch = computed(() => keyword.value.trim().length > 0);

function toggleTag(tag: string) {
  if (selectedTags.value.includes(tag)) {
    selectedTags.value = selectedTags.value.filter((item) => item !== tag);
    return;
  }
  selectedTags.value = [...selectedTags.value, tag];
}

async function submitSearch() {
  if (!canSearch.value) {
    return;
  }
}
</script>

<template>
  <section class="panel px-6 py-6">
    <div class="space-y-6">
      <div class="space-y-4">
        <form class="flex flex-col gap-3 xl:flex-row xl:items-center" @submit.prevent="submitSearch">
          <label class="relative block min-w-0 flex-1">
            <Search class="pointer-events-none absolute left-5 top-1/2 h-5 w-5 -translate-y-1/2 text-slate-400" />
            <input
              v-model="keyword"
              type="text"
              placeholder="搜索课程资料、讲义、实验报告或关键词"
              class="h-14 w-full rounded-lg border border-slate-300 bg-white pl-14 pr-5 text-[15px] text-slate-900 outline-none transition placeholder:text-slate-400 focus:border-slate-400 focus:ring-4 focus:ring-slate-100 dark:border-slate-700 dark:bg-slate-950 dark:text-slate-100 dark:placeholder:text-slate-500 dark:focus:border-slate-500 dark:focus:ring-slate-800"
            />
          </label>

          <button
            type="submit"
            class="h-11 rounded-lg px-6 text-sm font-medium transition xl:shrink-0"
            :class="
              canSearch
                ? 'bg-slate-900 text-white hover:bg-slate-800 dark:bg-slate-100 dark:text-slate-900 dark:hover:bg-white'
                : 'cursor-not-allowed bg-slate-200 text-slate-500 dark:bg-slate-800 dark:text-slate-500'
            "
            :disabled="!canSearch"
          >
            搜索
          </button>
        </form>

        <p class="pl-1 text-sm leading-6 text-slate-500 dark:text-slate-400">
          输入关键词后即可搜索公开资料，也可以叠加下方标签进一步筛选。
        </p>
      </div>

      <div v-if="props.tags.length > 0" class="border-t border-slate-100 pt-4 dark:border-slate-800">
        <div class="flex flex-wrap items-center gap-2.5">
          <span class="mr-1 text-xs font-semibold uppercase tracking-[0.08em] text-slate-400 dark:text-slate-500">Tags</span>
          <SearchTagChip
            v-for="tag in props.tags"
            :key="tag"
            :label="tag"
            :selected="selectedTags.includes(tag)"
            @click="toggleTag(tag)"
          />
        </div>
      </div>
    </div>
  </section>
</template>
